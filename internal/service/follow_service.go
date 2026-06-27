package service

import (
	"errors"
	"video_feed/internal/model"
	"video_feed/internal/repository"
)

// FollowService 关注业务逻辑层
type FollowService struct {
	followRepo *repository.FollowRepository
	userRepo   *repository.UserRepository
}

// NewFollowService 创建关注服务实例
func NewFollowService() *FollowService {
	return &FollowService{
		followRepo: repository.NewFollowRepository(),
		userRepo:   repository.NewUserRepository(),
	}
}

// FollowUser 关注用户
func (s *FollowService) FollowUser(followerID, followingID uint) error {
	// 1. 不能关注自己
	if followerID == followingID {
		return errors.New("不能关注自己")
	}

	// 2. 检查被关注用户是否存在
	_, err := s.userRepo.FindByID(followingID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 3. 检查是否已关注
	isFollowing, _ := s.followRepo.IsFollowing(followerID, followingID)
	if isFollowing {
		return errors.New("已经关注了该用户")
	}

	// 4. 关注用户
	follow := &model.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}
	return s.followRepo.FollowUser(follow)
}

// UnfollowUser 取消关注
func (s *FollowService) UnfollowUser(followerID, followingID uint) error {
	isFollowing, _ := s.followRepo.IsFollowing(followerID, followingID)
	if !isFollowing {
		return errors.New("还没有关注该用户")
	}

	return s.followRepo.UnfollowUser(followerID, followingID)
}

// GetFollowers 获取粉丝列表
func (s *FollowService) GetFollowers(userID uint, page, pageSize int) ([]UserInfo, int64, error) {
	users, total, err := s.followRepo.GetFollowers(userID, page, pageSize)
	if err != nil {
		return nil, 0, errors.New("获取粉丝列表失败")
	}

	var userInfos []UserInfo
	for _, u := range users {
		userInfos = append(userInfos, UserInfo{
			ID:       u.ID,
			Username: u.Username,
			Nickname: u.Nickname,
			Avatar:   u.Avatar,
			Bio:      u.Bio,
		})
	}

	return userInfos, total, nil
}

// GetFollowings 获取关注列表
func (s *FollowService) GetFollowings(userID uint, page, pageSize int) ([]UserInfo, int64, error) {
	users, total, err := s.followRepo.GetFollowings(userID, page, pageSize)
	if err != nil {
		return nil, 0, errors.New("获取关注列表失败")
	}

	var userInfos []UserInfo
	for _, u := range users {
		userInfos = append(userInfos, UserInfo{
			ID:       u.ID,
			Username: u.Username,
			Nickname: u.Nickname,
			Avatar:   u.Avatar,
			Bio:      u.Bio,
		})
	}

	return userInfos, total, nil
}
