// Copyright 2026 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"io"
	"log/slog"
)

// getLogger returns a slog.Logger configured with the given output and log level.
// If verbose is true, the log level is set to debug; otherwise, it defaults to info.
func getLogger(output io.Writer, verbose *bool) *slog.Logger {
	logLevel := slog.LevelInfo
	if verbose != nil && *verbose {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(
		slog.NewTextHandler(output, &slog.HandlerOptions{
			AddSource:   false,
			Level:       logLevel,
			ReplaceAttr: nil,
		}),
	)

	return logger
}
