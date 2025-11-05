package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type MetabaseSessionResponse struct {
	ID string `json:"id"`
}
type MetabaseUsersResponse struct {
	Data  []MetabaseUser `json:"data"`
	Total int            `json:"total"`
	Limit int            `json:"limit"`
}

type MetabaseUser struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

const METABASE_THRESHOLD_VERSION = "0.40.0"

func GetMetabaseRoles(hostname string, metabaseVersion string, apiKey string, sessionKey string, useAPIKey bool, verifyTLS bool, cfAccessClientID, cfAccessClientSecret string) (map[string]MetabaseUser, error) {
	baseUrl := "https://" + hostname + "/api/user"

	roles := map[string]MetabaseUser{}
	// In newer version of Metabase, the User API is paginated so the returned data is different, hence the difference of logic based on the version
	if metabaseVersion >= METABASE_THRESHOLD_VERSION {
		total := 1
		for len(roles) < total {
			url := baseUrl + "?offset=" + strconv.Itoa(len(roles))
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				return nil, err
			}
			req.Header.Set("Content-Type", "application/json")
			if useAPIKey {
				req.Header.Set("X-API-Key", apiKey)
			} else {
				req.Header.Set("X-Metabase-Session", sessionKey)
			}
			// Add Cloudflare Access headers if provided
			if cfAccessClientID != "" {
				req.Header.Set("CF-Access-Client-Id", cfAccessClientID)
			}
			if cfAccessClientSecret != "" {
				req.Header.Set("CF-Access-Client-Secret", cfAccessClientSecret)
			}

			// Send Request
			client := &http.Client{
				Timeout: 30 * time.Second,
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: !verifyTLS},
				},
			}
			resp, err := client.Do(req)
			if err != nil {
				return nil, err
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Error().Err(err).Msg("LOC getMetabaseRoles cannot read body")
				return nil, err
			}
			defer resp.Body.Close()

			var users []MetabaseUser
			var response MetabaseUsersResponse
			err = json.Unmarshal(body, &response)
			if err != nil {
				log.Debug().Msgf("Received body: %s", body)
				log.Error().Err(err).Msg("Error in getMetabaseRoles - cannot unmarshal body")
				err = json.Unmarshal(body, &users)
				if err != nil {
					log.Error().Err(err).Msg("Error in getMetabaseRoles - cannot unmarshal body")
					return nil, err
				}
			} else {
				users = response.Data
			}

			for _, user := range users {
				roles[user.Email] = user
			}

			total = response.Total
		}
	} else {
		req, err := http.NewRequest("GET", baseUrl, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		if useAPIKey {
			req.Header.Set("X-API-Key", apiKey)
		} else {
			req.Header.Set("X-Metabase-Session", sessionKey)
		}
		// Add Cloudflare Access headers if provided
		if cfAccessClientID != "" {
			req.Header.Set("CF-Access-Client-Id", cfAccessClientID)
		}
		if cfAccessClientSecret != "" {
			req.Header.Set("CF-Access-Client-Secret", cfAccessClientSecret)
		}

		// Send Request
		client := &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: !verifyTLS},
			},
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error().Err(err).Msg("LOC getMetabaseRoles cannot read body")
			return nil, err
		}
		defer resp.Body.Close()

		var users []MetabaseUser
		err = json.Unmarshal(body, &users)
		if err != nil {
			log.Error().Err(err).Msg("Error in getMetabaseRoles - cannot unmarshal body")
			err = json.Unmarshal(body, &users)
			if err != nil {
				log.Error().Err(err).Msg("Error in getMetabaseRoles - cannot unmarshal body")
				return nil, err
			}
		}

		for _, user := range users {
			roles[user.Email] = user
		}
	}
	return roles, nil
}

func RefreshMetabaseSessionKey(integration MetabaseIntegration, verifyTLS bool, cfAccessClientID, cfAccessClientSecret string) (string, error) {
	url := "https://" + integration.MetabaseHostname + "/api/session"

	payload := map[string]string{
		"username": integration.MetabaseUsername,
		"password": integration.MetabasePwd,
	}
	e, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	reader := strings.NewReader(string(e))
	req, err := http.NewRequest("POST", url, reader)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	// Add Cloudflare Access headers if provided
	if cfAccessClientID != "" {
		req.Header.Set("CF-Access-Client-Id", cfAccessClientID)
	}
	if cfAccessClientSecret != "" {
		req.Header.Set("CF-Access-Client-Secret", cfAccessClientSecret)
	}

	// Send Request
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: !verifyTLS},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Error().Msg("LOC refreshMetabaseSessionKey.StatusCode: " + string(body))
		return "", errors.New(string(body))
	}

	var response MetabaseSessionResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&response)
	if err != nil {
		log.Error().Err(err).Msg("LOC MetabaseSessionResponse")
		return "", err
	}

	return response.ID, nil
}
