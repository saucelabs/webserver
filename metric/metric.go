package metric

import (
	"github.com/saucelabs/webserver/internal/validation"
)

//////
// Definition.
//////

// Metric definition.
type Metric struct {
	// Name of the metric.
	Name string `json:"name" validate:"required"`

	// Var is a valid ExpVar.
	Value Var `json:"value" validate:"required"`
}

//////
// Metrics.
//////

// Server information.
func Server(address, name string, pid int) Func {
	return func() interface{} {
		return struct {
			// Server address.
			Address string `json:"Address"`

			// Server name.
			Name string `json:"Name"`

			// Server PID.
			PID int `json:"PID"`
		}{
			address, name, pid,
		}
	}
}

//////
// Factory.
//////

// New is the Metric factory.
func New(name string, value Var) (*Metric, error) {
	m := &Metric{
		Name:  name,
		Value: value,
	}

	if err := validation.ValidateStruct(m); err != nil {
		return nil, err
	}

	return m, nil
}
