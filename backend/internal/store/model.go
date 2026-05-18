package store

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusDraft       = "draft"
	StatusPublished   = "published"
	StatusUnpublished = "unpublished"
)

type Store struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	Name           string
	Slug           string
	Description    string
	LogoURL        string
	BannerURL      string
	Phone          string
	Whatsapp       string
	Email          string
	Address        string
	City           string
	Province       string
	PostalCode     string
	Status         string
	IsDiscoverable bool
	PublishedAt    *time.Time
}

type CreateParams struct {
	TenantID    uuid.UUID
	Name        string
	Slug        string
	Description string
	Phone       string
	Whatsapp    string
	Email       string
	Address     string
	City        string
	Province    string
	PostalCode  string
}

type UpdateProfileParams struct {
	TenantID       uuid.UUID
	StoreID        uuid.UUID
	Name           string
	Description    string
	Phone          string
	Whatsapp       string
	Email          string
	Address        string
	City           string
	Province       string
	PostalCode     string
	IsDiscoverable bool
}

type BusinessHour struct {
	DayOfWeek int
	OpenTime  string
	CloseTime string
	IsClosed  bool
}
