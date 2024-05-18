package repositoryImpl

import (
	pb "awesomeProject/emailService"
	_type "awesomeProject/type"
	"errors"
	"gorm.io/gorm"
)

type SubscriptionRepositoryImpl struct {
	DB *gorm.DB
}

func (r *SubscriptionRepositoryImpl) CreateSubscription(req *pb.CreateSubscription) error {
	subscription := &_type.Subscription{Email: req.Email}

	return r.DB.Create(subscription).Error
}

func (r *SubscriptionRepositoryImpl) IsSubscriptionExistsByEmail(email string) (bool, error) {
	var subscription _type.Subscription
	result := r.DB.Where("email = ?", email).First(&subscription)

	if result.Error == nil {
		return true, nil
	} else if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, result.Error
	}

	return false, nil
}
