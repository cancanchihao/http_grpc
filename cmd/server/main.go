package main

import (
	"http_grpc/internal/api/grpc"
	"log"
	"time"

	"http_grpc/internal/api/http"
	"http_grpc/internal/repository/model"
	"http_grpc/internal/repository/session"
	"http_grpc/pkg/config"
	"http_grpc/pkg/database"
	"http_grpc/pkg/pool"
)

func initRedisMysql() {
	c := config.AppConfig
	// 初始化 mysql 数据库
	err := database.InitDB(c.Mysql.DSN)
	if err != nil {
		panic("failed to connect database")
	}
	// 自动迁移表结构
	err = database.DB.AutoMigrate(&model.User{})
	if err != nil {
		panic("failed to migrating tables")
	}

	// 初始化 redis, 启动 Session GC
	err = session.InitRedis(c.Redis.Addr, c.Redis.Password, c.Redis.DB)
	if err != nil {
		log.Fatalf("Redis连接失败: %v", err)
	}
	err = session.LoadSessionsFromRedis()
	if err != nil {
		log.Fatalf("加载Session失败: %v", err)
	}
	session.StartGC(60 * 24 * time.Minute)
}

func init() {
	// 初始化 redis, mysql; 启动 Session GC
	initRedisMysql()
}

func main() {
	defer func() {
		pool.HandlerWorkerPool.Shutdown()
		pool.SessionPool.Shutdown()
	}()

	// 启动 http服务
	go http.StartHttpServer()

	// 启动 grpc服务
	grpc.StartGrpcServer()
}
