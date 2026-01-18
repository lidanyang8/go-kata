package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type StorageQuotaError struct {
	Operation string
	UserId    string
	FileId    string
	timestamp time.Time
	Err       error
}

func (e *StorageQuotaError) Error() string {
	return fmt.Sprintf("storage failed for user %q file %q during %s at %s %v", e.UserId,
		e.FileId, e.Operation, e.timestamp.Format(time.RFC3339), e.Err)
}

func (e *StorageQuotaError) Unwrap() error {
	return e.Err
}

func (e *StorageQuotaError) Timeout() bool {
	return errors.Is(e.Err, context.DeadlineExceeded)
}

func (e *StorageQuotaError) Temporary() bool {
	// 配额错误不是临时的，但超时错误是临时的
	if errors.Is(e.Err, context.DeadlineExceeded) {
		return true
	}
	// 配额超限不是临时错误，不能通过重试解决
	return false
}

func NewStorageQuotaError(op, userId, fileId string, err error) *StorageQuotaError {
	return &StorageQuotaError{
		Operation: op,
		UserId:    userId,
		FileId:    fileId,
		timestamp: time.Now(),
		Err:       err,
	}
}

type StorageService struct {
}

func (s *StorageService) UploadFile(ctx context.Context, userId, fileId string) error {
	if ctx.Err() != nil {
		return NewStorageQuotaError("UploadFile", userId, fileId, ctx.Err())
	}

	if userId == InvalidStorageUserId {
		return NewStorageQuotaError("UploadFile", userId, fileId, errors.New("invalid userId"))
	}
	if userId == TimeoutStorageUserId {
		return NewStorageQuotaError("UploadFile", userId, fileId, context.DeadlineExceeded)
	}

	return nil
}
