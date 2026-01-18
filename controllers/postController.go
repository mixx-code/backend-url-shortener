package controllers

import (
	"backend-go/models"

	"github.com/gin-gonic/gin"
)

func FindPosts(c *gin.Context) {
	var posts []models.Post
	models.DB.Find(&posts)
	c.JSON(200, gin.H{
		"status":  true,
		"message": "Posts retrieved successfully",
		"posts":   posts,
	})
}
