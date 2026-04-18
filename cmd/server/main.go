package main

// Реализованы згрузка конфигурации, инициализация БД, репозитория, сервисов и запуск HTTP-сервера.
// Поддержка закрытия greaceful shutdown.

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cdek/internal/handler"
	"cdek/internal/repository"
	"cdek/internal/service"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	dsn := os.Getenv("DB_DSN")

	jwtSecret := os.Getenv("JWT_SECRET")

	db, err := sql.Open("postgres", dsn)

	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}

	log.Println("Database connected successfully")

	repo := repository.NewPostgresRepository(db)

	svc := service.NewService(repo, jwtSecret)

	hndlr := handler.NewHandler(svc, jwtSecret)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: hndlr.InitRoutes(),
	}

	go func() {
		log.Printf("Server is running on port %s", port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
