package model

import (
	"errors"
	"gorm.io/gorm"
	"http_grpc/pkg/database"
	"time"
)

// User 数据库映射模型
type User struct {
	ID           int64     `gorm:"primaryKey;autoIncrement;comment:用户ID" json:"id"`
	Username     string    `gorm:"type:varchar(256);comment:用户昵称" json:"username"`
	UserAccount  string    `gorm:"column:userAccount;type:varchar(256);comment:账号" json:"userAccount"`
	AvatarUrl    string    `gorm:"column:avatarUrl;type:varchar(1024);comment:用户头像" json:"avatarUrl"`
	Gender       int8      `gorm:"type:tinyint;comment:性别" json:"gender"`
	UserPassword string    `gorm:"column:userPassword;type:varchar(512);not null;comment:密码" json:"userPassword"`
	Phone        string    `gorm:"type:varchar(128);comment:电话" json:"phone"`
	Email        string    `gorm:"type:varchar(512);comment:邮箱" json:"email"`
	UserStatus   int       `gorm:"column:userStatus;type:int;default:0;comment:用户状态 0-正常" json:"userStatus"`
	CreateTime   time.Time `gorm:"column:createTime;type:datetime;default:CURRENT_TIMESTAMP;comment:创建时间" json:"createTime"`
	UpdateTime   time.Time `gorm:"column:updateTime;type:datetime;default:CURRENT_TIMESTAMP;on update CURRENT_TIMESTAMP;comment:更新时间" json:"updateTime"`
	IsDelete     int8      `gorm:"column:isDelete;type:tinyint;default:0;comment:是否删除" json:"isDelete"`
	UserRole     int       `gorm:"column:userRole;type:int;not null;comment:用户角色 0-普通用户 1-管理员" json:"userRole"`
	PlanetCode   string    `gorm:"column:planetCode;type:varchar(512);comment:星球编号" json:"planetCode"`
}

func (User) TableName() string {
	return "user"
}

// AddUser 插入新用户
func AddUser(user *User) error {
	return database.DB.Create(user).Error
}

// GetUserByID 根据用户ID查询用户信息
func GetUserByID(id int64, user *User) error {
	return database.DB.First(&user, id).Error
}

// GetUserByAccount 根据账号查询用户信息
func GetUserByAccount(account string, user *User) error {
	result := database.DB.Where("userAccount = ?", account).First(user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil // 没查到
		}
		return result.Error // 非正常情况：数据库坏了
	}
	return nil
}

// FindUserByAccount 查找是否存在
func FindUserByAccount(account string) (bool, error) {
	var id int
	result := database.DB.Model(&User{}).Select("id").Where("userAccount = ?", account).First(&id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil // 正常情况：没查到
		}
		return false, result.Error // 出错
	}
	return true, nil // 找到了
}

// UpdateUserPassword 更新用户密码
func UpdateUserPassword(id int64, newPassword string) error {
	result := database.DB.Model(&User{}).Where("id = ?", id).Update("userPassword", newPassword)
	return result.Error
}

// DeleteUser 软删除用户
func DeleteUser(id int64) error {
	result := database.DB.Model(&User{}).Where("id = ?", id).Update("isDelete", 1)
	return result.Error
}

func UpdateUser(user *User, fields []string) error {
	return database.DB.Model(&User{}).
		Where("id = ?", user.ID).
		Select(fields).
		Updates(user).
		Error
}

// ListUsers 获取用户列表
func ListUsers(page, size int) ([]User, error) {
	var users []User
	result := database.DB.Where("isDelete = 0").Offset((page - 1) * size).Limit(size).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}
