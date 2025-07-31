package main

import (
	"embed"
	"log"
	"net/http"
	"time"

	"quiz-system/handlers"
	"quiz-system/services"
	"github.com/gin-gonic/gin"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	// 初始化数据库
	if err := services.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化缓存
	if err := services.InitCache(); err != nil {
		log.Fatalf("Failed to initialize cache: %v", err)
	}

	// 启动缓存清理定时器
	services.Cache.StartCleanupTimer()

	// 预加载热点题目到缓存
	if err := services.Cache.PreloadQuestions(); err != nil {
		log.Printf("Warning: Failed to preload questions to cache: %v", err)
	}

	// 初始化学习进度服务
	handlers.InitLearningProgressService()

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin路由器
	r := gin.New()

	// 添加中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	// 根路径重定向到静态文件
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		data, err := staticFiles.ReadFile("static/index.html")
		if err != nil {
			c.String(http.StatusNotFound, "File not found")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	// 静态文件路由
	r.GET("/app.js", func(c *gin.Context) {
		c.Header("Content-Type", "application/javascript")
		data, err := staticFiles.ReadFile("static/app.js")
		if err != nil {
			c.String(http.StatusNotFound, "File not found")
			return
		}
		c.Data(http.StatusOK, "application/javascript", data)
	})

	// API路由组
	api := r.Group("/api")
	{
		// 认证路由
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
			auth.POST("/logout", handlers.Logout)
		}

		// 需要认证的路由
		authenticated := api.Group("/")
		authenticated.Use(handlers.AuthMiddleware())
		{
			// 用户相关路由
			user := authenticated.Group("/user")
			{
				user.GET("/profile", handlers.GetProfile)
				user.GET("/stats", handlers.GetUserStatsHandler)
			}

			// 学习进度相关路由
			progress := authenticated.Group("/progress")
			{
				progress.GET("/dashboard", handlers.GetLearningDashboard)
				progress.GET("/summary", handlers.GetProgressSummary)
				progress.GET("/categories", handlers.GetCategoryProgress)
				progress.GET("/sessions", handlers.GetStudySessions)
				progress.GET("/insights", handlers.GetLearningInsights)
				progress.PUT("/insights/:insightId/read", handlers.MarkInsightAsRead)
				progress.GET("/weekly", handlers.GetWeeklyStats)
				progress.GET("/daily", handlers.GetDailyStats)
			}

			// 题目相关路由
			questions := authenticated.Group("/questions")
			{
				questions.GET("/", handlers.GetQuestions)
				questions.GET("/:id", handlers.GetQuestion)
				questions.GET("/categories", handlers.GetCategories)
				questions.GET("/search", handlers.SearchQuestions)
				questions.GET("/wrong", handlers.GetWrongQuestions)
				questions.POST("/submit", handlers.SubmitAnswer)
			}

			// 考试相关路由
			exam := authenticated.Group("/exam")
			{
				exam.POST("/start", handlers.StartExam)
				exam.GET("/:sessionId", handlers.GetExamSession)
				exam.POST("/:sessionId/answer", handlers.SubmitExamAnswer)
				exam.POST("/:sessionId/complete", handlers.CompleteExam)
				exam.GET("/history", handlers.GetExamHistory)
				
				// 答题卡相关路由
				exam.GET("/:examId/answer-sheet", handlers.GetAnswerSheet)
				exam.PUT("/:examId/answer-sheet/question/:questionNum", handlers.UpdateQuestionStatus)
				exam.POST("/:examId/answer-sheet/question/:questionNum/mark", handlers.ToggleQuestionMark)
				exam.GET("/:examId/answer-sheet/stats", handlers.GetAnswerSheetStats)
				exam.GET("/:examId/answer-sheet/question/:questionNum", handlers.JumpToQuestion)
			}
		}

		// 系统状态路由（无需认证）
		api.GET("/health", func(c *gin.Context) {
			stats := services.Cache.GetCacheStats()
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"timestamp": time.Now(),
				"cache_stats": stats,
			})
		})
	}

	// 启动服务器
	port := ":50442"
	log.Printf("Server starting on port %s", port)
	log.Printf("Access the application at: http://localhost%s", port)
	
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}