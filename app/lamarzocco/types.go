package lamarzocco

import "time"

type DoseMode string

const (
	DoseModeDose1      DoseMode = "Dose1"
	DoseModeDose2      DoseMode = "Dose2"
	DoseModeContinuous DoseMode = "Continuous"
)

func (d DoseMode) DisplayName() string {
	switch d {
	case DoseModeDose1:
		return "Dose 1"
	case DoseModeDose2:
		return "Dose 2"
	case DoseModeContinuous:
		return "Continuous"
	default:
		return string(d)
	}
}

func ParseDoseMode(s string) DoseMode {
	switch s {
	case "Dose1", "dose1":
		return DoseModeDose1
	case "Dose2", "dose2":
		return DoseModeDose2
	case "Continuous", "continuous", "Off", "off":
		return DoseModeContinuous
	default:
		return DoseModeContinuous
	}
}

type DoseInfo struct {
	Weight float64 `json:"weight"` // Weight in grams
}

type BoilerInfo struct {
	Ready           bool `json:"ready"`
	RemainingSeconds int  `json:"remainingSeconds,omitempty"` // Seconds until ready (0 if ready)
}

type ScaleInfo struct {
	Connected    bool `json:"connected"`
	BatteryLevel int  `json:"batteryLevel,omitempty"` // Battery percentage 0-100
}

type MachineStatus struct {
	Mode      DoseMode    `json:"mode"`
	Connected bool        `json:"connected"`
	Serial    string      `json:"serial,omitempty"`
	Model     string      `json:"model,omitempty"`
	Dose1     *DoseInfo   `json:"dose1,omitempty"`
	Dose2     *DoseInfo   `json:"dose2,omitempty"`
	MachineOn bool        `json:"machineOn"`
	Boiler    *BoilerInfo `json:"boiler,omitempty"`
	Scale     *ScaleInfo  `json:"scale,omitempty"`
}

type AuthResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type TokenInfo struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type ThingsResponse struct {
	Things []Thing `json:"things"`
}

type Thing struct {
	SerialNumber string `json:"serialNumber"`
	ModelName    string `json:"modelName"`
	Name         string `json:"name"`
}

type DashboardResponse struct {
	Widgets []Widget `json:"widgets"`
}

type Widget struct {
	Type   string      `json:"type"`
	Output interface{} `json:"output"`
}

type BrewByWeightOutput struct {
	Mode string `json:"mode"`
}

type SetModeRequest struct {
	Mode string `json:"mode"`
}

type SetDoseRequest struct {
	DoseId string  `json:"doseId"` // "Dose1" or "Dose2"
	Dose   float64 `json:"dose"`   // Weight in grams
}
