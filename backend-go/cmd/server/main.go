package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/config"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/handlers"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/middleware"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/models"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/repository"
	"github.com/chenxxianyi/NoteWeb/backend-go/internal/service"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	// Database
	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Document{}, &models.Annotation{}, &models.Note{}); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// Repositories
	userRepo := repository.NewUserRepo(db)
	docRepo := repository.NewDocumentRepo(db)
	annRepo := repository.NewAnnotationRepo(db)
	noteRepo := repository.NewNoteRepo(db)

	// Services
	authSvc := service.NewAuthService(userRepo, cfg.SecretKey, cfg.AccessTokenExpireMinutes)
	docSvc := service.NewDocumentService(docRepo, userRepo, "./uploads")
	annSvc := service.NewAnnotationService(annRepo)
	noteSvc := service.NewNoteService(noteRepo, docRepo)
	aiSvc := service.NewAIService()

	// Handlers
	authH := handlers.NewAuthHandler(authSvc)
	docH := handlers.NewDocumentHandler(docSvc)
	annH := handlers.NewAnnotationHandler(annSvc)
	noteH := handlers.NewNoteHandler(noteSvc)
	aiH := handlers.NewAIHandler(aiSvc)

	// Router
	r := gin.Default()
	api := r.Group("/api/v1")

	// Public
	auth := api.Group("/auth")
	auth.POST("/register", authH.Register)
	auth.POST("/login", authH.Login)

	// AI (public mock)
	ai := api.Group("/ai")
	ai.GET("/documents/:id/summary", aiH.Summary)
	ai.POST("/explain", aiH.Explain)
	ai.POST("/translate", aiH.Translate)
	ai.POST("/chat", aiH.Chat)

	// Protected
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(cfg.SecretKey))
	{
		protected.GET("/auth/me", authH.Me)

		docs := protected.Group("/documents")
		docs.GET("", docH.List)
		docs.GET("/:id", docH.Get)
		docs.GET("/:id/content", docH.GetContent)
		docs.GET("/:id/file", docH.GetFile)
		docs.POST("/upload", docH.Upload)
		docs.PATCH("/:id", docH.Rename)
		docs.PATCH("/:id/read-progress", docH.UpdateReadProgress)
		docs.DELETE("/:id", docH.Delete)

		protected.GET("/documents/:id/annotations", annH.ListByDocument)
		protected.POST("/annotations", annH.Create)
		protected.DELETE("/annotations/:id", annH.Delete)

		protected.GET("/notes", noteH.List)
		protected.GET("/notes/:id", noteH.Get)
		protected.POST("/notes", noteH.Create)
		protected.PATCH("/notes/:id", noteH.Update)
		protected.DELETE("/notes/:id", noteH.Delete)
	}

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "app": cfg.AppName})
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("NoteWeb API 启动于 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{})
}

func init() {
	os.MkdirAll("./uploads", 0755)
}
