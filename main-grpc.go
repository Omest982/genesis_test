package main

import (
	"awesomeProject/db"
	"awesomeProject/emailService"
	"awesomeProject/service"
	_type "awesomeProject/type"
	"context"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net"
	"os"
)

type myEmailServer struct {
	emailService.UnimplementedEmailServiceServer
}

func (s myEmailServer) Rate(context.Context, *emptypb.Empty) (*emailService.RateResponse, error) {
	usdRate, err := service.FetchUSDExchangeRate()
	if err != nil {
		log.Printf("Error fetching usd/uah rate: %s", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emailService.RateResponse{UsdRate: usdRate}, nil
}

func (s myEmailServer) AddSubscription(_ context.Context, createRequest *emailService.CreateSubscription) (*emptypb.Empty, error) {
	if !service.IsEmailValid(createRequest.Email) {
		log.Println("invalid email")
		return nil, status.Error(codes.InvalidArgument, "invalid mail")
	}

	var subscription _type.Subscription

	isSubscriptionExistsByEmail, dbError := db.IsSubscriptionExistsByEmail(createRequest.Email)

	if dbError != nil {
		return nil, status.Error(codes.Internal, dbError.Error())
	}

	if isSubscriptionExistsByEmail {
		log.Println("email already exists")
		return nil, status.Error(codes.AlreadyExists, "email already subscribed")
	}

	subscription.Email = createRequest.Email

	if err := db.DB.Create(&subscription).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s myEmailServer) SendEmails(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	if err := service.SendEmails(); err != nil {
		log.Println("Error sending emails")
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	serverPort := os.Getenv("SERVER_PORT")

	listener, err := net.Listen("tcp", ":"+serverPort)
	if err != nil {
		log.Fatalf("Cannot create listener: %s", err)
	}

	db.Init()

	serverRegistrar := grpc.NewServer()
	myEmailService := &myEmailServer{}

	emailService.RegisterEmailServiceServer(serverRegistrar, myEmailService)
	service.SetupDailyEmails()

	log.Println("gRPC server listening on port 8080")
	if err := serverRegistrar.Serve(listener); err != nil {
		log.Fatalf("Impossible to serve: %s", err)
	}
}
