package grpcinterceptor

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestUnaryServerLogging(t *testing.T) {
	interceptor := UnaryServerLogging("test-service")
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	t.Run("[Success] passthrough response unchanged", func(t *testing.T) {
		wantResp := "canned-response"
		resp, err := interceptor(context.Background(), "req", info, func(ctx context.Context, req interface{}) (interface{}, error) {
			return wantResp, nil
		})
		require.NoError(t, err)
		assert.Equal(t, wantResp, resp)
	})

	t.Run("[Error] passthrough handler error unchanged", func(t *testing.T) {
		wantErr := errors.New("handler failed")
		_, err := interceptor(context.Background(), "req", info, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, wantErr
		})
		assert.ErrorIs(t, err, wantErr)
	})
}

func TestUnaryServerRecovery(t *testing.T) {
	interceptor := UnaryServerRecovery()
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	t.Run("[Success] normal handler passes through unchanged", func(t *testing.T) {
		resp, err := interceptor(context.Background(), "req", info, func(ctx context.Context, req interface{}) (interface{}, error) {
			return "ok", nil
		})
		require.NoError(t, err)
		assert.Equal(t, "ok", resp)
	})

	t.Run("[Error] panic is recovered into codes.Internal instead of crashing", func(t *testing.T) {
		resp, err := interceptor(context.Background(), "req", info, func(ctx context.Context, req interface{}) (interface{}, error) {
			panic("boom")
		})
		assert.Nil(t, resp)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
	})
}

func TestUnaryClientLogging(t *testing.T) {
	interceptor := UnaryClientLogging("test-caller")

	t.Run("[Success] invokes with same args and injects a trace-id when absent", func(t *testing.T) {
		var capturedCtx context.Context
		invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			capturedCtx = ctx
			return nil
		}

		err := interceptor(context.Background(), "/test.Service/Method", "req", "reply", nil, invoker)
		require.NoError(t, err)

		md, ok := metadata.FromOutgoingContext(capturedCtx)
		require.True(t, ok)
		require.Len(t, md.Get(traceIDKey), 1)
		assert.NotEmpty(t, md.Get(traceIDKey)[0])
	})

	t.Run("[Success] keeps an existing trace-id instead of generating a new one", func(t *testing.T) {
		existing := metadata.Pairs(traceIDKey, "trace-abc-123")
		ctx := metadata.NewOutgoingContext(context.Background(), existing)

		var capturedCtx context.Context
		invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			capturedCtx = ctx
			return nil
		}

		err := interceptor(ctx, "/test.Service/Method", "req", "reply", nil, invoker)
		require.NoError(t, err)

		md, ok := metadata.FromOutgoingContext(capturedCtx)
		require.True(t, ok)
		assert.Equal(t, []string{"trace-abc-123"}, md.Get(traceIDKey))
	})

	t.Run("[Error] propagates invoker error", func(t *testing.T) {
		wantErr := errors.New("dial failed")
		invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return wantErr
		}
		err := interceptor(context.Background(), "/test.Service/Method", "req", "reply", nil, invoker)
		assert.ErrorIs(t, err, wantErr)
	})
}
