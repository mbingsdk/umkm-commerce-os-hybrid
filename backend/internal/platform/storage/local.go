package storage

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

type Local struct {
	rootDir   string
	publicURL string
	maxBytes  int64
	newID     func() string
}

func NewLocal(rootDir string, publicURL string, maxBytes int64) *Local {
	return &Local{
		rootDir:   rootDir,
		publicURL: strings.TrimRight(publicURL, "/"),
		maxBytes:  maxBytes,
		newID:     uuid.NewString,
	}
}

func (s *Local) Store(_ context.Context, input StoreInput) (Asset, error) {
	if input.TenantID == uuid.Nil {
		return Asset{}, validation("tenant_id", "Tenant is required")
	}
	if err := validateFolderSegment(input.Folder); err != nil {
		return Asset{}, err
	}

	mimeType, extension, err := validateImage(input.Data, s.maxBytes)
	if err != nil {
		return Asset{}, err
	}

	key := path.Join("tenants", input.TenantID.String(), input.Folder, s.newID()+extension)
	targetPath, err := s.safePath(key)
	if err != nil {
		return Asset{}, err
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return Asset{}, fmt.Errorf("create upload directory: %w", err)
	}
	if err := os.WriteFile(targetPath, input.Data, 0o644); err != nil {
		return Asset{}, fmt.Errorf("write upload file: %w", err)
	}

	return Asset{
		Key:      key,
		URL:      s.publicURL + "/" + key,
		MIMEType: mimeType,
		Size:     int64(len(input.Data)),
	}, nil
}

func (s *Local) Delete(_ context.Context, key string) error {
	targetPath, err := s.safePath(key)
	if err != nil {
		return err
	}
	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove upload file: %w", err)
	}
	return nil
}

func (s *Local) safePath(key string) (string, error) {
	rootDir, err := filepath.Abs(s.rootDir)
	if err != nil {
		return "", fmt.Errorf("resolve upload root: %w", err)
	}

	targetPath, err := filepath.Abs(filepath.Join(rootDir, filepath.FromSlash(key)))
	if err != nil {
		return "", fmt.Errorf("resolve upload target: %w", err)
	}

	rel, err := filepath.Rel(rootDir, targetPath)
	if err != nil {
		return "", fmt.Errorf("resolve upload relative path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", validation("file", "Upload path is invalid")
	}

	return targetPath, nil
}

func validateFolderSegment(folder string) error {
	folder = strings.TrimSpace(folder)
	if folder == "" || folder == "." || folder == ".." || strings.ContainsAny(folder, `/\`) {
		return validation("folder", "Folder is invalid")
	}
	return nil
}

func validation(field string, message string) error {
	return apperror.Validation("Validation failed", []map[string]string{
		{"field": field, "message": message},
	})
}
