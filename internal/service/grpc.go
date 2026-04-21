package service

import (
	"context"

	"autocheck-microservices/internal/contracts"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const serviceName = "autocheck.GatewayService"

type GatewayServer interface {
	SubmitRun(context.Context, *contracts.SubmitRunRequest) (*contracts.SubmitRunResponse, error)
	GetRun(context.Context, *contracts.GetRunRequest) (*contracts.GetRunResponse, error)
}

func RegisterGatewayServer(s grpc.ServiceRegistrar, srv GatewayServer) {
	s.RegisterService(&grpc.ServiceDesc{
		ServiceName: serviceName,
		HandlerType: (*GatewayServer)(nil),
		Methods: []grpc.MethodDesc{
			{MethodName: "SubmitRun", Handler: submitRunHandler(srv)},
			{MethodName: "GetRun", Handler: getRunHandler(srv)},
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: "manual",
	}, srv)
}

func submitRunHandler(srv GatewayServer) grpc.MethodHandler {
	return func(_ any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
		in := new(contracts.SubmitRunRequest)
		if err := dec(in); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		handler := func(ctx context.Context, req any) (any, error) {
			return srv.SubmitRun(ctx, req.(*contracts.SubmitRunRequest))
		}
		if interceptor == nil {
			return handler(ctx, in)
		}
		info := &grpc.UnaryServerInfo{FullMethod: "/" + serviceName + "/SubmitRun"}
		return interceptor(ctx, in, info, handler)
	}
}

func getRunHandler(srv GatewayServer) grpc.MethodHandler {
	return func(_ any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
		in := new(contracts.GetRunRequest)
		if err := dec(in); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		handler := func(ctx context.Context, req any) (any, error) {
			return srv.GetRun(ctx, req.(*contracts.GetRunRequest))
		}
		if interceptor == nil {
			return handler(ctx, in)
		}
		info := &grpc.UnaryServerInfo{FullMethod: "/" + serviceName + "/GetRun"}
		return interceptor(ctx, in, info, handler)
	}
}

type GatewayClient struct {
	cc *grpc.ClientConn
}

func NewGatewayClient(cc *grpc.ClientConn) *GatewayClient {
	return &GatewayClient{cc: cc}
}

func (c *GatewayClient) SubmitRun(ctx context.Context, in *contracts.SubmitRunRequest) (*contracts.SubmitRunResponse, error) {
	out := new(contracts.SubmitRunResponse)
	if err := c.cc.Invoke(ctx, "/"+serviceName+"/SubmitRun", in, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *GatewayClient) GetRun(ctx context.Context, in *contracts.GetRunRequest) (*contracts.GetRunResponse, error) {
	out := new(contracts.GetRunResponse)
	if err := c.cc.Invoke(ctx, "/"+serviceName+"/GetRun", in, out); err != nil {
		return nil, err
	}
	return out, nil
}
