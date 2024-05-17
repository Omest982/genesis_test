package service

import (
	_type "awesomeProject/type"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func FetchUSDExchangeRate() (float64, error) {
	resp, err := http.Get("https://bank.gov.ua/NBUStatService/v1/statdirectory/dollar_info?json")

	if err != nil {
		return 0, fmt.Errorf("error fetching exchange rate data: %v", err)
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response body: %v", err)
	}

	var rates []_type.ExchangeRate
	err = json.Unmarshal(data, &rates)
	if err != nil {
		return 0, fmt.Errorf("error unmarshalling JSON data: %v", err)
	}

	if len(rates) > 0 {
		return rates[0].Rate, nil
	}

	return 0, fmt.Errorf("no exchange rate data available")
}
