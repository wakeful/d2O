// Copyright 2026 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

type mockConfigLoader struct {
	mockConfig    aws.Config
	mockConfigErr error
}

//nolint:nonamedreturns
func (m *mockConfigLoader) LoadDefaultConfig(
	_ context.Context,
	_ ...func(*config.LoadOptions) error,
) (cfg aws.Config, err error) {
	return m.mockConfig, m.mockConfigErr
}

var _ ConfigLoader = (*mockConfigLoader)(nil)

func TestNewApp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		loader  *mockConfigLoader
		region  string
		wantApp bool
		wantErr bool
	}{
		{
			name:    "region cannot be empty",
			region:  "",
			wantApp: false,
			wantErr: true,
		},

		{
			name: "config loader error",
			loader: &mockConfigLoader{
				mockConfigErr: errors.New("test error"),
			},
			region:  "eu-west-1",
			wantApp: false,
			wantErr: true,
		},
		{
			name: "success setup",
			loader: &mockConfigLoader{
				mockConfig: aws.Config{},
			},
			region:  "eu-west-1",
			wantApp: true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewApp(t.Context(), tt.region, tt.loader)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewApp() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if (got != nil) != tt.wantApp {
				t.Errorf("got app = %v, want non-nil: %v", got, tt.wantApp)
			}
		})
	}
}

type MockServiceIAM struct {
	mockPolicies            []types.Policy
	mockPoliciesError       error
	mockPolicyVersions      map[string][]types.PolicyVersion
	mockPolicyVersionsError error
}

func (m MockServiceIAM) ListPolicies(
	_ context.Context,
	_ *iam.ListPoliciesInput,
	_ ...func(*iam.Options),
) (*iam.ListPoliciesOutput, error) {
	return &iam.ListPoliciesOutput{Policies: m.mockPolicies}, m.mockPoliciesError
}

func (m MockServiceIAM) ListPolicyVersions(
	_ context.Context,
	input *iam.ListPolicyVersionsInput,
	_ ...func(*iam.Options),
) (*iam.ListPolicyVersionsOutput, error) {
	arn := *input.PolicyArn

	v, ok := m.mockPolicyVersions[arn]
	if !ok {
		return nil, fmt.Errorf("no mock policy version found for arn %s", arn)
	}

	return &iam.ListPolicyVersionsOutput{Versions: v}, m.mockPolicyVersionsError
}

var _ ServiceIAM = (*MockServiceIAM)(nil)

func TestApp_getPolicies(t *testing.T) {
	t.Parallel()

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	manyVersions := []types.PolicyVersion{
		{
			CreateDate:       &now,
			IsDefaultVersion: true,
			VersionId:        aws.String("v2"),
		},
		{
			CreateDate:       &yesterday,
			IsDefaultVersion: false,
			VersionId:        aws.String("v1"),
		},
	}
	policyVersions := map[string][]types.PolicyVersion{
		"arn:aws:iam::0123456789:policy/many-versions": manyVersions,
		"arn:aws:iam::0123456789:policy/one-version": {
			{
				CreateDate:       &yesterday,
				IsDefaultVersion: true,
				VersionId:        aws.String("v1"),
			},
		},
	}
	tests := []struct {
		name    string
		client  ServiceIAM
		want    map[string][]PolicyVersion
		wantErr bool
	}{
		{
			name: "give back some data",
			client: &MockServiceIAM{
				mockPolicies: []types.Policy{
					{
						Arn: aws.String(
							"arn:aws:iam::0123456789:policy/many-versions",
						),
						AttachmentCount:               aws.Int32(0),
						CreateDate:                    &now,
						DefaultVersionId:              aws.String("v2"),
						PermissionsBoundaryUsageCount: aws.Int32(0),
					},
					{
						Arn: aws.String(
							"arn:aws:iam::0123456789:policy/one-version",
						),
						AttachmentCount:               aws.Int32(0),
						CreateDate:                    &yesterday,
						DefaultVersionId:              aws.String("v1"),
						PermissionsBoundaryUsageCount: aws.Int32(0),
					},
				},
				mockPoliciesError:  nil,
				mockPolicyVersions: policyVersions,
			},
			want: map[string][]PolicyVersion{
				"arn:aws:iam::0123456789:policy/many-versions": {
					{CreateDate: &yesterday, VersionID: aws.String("v1")},
				},
			},
			wantErr: false,
		},
		{
			name: "return error when accessing policy version",
			client: &MockServiceIAM{
				mockPolicies: []types.Policy{
					{
						Arn: aws.String(
							"arn:aws:iam::0123456789:policy/error-on-this",
						),
						AttachmentCount:               aws.Int32(0),
						CreateDate:                    &now,
						DefaultVersionId:              aws.String("v2"),
						PermissionsBoundaryUsageCount: aws.Int32(0),
					},
				},
				mockPolicyVersions:      make(map[string][]types.PolicyVersion),
				mockPolicyVersionsError: errors.New("unable to find any version for policy"),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "return error when listening policies",
			client: &MockServiceIAM{
				mockPoliciesError:  errors.New("unable to find any policies"),
				mockPolicyVersions: make(map[string][]types.PolicyVersion),
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a := &App{
				client: tt.client,
			}

			got, _, err := a.getPolicies(t.Context())
			if (err != nil) != tt.wantErr {
				t.Errorf("getPolicies() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getPolicies() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApp_getVersions(t *testing.T) {
	t.Parallel()

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	policyVersions := []types.PolicyVersion{
		{
			CreateDate:       &now,
			IsDefaultVersion: true,
			VersionId:        aws.String("v2"),
		},
		{
			CreateDate:       &yesterday,
			IsDefaultVersion: false,
			VersionId:        aws.String("v1"),
		},
	}
	client := &MockServiceIAM{
		mockPolicyVersions: map[string][]types.PolicyVersion{
			"arn:aws:iam::0123456789:policy/many-versions": policyVersions,
		},
	}
	tests := []struct {
		name      string
		client    ServiceIAM
		policyARN *string
		want      []PolicyVersion
		wantErr   bool
	}{
		{
			name:      "multiple versions for policy",
			client:    client,
			policyARN: aws.String("arn:aws:iam::0123456789:policy/many-versions"),
			want: []PolicyVersion{
				{CreateDate: &yesterday, VersionID: aws.String("v1")},
			},
			wantErr: false,
		},
		{
			name:      "no versions for policy",
			client:    client,
			policyARN: aws.String("arn:aws:iam::0123456789:policy/no-versions"),
			want:      nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a := &App{
				client: tt.client,
			}

			got, err := a.getVersions(t.Context(), tt.policyARN)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPolicyVersions() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getPolicyVersions() got = %v, want %v", got, tt.want)
			}
		})
	}
}
