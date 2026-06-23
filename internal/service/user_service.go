package service

//封装用户的业务逻辑
import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
	"video_feed/internal/middleware"
	"video_feed/internal/model"
	"video_feed/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService() *UserService {
	return &UserService{
		userRepo: repository.NewUserRepository(),
	}
}

// GetUserByID 根据ID查找用户
func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	return s.userRepo.FindByID(id)
}

// RegisterReq 注册请求
type RegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=20"` //binding标签中的内容表示这个字段不空且最小长度为3最大长度为20
	Password string `json:"password" binding:"required,min=6,max=20"`
	Nickname string `json:"nickname" binding:"max=20"`
}

// Register 用户注册
func (s *UserService) Register(req *RegisterReq) (*model.User, error) {
	// 1. 检查用户名是否已存在
	existing, _ := s.userRepo.FindByUsername(req.Username)
	if existing != nil {
		return nil, errors.New("用户名已存在")
	}

	// 2. 密码加密（使用bcrypt，成本为10）
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	// 3. 创建用户对象
	user := &model.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Nickname: req.Nickname,
	}

	// 4. 保存到数据库
	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("创建用户失败")
	}

	return user, nil
}

// LoginReq 登录请求
type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResp 登录响应
type LoginResp struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

// Login 用户登录
func (s *UserService) Login(req *LoginReq) (*LoginResp, error) {
	// 1. 查找用户
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 2. 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 3. 生成JWT Token
	token, err := middleware.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, errors.New("生成Token失败")
	}

	return &LoginResp{
		Token: token,
		User:  user,
	}, nil
}

// UpdateProfileReq 更新个人信息请求
type UpdateProfileReq struct {
	UserID   uint   `json:"user_id"`
	Nickname string `json:"nickname"`
	Bio      string `json:"bio"`
	Gender   int    `json:"gender"`
	Birthday string `json:"birthday"`
}

// UserInfo 用户个人信息
type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Bio      string `json:"bio"`
	Gender   int    `json:"gender"`
	Birthday string `json:"birthday"`
}

// UpdateProfile 更新用户个人信息
func (s *UserService) UpdateProfile(req *UpdateProfileReq) (*UserInfo, error) {
	// 1. 获取当前用户信息
	user, err := s.userRepo.FindByID(req.UserID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 2. 更新字段（只更新非空字段）
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}
	if req.Gender >= 0 && req.Gender <= 2 {
		user.Gender = req.Gender
	}
	if req.Birthday != "" {
		user.Birthday = req.Birthday
	}

	// 3. 保存到数据库
	if err := s.userRepo.Update(user); err != nil {
		return nil, errors.New("更新个人信息失败")
	}

	return &UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Bio:      user.Bio,
		Gender:   user.Gender,
		Birthday: user.Birthday,
	}, nil
}

// UpdateAvatar 更新用户头像
func (s *UserService) UpdateAvatar(userID uint, file *multipart.FileHeader) (string, error) {
	// 1. 创建头像存储目录
	avatarDir := "./uploads/avatars"
	if err := os.MkdirAll(avatarDir, 0755); err != nil {
		return "", errors.New("创建头像目录失败")
	}

	// 2. 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(avatarDir, fileName)

	// 3. 保存文件
	src, err := file.Open()
	if err != nil {
		return "", errors.New("打开上传文件失败")
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return "", errors.New("创建目标文件失败")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", errors.New("保存文件失败")
	}

	// 4. 更新数据库
	avatarURL := "/uploads/avatars/" + fileName
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return "", errors.New("用户不存在")
	}

	user.Avatar = avatarURL
	if err := s.userRepo.Update(user); err != nil {
		return "", errors.New("更新头像失败")
	}

	return avatarURL, nil
}
