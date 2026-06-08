// Copyright 2026 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func Test_getLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		verbose          *bool
		testMessage      string
		useLevel         slog.Level
		expectInOutput   bool
		testOutputSearch string
	}{
		{
			name:             "returns configured logger with verbose=false",
			verbose:          aws.Bool(false),
			testMessage:      "debug message",
			useLevel:         slog.LevelDebug,
			expectInOutput:   false,
			testOutputSearch: "debug message",
		},
		{
			name:             "enables debug logging when verbose=true",
			verbose:          aws.Bool(true),
			testMessage:      "debug message",
			useLevel:         slog.LevelDebug,
			expectInOutput:   true,
			testOutputSearch: "debug message",
		},
		{
			name:             "info messages are always logged",
			verbose:          aws.Bool(false),
			testMessage:      "info message",
			useLevel:         slog.LevelInfo,
			expectInOutput:   true,
			testOutputSearch: "info message",
		},
		{
			name:             "handles nil verbose flag",
			verbose:          nil,
			testMessage:      "info message",
			useLevel:         slog.LevelInfo,
			expectInOutput:   true,
			testOutputSearch: "info message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			logger := getLogger(&buf, tt.verbose)

			if logger == nil {
				t.Fatal("Expected logger to not be nil")
			}

			switch tt.useLevel {
			case slog.LevelDebug:
				logger.Debug(tt.testMessage)
			case slog.LevelInfo:
				logger.Info(tt.testMessage)
			case slog.LevelWarn:
				logger.Warn(tt.testMessage)
			case slog.LevelError:
				logger.Error(tt.testMessage)
			}

			logOutput := buf.String()
			containsMessage := strings.Contains(logOutput, tt.testOutputSearch)

			if containsMessage != tt.expectInOutput {
				if tt.expectInOutput {
					t.Errorf("Expected log output to contain '%s', but it did not. Got: %s",
						tt.testOutputSearch, logOutput)
				} else {
					t.Errorf("Expected log output to NOT contain '%s', but it did. Got: %s",
						tt.testOutputSearch, logOutput)
				}
			}
		})
	}
}
