package main

import (
	"api/config"
	"api/internal/handlers"
	"api/internal/middleware"
	"api/internal/models"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	ginprometheus "github.com/zsais/go-gin-prometheus"

	docs "api/docs"
)

// @title           User Management API
// @version         1.0
// @description     A complete RESTful API for user management with authentication, authorization, and logging.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func setupLogger(cfg *config.Config) *logrus.Logger {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Ensure log directory exists
	logDir := filepath.Dir(cfg.Log.File)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		logger.WithError(err).Fatal("Failed to create log directory")
	}

	// Set up log file
	file, err := os.OpenFile(cfg.Log.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logger.WithError(err).Fatal("Failed to open log file")
	}

	logger.SetOutput(file)
	logger.SetFormatter(&logrus.JSONFormatter{})

	return logger
}

func setupDatabase(cfg *config.DatabaseConfig, logger *logrus.Logger) *gorm.DB {
	// First, connect to the default postgres database to check if our database exists
	defaultDBInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.SSLMode)

	defaultDB, err := gorm.Open("postgres", defaultDBInfo)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to postgres database")
	}
	defer defaultDB.Close()

	// Check if database exists
	var dbExists bool
	err = defaultDB.Raw("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = ?)", cfg.DBName).Row().Scan(&dbExists)
	if err != nil {
		logger.WithError(err).Fatal("Failed to check if database exists")
	}

	// Create database if it doesn't exist
	if !dbExists {
		logger.Info("Creating database: ", cfg.DBName)
		err = defaultDB.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.DBName)).Error
		if err != nil {
			logger.WithError(err).Fatal("Failed to create database")
		}
		logger.Info("Database created successfully")
	}

	// Connect to the actual database
	dbInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := gorm.Open("postgres", dbInfo)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}

	// Auto-migrate models
	db.AutoMigrate(&models.User{}, &models.RefreshToken{}, &models.UserProfile{})

	return db
}

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Setup logger
	logger := setupLogger(cfg)

	// Setup database
	db := setupDatabase(&cfg.Database, logger)
	defer db.Close()

	// Initialize Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Initialize Prometheus middleware
	p := ginprometheus.NewPrometheus("gin")
	p.Use(router)

	// Initialize Swagger
	docs.SwaggerInfo.Title = "User Management API"
	docs.SwaggerInfo.Description = "A complete RESTful API for user management with authentication, authorization, and logging."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	// Middleware
	router.Use(gin.Recovery())
	router.Use(middleware.LoggingMiddleware(logger))

	// CORS configuration
	corsConfig := cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(corsConfig))

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, logger, &struct {
		AccessSecret  string
		RefreshSecret string
		AccessExpiry  int
		RefreshExpiry int
	}{
		AccessSecret:  cfg.JWT.AccessSecret,
		RefreshSecret: cfg.JWT.RefreshSecret,
		AccessExpiry:  cfg.JWT.AccessExpiry,
		RefreshExpiry: cfg.JWT.RefreshExpiry,
	})
	userHandler := handlers.NewUserHandler(db, logger)
	adminHandler := handlers.NewAdminHandler(db, logger)

	// Serve Scalar documentation
	// Serve the main documentation page
	router.StaticFile("/", "./statics/index.html") // Serve at root for better UX

	// Serve the OpenAPI/Swagger specification
	router.GET("/docs/swagger.json", func(c *gin.Context) {
		c.File("./docs/swagger.json")
	})

	// Legacy Swagger UI (optional)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Health check
		// @Summary Check API health
		// @Description Get the health status of the API
		// @Tags health
		// @Produce json
		// @Success 200 {object} map[string]string "status: OK"
		// @Router /health [get]
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "OK",
				"time":   time.Now().Format(time.RFC3339),
			})
		})

		// Auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", middleware.AuthMiddleware(cfg.JWT.AccessSecret), authHandler.Logout)
		}

		// Protected user routes
		user := v1.Group("/users")
		user.Use(middleware.AuthMiddleware(cfg.JWT.AccessSecret))
		{
			user.GET("/profile", userHandler.GetProfile)
			user.PUT("/profile", userHandler.UpdateProfile)
			user.PUT("/change-password", userHandler.ChangePassword)
			user.DELETE("/account", userHandler.DeleteAccount)
		}

		// Admin routes
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(cfg.JWT.AccessSecret), middleware.AdminMiddleware())
		{
			admin.GET("/users", adminHandler.ListUsers)
			admin.PUT("/users/:id/role", adminHandler.ChangeUserRole)
		}
	}

	// Start server
	logger.WithField("port", cfg.Server.Port).Info("Starting server")

	// Print startup message with links
	fmt.Printf("\n🚀 Server started successfully!\n\n")
	fmt.Printf("📡 API is running at: \033[36mhttp://localhost:%s/api/v1\033[0m\n", cfg.Server.Port)
	fmt.Printf("📚 API Documentation (Scalar UI): \033[36mhttp://localhost:%s\033[0m\n", cfg.Server.Port)
	fmt.Printf("📖 API Documentation (Swagger UI): \033[36mhttp://localhost:%s/swagger/index.html\033[0m\n\n", cfg.Server.Port)
	fmt.Printf("🏥 Health check: \033[36mhttp://localhost:%s/api/v1/health\033[0m\n\n", cfg.Server.Port)

	if err := router.Run(":" + cfg.Server.Port); err != nil {
		logger.WithError(err).Fatal("Failed to start server")
	}
}
