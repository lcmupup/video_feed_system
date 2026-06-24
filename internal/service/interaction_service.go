package service

import (
	"errors"
	"video_feed/internal/model"
	"video_feed/internal/repository"
)

// InteractionService 互动业务逻辑层
type InteractionService struct {
	interRepo *repository.InteractionRepository
	videoRepo *repository.VideoRepository
}

// NewInteractionService 创建互动服务实例
func NewInteractionService() *InteractionService {
	return &InteractionService{
		interRepo: repository.NewInteractionRepository(),
		videoRepo: repository.NewVideoRepository(),
	}
}

// ========== 点赞相关 ==========

// LikeVideo 点赞视频
func (s *InteractionService) LikeVideo(userID, videoID uint) error {
	// 1. 检查视频是否存在
	_, err := s.videoRepo.FindByID(videoID)
	if err != nil {
		return errors.New("视频不存在")
	}

	// 2. 检查是否已点赞
	isLiked, err := s.interRepo.IsLiked(userID, videoID)
	if err != nil {
		return errors.New("检查点赞状态失败")
	}
	if isLiked {
		return errors.New("已经点过赞了")
	}

	// 3. 添加点赞记录
	like := &model.Like{
		UserID:  userID,
		VideoID: videoID,
	}
	return s.interRepo.CreateLike(like)
}

// UnlikeVideo 取消点赞
func (s *InteractionService) UnlikeVideo(userID, videoID uint) error {
	// 1. 检查视频是否存在
	_, err := s.videoRepo.FindByID(videoID)
	if err != nil {
		return errors.New("视频不存在")
	}

	// 2. 检查是否已点赞
	isLiked, err := s.interRepo.IsLiked(userID, videoID)
	if err != nil {
		return errors.New("检查点赞状态失败")
	}
	if !isLiked {
		return errors.New("还没有点赞")
	}

	// 3. 取消点赞
	return s.interRepo.DeleteLike(userID, videoID)
}

// VideoLikeStatus 视频点赞状态
type VideoLikeStatus struct {
	IsLiked   bool  `json:"is_liked"`   // 是否已点赞
	LikeCount int64 `json:"like_count"` // 点赞数
}

// GetVideoLikeStatus 获取视频的点赞状态
func (s *InteractionService) GetVideoLikeStatus(userID, videoID uint) (*VideoLikeStatus, error) {
	isLiked, err := s.interRepo.IsLiked(userID, videoID)
	if err != nil {
		return nil, errors.New("获取点赞状态失败")
	}

	likeCount, err := s.interRepo.GetLikeCount(videoID)
	if err != nil {
		return nil, errors.New("获取点赞数失败")
	}

	return &VideoLikeStatus{
		IsLiked:   isLiked,
		LikeCount: likeCount,
	}, nil
}

// ========== 收藏相关 ==========

// FavoriteVideo 收藏视频
func (s *InteractionService) FavoriteVideo(userID, videoID uint) error {
	// 1. 检查视频是否存在
	_, err := s.videoRepo.FindByID(videoID)
	if err != nil {
		return errors.New("视频不存在")
	}

	// 2. 检查是否已收藏
	isFav, err := s.interRepo.IsFavorited(userID, videoID)
	if err != nil {
		return errors.New("检查收藏状态失败")
	}
	if isFav {
		return errors.New("已经收藏过了")
	}

	// 3. 添加收藏记录
	fav := &model.Favorite{
		UserID:  userID,
		VideoID: videoID,
	}
	return s.interRepo.CreateFavorite(fav)
}

// UnfavoriteVideo 取消收藏
func (s *InteractionService) UnfavoriteVideo(userID, videoID uint) error {
	// 1. 检查视频是否存在
	_, err := s.videoRepo.FindByID(videoID)
	if err != nil {
		return errors.New("视频不存在")
	}

	// 2. 检查是否已收藏
	isFav, err := s.interRepo.IsFavorited(userID, videoID)
	if err != nil {
		return errors.New("检查收藏状态失败")
	}
	if !isFav {
		return errors.New("还没有收藏")
	}

	// 3. 取消收藏
	return s.interRepo.DeleteFavorite(userID, videoID)
}

// VideoFavoriteStatus 视频收藏状态
type VideoFavoriteStatus struct {
	IsFavorited   bool  `json:"is_favorited"`   // 是否已收藏
	FavoriteCount int64 `json:"favorite_count"` // 收藏数
}

// GetVideoFavoriteStatus 获取视频的收藏状态
func (s *InteractionService) GetVideoFavoriteStatus(userID, videoID uint) (*VideoFavoriteStatus, error) {
	isFav, err := s.interRepo.IsFavorited(userID, videoID)
	if err != nil {
		return nil, errors.New("获取收藏状态失败")
	}

	favCount, err := s.interRepo.GetFavoriteCount(videoID)
	if err != nil {
		return nil, errors.New("获取收藏数失败")
	}

	return &VideoFavoriteStatus{
		IsFavorited:   isFav,
		FavoriteCount: favCount,
	}, nil
}

// GetUserFavorites 获取用户的收藏列表
func (s *InteractionService) GetUserFavorites(userID uint) ([]VideoInfo, error) {
	favVideos, err := s.interRepo.GetUserFavorites(userID)
	if err != nil {
		return nil, errors.New("获取收藏列表失败")
	}

	var videoInfos []VideoInfo
	for _, v := range favVideos {
		videoInfos = append(videoInfos, VideoInfo{
			ID:          v.ID,
			Title:       v.Title,
			Description: v.Description,
			FilePath:    v.FilePath,
			FileSize:    v.FileSize,
			CreatedAt:   v.CreatedAt.Format("2006-01-02 15:04:05"),
			Status:      v.Status,
		})
	}

	return videoInfos, nil
}
