package application

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Mussab2003/go-notes-app-auth_service.git/model"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	router *gin.Engine
	db     *gorm.DB
}

func New(ctx context.Context) *App {
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		host, user, pass, dbname, port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying DB: %v\n", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatalf("Database is unreachable: %v\n", err)
	}

	// ✅ Auto-migrate models
	if err := db.AutoMigrate(&model.User{}); err != nil {
		log.Fatalf("Failed to run migrations: %v\n", err)
	}

	// ✅ Connection pool settings (good defaults)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	app := &App{
		router: loadRoutes(),
		db:     db,
	}
	return app
}

func (a *App) Start(ctx context.Context) error {
	sqlDB, err := a.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying DB: %w", err)
	}
	defer sqlDB.Close()

	server := &http.Server{
		Addr:    ":8000",
		Handler: a.router,
	}

	ch := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			ch <- fmt.Errorf("unable to start server, %w", err)
		}
		close(ch)
	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		return server.Shutdown(timeout)
	}
}
