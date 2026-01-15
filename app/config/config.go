package config

import (
	"encoding/json"
	"os"

	"github.com/philipparndt/go-logger"
	"github.com/philipparndt/mqtt-gateway/config"
)

var cfg Config

type TriggerCondition struct {
	Selector string      `json:"selector"` // JSON path (e.g., "button", "event")
	Value    interface{} `json:"value"`    // Expected value (number, string, bool)
}

type TriggerAction struct {
	Mode string `json:"mode"` // Dose mode to set
}

type Trigger struct {
	Topic      string             `json:"topic"`
	Conditions []TriggerCondition `json:"conditions"`
	Action     TriggerAction      `json:"action"`
}

type Config struct {
	MQTT       config.MQTTConfig `json:"mqtt"`
	LaMarzocco LaMarzoccoConfig  `json:"lamarzocco"`
	Web        WebConfig         `json:"web"`
	Triggers   []Trigger         `json:"triggers,omitempty"`
	LogLevel   string            `json:"loglevel,omitempty"`
}

type WebConfig struct {
	Enabled bool `json:"enabled"`
	Port    int  `json:"port"`
}

type LaMarzoccoConfig struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	PollingInterval int    `json:"polling_interval"`
}

func LoadConfig(file string) (Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		logger.Error("Error reading config file", err)
		return Config{}, err
	}

	data = config.ReplaceEnvVariables(data)

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		logger.Error("Unmarshaling JSON:", err)
		return Config{}, err
	}

	// Set default values
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	if cfg.LaMarzocco.PollingInterval == 0 {
		cfg.LaMarzocco.PollingInterval = 30
	}

	if cfg.Web.Port == 0 {
		cfg.Web.Port = 8080
	}

	return cfg, nil
}

func Get() Config {
	return cfg
}
