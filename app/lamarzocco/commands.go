package lamarzocco

import (
	"encoding/json"
	"fmt"
)

type Command struct {
	Mode      string   `json:"mode,omitempty"`
	Dose1     *float64 `json:"dose1,omitempty"`     // Weight in grams for Dose1
	Dose2     *float64 `json:"dose2,omitempty"`     // Weight in grams for Dose2
	BackFlush *bool    `json:"backflush,omitempty"` // Start back flush cycle
}

func ParseCommand(payload []byte) (*Command, error) {
	var cmd Command
	if err := json.Unmarshal(payload, &cmd); err != nil {
		return nil, fmt.Errorf("failed to parse command: %w", err)
	}

	// At least one field must be set
	if cmd.Mode == "" && cmd.Dose1 == nil && cmd.Dose2 == nil && cmd.BackFlush == nil {
		return nil, fmt.Errorf("mode, dose1, dose2, or backflush is required")
	}

	return &cmd, nil
}

func (c *Command) GetDoseMode() DoseMode {
	return ParseDoseMode(c.Mode)
}

func (c *Command) HasMode() bool {
	return c.Mode != ""
}

func (c *Command) HasDose1() bool {
	return c.Dose1 != nil
}

func (c *Command) HasDose2() bool {
	return c.Dose2 != nil
}

func (c *Command) GetDose1() float64 {
	if c.Dose1 != nil {
		return *c.Dose1
	}
	return 0
}

func (c *Command) GetDose2() float64 {
	if c.Dose2 != nil {
		return *c.Dose2
	}
	return 0
}

func (c *Command) HasBackFlush() bool {
	return c.BackFlush != nil && *c.BackFlush
}
