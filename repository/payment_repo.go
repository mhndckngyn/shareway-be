package repository

import (
	"shareway/infra/db/migration"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type PaymentRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewPaymentRepository(db *gorm.DB, redis *redis.Client) IPaymentRepository {
	return &PaymentRepository{
		db:    db,
		redis: redis,
	}
}

type IPaymentRepository interface {
	StoreRequestID(requestID string, userID uuid.UUID, walletPhoneNumber string) error
	GetUserByID(userID uuid.UUID) (migration.User, error)
	GetRideOfferByID(rideOfferID uuid.UUID) (migration.RideOffer, error)
}

func (p *PaymentRepository) StoreRequestID(requestID string, userID uuid.UUID, walletPhoneNumber string) error {
	// Store request ID to db for later use
	var user migration.User
	if err := p.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return err
	}

	newRequestID, err := uuid.Parse(requestID)
	if err != nil {
		return err
	}
	user.MomoFirstRequestID = newRequestID
	user.MomoWalletID = walletPhoneNumber

	return p.db.Save(&user).Error
}

func (p *PaymentRepository) GetUserByID(userID uuid.UUID) (migration.User, error) {
	var user migration.User
	if err := p.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return user, err
	}

	return user, nil
}

func (p *PaymentRepository) GetRideOfferByID(rideOfferID uuid.UUID) (migration.RideOffer, error) {
	var rideOffer migration.RideOffer
	if err := p.db.Where("id = ?", rideOfferID).First(&rideOffer).Error; err != nil {
		return rideOffer, err
	}

	return rideOffer, nil
}
