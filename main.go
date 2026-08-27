package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	formalAPIKey := os.Getenv("FORMAL_API_KEY")
	verifyTLS, err := strconv.ParseBool(os.Getenv("VERIFY_TLS"))
	if err != nil {
		log.Warn().Msg("Invalid VERIFY_TLS value, defaulting to true")
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
		logLevel = "info"
	}

	frequency := os.Getenv("FREQUENCY")
	var duration time.Duration
	if frequency != "" {
		duration, err = time.ParseDuration(frequency)
		if err != nil {
			log.Fatal().Err(err).Msg("Invalid FREQUENCY format. Expected format: '1h', '30m', etc.")
		}
	}

	cfAccessClientID := os.Getenv("CF_ACCESS_CLIENT_ID")
	cfAccessClientSecret := os.Getenv("CF_ACCESS_CLIENT_SECRET")

	var metabaseEnabled bool
	var metabaseIntegration MetabaseIntegration
	metabaseHostname := os.Getenv("METABASE_HOSTNAME")
	if metabaseHostname != "" {
		metabaseUseApiKey, err := strconv.ParseBool(os.Getenv("METABASE_USE_API_KEY"))
		if err != nil {
			metabaseUseApiKey = false
		}
		metabaseIntegration = MetabaseIntegration{
			UseAPIKey:        metabaseUseApiKey,
			MetabaseAPIKey:   os.Getenv("METABASE_API_KEY"),
			MetabaseHostname: metabaseHostname,
			MetabaseUsername: os.Getenv("METABASE_USERNAME"),
			MetabasePwd:      os.Getenv("METABASE_PASSWORD"),
			Version:          os.Getenv("METABASE_VERSION"),
		}
		metabaseEnabled = true
		log.Info().Str("hostname", metabaseHostname).Msg("Metabase sync enabled")
	}

	var omniEnabled bool
	var omniIntegration OmniIntegration
	omniAPIKey := os.Getenv("OMNI_API_KEY")
	omniHostname := os.Getenv("OMNI_HOSTNAME")
	if omniAPIKey != "" && omniHostname != "" {
		omniIntegration = OmniIntegration{
			APIKey:        omniAPIKey,
			Hostname:      omniHostname,
			IntegrationID: os.Getenv("OMNI_BI_INTEGRATION_ID"),
		}
		omniEnabled = true
		log.Info().Str("hostname", omniHostname).Msg("Omni sync enabled")
	}

	if !metabaseEnabled && !omniEnabled {
		log.Fatal().Msg("No integrations configured. Set METABASE_HOSTNAME for Metabase or OMNI_API_KEY and OMNI_HOSTNAME for Omni.")
	}

	formalClient, err := New(formalAPIKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Formal SDK client")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	for {
		if err := ctx.Err(); err != nil {
			log.Info().Err(err).Msg("Shutting down")
			return
		}

		log.Info().Msg("Fetching users from Formal")
		users, err := formalClient.ListHumanFormalUsers(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Info().Err(err).Msg("Shutting down")
				return
			}
			log.Error().Err(err).Msg("Failed to fetch Formal users")
		} else {
			log.Info().Int("count", len(users)).Msg("Fetched users from Formal")

			if metabaseEnabled {
				log.Info().Msg("Starting Metabase sync")
				metabaseIntegrationID := os.Getenv("METABASE_BI_INTEGRATION_ID")
				err = MetabaseWorkflow(ctx, metabaseIntegration, formalClient, users, metabaseIntegrationID, verifyTLS, cfAccessClientID, cfAccessClientSecret)
				if err != nil {
					log.Error().Err(err).Msg("Metabase sync failed")
				}
			}

			if omniEnabled {
				log.Info().Msg("Starting Omni sync")
				err = OmniWorkflow(ctx, omniIntegration, formalClient, users)
				if err != nil {
					log.Error().Err(err).Msg("Omni sync failed")
				}
			}
		}

		if frequency == "" {
			break
		}
		log.Info().Msgf("Waiting %s before next sync", duration.String())
		timer := time.NewTimer(duration)
		select {
		case <-ctx.Done():
			timer.Stop()
			log.Info().Err(ctx.Err()).Msg("Shutting down")
			return
		case <-timer.C:
		}
	}
}
