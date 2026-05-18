package store

import (
	"testing"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

func TestValidateBusinessHours(t *testing.T) {
	t.Parallel()

	valid := []BusinessHour{
		{DayOfWeek: 1, OpenTime: "08:00", CloseTime: "17:00"},
		{DayOfWeek: 7, IsClosed: true},
	}
	if _, err := validateBusinessHours(valid); err != nil {
		t.Fatalf("validateBusinessHours(valid) error = %v", err)
	}

	invalid := []BusinessHour{
		{DayOfWeek: 1, OpenTime: "17:00", CloseTime: "08:00"},
		{DayOfWeek: 1, OpenTime: "08:00", CloseTime: "17:00"},
	}

	_, err := validateBusinessHours(invalid)
	appErr, ok := err.(*apperror.AppError)
	if !ok {
		t.Fatalf("validateBusinessHours(invalid) error type = %T, want *apperror.AppError", err)
	}
	if appErr.Code != apperror.CodeValidation {
		t.Fatalf("validateBusinessHours(invalid) code = %s, want %s", appErr.Code, apperror.CodeValidation)
	}
}
