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
	formalSdk := New(apiKey)

	sessionKey, err := RefreshMetabaseSessionKey(metabaseIntegration)
	if err != nil {
		return err
	}

	metabaseRoles, err := GetMetabaseRoles(metabaseIntegration.MetabaseHostname, metabaseIntegration.Version, sessionKey)
	if err != nil {
		return err
	}

	users, err := formalSdk.ListHumanFormalUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		metabaseUser, exists := metabaseRoles[user.Email]
		if exists {
			metabaseUserExternalId := strconv.Itoa(metabaseUser.Id)
			alreadyMapped := false
			for _, existingExternalId := range user.ExternalIds {
				if existingExternalId.ExternalId == metabaseUserExternalId && existingExternalId.AppId == integrationID {
					log.Info().Msg("Already mapped " + existingExternalId.AppId + " and " + existingExternalId.ExternalId + " to " + user.Id)
					alreadyMapped = true
					break
				}
			}

			if alreadyMapped {
				log.Info().Msg("Already mapped " + user.Id + " to " + metabaseUserExternalId)
				continue
			}

			err = formalSdk.MapUserToExternalId(user.Id, metabaseUserExternalId, integrationID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
