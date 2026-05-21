package plan

import (
	"errors"

	"github.com/google/uuid"
)

type Feature string

const (
	FeaturePOS       Feature = "pos"
	FeatureDiscovery Feature = "discovery"
	FeatureCourier   Feature = "courier"
)

var ErrPlanNotFound = errors.New("plan not found")

type Plan struct {
	ID                 uuid.UUID
	Code               string
	Name               string
	Description        string
	PriceMonthly       int64
	ProductLimit       *int
	StaffLimit         *int
	CanUsePOS          bool
	CanUseDiscovery    bool
	CanUseCourier      bool
	CanUseCustomDomain bool
	IsActive           bool
}
