package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"net/mail"
	"os"
	"os/exec"
	"strconv"
)

var db *gorm.DB

func initDB() {

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=disable", host, user, dbname, pass)

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

func fetchUSDExchangeRate() (float64, error) {
	resp, err := http.Get("https://bank.gov.ua/NBUStatService/v1/statdirectory/dollar_info?json")

	if err != nil {
		return 0, fmt.Errorf("error fetching exchange rate data: %v", err)
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response body: %v", err)
	}

	var rates []ExchangeRate
	err = json.Unmarshal(data, &rates)
	if err != nil {
		return 0, fmt.Errorf("error unmarshalling JSON data: %v", err)
	}

	if len(rates) > 0 {
		return rates[0].Rate, nil
	}

	return 0, fmt.Errorf("no exchange rate data available")
}

func GetRate(c *gin.Context) {

	usdRate, err := fetchUSDExchangeRate()
	if err != nil {
		log.Printf("Error: %v", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"usd_rate": usdRate})
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
	} else if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Println("Database error")
		c.IndentedJSON(http.StatusInternalServerError, nil)
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
	usdRate, err := fetchUSDExchangeRate()
	if err != nil {
		log.Printf("Error: %v", err)
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}

	usdRateString := fmt.Sprintf("%v", usdRate)

	var allRegisteredUsers []User

	result := db.Find(&allRegisteredUsers)

	if result.Error != nil {
		log.Println("Database error")
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}

	var allEmails []string

	for _, value := range allRegisteredUsers {
		allEmails = append(allEmails, value.Email)
	}

	if len(allEmails) == 0 {
		log.Println("No registered users found")
		c.IndentedJSON(http.StatusOK, "No registered users to send emails to")
		return
	}

	smtpUsername := os.Getenv("SMTP_USERNAME")

	m := gomail.NewMessage()
	m.SetHeader("From", smtpUsername)
	m.SetHeader("To", allEmails...)
	m.SetHeader("Subject", "Today's usd/uah rate")
	m.SetBody("text/plain", usdRateString)

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		log.Printf("Error: %v", err)
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUsername, smtpPassword)

	//d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		log.Fatalf("Error sending email: %v", err)
	}

	c.IndentedJSON(http.StatusOK, "E-mails send")

}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	initDB()
	runMigrations()

	controller := gin.New()
	controller.GET("/rate", GetRate)
	controller.POST("/subscribe", addSubscription)
	controller.POST("/sendEmails", sendEmails)
	controller.Run("localhost:8080")
}
