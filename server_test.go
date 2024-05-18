package main

import (
	"awesomeProject/db/dbService"
	"awesomeProject/mocks"
	"awesomeProject/service"
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net"
	"testing"

	pb "awesomeProject/emailService"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

var ctx = context.Background()
var mockRepo = new(mocks.MockSubscriptionRepository)

// bufconnDialer is a buffer-based Dialer to mock network connections
func bufconnDialer(ctx context.Context, _ string) (net.Conn, error) {
	return lis.DialContext(ctx)
}

func getGrpcClient() (pb.EmailServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufconnDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, nil, err
	}

	return pb.NewEmailServiceClient(conn), conn, nil
}

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(1024 * 1024)

	s := grpc.NewServer()

	subscriptionService := &dbService.SubscriptionService{Repo: mockRepo}
	emailServer := MyEmailServer{SubscriptionService: subscriptionService}

	pb.RegisterEmailServiceServer(s, emailServer)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func TestUsdRate_success(t *testing.T) {
	client, conn, err := getGrpcClient()
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	req := &emptypb.Empty{}
	res, err := client.Rate(ctx, req)
	if err != nil {
		t.Fatalf("Rate method failed: %v", err)
	}

	usdRate, err := service.FetchUSDExchangeRate()
	if err != nil {
		t.Errorf("Error fetching usd rate: %s", err)
	}

	// Assert your response
	if res.UsdRate != usdRate {
		t.Errorf("Expected SomeField to be %v, got %v", "expected value", res.UsdRate)
	}
}

func TestSubscriptionCreation_success(t *testing.T) {
	client, conn, err := getGrpcClient()
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	req := &pb.CreateSubscription{Email: "test@gmail.com"}

	mockRepo.On("IsSubscriptionExistsByEmail", "test@gmail.com").Return(false, nil)
	mockRepo.On("CreateSubscription", mock.Anything).Return(nil)

	_, err = client.AddSubscription(ctx, req)

	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestSubscriptionCreation_error_invalidEmail(t *testing.T) {
	client, conn, err := getGrpcClient()
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	req := &pb.CreateSubscription{Email: "test"}
	_, err = client.AddSubscription(ctx, req)
	if err != nil {
		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Fatalf("SubscriptionCreation method failed: %v", err)
		}
	}
}
