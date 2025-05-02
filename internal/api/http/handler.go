package http

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"http_grpc/internal/repository/session"
	"http_grpc/internal/service"
	"http_grpc/pkg/pool"
	"http_grpc/pkg/utils"
	"net/http"
	"strconv"
)

var (
	userService *service.UserService
)

func InitUserHandler(routinePool *pool.RoutinePool) {
	userService = service.NewUserService(routinePool)
}

// CreateUser 创建用户
func CreateUser(c *gin.Context) {
	taskData := pool.TaskDataPool.Get().(*pool.TaskData)
	defer pool.TaskDataPool.Put(taskData)
	taskData.Reset()

	c.Request.Header.Set("Content-Type", "application/json")
	if err := c.ShouldBind(&taskData.UserData); err != nil {
		utils.Fail(c, utils.BadRequestCode, "Invalid request payload")
		return
	}

	if err := userService.CreateUser(&taskData.UserData); err != nil {
		statusCode := utils.ServerErrorCode
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			statusCode = utils.DuplicateCode
		}
		utils.Fail(c, statusCode, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "User creation request accepted"})
}

// Login 用户登录
func Login(c *gin.Context) {
	var loginReq struct {
		Account  string `json:"userAccount"`
		Password string `json:"userPassword"`
	}
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		utils.Fail(c, utils.BadRequestCode, "Invalid request payload")
		return
	}

	id, account, err := userService.Login(loginReq.Account, loginReq.Password)
	if err != nil {
		statusCode := utils.UnauthorizedCode
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			statusCode = utils.ServerErrorCode
		}
		utils.Fail(c, statusCode, err.Error())
		return
	}

	store := session.GetSession(c)
	store.Values["userID"] = id
	store.Values["userAccount"] = account

	utils.Success(c, gin.H{
		"message":   "Login successful",
		"sessionID": store.ID,
	})
}

// GetUserByID 根据ID获取用户
func GetUserByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, utils.BadRequestCode, "Invalid user ID")
		return
	}

	if ok, _ := userService.CheckUserAuthorization(c, id); !ok {
		return
	}

	taskData := pool.TaskDataPool.Get().(*pool.TaskData)
	defer pool.TaskDataPool.Put(taskData)
	taskData.Reset()
	err = userService.GetUserByID(id, &taskData.UserData)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, utils.NotFoundCode, "User not found")
		} else {
			utils.Fail(c, utils.ServerErrorCode, "Database error")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": taskData.UserData})
}

// GetUserByAccount 根据账号获取用户
func GetUserByAccount(c *gin.Context) {
	account := c.Query("userAccount")
	if account == "" {
		utils.Fail(c, utils.BadRequestCode, "Account is required")
		return
	}

	taskData := pool.TaskDataPool.Get().(*pool.TaskData)
	defer pool.TaskDataPool.Put(taskData)
	taskData.Reset()

	err := userService.GetUserByAccount(account, &taskData.UserData)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, utils.NotFoundCode, "User not found")
		} else {
			utils.Fail(c, utils.ServerErrorCode, "Database error")
		}
		return
	}

	utils.Success(c, gin.H{"data": taskData.UserData})
}

// UpdateUserPassword 更新用户密码
func UpdateUserPassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, utils.BadRequestCode, "Invalid user ID")
		return
	}

	if ok, _ := userService.CheckUserAuthorization(c, id); !ok {
		return
	}

	var req struct {
		NewPassword string `json:"newPassword"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, utils.BadRequestCode, "Invalid request payload")
		return
	}

	userService.UpdatePassword(id, req.NewPassword)

	utils.Success(c, gin.H{"message": "Password updated"})
}

// ListUsers 获取用户列表
func ListUsers(c *gin.Context) {
	if ok, _ := userService.CheckUserAuthorization(c, -1); !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	users, err := userService.ListUsers(page, size)
	if err != nil {
		utils.Fail(c, utils.ServerErrorCode, "Failed to fetch users")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
		"page": page,
		"size": size,
	})
}

// DeleteUser 删除用户
func DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, utils.BadRequestCode, "Invalid user ID")
		return
	}

	if ok, _ := userService.CheckUserAuthorization(c, id); !ok {
		return
	}

	userService.DeleteUser(id)

	utils.Success(c, gin.H{"message": "User deletion request accepted"})
}

// UpdateUser 更新用户信息
func UpdateUser(c *gin.Context) {
	taskData := pool.TaskDataPool.Get().(*pool.TaskData)
	defer pool.TaskDataPool.Put(taskData)
	taskData.Reset()

	c.Request.Header.Set("Content-Type", "application/json")
	if err := c.ShouldBind(&taskData.UserData); err != nil {
		utils.Fail(c, utils.BadRequestCode, "Invalid request payload")
		return
	}

	if ok, _ := userService.CheckUserAuthorization(c, taskData.UserData.ID); !ok {
		utils.Fail(c, utils.UnauthorizedCode, "Invalid user ID")
		return
	}

	userService.UpdateUser(&taskData.UserData)

	utils.Success(c, gin.H{"message": "User update request accepted"})
}
