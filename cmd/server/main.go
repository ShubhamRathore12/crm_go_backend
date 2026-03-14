package main

import (
	"crm-backend/internal/config"
	"crm-backend/internal/db"
	"crm-backend/internal/handlers"
	"crm-backend/internal/middleware"
	"crm-backend/internal/websocket"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	database, err := db.NewDatabaseManager(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Run migrations
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "./migrations"
	}
	if err := db.RunMigrations(database.Primary(), migrationsDir); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize WebSocket manager
	wsManager := websocket.NewManager()

	// Create app state
	appState := &handlers.AppState{
		DB:               database,
		Config:           cfg,
		WebSocketManager: wsManager,
	}

	// Create Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Add state to context
	r.Use(func(c *gin.Context) {
		c.Set("state", appState)
		c.Set("jwt_secret", cfg.JWTSecret)
		c.Next()
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	// Setup routes
	setupRoutes(r, appState)

	// Start WebSocket manager
	go wsManager.Run()

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes(r *gin.Engine, state *handlers.AppState) {
	// Auth routes
	auth := r.Group("/auth")
	{
		auth.POST("/login", handlers.Login)
		auth.POST("/register", handlers.Register)
		auth.POST("/otp", handlers.OTP)
		auth.POST("/logout", handlers.Logout)
		auth.GET("/me", middleware.AuthMiddleware(), handlers.Me)
		auth.POST("/forgot-password", handlers.ForgotPassword)
		auth.POST("/reset-password", handlers.ResetPassword)
	}

	// Leads routes
	leads := r.Group("/leads")
	leads.Use(middleware.AuthMiddleware())
	{
		leads.GET("", handlers.GetLeads)
		leads.POST("", handlers.CreateLead)
		leads.GET("/:id", handlers.GetLead)
		leads.PUT("/:id", handlers.UpdateLead)
		leads.DELETE("/:id", handlers.DeleteLead)
	}

	// Contacts routes
	contacts := r.Group("/contacts")
	contacts.Use(middleware.AuthMiddleware())
	{
		contacts.GET("", handlers.GetContacts)
		contacts.POST("", handlers.CreateContact)
		contacts.GET("/:id", handlers.GetContact)
		contacts.PUT("/:id", handlers.UpdateContact)
		contacts.DELETE("/:id", handlers.DeleteContact)
	}

	// Interactions routes
	interactions := r.Group("/interactions")
	interactions.Use(middleware.AuthMiddleware())
	{
		interactions.GET("", handlers.GetInteractions)
		interactions.POST("", handlers.CreateInteraction)
		interactions.GET("/:id", handlers.GetInteraction)
	}

	// Messaging routes
	sms := r.Group("/sms")
	sms.Use(middleware.AuthMiddleware())
	{
		sms.POST("/send", handlers.SendSMS)
	}

	email := r.Group("/email")
	email.Use(middleware.AuthMiddleware())
	{
		email.POST("/send", handlers.SendEmail)
	}

	whatsapp := r.Group("/whatsapp")
	whatsapp.Use(middleware.AuthMiddleware())
	{
		whatsapp.POST("/send", handlers.SendWhatsApp)
	}

	// Email inbound routes
	emailInbound := r.Group("/email-inbound")
	{
		emailInbound.POST("/webhook", handlers.EmailWebhook)
	}

	// CTI routes
	cti := r.Group("/cti")
	cti.Use(middleware.AuthMiddleware())
	{
		cti.POST("/call", handlers.MakeCall)
		cti.POST("/log", handlers.LogCall)
	}

	// Workflow routes
	workflow := r.Group("/workflow")
	workflow.Use(middleware.AuthMiddleware())
	{
		workflow.GET("", handlers.GetWorkflows)
		workflow.POST("", handlers.CreateWorkflow)
		workflow.GET("/:id", handlers.GetWorkflow)
		workflow.PUT("/:id", handlers.UpdateWorkflow)
		workflow.DELETE("/:id", handlers.DeleteWorkflow)
		workflow.POST("/:id/execute", handlers.ExecuteWorkflow)
	}

	// Integrations routes
	integrations := r.Group("/integrations")
	integrations.Use(middleware.AuthMiddleware())
	{
		integrations.GET("", handlers.GetIntegrations)
		integrations.POST("", handlers.CreateIntegration)
		integrations.PUT("/:id", handlers.UpdateIntegration)
		integrations.DELETE("/:id", handlers.DeleteIntegration)
	}

	// AI scoring routes - TODO: Implement when AI module is ready
	// aiRoutes := r.Group("/ai")
	// aiRoutes.Use(middleware.AuthMiddleware())
	// {
	// 	aiRoutes.POST("/score", ai.ScoreHandler)
	// 	aiRoutes.GET("/scores", ai.GetScoresHandler)
	// }

	// Sales marketing routes
	salesMarketing := r.Group("/sales-marketing")
	salesMarketing.Use(middleware.AuthMiddleware())
	{
		salesMarketing.GET("/tasks", handlers.GetSalesTasks)
		salesMarketing.POST("/tasks", handlers.CreateSalesTask)
		salesMarketing.GET("/forms", handlers.GetSalesForms)
		salesMarketing.POST("/forms", handlers.CreateSalesForm)
	}

	// Sales forms routes
	salesForms := r.Group("/sales-marketing/forms")
	salesForms.Use(middleware.AuthMiddleware())
	{
		salesForms.GET("", handlers.GetForms)
		salesForms.POST("", handlers.CreateForm)
		salesForms.GET("/:id", handlers.GetForm)
		salesForms.PUT("/:id", handlers.UpdateForm)
		salesForms.DELETE("/:id", handlers.DeleteForm)
	}

	// Opportunities routes
	opportunities := r.Group("/opportunities")
	opportunities.Use(middleware.AuthMiddleware())
	{
		opportunities.GET("", handlers.GetOpportunities)
		opportunities.POST("", handlers.CreateOpportunity)
		opportunities.GET("/:id", handlers.GetOpportunity)
		opportunities.PUT("/:id", handlers.UpdateOpportunity)
		opportunities.DELETE("/:id", handlers.DeleteOpportunity)
	}

	// Attachments routes
	attachments := r.Group("/attachments")
	attachments.Use(middleware.AuthMiddleware())
	{
		attachments.GET("", handlers.GetAttachments)
		attachments.POST("", handlers.CreateAttachment)
		attachments.DELETE("/:id", handlers.DeleteAttachment)
	}

	// Bulk uploads routes
	bulkUploads := r.Group("/bulk-uploads")
	bulkUploads.Use(middleware.AuthMiddleware())
	{
		bulkUploads.POST("", handlers.BulkUpload)
		bulkUploads.GET("/:id/status", handlers.GetUploadStatus)
	}

	// Field definitions routes
	fields := r.Group("/fields")
	fields.Use(middleware.AuthMiddleware())
	{
		fields.GET("", handlers.GetFields)
		fields.POST("", handlers.CreateField)
		fields.PUT("/:id", handlers.UpdateField)
		fields.DELETE("/:id", handlers.DeleteField)
	}

	// Maintenance routes
	maintenance := r.Group("/maintenance")
	maintenance.Use(middleware.AuthMiddleware())
	{
		maintenance.POST("/archive", handlers.ArchiveData)
		maintenance.POST("/cleanup", handlers.CleanupData)
	}

	// Teams routes
	teams := r.Group("/teams")
	teams.Use(middleware.AuthMiddleware())
	{
		teams.GET("", handlers.GetTeams)
		teams.POST("", handlers.CreateTeam)
		teams.GET("/:id", handlers.GetTeam)
		teams.PUT("/:id", handlers.UpdateTeam)
		teams.DELETE("/:id", handlers.DeleteTeam)
	}

	// Analytics routes
	analytics := r.Group("/analytics")
	analytics.Use(middleware.AuthMiddleware())
	{
		analytics.GET("/dashboard", handlers.GetDashboard)
		analytics.GET("/reports", handlers.GetReports)
	}

	// Users routes
	users := r.Group("/users")
	users.Use(middleware.AuthMiddleware())
	{
		users.GET("", handlers.GetUsers)
		users.GET("/:id", handlers.GetUser)
		users.PUT("/:id", handlers.UpdateUser)
		users.DELETE("/:id", handlers.DeleteUser)
	}

	// WebSocket routes
	ws := r.Group("/ws")
	{
		ws.GET("", handlers.WebSocketHandler())
	}
}
