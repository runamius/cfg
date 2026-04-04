package main

import (
	"avito/iternal/config"
	"avito/iternal/repository/postgres"
	"avito/iternal/repository/service"
	transportHTTP "avito/transport"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	userRepo := postgres.NewUserRepo(pool)
	roomRepo := postgres.NewRoomRepo(pool)
	scheduleRepo := postgres.NewScheduleRepo(pool)
	slotRepo := postgres.NewSlotRepo(pool)
	bookingRepo := postgres.NewBookingRepo(pool)

	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)
	roomSvc := service.NewRoomService(roomRepo)
	scheduleSvc := service.NewScheduleService(scheduleRepo, roomRepo)
	slotSvc := service.NewSlotService(slotRepo, scheduleRepo, roomRepo)
	bookingSvc := service.NewBookingService(bookingRepo, slotRepo)

	router := transportHTTP.NewRouter(authSvc, roomSvc, scheduleSvc, slotSvc, bookingSvc)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting server on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("Server exited")
}
