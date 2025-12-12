package usecase

import (
	"apple_backend/order_service/internal/domain"
	"apple_backend/pkg/metrics"
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

type OrderRepository interface {
	GetOrderUserID(ctx context.Context, orderID string) (string, error)
	CreateOrder(ctx context.Context, userID string, isFast bool, comment string) (string, error)
	UpdateOrderStatus(ctx context.Context, orderID, status string) error
	GetOrder(ctx context.Context, orderID string) (*domain.OrderInfo, error)
	GetOrdersUser(ctx context.Context, filter *domain.OrderFilter) ([]*domain.Order, error)
	UpdateOrderTotal(ctx context.Context, orderID string, total float64) error
}

type OrderUsecase struct {
	repo      OrderRepository
	promoRepo PromoRepository
}

func NewOrderUsecase(repo OrderRepository, promoRepo PromoRepository) *OrderUsecase {
	return &OrderUsecase{repo: repo, promoRepo: promoRepo}
}

func (uc *OrderUsecase) CreateOrder(ctx context.Context, userID string, isFast bool, comment string, promo string) (*domain.OrderInfo, error) {
	orderID, err := uc.repo.CreateOrder(ctx, userID, isFast, comment)
	if err != nil {
		if errors.Is(err, domain.ErrCartEmpty) || errors.Is(err, domain.ErrRowsNotFound) {
			return nil, err
		}
		return nil, domain.ErrInternalServer
	}

	orderInfo, err := uc.repo.GetOrder(ctx, orderID)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			return nil, err
		}
		return nil, domain.ErrInternalServer
	}

	total := orderInfo.Total

	if promo != "" {
		p, err := uc.promoRepo.GetActiveForUserByCode(ctx, userID, promo, time.Now())
		if err != nil {
			if errors.Is(err, domain.ErrRowsNotFound) {
				return nil, domain.ErrRequestParams
			}
			return nil, domain.ErrInternalServer
		}

		var discount float64
		if p.RelativeDiscount > 0 {
			discount = total * p.RelativeDiscount / 100.0
		} else if p.AbsoluteDiscount > 0 {
			discount = p.AbsoluteDiscount
		}
		if discount > total {
			discount = total
		}

		total = total - discount

		if err := uc.promoRepo.MarkUsed(ctx, userID, p.ID); err != nil {
			return nil, domain.ErrInternalServer
		}
	}

	if isFast {
		total += 100
	}

	total = math.Round(total*100) / 100

	if err := uc.repo.UpdateOrderTotal(ctx, orderID, total); err != nil {
		return nil, domain.ErrInternalServer
	}

	orderInfo, err = uc.repo.GetOrder(ctx, orderID)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			return nil, err
		}
		return nil, domain.ErrInternalServer
	}

	storeID := "unknown"
	if len(orderInfo.Stores) > 0 {
		storeID = orderInfo.Stores[0].ID
	}
	metrics.OrdersCreatedTotal.WithLabelValues(storeID).Inc()

	return orderInfo, nil
}

func (uc *OrderUsecase) UpdateOrderStatus(ctx context.Context, orderID, userID, status string) error {
	allowed := map[string]bool{
		"pending":    true,
		"paid":       true,
		"delivered":  true,
		"cancelled":  true,
		"on_the_way": true,
	}
	if !allowed[status] {
		return domain.ErrRequestParams
	}

	realUserID, err := uc.repo.GetOrderUserID(ctx, orderID)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			return err
		}
		return domain.ErrInternalServer
	}
	if realUserID != userID {
		return domain.ErrForbidden
	}

	currentOrder, err := uc.repo.GetOrder(ctx, orderID)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			return err
		}
		return domain.ErrInternalServer
	}

	if status == "cancelled" {
		if currentOrder.Status == "pending" {
			if err := uc.repo.UpdateOrderStatus(ctx, orderID, "cancelled"); err != nil {
				if errors.Is(err, domain.ErrRowsNotFound) {
					return err
				}
				return domain.ErrInternalServer
			}

			storeID := "unknown"
			if currentOrder.Stores != nil && len(currentOrder.Stores) > 0 {
				storeID = currentOrder.Stores[0].ID
			}
			metrics.OrdersCanceledTotal.WithLabelValues(storeID).Inc()
			return nil
		}
		return fmt.Errorf("cannot cancel order in status '%s'", currentOrder.Status)
	}

	return domain.ErrForbidden
}

func (uc *OrderUsecase) GetOrder(ctx context.Context, orderID, userID string) (*domain.OrderInfo, error) {
	realUserID, err := uc.repo.GetOrderUserID(ctx, orderID)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			return nil, err
		}
		return nil, domain.ErrInternalServer
	}
	if realUserID != userID {
		return nil, domain.ErrForbidden
	}
	return uc.repo.GetOrder(ctx, orderID)
}

func (uc *OrderUsecase) GetOrdersUser(ctx context.Context, filter *domain.OrderFilter) ([]*domain.Order, error) {
	if filter == nil {
		return nil, domain.ErrRequestParams
	}
	if filter.Limit <= 0 || filter.Limit > 100 {
		return nil, domain.ErrRequestParams
	}
	if filter.UserID == "" {
		return nil, domain.ErrRequestParams
	}

	orders, err := uc.repo.GetOrdersUser(ctx, filter)
	if err != nil {
		if errors.Is(err, domain.ErrRowsNotFound) {
			return []*domain.Order{}, nil
		}
		return nil, domain.ErrInternalServer
	}
	return orders, nil
}
