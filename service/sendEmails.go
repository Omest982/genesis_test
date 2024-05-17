package service

import (
	"awesomeProject/db"
	_type "awesomeProject/type"
	"fmt"
	"gopkg.in/gomail.v2"
	"log"
	"os"
	"strconv"
)

func SendEmails() error {
	usdRate, err := FetchUSDExchangeRate()
	if err != nil {
		log.Printf("Error fetching USD exchange rate: %v", err)
		return err
	}

	usdRateString := fmt.Sprintf("%v", usdRate)

	var allRegisteredUsers []_type.Subscription

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
