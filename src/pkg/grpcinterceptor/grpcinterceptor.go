// Package grpcinterceptor provides the logging/recovery/trace-id
// interceptors shared by every service's internal gRPC client and server,
// per docs/01-architecture/system-architecture.md's note that gRPC calls
// use "interceptor cho auth, logging, trace-id propagation".
package grpcinterceptor

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const traceIDKey = "x-trace-id"

// UnaryServerLogging logs method, trace-id, duration and error for every
// incoming internal RPC.
func UnaryServerLogging(serviceName string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		traceID := traceIDFromIncoming(ctx)
		start := time.Now()
		resp, err := handler(ctx, req)
		log.Printf("[%s] grpc method=%s trace_id=%s duration=%s err=%v",
			serviceName, info.FullMethod, traceID, time.Since(start), err)
		return resp, err
	}
}

// UnaryServerRecovery turns a panic inside a handler into codes.Internal
// instead of crashing the process.
func UnaryServerRecovery() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("grpc panic method=%s recovered=%v", info.FullMethod, r)
				err = status.Errorf(codes.Internal, "internal error")
			}
		}()
		return handler(ctx, req)
	}
}

// UnaryClientLogging logs outgoing internal RPC calls made by one service
// against another (e.g. booking-service calling event-service).
func UnaryClientLogging(callerName string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = withTraceID(ctx)
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		log.Printf("[%s] grpc call method=%s duration=%s err=%v", callerName, method, time.Since(start), err)
		return err
	}
}

func traceIDFromIncoming(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get(traceIDKey); len(vals) > 0 {
			return vals[0]
		}
	}
	return "unknown"
}

func withTraceID(ctx context.Context) context.Context {
	if md, ok := metadata.FromOutgoingContext(ctx); ok {
		if len(md.Get(traceIDKey)) > 0 {
			return ctx
		}
	}
	return metadata.AppendToOutgoingContext(ctx, traceIDKey, uuid.NewString())
}
