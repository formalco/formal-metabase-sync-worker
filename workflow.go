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
		key, err := RefreshMetabaseSessionKey(metabaseIntegration, verifyTLS, cfAccessClientID, cfAccessClientSecret)
		if err != nil {
			return fmt.Errorf("error refreshing Metabase session key: %w", err)
		}

		sessionKey = key
	}

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
		return fmt.Errorf("error getting Metabase roles: %w", err)
	}

	users, err := client.ListHumanFormalUsers()
	if err != nil {
		return fmt.Errorf("error listing Formal users: %w", err)
	}

	mappedUserCount := 0
	for _, user := range users {
		metabaseUser, exists := metabaseRoles[user.Email]
		if exists {
			metabaseUserExternalId := strconv.Itoa(metabaseUser.Id)
			alreadyMapped := false
			for _, existingExternalId := range user.ExternalIds {
				if existingExternalId.ExternalId == metabaseUserExternalId && existingExternalId.AppId == integrationID {
					log.Debug().Msgf("Application %s has user %s already mapped to external ID %s", existingExternalId.AppId, user.Id, existingExternalId.ExternalId)
					alreadyMapped = true
					break
				}
			}
			if alreadyMapped {
				log.Debug().Msgf("User %s is already mapped to external ID %s", user.Email, metabaseUserExternalId)
				continue
			}

			err = client.MapUserToExternalId(user.Id, metabaseUserExternalId, integrationID)
			if err != nil {
				return fmt.Errorf("error mapping user %s to external ID %s: %w", user.Email, metabaseUserExternalId, err)
			}
			mappedUserCount++
		}
	}
	log.Info().Msgf("Metabase sync has finished. %d new user(s) mapped.", mappedUserCount)
	return nil
}
