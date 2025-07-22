package middleware

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func GRPCUnaryInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	fields := logrus.Fields{
		"grpc_req": req,
	}
	defer func() {
		fields["grpc_resp"] = resp
		if err != nil {
			fields["grpc_err"] = err.Error()
			logrus.WithContext(ctx).WithFields(fields).Error("grpc_request_out")
		}
	}()

	if md, exist := metadata.FromIncomingContext(ctx); exist {
		fields["grpc_metadata"] = md
	}

	logrus.WithContext(ctx).WithFields(fields).Info("grpc_request_in")
	return handler(ctx, req)
}
