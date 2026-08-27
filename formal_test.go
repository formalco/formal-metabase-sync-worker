package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"

	formal "github.com/formalco/go-sdk/v3"
	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/go-sdk/v3/core/v1/corev1connect"
)

type fakeUserService struct {
	corev1connect.UnimplementedUserServiceHandler
	listUsers        func(context.Context, *connect.Request[corev1.ListUsersRequest]) (*connect.Response[corev1.ListUsersResponse], error)
	listExternalIds  func(context.Context, *connect.Request[corev1.ListUserExternalIdsRequest]) (*connect.Response[corev1.ListUserExternalIdsResponse], error)
	createExternalId func(context.Context, *connect.Request[corev1.CreateUserExternalIdRequest]) (*connect.Response[corev1.CreateUserExternalIdResponse], error)
}

func (s *fakeUserService) ListUsers(ctx context.Context, req *connect.Request[corev1.ListUsersRequest]) (*connect.Response[corev1.ListUsersResponse], error) {
	return s.listUsers(ctx, req)
}

func (s *fakeUserService) ListUserExternalIds(ctx context.Context, req *connect.Request[corev1.ListUserExternalIdsRequest]) (*connect.Response[corev1.ListUserExternalIdsResponse], error) {
	return s.listExternalIds(ctx, req)
}

func (s *fakeUserService) CreateUserExternalId(ctx context.Context, req *connect.Request[corev1.CreateUserExternalIdRequest]) (*connect.Response[corev1.CreateUserExternalIdResponse], error) {
	return s.createExternalId(ctx, req)
}

func startFakeUserServer(t *testing.T, svc *fakeUserService) *Client {
	t.Helper()
	mux := http.NewServeMux()
	path, handler := corev1connect.NewUserServiceHandler(svc)
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client, err := New("test-api-key", formal.WithBaseURL(server.URL))
	require.NoError(t, err)
	return client
}

func TestListHumanFormalUsers(t *testing.T) {
	t.Parallel()

	client := startFakeUserServer(t, &fakeUserService{
		listUsers: func(_ context.Context, req *connect.Request[corev1.ListUsersRequest]) (*connect.Response[corev1.ListUsersResponse], error) {
			switch req.Msg.Cursor {
			case "":
				return connect.NewResponse(&corev1.ListUsersResponse{
					Users: []*corev1.User{
						{
							Id:   "usr_human",
							Type: "human",
							Info: &corev1.User_Human_{
								Human: &corev1.User_Human{Email: "ada@example.com"},
							},
						},
						{
							Id:   "usr_machine",
							Type: "machine",
							Info: &corev1.User_Machine_{
								Machine: &corev1.User_Machine{Name: "bot"},
							},
						},
					},
					ListMetadata: &corev1.ListMetadata{NextCursor: "page-2"},
				}), nil
			case "page-2":
				return connect.NewResponse(&corev1.ListUsersResponse{
					Users: []*corev1.User{
						{
							Id:   "usr_human_2",
							Type: "human",
							Info: &corev1.User_Human_{
								Human: &corev1.User_Human{Email: "grace@example.com"},
							},
						},
					},
					ListMetadata: &corev1.ListMetadata{},
				}), nil
			default:
				return nil, connect.NewError(connect.CodeInvalidArgument, nil)
			}
		},
		listExternalIds: func(_ context.Context, req *connect.Request[corev1.ListUserExternalIdsRequest]) (*connect.Response[corev1.ListUserExternalIdsResponse], error) {
			if req.Msg.Id == "usr_human" {
				return connect.NewResponse(&corev1.ListUserExternalIdsResponse{
					ExternalIds: []*corev1.ExternalId{
						{Id: "ext_1", ExternalId: "42", AppId: "app_metabase"},
					},
					ListMetadata: &corev1.ListMetadata{},
				}), nil
			}
			return connect.NewResponse(&corev1.ListUserExternalIdsResponse{
				ListMetadata: &corev1.ListMetadata{},
			}), nil
		},
	})

	users, err := client.ListHumanFormalUsers(t.Context())
	require.NoError(t, err)
	require.Equal(t, []User{
		{
			Id:    "usr_human",
			Email: "ada@example.com",
			ExternalIds: []ExternalId{
				{Id: "ext_1", ExternalId: "42", AppId: "app_metabase"},
			},
		},
		{
			Id:    "usr_human_2",
			Email: "grace@example.com",
		},
	}, users)
}

func TestCreateUserExternalId(t *testing.T) {
	t.Parallel()

	var got *corev1.CreateUserExternalIdRequest
	client := startFakeUserServer(t, &fakeUserService{
		createExternalId: func(_ context.Context, req *connect.Request[corev1.CreateUserExternalIdRequest]) (*connect.Response[corev1.CreateUserExternalIdResponse], error) {
			got = req.Msg
			return connect.NewResponse(&corev1.CreateUserExternalIdResponse{
				ExternalId: &corev1.ExternalId{Id: "ext_new"},
			}), nil
		},
	})

	err := client.CreateUserExternalId(t.Context(), "usr_human", "42", "app_metabase", "imported via Metabase")
	require.NoError(t, err)
	require.Equal(t, "usr_human", got.UserId)
	require.Equal(t, "42", got.ExternalId)
	require.Equal(t, "app_metabase", got.AppId)
	require.Equal(t, "imported via Metabase", got.Description)
}
