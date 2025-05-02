package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"http_grpc/internal/repository/model"
	"http_grpc/internal/repository/session"
	"http_grpc/pkg/pool"
	"http_grpc/pkg/utils"
)

type UserService struct {
	routinePool *pool.RoutinePool
}

func NewUserService(routinePool *pool.RoutinePool) *UserService {
	return &UserService{routinePool: routinePool}
}

func (s *UserService) CreateUser(user *model.User) error {
	if user.UserAccount == "" || user.UserPassword == "" {
		return errors.New("account and password are required")
	}

	s.routinePool.AddTask(pool.Task{
		Job: func() error {
			if result, err := model.FindUserByAccount(user.UserAccount); err != nil {
				return err
			} else if result {
				return gorm.ErrDuplicatedKey
			}
			return model.AddUser(user)
		},
	})
	return nil
}

func (s *UserService) Login(account, password string) (int64, string, error) {

	taskData := pool.TaskDataPool.Get().(*pool.TaskData)
	defer pool.TaskDataPool.Put(taskData)
	taskData.Reset()

	err := model.GetUserByAccount(account, &taskData.UserData)
	if err != nil {
		return -1, "", err
	}
	if taskData.UserData.UserPassword != password {
		return -1, "", errors.New("incorrect password")
	}
	return taskData.UserData.ID, taskData.UserData.UserAccount, nil
}

func (s *UserService) GetUserByID(id int64, user *model.User) error {
	return model.GetUserByID(id, user)
}

func (s *UserService) GetUserByAccount(account string, user *model.User) error {
	return model.GetUserByAccount(account, user)
}

func (s *UserService) UpdatePassword(id int64, newPassword string) {
	s.routinePool.AddTask(pool.Task{
		Job: func() error {
			return model.UpdateUserPassword(id, newPassword)
		},
	})
}

func (s *UserService) ListUsers(page, size int) ([]model.User, error) {
	return model.ListUsers(page, size)
}

func (s *UserService) DeleteUser(id int64) {
	s.routinePool.AddTask(pool.Task{
		Job: func() error {
			return model.DeleteUser(id)
		},
	})
}

func (s *UserService) UpdateUser(user *model.User) {
	s.routinePool.AddTask(pool.Task{
		Job: func() error {
			fields := selectNonZeroFields(user)
			return model.UpdateUser(user, fields)
		},
	})
}

func selectNonZeroFields(user *model.User) []string {
	var fields []string
	if user.Username != "" {
		fields = append(fields, "username")
	}
	if user.AvatarUrl != "" {
		fields = append(fields, "avatarUrl")
	}
	if user.Gender != 0 {
		fields = append(fields, "gender")
	}
	if user.Phone != "" {
		fields = append(fields, "phone")
	}
	if user.Email != "" {
		fields = append(fields, "email")
	}
	return fields
}

// CheckUserAuthorization 检查当前用户是否有权限
func (s *UserService) CheckUserAuthorization(c *gin.Context, targetUserID int64) (bool, error) {
	store := session.GetSession(c)

	// 对单个用户进行操作
	if targetUserID != -1 {
		currentUserID, ok := store.Values["userID"].(int64)
		if !ok {
			utils.Fail(c, utils.UnauthorizedCode, "Unauthorized")
			return false, nil
		}
		// 如果是本人，允许
		if currentUserID == targetUserID {
			return true, nil
		}
	}

	// 如果是管理员（userRole == 1），允许对全体用户操作
	userRole, ok := store.Values["userRole"].(int)
	if ok && userRole == 1 {
		return true, nil
	}

	// 否则无权限
	utils.Fail(c, utils.UnauthorizedCode, "Unauthorized")
	return false, nil
}
