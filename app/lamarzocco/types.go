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

type MachineStatus struct {
	Mode      DoseMode `json:"mode"`
	Connected bool     `json:"connected"`
	Serial    string   `json:"serial,omitempty"`
	Model     string   `json:"model,omitempty"`
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
