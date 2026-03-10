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

type OmniIntegration struct {
	APIKey        string
	Hostname      string
	IntegrationID string
}

func MetabaseWorkflow(metabaseIntegration MetabaseIntegration, formalClient *Client, users []User, integrationID string, verifyTLS bool, cfAccessClientID, cfAccessClientSecret string) error {
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

			err = formalClient.CreateUserExternalId(user.Id, metabaseUserExternalId, integrationID, "This External ID was imported for this user via Metabase.")
			if err != nil {
				return fmt.Errorf("failed to map user %s: %w", user.Email, err)
			}
			log.Debug().Str("email", user.Email).Str("external_id", metabaseUserExternalId).Msg("Mapped user")
			mappedUserCount++
		}
	}
	log.Info().Int("mapped", mappedUserCount).Int("skipped", skippedUserCount).Msg("Metabase sync completed")
	return nil
}

func OmniWorkflow(omniIntegration OmniIntegration, formalClient *Client, users []User) error {
	log.Info().Str("hostname", omniIntegration.Hostname).Msg("Fetching users from Omni")
	omniUsers, err := GetOmniUsers(omniIntegration.Hostname, omniIntegration.APIKey)
	if err != nil {
		return fmt.Errorf("failed to fetch Omni users: %w", err)
	}
	log.Info().Int("count", len(omniUsers)).Msg("Fetched users from Omni")

	log.Info().Msg("Mapping Omni users to Formal users")
	mappedUserCount := 0
	skippedUserCount := 0
	for _, user := range users {
		omniUser, exists := omniUsers[user.Email]
		if exists {
			omniUserExternalId := omniUser.Id
			alreadyMapped := false
			for _, existingExternalId := range user.ExternalIds {
				if existingExternalId.ExternalId == omniUserExternalId && existingExternalId.AppId == omniIntegration.IntegrationID {
					alreadyMapped = true
					break
				}
			}
			if alreadyMapped {
				log.Debug().Str("email", user.Email).Str("external_id", omniUserExternalId).Msg("User already mapped, skipping")
				skippedUserCount++
				continue
			}

			err = formalClient.CreateUserExternalId(user.Id, omniUserExternalId, omniIntegration.IntegrationID, "This External ID was imported for this user via Omni.")
			if err != nil {
				return fmt.Errorf("failed to map user %s: %w", user.Email, err)
			}
			log.Debug().Str("email", user.Email).Str("external_id", omniUserExternalId).Msg("Mapped user")
			mappedUserCount++
		}
	}
	log.Info().Int("mapped", mappedUserCount).Int("skipped", skippedUserCount).Msg("Omni sync completed")
	return nil
}
