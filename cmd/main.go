package main

import (
	"fmt"
	"log"
	"video_feed/internal/config"
	"video_feed/internal/handler"
	"video_feed/internal/middleware"
	"video_feed/internal/repository"
	"video_feed/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 加载配置
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 初始化数据库
	repository.InitDB(&cfg.Database)
	defer repository.CloseMySQL()

	// 3. 初始化 JWT（从配置文件读取 Secret 和过期时间）
	middleware.InitJWT(cfg.JWT.Secret, cfg.JWT.Expire)

	// 4. 初始化 Redis（连接失败不影响服务启动）
	repository.InitRedis(&cfg.Redis)
	defer repository.CloseRedis()

	// 5. 初始化 RabbitMQ（连接失败不影响服务启动）
	repository.InitRabbitMQ(&cfg.RabbitMQ)
	defer repository.CloseRabbitMQ()

	// 6. 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 7. 创建 Gin 引擎
	r := gin.Default()

	// 8. 注册路由
	userHandler := handler.NewUserHandler()
	videoHandler := handler.NewVideoHandler()
	interactionHandler := handler.NewInteractionHandler()
	commentHandler := handler.NewCommentHandler()
	followHandler := handler.NewFollowHandler()
	rankingHandler := handler.NewRankingHandler()

	// 冷启动：从 MySQL 同步点赞数到 Redis 排行榜（Redis 不可用时会跳过）
	service.RefreshRanking()

	api := r.Group("/api")
	{
		// 公开接口（无需登录）
		api.POST("/user/register", userHandler.Register)             // 用户注册
		api.POST("/user/login", userHandler.Login)                   // 用户登录
		api.GET("/feed", videoHandler.GetFeed)                       // 获取feed流
		api.GET("/video/play/:id", videoHandler.Play)                // 视频播放接口
		api.GET("/video/:id/comments", commentHandler.GetComments)   // 查看评论
		api.GET("/video/search", videoHandler.Search)                // 视频搜索
		api.GET("/user/:id/followers", followHandler.GetFollowers)   // 获取粉丝列表
		api.GET("/user/:id/followings", followHandler.GetFollowings) // 获取关注列表
		api.GET("/ranking", rankingHandler.GetRanking)                 // 点赞排行榜

		// 需要登录的接口（使用JWT中间件）
		auth := api.Group("")
		auth.Use(middleware.JWTAuth())
		{
			// 用户相关
			auth.GET("/user/profile", userHandler.GetProfile)    // 获取个人简介
			auth.PUT("/user/profile", userHandler.UpdateProfile) // 更新个人简介
			auth.POST("/user/avatar", userHandler.UpdateAvatar)  // 更新头像

			// 视频相关
			auth.POST("/video/upload", videoHandler.Upload) // 上传视频

			// 点赞相关
			auth.POST("/video/:id/like", interactionHandler.LikeVideo)     // 点赞
			auth.DELETE("/video/:id/like", interactionHandler.UnlikeVideo) // 取消点赞
			auth.GET("/video/:id/like", interactionHandler.GetLikeStatus)  // 获取点赞状态

			// 收藏相关
			auth.POST("/video/:id/favorite", interactionHandler.FavoriteVideo)     // 收藏
			auth.DELETE("/video/:id/favorite", interactionHandler.UnfavoriteVideo) // 取消收藏
			auth.GET("/video/:id/favorite", interactionHandler.GetFavoriteStatus)  // 获取收藏状态
			auth.GET("/user/favorites", interactionHandler.GetUserFavorites)       // 获取收藏列表

			// 评论相关
			auth.POST("/video/:id/comment", commentHandler.CreateComment) // 发表评论
			auth.DELETE("/comment/:id", commentHandler.DeleteComment)     // 删除评论

			// 关注相关
			auth.POST("/user/:id/follow", followHandler.FollowUser)     // 关注用户
			auth.DELETE("/user/:id/follow", followHandler.UnfollowUser) // 取消关注

			// 管理相关
			auth.POST("/admin/refresh-ranking", rankingHandler.RefreshRanking) // 手动刷新排行榜
		}
	}

	// 9. 健康检查
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// 10. 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("视频Feed流服务启动成功，监听端口: %d", cfg.Server.Port)
	err = r.Run(addr)
	if err != nil {
		log.Fatalln("启动服务器失败：", err)
	}
	log.Println("服务器已安全退出")
}
