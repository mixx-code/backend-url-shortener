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

	// Get filter parameters
	urlFilter := c.Query("url")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	period := c.DefaultQuery("period", "week") // week, month, year

	// Parse date filters
	var startTime, endTime time.Time
	var parseErr error

	if startDate != "" {
		startTime, parseErr = time.Parse("2006-01-02", startDate)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD"})
			return
		}
	} else {
		// Default to last 30 days if no start date provided
		startTime = time.Now().AddDate(0, 0, -30)
	}

	if endDate != "" {
		endTime, parseErr = time.Parse("2006-01-02", endDate)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD"})
			return
		}
		// Set end time to end of day
		endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, endTime.Location())
	} else {
		// Default to now if no end date provided
		endTime = time.Now()
	}

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

	// Calculate total clicks with filters (only user's URLs)
	var totalClicks int64
	clickQuery := models.DB.Table("clicks").
		Joins("JOIN urls ON clicks.url_id = urls.id").
		Where("urls.user_id = ?", userID).
		Where("clicks.clicked_at >= ?", startTime).
		Where("clicks.clicked_at <= ?", endTime)

	// Apply URL filter if provided
	if urlFilter != "" {
		clickQuery = clickQuery.Where("urls.short_code = ?", urlFilter)
	}

	clickQuery.Count(&totalClicks)

	// Generate time-based click data based on period and date range
	var timeBasedClicks []gin.H
	var timeLabels []string

	switch period {
	case "day":
		// Last 24 hours from date range
		timeBasedClicks = make([]gin.H, 24)
		timeLabels = make([]string, 24)
		for i := 0; i < 24; i++ {
			hour := startTime.Add(time.Duration(i) * time.Hour)
			timeLabels[i] = hour.Format("15:04")
			timeBasedClicks[i] = gin.H{
				"time":   timeLabels[i],
				"clicks": 0,
			}
		}
	case "week":
		// Use the actual date range provided by user
		days := int(endTime.Sub(startTime).Hours()/24) + 1
		if days > 365 { // Limit to prevent memory issues
			days = 365
		}
		timeBasedClicks = make([]gin.H, days)
		timeLabels = make([]string, days)
		for i := 0; i < days; i++ {
			date := startTime.AddDate(0, 0, i)
			timeLabels[i] = date.Format("2006-01-02")
			timeBasedClicks[i] = gin.H{
				"time":   timeLabels[i],
				"clicks": 0,
			}
		}
	case "month":
		// Use the actual date range provided by user
		days := int(endTime.Sub(startTime).Hours()/24) + 1
		if days > 365 { // Limit to prevent memory issues
			days = 365
		}
		timeBasedClicks = make([]gin.H, days)
		timeLabels = make([]string, days)
		for i := 0; i < days; i++ {
			date := startTime.AddDate(0, 0, i)
			timeLabels[i] = date.Format("2006-01-02")
			timeBasedClicks[i] = gin.H{
				"time":   timeLabels[i],
				"clicks": 0,
			}
		}
	case "year":
		// Last 12 months
		timeBasedClicks = make([]gin.H, 12)
		timeLabels = make([]string, 12)
		for i := 0; i < 12; i++ {
			month := startTime.AddDate(0, i, 0)
			if month.After(endTime) {
				break
			}
			timeLabels[i] = month.Format("2006-01")
			timeBasedClicks[i] = gin.H{
				"time":   timeLabels[i],
				"clicks": 0,
			}
		}
	default:
		// Default to week - use date range
		days := int(endTime.Sub(startTime).Hours()/24) + 1
		if days > 365 {
			days = 365
		}
		timeBasedClicks = make([]gin.H, days)
		timeLabels = make([]string, days)
		for i := 0; i < days; i++ {
			date := startTime.AddDate(0, 0, i)
			timeLabels[i] = date.Format("2006-01-02")
			timeBasedClicks[i] = gin.H{
				"time":   timeLabels[i],
				"clicks": 0,
			}
		}
	}

	// Get actual click data for the specified period (only user's URLs)
	var clicks []models.Click
	clickDataQuery := models.DB.Table("clicks").
		Joins("JOIN urls ON clicks.url_id = urls.id").
		Where("urls.user_id = ?", userID).
		Where("clicks.clicked_at >= ?", startTime).
		Where("clicks.clicked_at <= ?", endTime)

	// Apply URL filter if provided
	if urlFilter != "" {
		clickDataQuery = clickDataQuery.Where("urls.short_code = ?", urlFilter)
	}

	clickDataQuery.Find(&clicks)

	// Count clicks per time period
	for _, click := range clicks {
		var timeKey string
		var timeIndex int

		switch period {
		case "day":
			timeKey = click.ClickedAt.Format("15:04")
			for i, label := range timeLabels {
				if label == timeKey {
					timeIndex = i
					break
				}
			}
		case "week", "month":
			timeKey = click.ClickedAt.Format("2006-01-02")
			for i, label := range timeLabels {
				if label == timeKey {
					timeIndex = i
					break
				}
			}
		case "year":
			timeKey = click.ClickedAt.Format("2006-01")
			for i, label := range timeLabels {
				if label == timeKey {
					timeIndex = i
					break
				}
			}
		}

		if timeIndex < len(timeBasedClicks) {
			currentClicks := timeBasedClicks[timeIndex]["clicks"].(int)
			timeBasedClicks[timeIndex] = gin.H{
				"time":   timeLabels[timeIndex],
				"clicks": currentClicks + 1,
			}
		}
	}

	// Prepare URL stats with filters
	urlStats := make([]gin.H, len(urls))
	for i, url := range urls {
		// Get actual click count from clicks table for each URL within date range
		var clickCount int64
		urlClickQuery := models.DB.Table("clicks").Where("url_id = ?", url.ID).Where("clicked_at >= ?", startTime).Where("clicked_at <= ?", endTime)
		urlClickQuery.Count(&clickCount)

		urlStats[i] = gin.H{
			"id":           url.ID,
			"short_code":   url.ShortCode,
			"original_url": url.OriginalURL,
			"short_url":    "https://electric-hideously-drake.ngrok-free.app/" + url.ShortCode,
			"click_count":  int(clickCount),
			"created_at":   url.CreatedAt,
		}
	}

	// Determine response data key based on period
	var dataKey string
	switch period {
	case "day":
		dataKey = "hourlyClicks"
	case "week":
		dataKey = "dailyClicks"
	case "month":
		dataKey = "dailyClicks"
	case "year":
		dataKey = "monthlyClicks"
	default:
		dataKey = "dailyClicks"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Analytics retrieved successfully",
		"filters": gin.H{
			"url":        urlFilter,
			"start_date": startTime.Format("2006-01-02"),
			"end_date":   endTime.Format("2006-01-02"),
			"period":     period,
		},
		"data": gin.H{
			"totalClicks": totalClicks,
			dataKey:       timeBasedClicks,
			"urlStats":    urlStats,
		},
	})
}
