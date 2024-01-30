package main

import (
	"context"

	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
	connect_go "github.com/bufbuild/connect-go"
	"github.com/formalco/go-sdk/sdk"
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

type FormalSDK struct {
	client *sdk.FormalSDK
}

func New(apiKey string) *FormalSDK {
	client := sdk.New(apiKey)
	return &FormalSDK{
		client: client,
	}
}

func (sdk *FormalSDK) ListHumanFormalUsers() ([]User, error) {
	var cussor string

	var users []User
	for {
		resp, err := sdk.client.UserServiceClient.ListUsers(context.Background(), connect_go.NewRequest(&adminv1.ListUsersRequest{
			Limit:  100,
			Cursor: cussor,
		},
		))
		if err != nil {
			return nil, err
		}
		for _, user := range resp.Msg.Users {
			if user.Type == "human" {
				var externalIds []ExternalId
				for _, externalId := range user.ExternalIds {
					externalIds = append(externalIds, ExternalId{
						Id:         externalId.Id,
						ExternalId: externalId.ExternalId,
						AppId:      externalId.AppId,
					})
				}
				users = append(users, User{
					Id:          user.Id,
					Email:       user.Email,
					ExternalIds: externalIds,
				})
			}
		}
		if resp.Msg.ListMetadata.NextCursor == "" {
			break
		}
		cussor = resp.Msg.ListMetadata.NextCursor
	}

	return users, nil
}

func (sdk *FormalSDK) MapUserToExternalId(userId, metabaseUserExternalId, integrationID string) error {
	_, err := sdk.client.UserServiceClient.MapUserToExternalId(context.Background(), connect_go.NewRequest(&adminv1.MapUserToExternalIdRequest{
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
