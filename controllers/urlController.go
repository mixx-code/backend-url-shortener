package controllers

import (
	"backend-go/models"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type CreateURLRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url"`
	CustomCode  string `json:"custom_code"`
}

func CreateShortURL(c *gin.Context) {
	var req CreateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Invalid URL format",
			"error":   err.Error(),
		})
		return
	}

	// Get user ID from token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Use the same JWT key as authController
		return []byte("Kepo_banget_lo"), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	userID := uint(claims["user_id"].(float64))

	var shortCode string

	// Use custom code if provided, otherwise generate random
	if req.CustomCode != "" {
		// Check if custom code already exists
		var existingURL models.URL
		if err := models.DB.Where("short_code = ?", req.CustomCode).First(&existingURL).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{
				"status":  false,
				"message": "Custom short code already exists",
			})
			return
		}
		shortCode = req.CustomCode
	} else {
		shortCode = models.GenerateShortCode()
	}

	url := models.URL{
		OriginalURL: req.OriginalURL,
		ShortCode:   shortCode,
		UserID:      int(userID),
	}

	if err := models.DB.Create(&url).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Failed to create short URL",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  true,
		"message": "Short URL created successfully",
		"data": gin.H{
			"id":           url.ID,
			"original_url": url.OriginalURL,
			"short_code":   url.ShortCode,
			"short_url":    "https://electric-hideously-drake.ngrok-free.app/" + url.ShortCode,
			"click_count":  url.ClickCount,
			"created_at":   url.CreatedAt,
		},
	})
}

func GetURLs(c *gin.Context) {
	// Get pagination parameters from query string
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "5")

	// Convert to integers
	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	limitNum, err := strconv.Atoi(limit)
	if err != nil || limitNum < 1 || limitNum > 100 {
		limitNum = 10
	}

	// Calculate offset
	offset := (pageNum - 1) * limitNum

	// Get user ID from token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Use the same JWT key as authController
		return []byte("Kepo_banget_lo"), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	userID := uint(claims["user_id"].(float64))

	// Get total count for pagination info (filtered by user)
	var totalCount int64
	models.DB.Model(&models.URL{}).Where("user_id = ?", userID).Count(&totalCount)

	// Get URLs with pagination (filtered by user)
	var urls []models.URL
	result := models.DB.Where("user_id = ?", userID).Offset(offset).Limit(limitNum).Order("created_at DESC").Find(&urls)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Failed to retrieve URLs",
		})
		return
	}

	// Calculate click counts from clicks table for each URL
	for i, url := range urls {
		var clickCount int64
		models.DB.Table("clicks").Where("url_id = ?", url.ID).Count(&clickCount)
		urls[i].ClickCount = int(clickCount)
	}

	// Calculate pagination info
	totalPages := (totalCount + int64(limitNum) - 1) / int64(limitNum)

	// Build URLs with full short_url
	urlsWithFullURL := make([]gin.H, len(urls))
	for i, url := range urls {
		urlsWithFullURL[i] = gin.H{
			"id":           url.ID,
			"original_url": url.OriginalURL,
			"short_code":   url.ShortCode,
			"short_url":    "https://electric-hideously-drake.ngrok-free.app/" + url.ShortCode,
			"click_count":  url.ClickCount,
			"created_at":   url.CreatedAt,
			"updated_at":   url.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "URLs retrieved successfully",
		"data": gin.H{
			"urls": urlsWithFullURL,
			"pagination": gin.H{
				"current_page": pageNum,
				"per_page":     limitNum,
				"total":        totalCount,
				"total_pages":  totalPages,
				"has_next":     pageNum < int(totalPages),
				"has_prev":     pageNum > 1,
			},
		},
	})
}

