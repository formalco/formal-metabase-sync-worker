package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

type OmniSCIMResponse struct {
	Resources    []OmniUser `json:"Resources"`
	TotalResults int        `json:"totalResults"`
	StartIndex   int        `json:"startIndex"`
	ItemsPerPage int        `json:"itemsPerPage"`
}

type OmniUser struct {
	Id       string `json:"id"`
	UserName string `json:"userName"`
	Active   bool   `json:"active"`
}

func GetOmniUsers(hostname, apiKey string) (map[string]OmniUser, error) {
	baseURL := "https://" + hostname + "/api/scim/v2/users"
	users := map[string]OmniUser{}
	startIndex := 1
	count := 100

	httpClient := &http.Client{Timeout: 30 * time.Second}

	for {
		url := baseURL + "?startIndex=" + strconv.Itoa(startIndex) + "&count=" + strconv.Itoa(count)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			log.Debug().Str("body", string(body)).Int("status", resp.StatusCode).Msg("Unexpected response from Omni API")
			return nil, errors.New("Omni API returned status " + strconv.Itoa(resp.StatusCode))
		}

		var response OmniSCIMResponse
		err = json.Unmarshal(body, &response)
		if err != nil {
			log.Debug().Str("body", string(body)).Msg("Failed to parse Omni API response")
			return nil, err
		}

		for _, user := range response.Resources {
			users[user.UserName] = user
		}

		if len(response.Resources) == 0 || startIndex+len(response.Resources) > response.TotalResults {
			break
		}
		startIndex += len(response.Resources)
	}

	return users, nil
}
