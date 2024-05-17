package main

import (
	"awesomeProject/db"
	"awesomeProject/emailService"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net"
)

type myEmailServer struct {
	emailService.UnimplementedEmailServiceServer
}

func (s myEmailServer) Rate(context.Context, *emptypb.Empty) (*emailService.RateResponse, error) {
	usdRate, err := fetchUSDExchangeRate()
	if err != nil {
		log.Printf("Error fetching usd/uah rate: %s", err)
		return nil, err
	}

	return &emailService.RateResponse{UsdRate: usdRate}, nil
}

func (s myEmailServer) AddSubscription(context.Context, *emailService.CreateSubscription) (*emptypb.Empty, error) {

	return
}

func (s myEmailServer) SendEmails(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {

	return
}

type ExchangeRate struct {
	R030         int     `json:"r030"`
	Txt          string  `json:"txt"`
	Rate         float64 `json:"rate"`
	Cc           string  `json:"cc"`
	ExchangeDate string  `json:"exchangedate"`
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Cannot create listener: %s", err)
	}

	db.Init()

	serverRegistrar := grpc.NewServer()
	service := &myEmailServer{}

	emailService.RegisterEmailServiceServer(serverRegistrar, service)
	if err := serverRegistrar.Serve(listener); err != nil {
		log.Fatalf("Impossible to serve: %s", err)
	}

	fmt.Println("Program started!")
}
