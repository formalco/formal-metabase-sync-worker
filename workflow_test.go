package main

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHasExternalID(t *testing.T) {
	t.Parallel()

	user := User{
		ExternalIds: []ExternalId{
			{ExternalId: "42", AppId: "app_metabase"},
			{ExternalId: "omni-1", AppId: "app_omni"},
		},
	}

	require.True(t, hasExternalID(user, "42", "app_metabase"))
	require.False(t, hasExternalID(user, "42", "app_omni"))
	require.False(t, hasExternalID(user, "99", "app_metabase"))
}

type stubExternalIDClient struct {
	creates []createdExternalID
	err     error
}

type createdExternalID struct {
	userID        string
	externalID    string
	integrationID string
	description   string
}

func (s *stubExternalIDClient) CreateUserExternalId(_ context.Context, userId, externalId, integrationID, description string) error {
	s.creates = append(s.creates, createdExternalID{
		userID:        userId,
		externalID:    externalId,
		integrationID: integrationID,
		description:   description,
	})
	return s.err
}

func TestMapUsersToIntegration(t *testing.T) {
	t.Parallel()

	users := []User{
		{Id: "usr_ada", Email: "ada@example.com"},
		{
			Id:    "usr_grace",
			Email: "grace@example.com",
			ExternalIds: []ExternalId{
				{ExternalId: "7", AppId: "app_metabase"},
			},
		},
		{Id: "usr_unknown", Email: "nobody@example.com"},
	}
	client := &stubExternalIDClient{}

	mapped, skipped, err := mapUsersToIntegration(
		t.Context(),
		client,
		users,
		map[string]string{
			"ada@example.com":   "42",
			"grace@example.com": "7",
		},
		"app_metabase",
		"Metabase",
	)
	require.NoError(t, err)
	require.Equal(t, 1, mapped)
	require.Equal(t, 1, skipped)
	require.Equal(t, []createdExternalID{{
		userID:        "usr_ada",
		externalID:    "42",
		integrationID: "app_metabase",
		description:   "This External ID was imported for this user via Metabase.",
	}}, client.creates)
}

func TestMapUsersToIntegrationCreateError(t *testing.T) {
	t.Parallel()

	client := &stubExternalIDClient{err: errors.New("boom")}
	_, _, err := mapUsersToIntegration(
		t.Context(),
		client,
		[]User{{Id: "usr_ada", Email: "ada@example.com"}},
		map[string]string{"ada@example.com": "42"},
		"app_metabase",
		"Metabase",
	)
	require.ErrorContains(t, err, "failed to map user ada@example.com")
}
