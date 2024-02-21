package main

import (
	"context"
	"net/http"

	core_connect "buf.build/gen/go/formal/core/connectrpc/go/core/v1/corev1connect"
	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
)

type User struct {
	Id          string
	Email       string
	ExternalIds []ExternalId
}

type ExternalId struct {
	Id         string
	ExternalId string
	AppId      string
}

type transport struct {
	underlyingTransport http.RoundTripper
	apiKey              string
}

type Client struct {
	client core_connect.UserServiceHandler
}

const (
	FORMAL_HOST_URL string = "https://v2api.formalcloud.net"
)

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("X-Api-Key", t.apiKey)
	return t.underlyingTransport.RoundTrip(req)
}

func New(apiKey string) *Client {
	httpClient := &http.Client{Transport: &transport{
		underlyingTransport: http.DefaultTransport,
		apiKey:              apiKey,
	}}
	return &Client{
		client: core_connect.NewUserServiceClient(httpClient, FORMAL_HOST_URL),
	}
}

func (c *Client) ListHumanFormalUsers() ([]User, error) {
	var cursor string

	var users []User
	for {
		resp, err := c.client.ListUsers(context.Background(), connect.NewRequest(&corev1.ListUsersRequest{
			Limit:  100,
			Cursor: cursor,
		},
		))
		if err != nil {
			return nil, err
		}
		for _, user := range resp.Msg.Users {
			if user.Type == "human" {
				var externalIds []ExternalId
				resp, err := c.client.ListUserExternalIds(context.Background(), connect.NewRequest(&corev1.ListUserExternalIdsRequest{
					Id: user.Id,
				},
				))
				if err != nil {
					return nil, err
				}
				for _, externalId := range resp.Msg.ExternalIds {
					externalIds = append(externalIds, ExternalId{
						Id:         externalId.Id,
						ExternalId: externalId.ExternalId,
						AppId:      externalId.AppId,
					})
				}
				users = append(users, User{
					Id:          user.Id,
					Email:       user.GetHuman().Email,
					ExternalIds: externalIds,
				})
			}
		}
		if resp.Msg.ListMetadata.NextCursor == "" {
			break
		}
		cursor = resp.Msg.ListMetadata.NextCursor
	}

	return users, nil
}

func (c *Client) MapUserToExternalId(userId, metabaseUserExternalId, integrationID string) error {
	_, err := c.client.CreateUserExternalId(context.Background(), connect.NewRequest(&corev1.CreateUserExternalIdRequest{
		UserId:      userId,
		ExternalId:  metabaseUserExternalId,
		AppId:       integrationID,
		Description: "This External ID was imported for this role via Metabase.",
	}))
	if err != nil {
		return err
	}

	return nil
}
