package main

import (
	"awesomeProject/db"
	"awesomeProject/db/dbService"
	"awesomeProject/db/repository/repositoryImpl"
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

type MyEmailServer struct {
	emailService.UnimplementedEmailServiceServer
	SubscriptionService *dbService.SubscriptionService
}

func (s MyEmailServer) Rate(context.Context, *emptypb.Empty) (*emailService.RateResponse, error) {
	usdRate, err := service.FetchUSDExchangeRate()
	if err != nil {
		log.Printf("Error fetching usd/uah rate: %s", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emailService.RateResponse{UsdRate: usdRate}, nil
}

func (s MyEmailServer) AddSubscription(_ context.Context, createRequest *emailService.CreateSubscription) (*emptypb.Empty, error) {
	if !service.IsEmailValid(createRequest.Email) {
		log.Println("invalid email")
		return nil, status.Error(codes.InvalidArgument, "invalid mail")
	}

	isSubscriptionExistsByEmail, dbError := s.SubscriptionService.IsSubscriptionExistsByEmail(createRequest.Email)

	if dbError != nil {
		return nil, status.Error(codes.Internal, dbError.Error())
	}

	if isSubscriptionExistsByEmail {
		log.Println("email already exists")
		return nil, status.Error(codes.AlreadyExists, "email already subscribed")
	}

	var subscription _type.Subscription

	subscription.Email = createRequest.Email

	if err := s.SubscriptionService.AddSubscription(createRequest); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s MyEmailServer) SendEmails(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
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

	subscriptionRepo := &repositoryImpl.SubscriptionRepositoryImpl{DB: db.DB}
	subscriptionService := &dbService.SubscriptionService{Repo: subscriptionRepo}

	serverRegistrar := grpc.NewServer()
	myEmailService := &MyEmailServer{
		SubscriptionService: subscriptionService,
	}

	emailService.RegisterEmailServiceServer(serverRegistrar, myEmailService)
	service.SetupDailyEmails()

	log.Println("gRPC server listening on port 8080")
	if err := serverRegistrar.Serve(listener); err != nil {
		log.Fatalf("Impossible to serve: %s", err)
	}
}
