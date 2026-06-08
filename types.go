// Copyright 2026 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

type PolicyVersion struct {
	CreateDate *time.Time `json:"createDate"`
	VersionID  *string    `json:"versionID"` //nolint:tagliatelle
}

type Output struct {
	UnusedPolicies       []string                   `json:"unused"`
	WithMultipleVersions map[string][]PolicyVersion `json:"withVersions"`
}

// ServiceIAM lists IAM policies and their versions using the AWS SDK.
type ServiceIAM interface {
	iam.ListPoliciesAPIClient
	iam.ListPolicyVersionsAPIClient
}

var (
	_ iam.ListPoliciesAPIClient       = (ServiceIAM)(nil)
	_ iam.ListPolicyVersionsAPIClient = (ServiceIAM)(nil)
)

// App provides functionality for interacting with the AWS IAM service.
type App struct {
	client ServiceIAM
}

// ConfigLoader defines an interface for loading AWS SDK config with customizable options.
type ConfigLoader interface {
	LoadDefaultConfig(
		ctx context.Context,
		optFns ...func(*config.LoadOptions) error,
	) (cfg aws.Config, err error)
}

// DefaultConfigLoader loads default AWS SDK config with optional customization.
type DefaultConfigLoader struct{}

// LoadDefaultConfig loads the default AWS SDK config with optional modifiers.
//
//nolint:nonamedreturns
func (d DefaultConfigLoader) LoadDefaultConfig(
	ctx context.Context,
	optFns ...func(*config.LoadOptions) error,
) (cfg aws.Config, err error) {
	return config.LoadDefaultConfig(ctx, optFns...) //nolint:wrapcheck
}

var _ ConfigLoader = (*DefaultConfigLoader)(nil)
