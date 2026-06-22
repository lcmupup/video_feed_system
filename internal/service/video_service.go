package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
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
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	FilePath    string `json:"file_path"`
	CoverPath   string `json:"cover_path"`
	Duration    int64  `json:"duration"`
	FileSize    int64  `json:"file_size"`
	CreatedAt   string `json:"created_at"`
	Status      int    `json:"status"`
}

// UploadVideo 上传视频
func (s *VideoService) UploadVideo(req *UploadVideoReq) (*VideoInfo, error) {
	// 1. 验证文件类型
	ext := filepath.Ext(req.File.Filename) // 从文件中提取扩展名（包含前导点）
	allowedExts := map[string]bool{        // 支持的视频格式
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
	maxSize := int64(100 * 1024 * 1024) // 100MB
	if req.File.Size > maxSize {
		return nil, errors.New("视频文件过大，最大支持100MB")
	}

	// 3. 创建存储目录
	uploadDir := "./uploads/videos/" + time.Now().Format("2006/01/02")
	err := os.MkdirAll(uploadDir, 0755) // 递归创建目录，如果父目录不存在，会自动创建所有缺失的父目录，0755表示权限值
	if err != nil {
		return nil, errors.New("创建存储目录失败")
	}

	// 4. 生成唯一文件名
	// time.Now().UnixNano()获取从1970-01-01到现在的纳秒数，因为纳秒精度很高，两次调用几乎不会重复，适合用来生成临时的唯一ID
	fileName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), req.Title, ext)
	filePath := filepath.Join(uploadDir, fileName)

	// 5. 保存文件
	// 打开源文件(用户上传的文件)
	src, err := req.File.Open()
	if err != nil {
		return nil, errors.New("打开上传文件失败")
	}
	defer src.Close()

	// 创建目标文件(要保存在服务器的文件)
	dst, err := os.Create(filePath) // 先在服务器上创建一个空文件
	if err != nil {
		return nil, errors.New("创建目标文件失败")
	}
	defer dst.Close()

	_, err = io.Copy(dst, src) // 再把用户文件的内容写到服务器文件里
	if err != nil {
		return nil, errors.New("保存文件失败")
	}

	// 6. 保存视频信息到数据库
	video := &model.Video{
		Title:       req.Title,
		Description: req.Description,
		FilePath:    filePath,
		FileSize:    req.File.Size,
		UserID:      req.UserID,
		Status:      1, // 默认已发布
	}

	err = s.videoRepo.Create(video)
	if err != nil {
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

// GetFeed 获取Feed流
func (s *VideoService) GetFeed(page, pageSize int) ([]VideoInfo, int64, error) {
	videos, total, err := s.videoRepo.GetFeed(page, pageSize)
	if err != nil {
		return nil, 0, errors.New("获取Feed流失败")
	}

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
