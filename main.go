package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

var (
	version   = "1.0.0"
	buildTime = ""
)

func main() {
	versionFlag := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Version: %s\nBuild Time: %s\n", version, buildTime)
		return
	}

	cfg := loadConfig()
	log.Printf("Starting instance %s", cfg.InstanceID)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cluster, err := NewCluster(cfg)
	if err != nil {
		log.Fatalf("Failed to create cluster: %v", err)
	}

	services, err := SetupServices(ctx, cfg, cluster)
	if err != nil {
		log.Fatalf("Failed to setup services: %v", err)
	}

	gin.SetMode(gin.ReleaseMode)
	router := setupGinRoutes(services)
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: router,
	}

	go func() {
		log.Printf("HTTP server starting on port %d", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	grpcServer := grpc.NewServer()
	setupGRPCServices(grpcServer, services)
	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	go func() {
		log.Printf("gRPC server starting on port %d", cfg.GRPCPort)
		if err := grpcServer.Serve(grpcLis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	waitForShutdown(ctx, httpServer, grpcServer, cluster, services)
	log.Println("Application stopped gracefully")
}

func waitForShutdown(ctx context.Context, httpServer *http.Server, grpcServer *grpc.Server, cluster *Cluster, services *Services) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	grpcServer.GracefulStop()

	if err := services.Shutdown(shutdownCtx); err != nil {
		log.Printf("Services shutdown error: %v", err)
	}

	if err := cluster.Leave(); err != nil {
		log.Printf("Cluster leave error: %v", err)
	}
}
