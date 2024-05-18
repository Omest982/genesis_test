package dbService

import (
	"awesomeProject/db/repository"
	pb "awesomeProject/emailService"
)

type SubscriptionService struct {
	Repo repository.SubscriptionRepository
}

func (s *SubscriptionService) AddSubscription(req *pb.CreateSubscription) error {
	return s.Repo.CreateSubscription(req)
}

func (s *SubscriptionService) IsSubscriptionExistsByEmail(email string) (bool, error) {
	return s.Repo.IsSubscriptionExistsByEmail(email)
}
