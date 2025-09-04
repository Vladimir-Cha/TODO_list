package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/Vladimir-Cha/TODO_list/internal/config"
	"github.com/Vladimir-Cha/TODO_list/internal/db"
	"github.com/Vladimir-Cha/TODO_list/internal/repository"
	"github.com/Vladimir-Cha/TODO_list/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// флаг для миграций
	migrateOnly := flag.Bool("migrate-only", false, "Run migrations and exit")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// подключение к БД
	pool, err := db.InitNew(ctx, *cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// опциональные миграции
	if cfg.RunMigrations || *migrateOnly {
		if err := db.GooseMigrationsWithPool(ctx, pool); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Migrations applied successfully")

		if *migrateOnly {
			log.Println("Migration only mode - exiting")
			os.Exit(0) // Завершаем работу
		}
	}

	// создание repository с pgxpool
	taskRepo := repository.NewTaskRepository(pool)
	taskService := service.NewTaskService(taskRepo)
	handlers := service.NewHandlers(taskService)

	// Настраиваем Echo
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.POST("/api/task", handlers.CreateTask)
	e.GET("/api/task", handlers.GetTaskByID)
	e.PUT("/api/task", handlers.UpdateTask)
	e.DELETE("/api/task", handlers.DeleteTask)
	e.GET("/api/tasks", handlers.ListTasks)
	e.POST("/api/task/done", handlers.MarkTaskAsDone)
	e.GET("/api/nextdate", handlers.GetNextDate)

	// запуск сервера
	go func() {
		addr := ":" + strconv.Itoa(cfg.Server.Port)
		log.Printf("Server starting on %s", addr)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Starting server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server shutdown error:", err)
	}
	log.Println("Server shutdown completed gracefully")
}
