package mocks

import (
	pb "awesomeProject/emailService"
	"github.com/stretchr/testify/mock"
)

type MockSubscriptionRepository struct {
	mock.Mock
}

func (m *MockSubscriptionRepository) CreateSubscription(createRequest *pb.CreateSubscription) error {
	args := m.Called(createRequest)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) IsSubscriptionExistsByEmail(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}
