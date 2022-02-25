package metric

import (
	"github.com/saucelabs/webserver/internal/expvar"
	"github.com/saucelabs/webserver/internal/validation"
)

//////
// Consts, and vars.
//////

var (
	// CommandLine metric.
	CommandLine = expvar.CommandLine

	// MemoryStats metric.
	MemoryStats = expvar.MemoryStats
)

//////
// Definition.
//////

// Metric definition.
type Metric struct {
	// Name of the metric.
	Name string `json:"name" validate:"required"`

	// Var is a valid ExpVar.
	Var expvar.Var `json:"var" validate:"required"`
}

//////
// Metrics.
//////

// Server information.
func Server(address, name string, pid int) expvar.Func {
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
func New(name string, v expvar.Var) (*Metric, error) {
	m := &Metric{
		Name: name,
		Var:  v,
	}

	if err := validation.ValidateStruct(m); err != nil {
		return nil, err
	}

	return m, nil
}
