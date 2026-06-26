package repository

import (
	"video_feed/internal/model"

	"github.com/jinzhu/gorm"
)

// CommentRepository 评论数据访问层
type CommentRepository struct{}

// NewCommentRepository 创建评论仓库实例
func NewCommentRepository() *CommentRepository {
	return &CommentRepository{}
}

// Create 创建评论
func (r *CommentRepository) Create(comment *model.Comment) error {
	return DB.Create(comment).Error
}

// FindByID 根据ID查找评论
func (r *CommentRepository) FindByID(id uint) (*model.Comment, error) {
	var comment model.Comment
	// 把每条评论关联的User数据也一并查出来，这样comment.User字段才会被填充完整的用户信息
	err := DB.Preload("User").Where("id = ?", id).First(&comment).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// DeleteByID 根据ID删除评论及其所有子评论(硬删除)
func (r *CommentRepository) DeleteByID(id uint) error {
	// 1. 先查找所有子评论ID（包括已被软删除的子评论 —— 用Unscoped确保找到全部）
	var childIDs []uint
	// Model(&model.Comment{})表示操作comments表，Pluck("id", &childIDs)表示只取 id 列，存入 childIDs 切片
	if err := DB.Unscoped().Model(&model.Comment{}).Where("parent_id = ?", id).Pluck("id", &childIDs).Error; err != nil {
		return err
	}

	// 2. 递归删除每个子评论
	for _, childID := range childIDs {
		if err := r.DeleteByID(childID); err != nil {
			return err
		}
	}

	// 3. 硬删除当前评论
	return DB.Unscoped().Where("id = ?", id).Delete(&model.Comment{}).Error
}

// GetCommentsByVideoID 获取视频的根评论（分页，包含回复）
func (r *CommentRepository) GetCommentsByVideoID(videoID uint, page, pageSize int) ([]model.Comment, int64, error) {
	var comments []model.Comment
	var total int64

	// 统计总数（只统计根评论）
	DB.Model(&model.Comment{}).
		Where("video_id = ? AND parent_id IS NULL", videoID).
		Count(&total)

	// 分页查询根评论，并预加载用户信息和回复
	err := DB.Where("video_id = ? AND parent_id IS NULL", videoID).
		Preload("User").                                // 预加载评论作者
		Preload("Replies", func(db *gorm.DB) *gorm.DB { // 预加载评论的回复，第二个参数是一个自定义函数，用来控制如何加载回复
			return db.Order("created_at ASC").Preload("User") // 回复按时间正序排列，并且每个回复也预加载用户信息
		}).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&comments).Error

	return comments, total, err
}

// GetRepliesByParentID 获取某个评论的所有回复
func (r *CommentRepository) GetRepliesByParentID(parentID uint) ([]model.Comment, error) {
	var replies []model.Comment
	err := DB.Where("parent_id = ?", parentID).
		Preload("User"). // 每个回复预加载用户信息
		Order("created_at ASC").
		Find(&replies).Error
	return replies, err
}
