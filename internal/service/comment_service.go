package service

import (
	"errors"
	"video_feed/internal/model"
	"video_feed/internal/repository"
)

// CommentService 评论业务逻辑层
type CommentService struct {
	commentRepo *repository.CommentRepository
	videoRepo   *repository.VideoRepository
	userRepo    *repository.UserRepository
}

// NewCommentService 创建评论服务实例
func NewCommentService() *CommentService {
	return &CommentService{
		commentRepo: repository.NewCommentRepository(),
		videoRepo:   repository.NewVideoRepository(),
		userRepo:    repository.NewUserRepository(),
	}
}

// CreateCommentReq 创建评论请求
type CreateCommentReq struct {
	Content  string `json:"content" binding:"required"` // 评论内容
	VideoID  uint   `json:"video_id"`                   // 评论的视频ID
	ParentID *uint  `json:"parent_id"`                  // 可选，回复时填写父评论ID
	UserID   uint   `json:"-"`
}

// CommentInfo 评论信息响应
type CommentInfo struct {
	ID        uint          `json:"id"`
	Content   string        `json:"content"`
	UserID    uint          `json:"user_id"`
	Username  string        `json:"username"`
	Avatar    string        `json:"avatar"`
	VideoID   uint          `json:"video_id"`
	ParentID  *uint         `json:"parent_id"` // 父评论ID
	CreatedAt string        `json:"created_at"`
	Replies   []CommentInfo `json:"replies,omitempty"` // 子评论列表
}

// CreateComment 创建评论
func (s *CommentService) CreateComment(req *CreateCommentReq) (*CommentInfo, error) {
	// 1. 检查视频是否存在
	_, err := s.videoRepo.FindByID(req.VideoID)
	if err != nil {
		return nil, errors.New("视频不存在")
	}

	// 2. 如果是回复，检查父评论是否存在
	if req.ParentID != nil {
		_, err := s.commentRepo.FindByID(*req.ParentID)
		if err != nil {
			return nil, errors.New("父评论不存在")
		}
	}

	// 3. 校验内容长度
	if len(req.Content) > 1000 {
		return nil, errors.New("评论内容不能超过1000个字符")
	}

	// 4. 创建评论
	comment := &model.Comment{
		Content:  req.Content,
		UserID:   req.UserID,
		VideoID:  req.VideoID,
		ParentID: req.ParentID,
	}
	err = s.commentRepo.Create(comment)
	if err != nil {
		return nil, errors.New("发表评论失败")
	}

	// 5. 获取用户信息
	user, _ := s.userRepo.FindByID(req.UserID)

	return &CommentInfo{
		ID:        comment.ID,
		Content:   comment.Content,
		UserID:    comment.UserID,
		Username:  user.Username,
		Avatar:    user.Avatar,
		VideoID:   comment.VideoID,
		ParentID:  comment.ParentID,
		CreatedAt: comment.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// DeleteComment 删除评论
func (s *CommentService) DeleteComment(commentID, userID uint) error {
	// 1. 查找评论
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		return errors.New("评论不存在")
	}

	// 2. 检查权限（只有评论作者才能删除）
	if comment.UserID != userID {
		return errors.New("无权删除此评论")
	}

	// 3. 删除评论
	return s.commentRepo.DeleteByID(commentID)
}

// GetCommentsByVideoID 获取视频的评论列表
func (s *CommentService) GetCommentsByVideoID(videoID uint, page, pageSize int) ([]CommentInfo, int64, error) {
	// 1. 检查视频是否存在
	_, err := s.videoRepo.FindByID(videoID)
	if err != nil {
		return nil, 0, errors.New("视频不存在")
	}

	// 2. 获取评论
	comments, total, err := s.commentRepo.GetCommentsByVideoID(videoID, page, pageSize)
	if err != nil {
		return nil, 0, errors.New("获取评论失败")
	}

	// 3. 将model.Comment转换为CommentInfo格式
	var commentInfos []CommentInfo
	for _, c := range comments {
		commentInfo := s.convertToCommentInfo(&c)
		commentInfos = append(commentInfos, *commentInfo)
	}

	return commentInfos, total, nil
}

// convertToCommentInfo 将model.Comment转换为CommentInfo格式
func (s *CommentService) convertToCommentInfo(comment *model.Comment) *CommentInfo {
	info := &CommentInfo{
		ID:        comment.ID,
		Content:   comment.Content,
		UserID:    comment.User.ID,
		Username:  comment.User.Username,
		Avatar:    comment.User.Avatar,
		VideoID:   comment.VideoID,
		ParentID:  comment.ParentID,
		CreatedAt: comment.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	// 如果有回复，递归转换
	if len(comment.Replies) > 0 {
		for _, reply := range comment.Replies {
			replyInfo := s.convertToCommentInfo(&reply)
			info.Replies = append(info.Replies, *replyInfo)
		}
	}

	return info
}
