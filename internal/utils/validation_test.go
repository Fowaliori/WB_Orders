package utils

import (
	"testing"
	"time"

	"l0/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestValidateStruct_ValidOrder(t *testing.T) {
	order := models.Order{
		OrderUID:    "valid-123",
		TrackNumber: "TRACK123",
		Entry:       "web",
		Delivery: models.Delivery{
			Name:    "John Doe",
			Phone:   "+79161234567",
			Zip:     "123456",
			City:    "Moscow",
			Address: "Red Square",
			Region:  "Moscow",
			Email:   "test@example.com",
		},
		Payment: models.Payment{
			Transaction: "trx-123",
			Currency:    "RUB",
			Provider:    "visa",
			Amount:      1000,
			PaymentDT:   time.Now().Unix(),
			Bank:        "Sber",
		},
		Items: []models.Item{
			{
				ChrtID:      1,
				TrackNumber: "TRACK123",
				Price:       500,
				RID:         "rid-1",
				Name:        "Item1",
				Size:        "M",
				TotalPrice:  500,
				NMID:        1,
				Brand:       "Nike",
				Status:      1,
			},
		},
		Locale:          "ru",
		CustomerID:      "cust-1",
		DeliveryService: "WB",
		ShardKey:        "1",
		SMID:            1,
		DateCreated:     time.Now(),
		OOFShard:        "1",
	}

	err := ValidateStruct(order)
	assert.NoError(t, err, "Valid order should not return validation errors")
}

func TestValidateStruct_InvalidOrder(t *testing.T) {
	order := models.Order{
		OrderUID: "",
	}

	err := ValidateStruct(order)
	assert.Error(t, err, "Invalid order should return validation errors")

	validationErrors := GetValidationErrors(err)
	assert.Contains(t, validationErrors, "OrderUID", "OrderUID should be required")
}
