package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type DeadlockError struct {
}

func (e *DeadlockError) Error() string {
	return "database deadlock"
}

type MetadataError struct {
	Operation string
	UserId    string
	FileId    string
	timestamp time.Time
	Err       error
}

func NewMetadataError(op, userId, fileId string, err error) *MetadataError {
	return &MetadataError{
		Operation: op,
		UserId:    userId,
		FileId:    fileId,
		timestamp: time.Now(),
		Err:       err,
	}
}

func (e *MetadataError) Error() string {
	return fmt.Sprintf("metadata failed for user %q file %q during %s at %s", e.UserId,
		e.FileId, e.Operation, e.timestamp.Format(time.RFC3339))
}

func (e *MetadataError) Unwrap() error {
	return e.Err
}

func (e *MetadataError) Timeout() bool {
	return errors.Is(e.Err, context.DeadlineExceeded)
}

func (e *MetadataError) Temporary() bool {
	if e.Err == nil {
		return false
	}

	if errors.Is(e.Err, context.DeadlineExceeded) {
		return true
	}

	var deadlock *DeadlockError
	if errors.As(e.Err, &deadlock) {
		return true
	}

	var t interface{ Temporary() bool }
	if errors.As(e.Err, &t) {
		return t.Temporary()
	}

	return false
}

type MetadataService struct {
}

func NewMetadataService() *MetadataService {
	return &MetadataService{}
}

func (m *MetadataService) SaveMetadata(ctx context.Context, userId, fileId string) error {
	if ctx.Err() != nil {
		return NewMetadataError("SaveMetadata", userId, fileId, ctx.Err())
	}
	if fileId == InvalidFileId {
		return NewMetadataError("SaveMetadata", userId, fileId, errors.New("invalid fileId"))
	}
	if fileId == TimeoutFileId {
		return NewMetadataError("SaveMetadata", userId, fileId, context.DeadlineExceeded)
	}
	if fileId == DeadlockFileId {
		return NewMetadataError("SaveMetadata", userId, fileId, &DeadlockError{})
	}

	return nil
}
