package http

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"http_grpc/pkg/config"
	"http_grpc/pkg/pool"
	"log"
)

func StartHttpServer() {
	InitUserHandler(pool.HandlerWorkerPool)

	c := config.AppConfig
	port := c.Http.Port
	// 创建 Gin 引擎，启动服务
	router := gin.Default()
	if err := router.SetTrustedProxies([]string{"127.0.0.1"}); err != nil {
		log.Fatalf("设置代理失败: %v", err)
	}
	SetupRoutes(router)
	if err := router.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
