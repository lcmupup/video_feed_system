package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"video_feed/internal/model"
	"video_feed/internal/repository"
)

// VideoService 视频业务逻辑层
type VideoService struct {
	videoRepo *repository.VideoRepository
}

// NewVideoService 创建视频服务实例
func NewVideoService() *VideoService {
	return &VideoService{
		videoRepo: repository.NewVideoRepository(),
	}
}

// UploadVideoReq 视频上传请求
type UploadVideoReq struct {
	Title       string                `form:"title" binding:"required"`
	Description string                `form:"description"`
	File        *multipart.FileHeader `form:"file" binding:"required"` //不直接包含文件内容，只记录文件名、文件大小、HTTP头部信息等
	UserID      uint
}

// VideoInfo 视频信息响应
type VideoInfo struct {
	ID             uint   `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	FilePath       string `json:"file_path,omitempty"` // 搜索时不返回文件路径，omitempty表示当字段值为零值时，序列化时忽略该字段
	FileSize       int64  `json:"file_size"`
	CreatedAt      string `json:"created_at"`
	URL            string `json:"url"`                       // 视频播放地址
	Status         int    `json:"status"`                    // 视频状态，0-待审核，1-已发布，2-已下架
	HighlightTitle string `json:"highlight_title,omitempty"` // 搜索高亮（仅搜索时返回）
	HighlightDesc  string `json:"highlight_desc,omitempty"`  // 搜索高亮（仅搜索时返回）
	IsLiked        bool   `json:"is_liked"`                  // 是否点赞
	IsFavored      bool   `json:"is_favored"`                // 是否收藏
}

// VideoUploadMessage 视频上传 RabbitMQ 消息体
type VideoUploadMessage struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	FilePath    string `json:"file_path"`
	FileSize    int64  `json:"file_size"`
	UserID      uint   `json:"user_id"`
	Status      int    `json:"status"`
}

// UploadVideo 上传视频
// 文件落盘后发布到 RabbitMQ 即返回，消费者异步写入 MySQL
func (s *VideoService) UploadVideo(req *UploadVideoReq) (*VideoInfo, error) {
	// 1. 验证文件类型
	ext := filepath.Ext(req.File.Filename)
	allowedExts := map[string]bool{
		".mp4": true,
		".avi": true,
		".mov": true,
		".wmv": true,
		".flv": true,
		".mkv": true,
	}
	if !allowedExts[ext] {
		return nil, errors.New("不支持的视频格式，仅支持 mp4/avi/mov/wmv/flv/mkv")
	}

	// 2. 限制文件大小（最大100MB）
	maxSize := int64(100 * 1024 * 1024)
	if req.File.Size > maxSize {
		return nil, errors.New("视频文件过大，最大支持100MB")
	}

	// 3. 创建存储目录
	uploadDir := "./uploads/videos/" + time.Now().Format("2006/01/02")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, errors.New("创建存储目录失败")
	}

	// 4. 生成唯一文件名并保存文件
	fileName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), req.Title, ext)
	filePath := filepath.Join(uploadDir, fileName)

	src, err := req.File.Open()
	if err != nil {
		return nil, errors.New("打开上传文件失败")
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return nil, errors.New("创建目标文件失败")
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return nil, errors.New("保存文件失败")
	}

	// 5. 发布到 RabbitMQ（异步写入 MySQL），不阻塞用户
	msg := VideoUploadMessage{
		Title:       req.Title,
		Description: req.Description,
		FilePath:    filePath,
		FileSize:    req.File.Size,
		UserID:      req.UserID,
		Status:      1,
	}

	jsonBody, err := json.Marshal(msg)
	if err != nil {
		return nil, errors.New("消息序列化失败")
	}

	if err := repository.PublishVideoUpload(jsonBody); err != nil {
		// RabbitMQ 不可用时，降级为同步写 MySQL
		log.Printf("RabbitMQ 发布失败，降级为同步写入 MySQL: %v", err)
		video := &model.Video{
			Title:       req.Title,
			Description: req.Description,
			FilePath:    filePath,
			FileSize:    req.File.Size,
			UserID:      req.UserID,
			Status:      1,
		}
		if err := s.videoRepo.Create(video); err != nil {
			return nil, errors.New("保存视频信息失败")
		}
		return &VideoInfo{
			ID:          video.ID,
			Title:       video.Title,
			Description: video.Description,
			FilePath:    video.FilePath,
			FileSize:    video.FileSize,
			CreatedAt:   video.CreatedAt.Format("2006-01-02 15:04:05"),
			Status:      video.Status,
		}, nil
	}

	// 异步处理，立即返回（此时尚未写入 MySQL，无 ID）
	return &VideoInfo{
		Title:       req.Title,
		Description: req.Description,
		FilePath:    filePath,
		FileSize:    req.File.Size,
		Status:      0, // 0 = 处理中
	}, nil
}

// SaveVideoToDB 将视频信息写入 MySQL（Consumer 调用）
func SaveVideoToDB(msg *VideoUploadMessage) (*model.Video, error) {
	video := &model.Video{
		Title:       msg.Title,
		Description: msg.Description,
		FilePath:    msg.FilePath,
		FileSize:    msg.FileSize,
		UserID:      msg.UserID,
		Status:      msg.Status,
	}

	repo := repository.NewVideoRepository()
	if err := repo.Create(video); err != nil {
		return nil, err
	}
	return video, nil
}

// StartVideoConsumer 启动视频上传消费者
func StartVideoConsumer() {
	if repository.MQChannel == nil {
		log.Println("RabbitMQ 未连接，跳过视频上传消费者启动")
		return
	}

	msgs, err := repository.ConsumeVideoUploads()
	if err != nil {
		log.Printf("启动视频上传消费者失败: %v", err)
		return
	}

	log.Println("视频上传消费者已启动")

	for d := range msgs {
		var msg VideoUploadMessage
		if err := json.Unmarshal(d.Body, &msg); err != nil {
			log.Printf("视频消息解析失败: %v", err)
			d.Nack(false, false)
			continue
		}

		video, err := SaveVideoToDB(&msg)
		if err != nil {
			log.Printf("视频信息写入 MySQL 失败: %v", err)
			d.Nack(false, true) // 重新入队，稍后重试
			continue
		}

		log.Printf("视频信息已写入 MySQL: id=%d, title=%s", video.ID, video.Title)
		d.Ack(false)
	}
}

// GetFeed 获取Feed流
func (s *VideoService) GetFeed(page, pageSize int) ([]VideoInfo, int64, error) {
	videos, total, err := s.videoRepo.GetFeed(page, pageSize)
	if err != nil {
		return nil, 0, errors.New("获取Feed流失败")
	}

	// 转换成响应格式
	var videoInfos []VideoInfo
	for _, v := range videos {
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

	return videoInfos, total, nil
}

// GetVideoByID 通过视频ID获取视频
func (s *VideoService) GetVideoByID(id uint) (*model.Video, error) {
	video, err := s.videoRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("视频不存在")
	}
	return video, nil
}

// SearchVideos 搜索视频
func (s *VideoService) SearchVideos(keyword string, page, pageSize int) ([]VideoInfo, int64, error) {
	// 1. 校验关键词长度
	if len(keyword) < 1 || len(keyword) > 100 {
		return nil, 0, errors.New("关键词长度应在1-100个字符之间")
	}

	// 2. 调用仓库层搜索
	videos, total, err := s.videoRepo.SearchVideos(keyword, page, pageSize)
	if err != nil {
		return nil, 0, errors.New("搜索视频失败")
	}

	// 3. 转换为响应格式
	var videoInfos []VideoInfo
	for _, v := range videos {
		videoInfo := VideoInfo{
			ID:          v.ID,
			Title:       v.Title,
			Description: v.Description,
			FilePath:    v.FilePath,
			FileSize:    v.FileSize,
			CreatedAt:   v.CreatedAt.Format("2006-01-02 15:04:05"),
			Status:      v.Status,
			// 高亮关键词
			HighlightTitle: highlightKeyword(v.Title, keyword),
			HighlightDesc:  highlightKeyword(v.Description, keyword),
		}
		videoInfos = append(videoInfos, videoInfo)
	}

	return videoInfos, total, nil
}

// highlightKeyword 高亮关键词（用HTML标记包裹）
func highlightKeyword(text, keyword string) string {
	return strings.ReplaceAll(text, keyword, "<em>"+keyword+"</em>") // 将text中所有匹配keyword的子串替换为新的内容
}
