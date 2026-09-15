package repository

import (
	"context"

	"gorm.io/gorm"

	"ticketflow/services/payment-service/internal/model"
)

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, p *model.Payment) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, id, status string, providerTxnID *string) error {
	return r.db.WithContext(ctx).Model(&model.Payment{}).Where("id = ?", id).
		Updates(map[string]any{"status": status, "provider_txn_id": providerTxnID}).Error
}

func (r *PaymentRepository) GetLatestByOrderID(ctx context.Context, orderID string) (*model.Payment, error) {
	var p model.Payment
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Order("created_at DESC").First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}
