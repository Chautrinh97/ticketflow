// Package identityclient is the API Gateway's gRPC client to
// identity-service's CheckSession RPC (src/proto/identity.proto) — the
// gateway owns no Postgres/Redis business state itself (per its own
// README), so this is the only way it can enforce the live
// blacklist/banned-account check on every authenticated request.
package identityclient

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"ticketflow/pkg/grpcinterceptor"
	"ticketflow/proto/identitypb"
)

type Client struct {
	conn *grpc.ClientConn
	api  identitypb.IdentityServiceClient
}

func Dial(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(grpcinterceptor.UnaryClientLogging("api-gateway")),
	)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, api: identitypb.NewIdentityServiceClient(conn)}, nil
}

func (c *Client) Close() error { return c.conn.Close() }

func (c *Client) CheckSession(ctx context.Context, jti, userID string) (blacklisted bool, status string, err error) {
	resp, err := c.api.CheckSession(ctx, &identitypb.CheckSessionRequest{Jti: jti, UserId: userID})
	if err != nil {
		return false, "", err
	}
	return resp.GetBlacklisted(), resp.GetUserStatus(), nil
}
