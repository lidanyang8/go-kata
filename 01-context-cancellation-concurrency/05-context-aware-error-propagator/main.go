package main

import (
	"context"
	"fmt"
	"log/slog"
)

type Option func(*UploadFileService)

type UploadFileService struct {
	auth    *AuthService
	meta    *MetadataService
	storage *StorageService
}

func NewUploadFileService(opts ...Option) *UploadFileService {
	upload := &UploadFileService{}

	for _, opt := range opts {
		opt(upload)
	}
	return upload
}

func WithAuth() Option {
	return func(upload *UploadFileService) {
		upload.auth = NewAuthService()
	}
}

func WithMeta() Option {
	return func(upload *UploadFileService) {
		upload.meta = NewMetadataService()
	}
}

func WithStorage() Option {
	return func(upload *UploadFileService) {
		upload.storage = &StorageService{}
	}
}

func (u *UploadFileService) uploadFile(ctx context.Context, userId, apiKey, fileId string) error {
	// 认证
	if err := u.auth.Authenticate(ctx, userId, apiKey); err != nil {
		return fmt.Errorf("upload file failed: %w", err)
	}

	// 元数据
	if err := u.meta.SaveMetadata(ctx, userId, fileId); err != nil {
		return fmt.Errorf("upload file failed: %w", err)
	}

	// 存储
	if err := u.storage.UploadFile(ctx, userId, fileId); err != nil {
		return fmt.Errorf("upload file failed: %w", err)
	}

	return nil
}

func main() {

	ctx := context.Background()
	upload := NewUploadFileService(WithAuth(), WithMeta(), WithStorage())
	if err := upload.uploadFile(ctx, "user123", "valid_api_key", "file123"); err != nil {
		slog.Error("upload file failed")
		return
	}
	slog.Info("upload file success")
}
