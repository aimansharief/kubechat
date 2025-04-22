package main

import (
	"kubechat-api/api"
	"kubechat-api/config"
	"kubechat-api/kube"
	"os"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

func main() {
	cwd, _ := os.Getwd()
	fmt.Println("DEBUG: Running from directory:", cwd)

	// Initialize components
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	router := gin.Default()

	// Cache the request body for all handlers/middlewares
	router.Use(api.BodyCacheMiddleware())

	// Load config
	cfg := config.Load("config/default.yaml")

	// Load kube client for first available cluster (support any kubeconfig)
	var clusterCfg *config.Cluster
	if len(cfg.Clusters) == 0 {
		panic("No clusters found in config. Please check your config file.")
	}
	// Use first cluster if only one, else prefer 'dev-cluster' if present
	clusterCfg = &cfg.Clusters[0]
	for _, cl := range cfg.Clusters {
		if cl.Name == "dev-cluster" {
			clusterCfg = &cl
			break
		}
	}
	kubeClient, err := kube.NewKubeClient(clusterCfg.Kubeconfig, false)
	if err != nil {
		logger.Warn("Kubeconfig not found or invalid, running in degraded mode", zap.Error(err))
		kubeClient = nil
	}
	if kubeClient == nil {
		logger.Error("[STARTUP] kubeClient is nil! Most API endpoints will return errors. Please check that kubeconfig.yaml exists and is valid. See README for setup instructions.")
	}

	// Register middleware
	router.Use(api.CORSMiddleware())
	router.Use(api.RequestLogger(logger))
	router.Use(func(c *gin.Context) { c.Set("logger", logger); c.Next() })
	// Inject kubeClient into context for handlers
	router.Use(func(c *gin.Context) { c.Set("kubeClient", kubeClient); c.Next() })

	router.Use(api.NewRateLimiter(10, time.Minute).Limit())

	// Register routes
	api.RegisterRoutes(router, cfg, logger, kubeClient)

	// Start server
	router.Run(":8080")
}
