package handler

import (
	"strconv"

	"video_feed/internal/service"
	"video_feed/pkg/errcode"

	"github.com/gin-gonic/gin"
)

// RankingHandler 排行榜 HTTP 处理器
type RankingHandler struct {
	rankingService *service.RankingService
}

// NewRankingHandler 创建排行榜处理器实例
func NewRankingHandler() *RankingHandler {
	return &RankingHandler{
		rankingService: service.NewRankingService(),
	}
}

// GetRanking 获取点赞排行榜
// GET /api/ranking?limit=50
func (h *RankingHandler) GetRanking(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 50
	}
	if limit > 100 {
		limit = 100 // 最多返回 100 条
	}

	items, err := h.rankingService.GetTopVideos(limit)
	if err != nil {
		errcode.Error(c, errcode.CodeInternalError, "获取排行榜失败: "+err.Error())
		return
	}

	errcode.SuccessWithData(c, "获取成功", items)
}

// RefreshRanking 手动刷新排行榜（冷启动/数据修复）
// POST /api/admin/refresh-ranking
func (h *RankingHandler) RefreshRanking(c *gin.Context) {
	if err := service.RefreshRanking(); err != nil {
		errcode.Error(c, errcode.CodeInternalError, "刷新排行榜失败: "+err.Error())
		return
	}

	errcode.Success(c, "排行榜刷新成功")
}
