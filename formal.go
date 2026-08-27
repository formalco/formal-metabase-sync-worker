package main

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	formal "github.com/formalco/go-sdk/v3"
	corev1 "github.com/formalco/go-sdk/v3/core/v1"
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

type Client struct {
	sdk *formal.Client
}

func New(apiKey string, opts ...formal.Option) (*Client, error) {
	all := make([]formal.Option, 0, len(opts)+1)
	all = append(all, formal.WithAPIKey(apiKey))
	all = append(all, opts...)
	sdk, err := formal.New(all...)
	if err != nil {
		return nil, fmt.Errorf("create Formal SDK client: %w", err)
	}
	return &Client{sdk: sdk}, nil
}

func (c *Client) ListHumanFormalUsers(ctx context.Context) ([]User, error) {
	var cursor string
	var users []User
	for {
		resp, err := c.sdk.UserServiceClient.ListUsers(ctx, &corev1.ListUsersRequest{
			Limit:  100,
			Cursor: cursor,
		})
		if err != nil {
			return nil, err
		}
		humans := lo.Filter(resp.Users, func(user *corev1.User, _ int) bool {
			return user.Type == "human" && user.GetHuman() != nil
		})
		pageUsers, err := lo.MapErr(humans, func(user *corev1.User, _ int) (User, error) {
			externalIds, err := c.listUserExternalIds(ctx, user.Id)
			if err != nil {
				return User{}, err
			}
			return User{
				Id:          user.Id,
				Email:       user.GetHuman().Email,
				ExternalIds: externalIds,
			}, nil
		})
		if err != nil {
			return nil, err
		}
		users = append(users, pageUsers...)
		if resp.ListMetadata == nil || resp.ListMetadata.NextCursor == "" {
			break
		}
		cursor = resp.ListMetadata.NextCursor
	}

	return users, nil
}

func (c *Client) listUserExternalIds(ctx context.Context, userID string) ([]ExternalId, error) {
	var cursor string
	var externalIds []ExternalId
	for {
		resp, err := c.sdk.UserServiceClient.ListUserExternalIds(ctx, &corev1.ListUserExternalIdsRequest{
			Id:     userID,
			Limit:  500,
			Cursor: cursor,
		})
		if err != nil {
			return nil, err
		}
		externalIds = append(externalIds, lo.Map(resp.ExternalIds, func(externalId *corev1.ExternalId, _ int) ExternalId {
			return ExternalId{
				Id:         externalId.Id,
				ExternalId: externalId.ExternalId,
				AppId:      externalId.AppId,
			}
		})...)
		if resp.ListMetadata == nil || resp.ListMetadata.NextCursor == "" {
			break
		}
		cursor = resp.ListMetadata.NextCursor
	}
	return externalIds, nil
}

func (c *Client) CreateUserExternalId(ctx context.Context, userId, externalId, integrationID, description string) error {
	_, err := c.sdk.UserServiceClient.CreateUserExternalId(ctx, &corev1.CreateUserExternalIdRequest{
		UserId:      userId,
		ExternalId:  externalId,
		AppId:       integrationID,
		Description: description,
	})
	return err
}
