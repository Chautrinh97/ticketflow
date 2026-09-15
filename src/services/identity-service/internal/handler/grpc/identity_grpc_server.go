// Package grpc implements identitypb.IdentityServiceServer — the internal
// RPC surface described in src/proto/identity.proto, consumed by the API
// Gateway's auth middleware.
package grpc

import (
	"context"

	"ticketflow/proto/identitypb"
	"ticketflow/services/identity-service/internal/service"
)

type IdentityGRPCServer struct {
	identitypb.UnimplementedIdentityServiceServer
	auth *service.AuthService
}

func NewIdentityGRPCServer(auth *service.AuthService) *IdentityGRPCServer {
	return &IdentityGRPCServer{auth: auth}
}

func (s *IdentityGRPCServer) CheckSession(ctx context.Context, req *identitypb.CheckSessionRequest) (*identitypb.CheckSessionResponse, error) {
	blacklisted, status, err := s.auth.CheckSession(ctx, req.GetJti(), req.GetUserId())
	if err != nil {
		return nil, err
	}
	return &identitypb.CheckSessionResponse{Blacklisted: blacklisted, UserStatus: status}, nil
}
