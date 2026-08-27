package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

type MetabaseIntegration struct {
	UseAPIKey        bool
	MetabaseAPIKey   string
	MetabaseHostname string
	MetabaseUsername string
	MetabasePwd      string
	Version          string
}

type OmniIntegration struct {
	APIKey        string
	Hostname      string
	IntegrationID string
}

func MetabaseWorkflow(ctx context.Context, metabaseIntegration MetabaseIntegration, formalClient *Client, users []User, integrationID string, verifyTLS bool, cfAccessClientID, cfAccessClientSecret string) error {
	sessionKey := ""

	if !metabaseIntegration.UseAPIKey {
		log.Info().Str("hostname", metabaseIntegration.MetabaseHostname).Msg("Authenticating to Metabase")
		key, err := RefreshMetabaseSessionKey(metabaseIntegration, verifyTLS, cfAccessClientID, cfAccessClientSecret)
		if err != nil {
			return fmt.Errorf("failed to authenticate to Metabase: %w", err)
		}

		sessionKey = key
	}

	log.Info().Str("hostname", metabaseIntegration.MetabaseHostname).Msg("Fetching users from Metabase")
	metabaseRoles, err := GetMetabaseRoles(
		metabaseIntegration.MetabaseHostname,
		metabaseIntegration.Version,
		metabaseIntegration.MetabaseAPIKey,
		sessionKey,
		metabaseIntegration.UseAPIKey,
		verifyTLS,
		cfAccessClientID,
		cfAccessClientSecret,
	)
	if err != nil {
		return fmt.Errorf("failed to fetch Metabase users: %w", err)
	}
	log.Info().Int("count", len(metabaseRoles)).Msg("Fetched users from Metabase")

	log.Info().Msg("Mapping Metabase users to Formal users")
	mappedUserCount, skippedUserCount, err := mapUsersToIntegration(
		ctx,
		formalClient,
		users,
		lo.MapValues(metabaseRoles, func(user MetabaseUser, _ string) string {
			return strconv.Itoa(user.Id)
		}),
		integrationID,
		"Metabase",
	)
	if err != nil {
		return err
	}
	log.Info().Int("mapped", mappedUserCount).Int("skipped", skippedUserCount).Msg("Metabase sync completed")
	return nil
}

func OmniWorkflow(ctx context.Context, omniIntegration OmniIntegration, formalClient *Client, users []User) error {
	log.Info().Str("hostname", omniIntegration.Hostname).Msg("Fetching users from Omni")
	omniUsers, err := GetOmniUsers(omniIntegration.Hostname, omniIntegration.APIKey)
	if err != nil {
		return fmt.Errorf("failed to fetch Omni users: %w", err)
	}
	log.Info().Int("count", len(omniUsers)).Msg("Fetched users from Omni")

	log.Info().Msg("Mapping Omni users to Formal users")
	mappedUserCount, skippedUserCount, err := mapUsersToIntegration(
		ctx,
		formalClient,
		users,
		lo.MapValues(omniUsers, func(user OmniUser, _ string) string {
			return user.Id
		}),
		omniIntegration.IntegrationID,
		"Omni",
	)
	if err != nil {
		return err
	}
	log.Info().Int("mapped", mappedUserCount).Int("skipped", skippedUserCount).Msg("Omni sync completed")
	return nil
}

type userExternalIDClient interface {
	CreateUserExternalId(ctx context.Context, userId, externalId, integrationID, description string) error
}

func mapUsersToIntegration(ctx context.Context, formalClient userExternalIDClient, users []User, externalIDsByEmail map[string]string, integrationID, source string) (mapped, skipped int, err error) {
	for _, user := range users {
		externalID, exists := externalIDsByEmail[user.Email]
		if !exists {
			continue
		}
		if hasExternalID(user, externalID, integrationID) {
			log.Debug().Str("email", user.Email).Str("external_id", externalID).Msg("User already mapped, skipping")
			skipped++
			continue
		}

		err = formalClient.CreateUserExternalId(ctx, user.Id, externalID, integrationID, fmt.Sprintf("This External ID was imported for this user via %s.", source))
		if err != nil {
			return mapped, skipped, fmt.Errorf("failed to map user %s: %w", user.Email, err)
		}
		log.Debug().Str("email", user.Email).Str("external_id", externalID).Msg("Mapped user")
		mapped++
	}
	return mapped, skipped, nil
}

func hasExternalID(user User, externalID, appID string) bool {
	return lo.ContainsBy(user.ExternalIds, func(existing ExternalId) bool {
		return existing.ExternalId == externalID && existing.AppId == appID
	})
}
