package interceptors

import (
	"context"

	tenantCtx "github.com/keshvan/rag-as-service/backend/pkg/common/grpc/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func ClientTenantInterceptor(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	if orgID, ok := tenantCtx.FromContext(ctx); ok {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-organization-id", orgID)
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

func ServerTenantInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		orgIDs := md.Get("x-organization-id")
		if len(orgIDs) > 0 {
			ctx = tenantCtx.ToContext(ctx, orgIDs[0])
		}
	}
	return handler(ctx, req)
}
