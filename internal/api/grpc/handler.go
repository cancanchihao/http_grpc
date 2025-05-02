package grpc

import (
	"context"
	"http_grpc/internal/service"
	"http_grpc/pkg/pool"
	userpb "http_grpc/proto/user"
)

type UserGrpcHandler struct {
	userpb.UnimplementedUserServiceServer
	userService *service.UserService
}

func NewUserGrpcHandler(routinePool *pool.RoutinePool) *UserGrpcHandler {
	return &UserGrpcHandler{
		userService: service.NewUserService(routinePool),
	}
}

func (h *UserGrpcHandler) CreateUser(ctx context.Context, req *userpb.User) (*userpb.CommonResponse, error) {
	taskData := pool.TaskDataPool.Get().(*pool.TaskData)
	defer pool.TaskDataPool.Put(taskData)
	taskData.Reset()
	taskData.UserData.UserAccount = req.UserAccount
	taskData.UserData.UserPassword = req.UserPassword

	err := h.userService.CreateUser(&taskData.UserData)
	if err != nil {
		return nil, err
	}
	return &userpb.CommonResponse{Message: "User creation request accepted"}, nil
}

func (h *UserGrpcHandler) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	id, account, err := h.userService.Login(req.UserAccount, req.UserPassword)
	if err != nil {
		return nil, err
	}
	return &userpb.LoginResponse{
		UserId:      id,
		UserAccount: account,
		Message:     "Login successful",
	}, nil
}

func (h *UserGrpcHandler) GetUserByID(ctx context.Context, req *userpb.IdRequest) (*userpb.User, error) {
	taskData := pool.TaskDataPool.Get().(*pool.TaskData)
	defer pool.TaskDataPool.Put(taskData)
	taskData.Reset()

	err := h.userService.GetUserByID(req.Id, &taskData.UserData)
	if err != nil {
		return nil, err
	}

	return &userpb.User{
		Id:           taskData.UserData.ID,
		UserAccount:  taskData.UserData.UserAccount,
		UserPassword: taskData.UserData.UserPassword,
	}, nil
}

func (h *UserGrpcHandler) GetUserByAccount(ctx context.Context, req *userpb.AccountRequest) (*userpb.User, error) {
	taskData := pool.TaskDataPool.Get().(*pool.TaskData)
	defer pool.TaskDataPool.Put(taskData)
	taskData.Reset()

	err := h.userService.GetUserByAccount(req.UserAccount, &taskData.UserData)
	if err != nil {
		return nil, err
	}

	return &userpb.User{
		Id:           taskData.UserData.ID,
		UserAccount:  taskData.UserData.UserAccount,
		UserPassword: taskData.UserData.UserPassword,
	}, nil
}

func (h *UserGrpcHandler) UpdatePassword(ctx context.Context, req *userpb.UpdatePasswordRequest) (*userpb.CommonResponse, error) {
	h.userService.UpdatePassword(req.Id, req.NewPassword)
	return &userpb.CommonResponse{Message: "Password updated"}, nil
}

func (h *UserGrpcHandler) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	users, err := h.userService.ListUsers(int(req.Page), int(req.Size))
	if err != nil {
		return nil, err
	}

	res := &userpb.ListUsersResponse{
		Page: req.Page,
		Size: req.Size,
	}
	for _, u := range users {
		res.Users = append(res.Users, &userpb.User{
			Id:           u.ID,
			UserAccount:  u.UserAccount,
			UserPassword: u.UserPassword,
		})
	}
	return res, nil
}

func (h *UserGrpcHandler) DeleteUser(ctx context.Context, req *userpb.IdRequest) (*userpb.CommonResponse, error) {
	h.userService.DeleteUser(req.Id)
	return &userpb.CommonResponse{Message: "User deletion request accepted"}, nil
}

func (h *UserGrpcHandler) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.CommonResponse, error) {
	taskData := pool.TaskDataPool.Get().(*pool.TaskData)
	defer pool.TaskDataPool.Put(taskData)
	taskData.Reset()

	// 1. 将gRPC请求转换为模型对象（只复制允许更新的字段）
	taskData.UserData.ID = req.Id
	if req.Username != nil {
		taskData.UserData.Username = req.Username.Value
	}
	if req.AvatarUrl != nil {
		taskData.UserData.AvatarUrl = req.AvatarUrl.Value
	}
	if req.Gender != nil {
		taskData.UserData.Gender = int8(req.Gender.Value)
	}
	if req.Phone != nil {
		taskData.UserData.Phone = req.Phone.Value
	}
	if req.Email != nil {
		taskData.UserData.Email = req.Email.Value
	}

	// 2. 调用现有Service（保持您的协程池逻辑）
	h.userService.UpdateUser(&taskData.UserData)

	// 3. 返回异步接受响应
	return &userpb.CommonResponse{Message: "User update request accepted"}, nil
}
