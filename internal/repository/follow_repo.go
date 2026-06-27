package repository

import (
	"video_feed/internal/model"
)

// FollowRepository 关注数据访问层
type FollowRepository struct{}

// NewFollowRepository 创建关注仓库实例
func NewFollowRepository() *FollowRepository {
	return &FollowRepository{}
}

// FollowUser 关注用户
func (r *FollowRepository) FollowUser(follow *model.Follow) error {
	return DB.Create(follow).Error
}

// UnfollowUser 取消关注（硬删除）
func (r *FollowRepository) UnfollowUser(followerID, followingID uint) error {
	return DB.Unscoped().Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Delete(&model.Follow{}).Error
}

// IsFollowing 检查是否已关注
func (r *FollowRepository) IsFollowing(followerID, followingID uint) (bool, error) {
	var count int64
	err := DB.Model(&model.Follow{}).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Count(&count).Error
	return count > 0, err
}

// GetFollowers 获取用户的粉丝列表（分页）
func (r *FollowRepository) GetFollowers(userID uint, page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := DB.Model(&model.User{}).
		Joins("JOIN follows ON follows.follower_id = users.id").
		Where("follows.following_id = ?", userID)

	query.Count(&total)
	err := query.
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&users).Error

	return users, total, err
}

// GetFollowings 获取用户的关注列表（分页）
func (r *FollowRepository) GetFollowings(userID uint, page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := DB.Model(&model.User{}).
		Joins("JOIN follows ON follows.following_id = users.id").
		Where("follows.follower_id = ?", userID)

	query.Count(&total)
	err := query.
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&users).Error

	return users, total, err
}