func RedirectURL(c *gin.Context) {
	shortCode := c.Param("shortCode")

	var url models.URL
	if err := models.DB.Where("short_code = ?", shortCode).First(&url).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "Short URL not found",
		})
		return
	}

	// Track the click
	click := models.Click{
		URLID:     url.ID,
		ClickedAt: time.Now(),
	}

	// Save click record
	if err := models.DB.Create(&click).Error; err != nil {
		// Log error but don't block redirect
		fmt.Printf("Failed to track click: %v", err)
	}

	// Increment click count in URL table
	models.DB.Model(&url).Update("click_count", url.ClickCount+1)

	c.Redirect(http.StatusMovedPermanently, url.OriginalURL)
}

func GetURLStats(c *gin.Context) {
	shortCode := c.Param("shortCode")

	var url models.URL
	if err := models.DB.Where("short_code = ?", shortCode).First(&url).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "Short URL not found",
		})
		return
	}

	// Get actual click count from clicks table
	var clickCount int64
	models.DB.Table("clicks").Where("url_id = ?", url.ID).Count(&clickCount)

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "URL stats retrieved successfully",
		"data": gin.H{
			"id":           url.ID,
			"original_url": url.OriginalURL,
			"short_code":   url.ShortCode,
			"short_url":    "https://electric-hideously-drake.ngrok-free.app/" + url.ShortCode,
			"click_count":  int(clickCount),
			"created_at":   url.CreatedAt,
			"updated_at":   url.UpdatedAt,
		},
	})
}

