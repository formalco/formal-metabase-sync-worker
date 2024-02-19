package main

import (
	"os"

	"github.com/rs/zerolog/log"
)

func main() {
	metabaseIntegration := MetabaseIntegration{
		MetabaseHostname: os.Getenv("METABASE_HOSTNAME"),
		MetabaseUsername: os.Getenv("METABASE_USERNAME"),
		MetabasePwd:      os.Getenv("METABASE_PASSWORD"),
		Version:          os.Getenv("METABASE_VERSION"),
	}
	formalAPIKey := os.Getenv("FORMAL_API_KEY")
	integrationID := os.Getenv("FORMAL_APP_ID")

	err := MetabaseWorkflow(metabaseIntegration, formalAPIKey, integrationID)
	if err != nil {
		log.Fatal().Err(err).Msg("Error in MetabaseWorkflow")
	}
}
