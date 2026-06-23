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

	api := r.Group("/api")
	{
		// 公开接口（无需登录）
		api.POST("/user/register", userHandler.Register)
		api.POST("/user/login", userHandler.Login)
		api.GET("/feed", videoHandler.GetFeed)
		api.GET("/video/play/:id", videoHandler.Play)  // 视频播放接口

		// 需要登录的接口（使用JWT中间件）
		auth := api.Group("")
		auth.Use(middleware.JWTAuth())
		{
			// 只有JWT放行，才能到这里
			// 用户相关
			auth.GET("/user/profile", userHandler.GetProfile)
			auth.PUT("/user/profile", userHandler.UpdateProfile) // 更新个人信息
			auth.POST("/user/avatar", userHandler.UpdateAvatar)  // 更新头像

			// 视频相关
			auth.POST("/video/upload", videoHandler.Upload)
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
	r.Run(addr)
}
