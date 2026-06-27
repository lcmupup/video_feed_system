package repository

//封装数据库中的视频操作
import (
	"video_feed/internal/model"
)

// VideoRepository 视频数据访问层，定义一个空结构体用来挂载方法
type VideoRepository struct{}

// NewVideoRepository 创建视频仓库实例
func NewVideoRepository() *VideoRepository {
	return &VideoRepository{}
}

// Create 保存视频信息
func (r *VideoRepository) Create(video *model.Video) error {
	return DB.Create(video).Error
}

// FindByID 根据ID查找视频
func (r *VideoRepository) FindByID(id uint) (*model.Video, error) {
	var video model.Video
	err := DB.Where("id = ?", id).First(&video).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

// FindByUserID 查找用户的所有视频
func (r *VideoRepository) FindByUserID(userID uint) ([]model.Video, error) {
	var videos []model.Video
	err := DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&videos).Error
	return videos, err
}

// GetFeed 获取Feed流（分页），参数表明要查找第几页以及每页的大小
func (r *VideoRepository) GetFeed(page, pageSize int) ([]model.Video, int64, error) {
	var videos []model.Video
	var total int64

	// 计算总数
	DB.Model(&model.Video{}).Where("status = ?", 1).Count(&total)

	// 分页查询
	err := DB.Where("status = ?", 1).
		Order("created_at DESC").      // 按创建时间降序排序
		Offset((page - 1) * pageSize). // 偏移量，表示从第几条记录开始
		Limit(pageSize).               // 限制查找的记录条数
		Find(&videos).Error

	return videos, total, err
}

// SearchVideos 搜索视频（按视频标题或描述模糊匹配）
func (r *VideoRepository) SearchVideos(keyword string, page, pageSize int) ([]model.Video, int64, error) {
	var videos []model.Video
	var total int64 // 搜索结果条数

	// 构建查询条件
	query := DB.Model(&model.Video{}).
		Where("status = ?", 1). // 只搜索已发布的视频
		Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")

	// 统计总数
	query.Count(&total)

	// 分页查询
	err := query.
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&videos).Error

	return videos, total, err
}
