package lamarzocco

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/philipparndt/go-logger"
)

const (
	BaseURL = "https://lion.lamarzocco.io/api/customer-app"
)

type Client struct {
	httpClient *http.Client
	username   string
	password   string

	installKey *InstallationKey
	keyLock    sync.RWMutex

	token     *TokenInfo
	tokenLock sync.RWMutex

	serial string
	model  string

	currentMode DoseMode
	dose1       *DoseInfo
	dose2       *DoseInfo
	machineOn   bool
	boiler      *BoilerInfo
	scale       *ScaleInfo
	modeLock    sync.RWMutex

	onStatusChange func(MachineStatus)
}

func NewClient(username, password string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		username:    username,
		password:    password,
		currentMode: DoseModeContinuous,
	}
}

func (c *Client) SetStatusChangeCallback(callback func(MachineStatus)) {
	c.onStatusChange = callback
}

// registerClient performs the initial registration with /auth/init
func (c *Client) registerClient() error {
	// Generate new installation key
	installKey, err := GenerateInstallationKey()
	if err != nil {
		return fmt.Errorf("failed to generate installation key: %w", err)
	}

	// Get public key for registration
	pubKeyB64, err := installKey.PublicKeyB64()
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	url := BaseURL + "/auth/init"

	payload := map[string]string{
		"pk": pubKeyB64,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal init payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create init request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Generate request proof
	baseString := installKey.BaseString()
	proof := GenerateRequestProof(baseString, installKey.Secret)

	req.Header.Set("X-App-Installation-Id", installKey.InstallationID)
	req.Header.Set("X-Request-Proof", proof)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("init request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("init failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.keyLock.Lock()
	c.installKey = installKey
	c.keyLock.Unlock()

	logger.Info("Client registered successfully", "installation_id", installKey.InstallationID)
	return nil
}

func (c *Client) authenticate() error {
	// Ensure we have an installation key
	c.keyLock.RLock()
	installKey := c.installKey
	c.keyLock.RUnlock()

	if installKey == nil {
		if err := c.registerClient(); err != nil {
			return err
		}
		c.keyLock.RLock()
		installKey = c.installKey
		c.keyLock.RUnlock()
	}

	url := BaseURL + "/auth/signin"

	payload := map[string]string{
		"username": c.username,
		"password": c.password,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal auth payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add authentication headers (only the extra headers for signin)
	extraHeaders, err := installKey.GenerateExtraHeaders()
	if err != nil {
		return fmt.Errorf("failed to generate extra headers: %w", err)
	}
	for key, value := range extraHeaders {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("auth failed with status %d: %s", resp.StatusCode, string(body))
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	// Token expires in 1 hour based on JWT exp claim
	expiresAt := time.Now().Add(1 * time.Hour)
	c.tokenLock.Lock()
	c.token = &TokenInfo{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		ExpiresAt:    expiresAt,
	}
	c.tokenLock.Unlock()

	logger.Info("Successfully authenticated with La Marzocco API", "expires_at", expiresAt)
	return nil
}

func (c *Client) refreshToken() error {
	c.tokenLock.RLock()
	refreshToken := ""
	if c.token != nil {
		refreshToken = c.token.RefreshToken
	}
	c.tokenLock.RUnlock()

	if refreshToken == "" {
		return c.authenticate()
	}

	c.keyLock.RLock()
	installKey := c.installKey
	c.keyLock.RUnlock()

	if installKey == nil {
		return c.authenticate()
	}

	url := BaseURL + "/auth/refreshtoken"

	payload := map[string]string{
		"username":      c.username,
		"refresh_token": refreshToken,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal refresh payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add authentication headers (only the extra headers for refresh)
	extraHeaders, err := installKey.GenerateExtraHeaders()
	if err != nil {
		return fmt.Errorf("failed to generate extra headers: %w", err)
	}
	for key, value := range extraHeaders {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Warn("Token refresh failed, re-authenticating")
		return c.authenticate()
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode refresh response: %w", err)
	}

	// Token expires in 1 hour based on JWT exp claim
	expiresAt := time.Now().Add(1 * time.Hour)
	c.tokenLock.Lock()
	c.token = &TokenInfo{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		ExpiresAt:    expiresAt,
	}
	c.tokenLock.Unlock()

	logger.Debug("Token refreshed successfully", "expires_at", expiresAt)
	return nil
}

func (c *Client) ensureValidToken() error {
	c.tokenLock.RLock()
	token := c.token
	c.tokenLock.RUnlock()

	if token == nil {
		return c.authenticate()
	}

	// Refresh 5 minutes before expiry
	if time.Now().Add(5 * time.Minute).After(token.ExpiresAt) {
		logger.Debug("Token expiring soon, refreshing", "expires_at", token.ExpiresAt)
		return c.refreshToken()
	}

	return nil
}

func (c *Client) doAuthenticatedRequest(method, url string, body interface{}) (*http.Response, error) {
	return c.doAuthenticatedRequestWithRetry(method, url, body, true)
}

func (c *Client) doAuthenticatedRequestWithRetry(method, url string, body interface{}, allowRetry bool) (*http.Response, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	var reqBody io.Reader
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.tokenLock.RLock()
	accessToken := c.token.AccessToken
	c.tokenLock.RUnlock()

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Add installation headers for all requests
	c.keyLock.RLock()
	installKey := c.installKey
	c.keyLock.RUnlock()

	if installKey != nil {
		extraHeaders, err := installKey.GenerateExtraHeaders()
		if err == nil {
			for key, value := range extraHeaders {
				req.Header.Set(key, value)
			}
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Handle 401 by re-authenticating and retrying once
	if resp.StatusCode == http.StatusUnauthorized && allowRetry {
		resp.Body.Close()
		logger.Info("Received 401, re-authenticating")
		if err := c.authenticate(); err != nil {
			return nil, fmt.Errorf("re-authentication failed: %w", err)
		}
		return c.doAuthenticatedRequestWithRetry(method, url, body, false)
	}

	return resp, nil
}

func (c *Client) Connect() error {
	if err := c.authenticate(); err != nil {
		return err
	}

	// Fetch machine info
	if err := c.fetchMachineInfo(); err != nil {
		return err
	}

	// Get initial status
	if err := c.fetchCurrentMode(); err != nil {
		return err
	}

	return nil
}

func (c *Client) fetchMachineInfo() error {
	url := BaseURL + "/things"

	resp, err := c.doAuthenticatedRequest("GET", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to fetch things: %d - %s", resp.StatusCode, string(body))
	}

	// API returns an array directly, not wrapped in an object
	var things []Thing
	if err := json.NewDecoder(resp.Body).Decode(&things); err != nil {
		return fmt.Errorf("failed to decode things response: %w", err)
	}

	if len(things) == 0 {
		return fmt.Errorf("no machines found in account")
	}

	c.serial = things[0].SerialNumber
	c.model = things[0].ModelName

	logger.Info("Found machine", "serial", c.serial, "model", c.model)
	return nil
}

func (c *Client) fetchCurrentMode() error {
	url := fmt.Sprintf("%s/things/%s/dashboard", BaseURL, c.serial)

	resp, err := c.doAuthenticatedRequest("GET", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to fetch dashboard: %d - %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read dashboard response: %w", err)
	}

	logger.Debug("Dashboard response", "body", string(body))

	// Extract mode and dose info from dashboard
	data := c.extractDataFromDashboard(body)

	c.modeLock.Lock()
	oldMode := c.currentMode
	oldDose1 := c.dose1
	oldDose2 := c.dose2
	oldMachineOn := c.machineOn
	oldBoiler := c.boiler
	oldScale := c.scale
	c.currentMode = data.mode
	c.dose1 = data.dose1
	c.dose2 = data.dose2
	c.machineOn = data.machineOn
	c.boiler = data.boiler
	c.scale = data.scale
	c.modeLock.Unlock()

	// Check if anything changed
	changed := oldMode != data.mode || oldMachineOn != data.machineOn
	if !changed && data.dose1 != nil && (oldDose1 == nil || oldDose1.Weight != data.dose1.Weight) {
		changed = true
	}
	if !changed && data.dose2 != nil && (oldDose2 == nil || oldDose2.Weight != data.dose2.Weight) {
		changed = true
	}
	if !changed && data.boiler != nil && (oldBoiler == nil || oldBoiler.Ready != data.boiler.Ready || oldBoiler.RemainingSeconds != data.boiler.RemainingSeconds) {
		changed = true
	}
	if !changed && data.scale != nil && (oldScale == nil || oldScale.Connected != data.scale.Connected || oldScale.BatteryLevel != data.scale.BatteryLevel) {
		changed = true
	}

	if changed {
		c.notifyStatusChange()
	}

	logger.Debug("Current mode", "mode", data.mode, "dose1", data.dose1, "dose2", data.dose2, "machineOn", data.machineOn, "boiler", data.boiler, "scale", data.scale)
	return nil
}

type dashboardData struct {
	mode      DoseMode
	dose1     *DoseInfo
	dose2     *DoseInfo
	machineOn bool
	boiler    *BoilerInfo
	scale     *ScaleInfo
}

func (c *Client) extractDataFromDashboard(body []byte) dashboardData {
	result := dashboardData{mode: DoseModeContinuous}

	// Parse JSON to find the mode and dose info
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return result
	}

	// Check top-level connected field
	if connected, ok := data["connected"].(bool); ok && connected {
		result.machineOn = true
	}

	// Try to find mode, doses, and machine status in widgets
	if widgets, ok := data["widgets"].([]interface{}); ok {
		for _, w := range widgets {
			widget, ok := w.(map[string]interface{})
			if !ok {
				continue
			}

			// Widget uses "code" field, not "type"
			widgetCode, _ := widget["code"].(string)

			// Extract machine power status from CMMachineStatus widget
			if widgetCode == "CMMachineStatus" {
				if output, ok := widget["output"].(map[string]interface{}); ok {
					if status, ok := output["status"].(string); ok {
						result.machineOn = status == "PoweredOn"
					}
				}
			}

			// Extract brew by weight mode and doses
			if widgetCode == "CMBrewByWeightDoses" || widgetCode == "BrewByWeightDoses" {
				if output, ok := widget["output"].(map[string]interface{}); ok {
					// Extract mode
					if mode, ok := output["mode"].(string); ok {
						result.mode = ParseDoseMode(mode)
					}

					// Extract doses object (e.g., {"Dose1": {"dose": 15.00}, "Dose2": {"dose": 34.00}})
					if doses, ok := output["doses"].(map[string]interface{}); ok {
						if dose1Data, ok := doses["Dose1"].(map[string]interface{}); ok {
							if weight, ok := dose1Data["dose"].(float64); ok && weight > 0 {
								result.dose1 = &DoseInfo{Weight: weight}
							}
						}
						if dose2Data, ok := doses["Dose2"].(map[string]interface{}); ok {
							if weight, ok := dose2Data["dose"].(float64); ok && weight > 0 {
								result.dose2 = &DoseInfo{Weight: weight}
							}
						}
					}
				}
			}

			// Extract boiler status from CMCoffeeBoiler widget
			if widgetCode == "CMCoffeeBoiler" || widgetCode == "CMBoilerStatus" {
				if output, ok := widget["output"].(map[string]interface{}); ok {
					boiler := &BoilerInfo{}
					// Check status string (Ready, Heating, etc.)
					if status, ok := output["status"].(string); ok {
						boiler.Ready = status == "Ready"
					}
					// Get remaining seconds until ready (if heating)
					if remaining, ok := output["remainingSeconds"].(float64); ok {
						boiler.RemainingSeconds = int(remaining)
					}
					if remaining, ok := output["readyStartTime"].(float64); ok && remaining > 0 {
						// Calculate remaining time from ready start time
						now := float64(time.Now().UnixMilli())
						if remaining > now {
							boiler.RemainingSeconds = int((remaining - now) / 1000)
						}
					}
					result.boiler = boiler
				}
			}

			// Extract scale info from ThingScale widget
			if widgetCode == "ThingScale" {
				if output, ok := widget["output"].(map[string]interface{}); ok {
					scale := &ScaleInfo{}
					// Check connected status
					if connected, ok := output["connected"].(bool); ok {
						scale.Connected = connected
					}
					// Get battery level
					if battery, ok := output["batteryLevel"].(float64); ok {
						scale.BatteryLevel = int(battery)
					}
					result.scale = scale
				}
			}
		}
	}

	// Try direct mode field as fallback
	if result.mode == DoseModeContinuous {
		if mode, ok := data["mode"].(string); ok {
			result.mode = ParseDoseMode(mode)
		}
	}

	return result
}

func (c *Client) SetMode(mode DoseMode) error {
	url := fmt.Sprintf("%s/things/%s/command/CoffeeMachineBrewByWeightChangeMode", BaseURL, c.serial)

	payload := SetModeRequest{
		Mode: string(mode),
	}

	resp, err := c.doAuthenticatedRequest("POST", url, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to set mode: %d - %s", resp.StatusCode, string(body))
	}

	c.modeLock.Lock()
	c.currentMode = mode
	c.modeLock.Unlock()

	c.notifyStatusChange()

	logger.Info("Mode set successfully", "mode", mode)
	return nil
}

func (c *Client) SetDose(doseId string, weight float64) error {
	// Use CoffeeMachineBrewByWeightSettingDoses command (from pylamarzocco)
	url := fmt.Sprintf("%s/things/%s/command/CoffeeMachineBrewByWeightSettingDoses", BaseURL, c.serial)

	// Get current dose values
	c.modeLock.RLock()
	dose1Val := 0.0
	dose2Val := 0.0
	if c.dose1 != nil {
		dose1Val = c.dose1.Weight
	}
	if c.dose2 != nil {
		dose2Val = c.dose2.Weight
	}
	c.modeLock.RUnlock()

	// Update the target dose, rounded to 1 decimal
	roundedWeight := float64(int(weight*10)) / 10
	if doseId == "Dose1" {
		dose1Val = roundedWeight
	} else if doseId == "Dose2" {
		dose2Val = roundedWeight
	}

	// Payload requires both doses: {"doses": {"Dose1": 15.0, "Dose2": 34.0}}
	payload := map[string]interface{}{
		"doses": map[string]interface{}{
			"Dose1": dose1Val,
			"Dose2": dose2Val,
		},
	}

	resp, err := c.doAuthenticatedRequest("POST", url, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to set dose: %d - %s", resp.StatusCode, string(body))
	}

	// Update local state
	c.modeLock.Lock()
	if doseId == "Dose1" {
		c.dose1 = &DoseInfo{Weight: roundedWeight}
	} else if doseId == "Dose2" {
		c.dose2 = &DoseInfo{Weight: roundedWeight}
	}
	c.modeLock.Unlock()

	c.notifyStatusChange()

	logger.Info("Dose set successfully", "doseId", doseId, "weight", weight)
	return nil
}

func (c *Client) StartBackFlush() error {
	// Use CoffeeMachineBackFlushStartCleaning command (from pylamarzocco)
	url := fmt.Sprintf("%s/things/%s/command/CoffeeMachineBackFlushStartCleaning", BaseURL, c.serial)

	// Payload format: {"enabled": true}
	payload := map[string]interface{}{
		"enabled": true,
	}

	resp, err := c.doAuthenticatedRequest("POST", url, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to start back flush: %d - %s", resp.StatusCode, string(body))
	}

	logger.Info("Back flush started successfully")
	return nil
}

func (c *Client) GetStatus() MachineStatus {
	c.modeLock.RLock()
	mode := c.currentMode
	dose1 := c.dose1
	dose2 := c.dose2
	machineOn := c.machineOn
	boiler := c.boiler
	scale := c.scale
	c.modeLock.RUnlock()

	return MachineStatus{
		Mode:      mode,
		Connected: c.token != nil,
		Serial:    c.serial,
		Model:     c.model,
		Dose1:     dose1,
		Dose2:     dose2,
		MachineOn: machineOn,
		Boiler:    boiler,
		Scale:     scale,
	}
}

func (c *Client) notifyStatusChange() {
	if c.onStatusChange != nil {
		c.onStatusChange(c.GetStatus())
	}
}

func (c *Client) StartPolling(interval time.Duration, stopCh <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.fetchCurrentMode(); err != nil {
				logger.Error("Failed to poll status", "error", err)
			}
		case <-stopCh:
			return
		}
	}
}
