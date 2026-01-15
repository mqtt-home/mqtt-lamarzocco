package lamarzocco

import (
	"encoding/json"
	"fmt"
)

type Command struct {
	Mode string `json:"mode"`
}

func ParseCommand(payload []byte) (*Command, error) {
	var cmd Command
	if err := json.Unmarshal(payload, &cmd); err != nil {
		return nil, fmt.Errorf("failed to parse command: %w", err)
	}

	if cmd.Mode == "" {
		return nil, fmt.Errorf("mode is required")
	}

	return &cmd, nil
}

func (c *Command) GetDoseMode() DoseMode {
	return ParseDoseMode(c.Mode)
}
