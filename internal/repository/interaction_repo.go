package repository

import (
	"video_feed/internal/model"
)

// InteractionRepository 互动数据访问层
type InteractionRepository struct{}

// NewInteractionRepository 创建互动仓库实例
func NewInteractionRepository() *InteractionRepository {
	return &InteractionRepository{}
}

// ========== 点赞相关 ==========

// CreateLike 添加点赞
func (r *InteractionRepository) CreateLike(like *model.Like) error {
	return DB.Create(like).Error
}

// DeleteLike 取消点赞（硬删除）
func (r *InteractionRepository) DeleteLike(userID, videoID uint) error {
	return DB.Unscoped().Where("user_id = ? AND video_id = ?", userID, videoID).
		Delete(&model.Like{}).Error
}

// IsLiked 检查是否已点赞
func (r *InteractionRepository) IsLiked(userID, videoID uint) (bool, error) {
	var count int64
	err := DB.Model(&model.Like{}).
		Where("user_id = ? AND video_id = ?", userID, videoID).
		Count(&count).Error
	return count > 0, err
}

// GetLikeCount 获取视频点赞数
func (r *InteractionRepository) GetLikeCount(videoID uint) (int64, error) {
	var count int64
	err := DB.Model(&model.Like{}).
		Where("video_id = ?", videoID).
		Count(&count).Error
	return count, err
}

// ========== 收藏相关 ==========

// CreateFavorite 添加收藏
func (r *InteractionRepository) CreateFavorite(fav *model.Favorite) error {
	return DB.Create(fav).Error
}

// DeleteFavorite 取消收藏（硬删除）
func (r *InteractionRepository) DeleteFavorite(userID, videoID uint) error {
	return DB.Unscoped().Where("user_id = ? AND video_id = ?", userID, videoID).
		Delete(&model.Favorite{}).Error
}

// IsFavorited 检查是否已收藏
func (r *InteractionRepository) IsFavorited(userID, videoID uint) (bool, error) {
	var count int64
	err := DB.Model(&model.Favorite{}).
		Where("user_id = ? AND video_id = ?", userID, videoID).
		Count(&count).Error
	return count > 0, err
}

// GetFavoriteCount 获取视频收藏数
func (r *InteractionRepository) GetFavoriteCount(videoID uint) (int64, error) {
	var count int64
	err := DB.Model(&model.Favorite{}).
		Where("video_id = ?", videoID).
		Count(&count).Error
	return count, err
}

// GetUserFavorites 获取用户的收藏列表
func (r *InteractionRepository) GetUserFavorites(userID uint) ([]model.Video, error) {
	var videos []model.Video
	err := DB.Joins("JOIN favorites ON favorites.video_id = videos.id").
		Where("favorites.user_id = ?", userID).
		Order("favorites.created_at DESC").
		Find(&videos).Error
	return videos, err
}
