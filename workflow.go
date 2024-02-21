package main

import (
	"strconv"

	"github.com/rs/zerolog/log"
)

type MetabaseIntegration struct {
	MetabaseHostname string
	MetabaseUsername string
	MetabasePwd      string
	Version          string
}

func MetabaseWorkflow(metabaseIntegration MetabaseIntegration, apiKey, integrationID string) error {
	client := New(apiKey)

	sessionKey, err := RefreshMetabaseSessionKey(metabaseIntegration)
	if err != nil {
		return err
	}

	metabaseRoles, err := GetMetabaseRoles(metabaseIntegration.MetabaseHostname, metabaseIntegration.Version, sessionKey)
	if err != nil {
		return err
	}

	users, err := client.ListHumanFormalUsers()
	if err != nil {
		return err
	}

	mappedUserCount := 0
	for _, user := range users {
		metabaseUser, exists := metabaseRoles[user.Email]
		if exists {
			metabaseUserExternalId := strconv.Itoa(metabaseUser.Id)
			alreadyMapped := false
			for _, existingExternalId := range user.ExternalIds {
				if existingExternalId.ExternalId == metabaseUserExternalId && existingExternalId.AppId == integrationID {
					log.Info().Msgf("Application %s has user %s already mapped to external ID %s", existingExternalId.AppId, user.Id, existingExternalId.ExternalId)
					alreadyMapped = true
					break
				}
			}
			if alreadyMapped {
				log.Info().Msgf("User %s is already mapped to external ID %s", user.Email, metabaseUserExternalId)
				continue
			}

			err = client.MapUserToExternalId(user.Id, metabaseUserExternalId, integrationID)
			if err != nil {
				log.Error().Err(err)
				return err
			}
			mappedUserCount++
		}
	}
	log.Info().Msgf("Metabase sync has finished. %d new user(s) mapped.", mappedUserCount)
	return nil
}
