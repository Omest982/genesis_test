package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"net/mail"
	"os/exec"
	//"errors"
)

var db *gorm.DB

func initDB() {
	dsn := "host=localhost user=postgres dbname=genesis password=root sslmode=disable"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database")
	}
}

func runMigrations() {
	cmd := exec.Command("flyway", "migrate")
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to execute Flyway migrations: %v", err)
	}
	log.Println("Database migrations applied successfully.")
}

type ExchangeRate struct {
	R030         int     `json:"r030"`
	Txt          string  `json:"txt"`
	Rate         float64 `json:"rate"`
	Cc           string  `json:"cc"`
	ExchangeDate string  `json:"exchangedate"`
}

type UserCreateDto struct {
	Email string `json:"email"`
}

type User struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

func GetRate(c *gin.Context) {

	resp, err := http.Get("https://bank.gov.ua/NBUStatService/v1/statdirectory/dollar_info?json")

	if err != nil {
		log.Printf("Error fetching exchange rate data: %v", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch exchange rate data"})
		return
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Printf("Error reading response body: %v", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Unable to read response body"})
		return
	}

	var rates []ExchangeRate
	err = json.Unmarshal(data, &rates)

	if err != nil {
		log.Printf("Error unmarshalling JSON data: %v", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Unable to parse exchange rate data"})
		return
	}

	if len(rates) > 0 {
		usdRate := rates[0].Rate
		c.IndentedJSON(http.StatusOK, usdRate)
	} else {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "No exchange rate data available"})
	}
}

func addSubscription(c *gin.Context) {
	var createRequest UserCreateDto

	if err := c.BindJSON(&createRequest); err != nil {
		return
	}

	if _, err := mail.ParseAddress(createRequest.Email); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	var user User

	result := db.Where("email = ?", createRequest.Email).First(&user)

	if result.Error == nil {
		c.IndentedJSON(http.StatusConflict, gin.H{"error": "This email is already registered"})
		return
	} else if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	user.Email = createRequest.Email

	if err := db.Create(&user).Error; err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "E-mail added"})
}

func sendEmails(c *gin.Context) {

}

func main() {
	initDB()
	runMigrations()

	controller := gin.New()
	controller.GET("/rate", GetRate)
	controller.POST("/subscribe", addSubscription)
	controller.POST("/sendEmails", sendEmails)
	controller.Run("localhost:8080")
}
