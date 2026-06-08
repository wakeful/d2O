// Copyright 2026 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

// ErrEmptyRegion indicates a missing or empty region parameter.
var ErrEmptyRegion = errors.New("region cannot be empty")

// NewApp returns a new App configured with the given region.
func NewApp(ctx context.Context, region string, loader ConfigLoader) (*App, error) {
	if region == "" {
		return nil, ErrEmptyRegion
	}

	if loader == nil {
		loader = DefaultConfigLoader{}
	}

	cfg, err := loader.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %w", err)
	}

	return &App{
		client: iam.NewFromConfig(cfg),
	}, nil
}

func (a *App) getPolicies(ctx context.Context) (map[string][]PolicyVersion, []string, error) {
	paginator := iam.NewListPoliciesPaginator(a.client, &iam.ListPoliciesInput{
		Scope: types.PolicyScopeTypeLocal,
	})

	policyVersions := make(map[string][]PolicyVersion)

	unusedPolicies := make([]string, 0)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list policies: %w", err)
		}

		slog.Debug("found policies", slog.Int("count", len(page.Policies)))

		for _, policy := range page.Policies {
			slog.Debug(
				"found policy",
				slog.String("arn", *policy.Arn),
				slog.String("default_version", *policy.DefaultVersionId),
				slog.Int("n_attach", int(*policy.AttachmentCount)),
				slog.Int("n_perm_bound", int(*policy.PermissionsBoundaryUsageCount)),
			)

			if int(*policy.AttachmentCount) == 0 &&
				int(*policy.PermissionsBoundaryUsageCount) == 0 {
				slog.Debug("policy is not used", slog.String("arn", *policy.Arn))
				unusedPolicies = append(unusedPolicies, *policy.Arn)
			}

			versions, err := a.getVersions(ctx, policy.Arn)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get versions: %w", err)
			}

			if len(versions) < 1 {
				slog.Debug("no versions found", slog.String("arn", *policy.Arn))

				continue
			}

			policyVersions[*policy.Arn] = versions
		}
	}

	return policyVersions, unusedPolicies, nil
}

func (a *App) getVersions(
	ctx context.Context,
	policyARN *string,
) ([]PolicyVersion, error) {
	paginator := iam.NewListPolicyVersionsPaginator(
		a.client,
		&iam.ListPolicyVersionsInput{
			PolicyArn: policyARN,
		},
	)

	var versions []PolicyVersion

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			slog.Error("failed to list policy versions",
				slog.Any("error", err),
			)

			return nil, fmt.Errorf("failed to list policy versions: %w", err)
		}

		for _, version := range page.Versions {
			slog.Debug(
				"found version",
				slog.String("arn", *policyARN),
				slog.String("created", version.CreateDate.Format(time.RFC822)),
				slog.Bool("current_version", version.IsDefaultVersion),
				slog.String("version", *version.VersionId),
			)

			if version.IsDefaultVersion {
				continue
			}

			versions = append(versions, PolicyVersion{
				CreateDate: version.CreateDate,
				VersionID:  version.VersionId,
			})
		}
	}

	return versions, nil
}
