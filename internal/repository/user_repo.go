package repository

//封装数据库中的用户操作
import (
	"video_feed/internal/model"
)

// UserRepository 定义一个空结构体用来挂载方法
type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// Create 创建用户
func (r *UserRepository) Create(user *model.User) error {
	return DB.Create(user).Error
}

// FindByUsername 根据用户名查找用户
func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID 根据ID查找用户
func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := DB.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
