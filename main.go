// Copyright 2026 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

// Package main provides a tool for analyzing AWS IAM policy versions.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"os"
)

var version = "dev"

func main() {
	region := flag.String("region", "eu-west-1", "AWS region used for IAM communication")
	showVersion := flag.Bool("version", false, "show version")
	verbose := flag.Bool("verbose", false, "verbose log output") // @todo flip

	flag.Parse()

	slog.SetDefault(getLogger(os.Stderr, verbose))

	if *showVersion {
		slog.Info(
			"d2O",
			slog.String("repo", "https://github.com/wakeful/d2O"),
			slog.String("version", version),
		)

		return
	}

	ctx := context.Background()

	app, err := NewApp(ctx, *region, DefaultConfigLoader{})
	if err != nil {
		slog.Error("failed to create app", slog.Any("error", err))

		return
	}

	multipleVersions, unusedPolicies, err := app.getPolicies(ctx)
	if err != nil {
		slog.Error("failed to run", slog.Any("error", err))
	}

	output := Output{
		UnusedPolicies:       unusedPolicies,
		WithMultipleVersions: multipleVersions,
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		slog.Error("failed to marshal output", slog.Any("error", err))

		return
	}

	_, _ = os.Stdout.Write(jsonData)
}
