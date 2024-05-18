package repository

import pb "awesomeProject/emailService"

type SubscriptionRepository interface {
	CreateSubscription(req *pb.CreateSubscription) error
	IsSubscriptionExistsByEmail(email string) (bool, error)
}
