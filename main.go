package main

import (
	"os"
	"strconv"

	"github.com/rs/zerolog"
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
	verifyTLS, err := strconv.ParseBool(os.Getenv("VERIFY_TLS"))
	if err != nil {
		log.Error().Msg("Error parsing VERIFY_TLS environment variable. Defaulting to TRUE")
		verifyTLS = true
	}

	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case "disabled":
		zerolog.SetGlobalLevel(zerolog.Disabled)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Info().Msg("No log level set, defaulting to info")
		logLevel = "info"
	}

	err = MetabaseWorkflow(metabaseIntegration, formalAPIKey, integrationID, verifyTLS)
	if err != nil {
		log.Fatal().Err(err).Msg("Error in MetabaseWorkflow")
	}
}
