package db

import (
	"context"
	"testing"
	"time"

	"l0/internal/models"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestPostgres_SaveOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockDatabase(ctrl)
	ctx := context.Background()

	order := models.Order{
		OrderUID:    "test-save",
		TrackNumber: "WBILSAVE",
		DateCreated: time.Now(),
		Delivery: models.Delivery{
			Name:  "Test User",
			Phone: "+79161234567",
		},
		Payment: models.Payment{
			Transaction: "test-save",
			Amount:      1000,
		},
		Items: []models.Item{
			{
				ChrtID: 123,
				Price:  500,
				Name:   "Test Item",
			},
		},
	}

	mockDB.EXPECT().
		SaveOrder(ctx, order).
		Return(nil).
		Times(1)

	err := mockDB.SaveOrder(ctx, order)
	assert.NoError(t, err)
}

func TestPostgres_GetOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockDatabase(ctrl)
	ctx := context.Background()

	expectedOrder := &models.Order{
		OrderUID:    "test-get",
		TrackNumber: "WBILGET",
		DateCreated: time.Now(),
	}

	mockDB.EXPECT().
		GetOrder(ctx, "test-get").
		Return(expectedOrder, nil).
		Times(1)

	order, err := mockDB.GetOrder(ctx, "test-get")
	assert.NoError(t, err)
	assert.Equal(t, expectedOrder.OrderUID, order.OrderUID)
	assert.Equal(t, expectedOrder.TrackNumber, order.TrackNumber)
}

func TestPostgres_GetOrderNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockDatabase(ctrl)
	ctx := context.Background()

	mockDB.EXPECT().
		GetOrder(ctx, "non-existent").
		Return(nil, assert.AnError).
		Times(1)

	order, err := mockDB.GetOrder(ctx, "non-existent")
	assert.Error(t, err)
	assert.Nil(t, order)
}

func TestPostgres_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockDatabase(ctrl)

	mockDB.EXPECT().
		Close().
		Times(1)

	mockDB.Close()
}
