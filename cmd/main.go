package main

import (
	"fmt"
	"log"
	"video_feed/internal/config"
	"video_feed/internal/handler"
	"video_feed/internal/middleware"
	"video_feed/internal/repository"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 加载配置
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		//打印完内容后会立即终止程序，defer后的语句都不会执行
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 初始化数据库
	repository.InitDB(&cfg.Database)

	// 3. 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 4. 创建Gin引擎
	r := gin.Default()

	// 5. 注册路由
	userHandler := handler.NewUserHandler()
	videoHandler := handler.NewVideoHandler()
	interactionHandler := handler.NewInteractionHandler()
	commentHandler := handler.NewCommentHandler()
	followHandler := handler.NewFollowHandler()

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

		// 需要登录的接口（使用JWT中间件）
		auth := api.Group("")
		auth.Use(middleware.JWTAuth())
		{
			// 只有JWT放行，才能到这里
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
		}
	}

	// 6. 健康检查
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// 7. 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("视频Feed流服务启动成功，监听端口: %d", cfg.Server.Port)
	err = r.Run(addr)
	if err != nil {
		log.Fatalln("启动服务器失败：", err)
	}
}
