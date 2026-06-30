package service

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"video_feed/internal/model"
	"video_feed/internal/repository"

	"github.com/redis/go-redis/v9"
)

// RankingService 排行榜业务逻辑层
type RankingService struct {
	videoRepo *repository.VideoRepository
	userRepo  *repository.UserRepository
}

// NewRankingService 创建排行榜服务实例
func NewRankingService() *RankingService {
	return &RankingService{
		videoRepo: repository.NewVideoRepository(),
		userRepo:  repository.NewUserRepository(),
	}
}

// RankingItem 排行榜单项
type RankingItem struct {
	Rank      int      `json:"rank"`
	VideoID   uint     `json:"video_id"`
	Title     string   `json:"title"`
	LikeCount int64    `json:"like_count"`
	CoverPath string   `json:"cover_path"`
	Author    UserInfo `json:"author"`
}

const rankingKey = "ranking:likes"

// GetTopVideos 获取排行榜 Top N
// 优先从 Redis Sorted Set 读取，Redis 不可用时降级到 MySQL
func (s *RankingService) GetTopVideos(limit int) ([]RankingItem, error) {
	// 1. 尝试从 Redis 获取
	if repository.RDB != nil {
		items, err := s.getTopFromRedis(limit)
		if err == nil {
			return items, nil
		}
		log.Printf("从 Redis 获取排行榜失败，降级到 MySQL: %v", err)
	}

	// 2. 降级：直接从 MySQL 查询
	return s.getTopFromMySQL(limit)
}

// getTopFromRedis 从 Redis Sorted Set 获取排行榜
func (s *RankingService) getTopFromRedis(limit int) ([]RankingItem, error) {
	ctx := context.Background()

	// ZREVRANGE ranking:likes 0 limit-1 WITHSCORES
	results, err := repository.RDB.ZRevRangeWithScores(ctx, rankingKey, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("Redis ZRevRangeWithScores 失败: %w", err)
	}

	if len(results) == 0 {
		return s.getTopFromMySQL(limit)
	}

	// 收集 video_id 列表
	videoIDs := make([]uint, 0, len(results))
	for _, z := range results {
		id, err := strconv.ParseUint(z.Member.(string), 10, 64)
		if err != nil {
			continue
		}
		videoIDs = append(videoIDs, uint(id))
	}

	// 批量从 MySQL 查视频详情
	items := make([]RankingItem, 0, len(videoIDs))
	for i, videoID := range videoIDs {
		video, err := s.videoRepo.FindByID(videoID)
		if err != nil {
			continue
		}

		author, _ := s.userRepo.FindByID(video.UserID)
		authorInfo := UserInfo{}
		if author != nil {
			authorInfo = UserInfo{
				ID:       author.ID,
				Username: author.Username,
				Nickname: author.Nickname,
				Avatar:   author.Avatar,
			}
		}

		items = append(items, RankingItem{
			Rank:      i + 1,
			VideoID:   video.ID,
			Title:     video.Title,
			LikeCount: int64(results[i].Score),
			CoverPath: video.CoverPath,
			Author:    authorInfo,
		})
	}

	return items, nil
}

// getTopFromMySQL 从 MySQL 降级查询排行榜
func (s *RankingService) getTopFromMySQL(limit int) ([]RankingItem, error) {
	var videos []model.Video
	err := repository.DB.Where("status = ?", 1).
		Order("like_count DESC").
		Limit(limit).
		Find(&videos).Error
	if err != nil {
		return nil, fmt.Errorf("MySQL 查询排行榜失败: %w", err)
	}

	items := make([]RankingItem, 0, len(videos))
	for i, video := range videos {
		author, _ := s.userRepo.FindByID(video.UserID)
		authorInfo := UserInfo{}
		if author != nil {
			authorInfo = UserInfo{
				ID:       author.ID,
				Username: author.Username,
				Nickname: author.Nickname,
				Avatar:   author.Avatar,
			}
		}

		items = append(items, RankingItem{
			Rank:      i + 1,
			VideoID:   video.ID,
			Title:     video.Title,
			LikeCount: video.LikeCount,
			CoverPath: video.CoverPath,
			Author:    authorInfo,
		})
	}

	return items, nil
}

// ========== 点赞联动 Redis ==========

// UpdateRankingLikeCount 点赞/取消点赞时更新 Redis 排行榜
// delta: +1 表示点赞，-1 表示取消点赞
func UpdateRankingLikeCount(videoID uint, delta int64) {
	if repository.RDB == nil {
		return
	}
	ctx := context.Background()
	member := strconv.FormatUint(uint64(videoID), 10)
	if err := repository.RDB.ZIncrBy(ctx, rankingKey, float64(delta), member).Err(); err != nil {
		log.Printf("更新 Redis 排行榜失败 (video_id=%d, delta=%d): %v", videoID, delta, err)
	}
}

// RefreshRanking 冷启动：从 MySQL 全量同步视频点赞数到 Redis Sorted Set
func RefreshRanking() error {
	if repository.RDB == nil {
		return fmt.Errorf("Redis 未连接")
	}

	var videos []model.Video
	if err := repository.DB.Where("status = ?", 1).Find(&videos).Error; err != nil {
		return fmt.Errorf("查询视频列表失败: %w", err)
	}

	ctx := context.Background()
	// 先清空旧排行榜
	repository.RDB.Del(ctx, rankingKey)

	if len(videos) == 0 {
		log.Println("排行榜刷新完成，共 0 条视频（数据库为空）")
		return nil
	}

	// 批量 ZADD
	members := make([]redis.Z, 0, len(videos))
	for _, video := range videos {
		members = append(members, redis.Z{
			Score:  float64(video.LikeCount),
			Member: strconv.FormatUint(uint64(video.ID), 10),
		})
	}

	if err := repository.RDB.ZAdd(ctx, rankingKey, members...).Err(); err != nil {
		return fmt.Errorf("Redis ZADD 失败: %w", err)
	}

	log.Printf("排行榜刷新完成，共 %d 条视频", len(videos))
	return nil
}
