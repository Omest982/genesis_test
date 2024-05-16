package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"os/exec"
	//"errors"
)

var db *gorm.DB

func dbConnect() *gorm.DB {
	dsn := "host=localhost user=postgres dbname=genesis password=root sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database")
		return nil
	}

	return db
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

	var usdRate = rates[0].Rate

	c.IndentedJSON(http.StatusOK, usdRate)
}

func addSubscription(c *gin.Context) {

}

func sendEmails(c *gin.Context) {

}

func main() {
	runMigrations()
	//db := dbConnect()

	controller := gin.New()
	controller.GET("/rate", GetRate)
	controller.POST("/subscribe", addSubscription)
	controller.POST("/sendEmails", sendEmails)
	controller.Run("localhost:8080")
}
