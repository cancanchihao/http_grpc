package grpc

import (
	"fmt"
	"google.golang.org/grpc"
	"http_grpc/pkg/config"
	"http_grpc/pkg/pool"
	userpb "http_grpc/proto/user"
	"log"
	"net"
)

func StartGrpcServer() {
	c := config.AppConfig
	port := c.Grpc.Port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// 创建新的 gRPC 服务器实例
	grpcServer := grpc.NewServer()

	// 初始化你的 gRPC handler
	handler := NewUserGrpcHandler(pool.HandlerWorkerPool) // 请根据你的需求传入合适的参数

	// 注册 UserService 服务
	userpb.RegisterUserServiceServer(grpcServer, handler)

	// 启动 gRPC 服务器
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
