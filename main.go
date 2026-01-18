package main

import (
	"backend-go/controllers"
	"backend-go/middlewares"
	"backend-go/models"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// CORS Middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3001", "http://127.0.0.1:3001"}, // Frontend URLs
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	models.ConnectDB()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// URL Shortener Routes
	api := r.Group("/api")
	{
		api.POST("/register", controllers.Register)
		api.POST("/login", controllers.Login)

		protected := api.Group("/")
		protected.Use(middlewares.AuthMiddleware())
		{
			protected.POST("/shorten", controllers.CreateShortURL)
			protected.GET("/urls", controllers.GetURLs)
			protected.GET("/stats/:shortCode", controllers.GetURLStats)
			protected.PUT("/urls/:id", controllers.UpdateURL)
			protected.DELETE("/urls/:id", controllers.DeleteURL)
			protected.POST("/change-password", controllers.ChangePassword)
			protected.GET("/analytics", controllers.GetAnalytics)
		}
	}

	// Redirect route
	r.GET("/:shortCode", controllers.RedirectURL)

	r.Run(":3000") // listen and serve on 0.0.0.0:3000 (for windows "localhost:3000")
}
