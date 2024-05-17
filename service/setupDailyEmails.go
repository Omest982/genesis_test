package service

import (
	"github.com/robfig/cron/v3"
	"log"
)

func SetupDailyEmails() {
	c := cron.New()
	//Schedule the sendEmails function to run at 8 AM every day
	_, err := c.AddFunc("0 8 * * *", func() {
		if err := SendEmails(); err != nil {
			log.Fatalf("Error in scheduled email sending: %v", err)
		}
	})

	if err != nil {
		log.Fatalf("Error scheduling cron job: %v", err)
	}

	c.Start()
}
