package http

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes 初始化所有路由
func SetupRoutes(router *gin.Engine) {

	// 用户相关路由
	userRoutes := router.Group("/users")
	{
		userRoutes.POST("/register", CreateUser)
		userRoutes.POST("/update", UpdateUser)
		userRoutes.POST("/login", Login)
		userRoutes.GET("/:id", GetUserByID)
		userRoutes.GET("/by-account", GetUserByAccount)
		userRoutes.PUT("/:id/password", UpdateUserPassword)
		userRoutes.GET("/list", ListUsers)
		userRoutes.DELETE("/:id", DeleteUser)
	}

}
