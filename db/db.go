package db

import (
	_type "awesomeProject/type"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"os/exec"
)

var DB *gorm.DB

func runMigrations() {
	cmd := exec.Command("flyway", "migrate")
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to execute Flyway migrations: %v", err)
	}
	log.Println("Database migrations applied successfully.")
}

func IsSubscriptionExistsByEmail(email string) (bool, error) {
	var subscription _type.Subscription
	result := DB.Where("email = ?", email).First(&subscription)

	if result.Error == nil {
		return true, nil
	} else if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, result.Error
	}

	return false, nil
}

func Init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", host, port, user, dbname, pass)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database")
	}

	//runMigrations()
}
