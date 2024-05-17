package rest

import (
	"awesomeProject/db"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"net/mail"
	"os"
	"strconv"
)

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
		log.Printf("Error fetching usd/uah rate: %v", err)
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

	result := db.DB.Where("email = ?", createRequest.Email).First(&user)

	if result.Error == nil {
		c.IndentedJSON(http.StatusConflict, gin.H{"error": "This email is already registered"})
		return
	} else if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Println("Database error")
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}

	user.Email = createRequest.Email

	if err := db.DB.Create(&user).Error; err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "E-mail added"})
}

func sendEmails() error {
	usdRate, err := fetchUSDExchangeRate()
	if err != nil {
		log.Printf("Error fetching USD exchange rate: %v", err)
		return err
	}

	usdRateString := fmt.Sprintf("%v", usdRate)

	var allRegisteredUsers []User

	result := db.DB.Find(&allRegisteredUsers)

	if result.Error != nil {
		log.Println("Database error")
		return result.Error
	}

	var allEmails []string

	for _, value := range allRegisteredUsers {
		allEmails = append(allEmails, value.Email)
	}

	if len(allEmails) == 0 {
		log.Println("No registered users found")
		return nil
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
		log.Printf("Invalid SMTP port: %v", err)
		return nil
	}
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUsername, smtpPassword)

	if err := d.DialAndSend(m); err != nil {
		log.Fatalf("Error sending email: %v", err)
		return err
	}

	return nil
}

func sendEmailsHandler(c *gin.Context) {
	if err := sendEmails(); err != nil {
		log.Println("Error sending emails")
		c.IndentedJSON(http.StatusInternalServerError, nil)
		return
	}
	c.IndentedJSON(http.StatusOK, "Emails sent successfully")
}

func setupDailyEmails() {
	c := cron.New()
	//Schedule the sendEmails function to run at 8 AM every day
	_, err := c.AddFunc("0 8 * * *", func() {
		if err := sendEmails(); err != nil {
			log.Fatalf("Error in scheduled email sending: %v", err)
		}
	})

	if err != nil {
		log.Fatalf("Error scheduling cron job: %v", err)
	}

	c.Start()
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	db.Init()

	controller := gin.New()
	controller.GET("/rate", GetRate)
	controller.POST("/subscribe", addSubscription)
	controller.POST("/sendEmails", sendEmailsHandler)

	setupDailyEmails()

	if err := controller.Run("localhost:8080"); err != nil {
		log.Fatalf("Sever run error")
	}
}