func UpdateURL(c *gin.Context) {
	id := c.Param("id")

	var url models.URL
	if err := models.DB.First(&url, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "URL not found",
		})
		return
	}

	var input struct {
		OriginalURL string `json:"original_url" binding:"required"`
		ShortCode   string `json:"short_code"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	// Check if new short_code is already taken by another URL
	if input.ShortCode != "" && input.ShortCode != url.ShortCode {
		var existingURL models.URL
		if err := models.DB.Where("short_code = ?", input.ShortCode).First(&existingURL).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{
				"status":  false,
				"message": "Short code already exists",
			})
			return
		}
	}

	// Update URL
	url.OriginalURL = input.OriginalURL
	if input.ShortCode != "" {
		url.ShortCode = input.ShortCode
	}

	if err := models.DB.Save(&url).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Failed to update URL",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "URL updated successfully",
		"data": gin.H{
			"id":           url.ID,
			"original_url": url.OriginalURL,
			"short_code":   url.ShortCode,
			"short_url":    "https://electric-hideously-drake.ngrok-free.app/" + url.ShortCode,
			"click_count":  url.ClickCount,
			"created_at":   url.CreatedAt,
			"updated_at":   url.UpdatedAt,
		},
	})
}

func DeleteURL(c *gin.Context) {
	id := c.Param("id")

	var url models.URL
	if err := models.DB.First(&url, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "URL not found",
		})
		return
	}

	if err := models.DB.Delete(&url).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Failed to delete URL",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "URL deleted successfully",
	})
}

func GetAnalytics(c *gin.Context) {
	// Get user ID from token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Use the same JWT key as authController
		return []byte("Kepo_banget_lo"), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	userID := uint(claims["user_id"].(float64))

	// Get URL filter parameter
	urlFilter := c.Query("url")

	// Get user's URLs
	var urls []models.URL
	query := models.DB.Where("user_id = ?", userID)

	// Apply URL filter if provided
	if urlFilter != "" {
		query = query.Where("short_code = ?", urlFilter)
	}

	if err := query.Find(&urls).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Failed to retrieve URLs for analytics",
		})
		return
	}

	// Calculate total clicks from clicks table
	var totalClicks int64
	models.DB.Table("clicks").Count(&totalClicks)

	// Generate daily click data (last 7 days - Monday to Sunday)
	dailyClicks := make([]gin.H, 7)

	// Get current date and find Monday
	now := time.Now()
	currentWeekday := int(now.Weekday())
	mondayOffset := (currentWeekday + 6) % 7 // Calculate days since Monday

	for i := 0; i < 7; i++ {
		date := now.AddDate(0, 0, -mondayOffset+i)
		dateStr := date.Format("2006-01-02")
		dailyClicks[i] = gin.H{
			"date":   dateStr,
			"clicks": 0,
		}
	}

	// Get actual click data for the week
	var clicks []models.Click
	query = models.DB.Where("clicked_at >= ?", now.AddDate(0, 0, -mondayOffset).Format("2006-01-02")+" 00:00:00")
	query = query.Where("clicked_at <= ?", now.AddDate(0, 0, -mondayOffset+6).Format("2006-01-02")+" 23:59:59")

	// Apply URL filter if provided
	if urlFilter != "" {
		var url models.URL
		models.DB.Where("short_code = ?", urlFilter).First(&url)
		query = query.Where("url_id = ?", url.ID)
	}

	query.Find(&clicks)

	// Count clicks per day
	for _, click := range clicks {
		clickDate := click.ClickedAt.Format("2006-01-02")
		for i, day := range dailyClicks {
			if day["date"].(string) == clickDate {
				currentClicks := day["clicks"].(int)
				dailyClicks[i] = gin.H{
					"date":   clickDate,
					"clicks": currentClicks + 1,
				}
				break
			}
		}
	}

	// Generate monthly click data (last 12 months)
	monthlyClicks := make([]gin.H, 12)

	// Create map to store clicks per month
	clicksPerMonth := make(map[string]int)

	// Initialize all 12 months
	for i := 0; i < 12; i++ {
		// Calculate the month (going backwards from current month)
		targetDate := now.AddDate(0, -i, 0)
		monthKey := targetDate.Format("2006-01") // Format: YYYY-MM
		monthNum := int(targetDate.Month())

		clicksPerMonth[monthKey] = 0

		// Store in reverse order so index 0 is oldest month, index 11 is current month
		monthlyClicks[11-i] = gin.H{
			"month":  fmt.Sprintf("%d", monthNum),
			"clicks": 0,
		}
	}

	// Get actual click data for last 12 months
	var monthlyClickData []models.Click
	// Start from beginning of month 12 months ago
	startOfPeriod := now.AddDate(0, -11, 0)
	startOfPeriod = time.Date(startOfPeriod.Year(), startOfPeriod.Month(), 1, 0, 0, 0, 0, startOfPeriod.Location())

	monthQuery := models.DB.Where("clicked_at >= ?", startOfPeriod)
	monthQuery = monthQuery.Where("clicked_at <= ?", now)

	// Apply URL filter if provided
	if urlFilter != "" {
		var url models.URL
		models.DB.Where("short_code = ?", urlFilter).First(&url)
		monthQuery = monthQuery.Where("url_id = ?", url.ID)
	}

	monthQuery.Find(&monthlyClickData)

	// Count clicks per month
	for _, click := range monthlyClickData {
		monthKey := click.ClickedAt.Format("2006-01")
		if _, exists := clicksPerMonth[monthKey]; exists {
			clicksPerMonth[monthKey]++
		}
	}

	// Update monthlyClicks array with actual counts
	for i := 0; i < 12; i++ {
		targetDate := now.AddDate(0, -(11 - i), 0)
		monthKey := targetDate.Format("2006-01")
		monthNum := int(targetDate.Month())

		monthlyClicks[i] = gin.H{
			"month":  fmt.Sprintf("%d", monthNum),
			"clicks": clicksPerMonth[monthKey],
		}
	}
	// Prepare URL stats
	urlStats := make([]gin.H, len(urls))
	for i, url := range urls {
		// Get actual click count from clicks table for each URL
		var clickCount int64
		models.DB.Table("clicks").Where("url_id = ?", url.ID).Count(&clickCount)

		urlStats[i] = gin.H{
			"id":           url.ID,
			"short_code":   url.ShortCode,
			"original_url": url.OriginalURL,
			"short_url":    "https://electric-hideously-drake.ngrok-free.app/" + url.ShortCode,
			"click_count":  int(clickCount),
			"created_at":   url.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Analytics retrieved successfully",
		"data": gin.H{
			"totalClicks":   totalClicks,
			"dailyClicks":   dailyClicks,
			"monthlyClicks": monthlyClicks,
			"urlStats":      urlStats,
		},
	})
}
