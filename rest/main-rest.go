package rest

import (
	"awesomeProject/db"
	"awesomeProject/service"
	_type "awesomeProject/type"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"log"
	"net/http"
)

func getRate(c *gin.Context) {

	usdRate, err := service.FetchUSDExchangeRate()
	if err != nil {
		log.Printf("Error fetching usd/uah rate: %v", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"usd_rate": usdRate})
}

func addSubscription(c *gin.Context) {
	var createRequest _type.SubscriptionCreateDto

	if err := c.BindJSON(&createRequest); err != nil {
		return
	}

	if !service.IsEmailValid(createRequest.Email) {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Enter valid email"})
	}

	var subscription _type.Subscription

	result := db.DB.Where("email = ?", createRequest.Email).First(&subscription)

	if result.Error == nil {
		c.IndentedJSON(http.StatusConflict, gin.H{"error": "This email is already registered"})
		return
	} else if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Println("Database error")
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}

	subscription.Email = createRequest.Email

	if err := db.DB.Create(&subscription).Error; err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "E-mail added"})
}

func sendEmailsHandler(c *gin.Context) {
	if err := service.SendEmails(); err != nil {
		log.Println("Error sending emails")
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}
	c.IndentedJSON(http.StatusOK, "Emails sent successfully")
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	db.Init()

	controller := gin.New()
	controller.GET("/rate", getRate)
	controller.POST("/subscribe", addSubscription)
	controller.POST("/sendEmails", sendEmailsHandler)

	service.SetupDailyEmails()

	if err := controller.Run("localhost:8080"); err != nil {
		log.Fatalf("Sever run error")
	}
}