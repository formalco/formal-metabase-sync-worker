package main

import (
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"
)

type MetabaseIntegration struct {
	UseAPIKey        bool
	MetabaseAPIKey   string
	MetabaseHostname string
	MetabaseUsername string
	MetabasePwd      string
	Version          string
}

func MetabaseWorkflow(metabaseIntegration MetabaseIntegration, apiKey, integrationID string, verifyTLS bool, cfAccessClientID, cfAccessClientSecret string) error {
	client := New(apiKey)

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

	log.Info().Msg("Fetching users from Formal")
	users, err := client.ListHumanFormalUsers()
	if err != nil {
		return fmt.Errorf("failed to fetch Formal users: %w", err)
	}
	log.Info().Int("count", len(users)).Msg("Fetched users from Formal")

	log.Info().Msg("Mapping Metabase users to Formal users")
	mappedUserCount := 0
	skippedUserCount := 0
	for _, user := range users {
		metabaseUser, exists := metabaseRoles[user.Email]
		if exists {
			metabaseUserExternalId := strconv.Itoa(metabaseUser.Id)
			alreadyMapped := false
			for _, existingExternalId := range user.ExternalIds {
				if existingExternalId.ExternalId == metabaseUserExternalId && existingExternalId.AppId == integrationID {
					alreadyMapped = true
					break
				}
			}
			if alreadyMapped {
				log.Debug().Str("email", user.Email).Str("external_id", metabaseUserExternalId).Msg("User already mapped, skipping")
				skippedUserCount++
				continue
			}

			err = client.MapUserToExternalId(user.Id, metabaseUserExternalId, integrationID)
			if err != nil {
				return fmt.Errorf("failed to map user %s: %w", user.Email, err)
			}
			log.Debug().Str("email", user.Email).Str("external_id", metabaseUserExternalId).Msg("Mapped user")
			mappedUserCount++
		}
	}
	log.Info().Int("mapped", mappedUserCount).Int("skipped", skippedUserCount).Msg("Sync completed")
	return nil
}
