// Copyright 2021 The webserver Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package logger

import (
	"log"

	"github.com/saucelabs/sypl"
	"github.com/saucelabs/sypl/level"
	"github.com/saucelabs/sypl/output"
	"github.com/saucelabs/sypl/processor"
	"github.com/saucelabs/sypl/status"
)

// Global, singleton, cached logger. It's safe to be retrieved via `Get`.
var l *sypl.Sypl

// Get safely returns the global application logger.
func Get() *sypl.Sypl {
	if l != nil {
		return l
	}

	log.Fatalln("Logger isn't setup")

	return nil
}

// Setup logger.
func Setup(name, logLevel, logFilePath string) *sypl.Sypl {
	logLevelAsLevel := level.MustFromString(logLevel)

	l = sypl.NewDefault(
		name,
		logLevelAsLevel,
		processor.ChangeFirstCharCase(processor.Lowercase),
	)

	// Only enable File output if path is set.
	if logFilePath != "" {
		l.AddOutputs(output.File(
			logFilePath,
			logLevelAsLevel,
			processor.ChangeFirstCharCase(processor.Lowercase),
		))

		// "-" special case makes the File Output behave as Console, also
		// writing to `stdout` causing duplicated messages.
		if logFilePath == "-" {
			l.GetOutput("Console").SetStatus(status.Disabled)
		}
	}

	return l
}
