package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type AuthError struct {
	Operation string
	UserId    string
	ApiKey    string
	Err       error
	timestamp time.Time
}

func NewAuthError(op, userId, apiKey string, err error) *AuthError {
	return &AuthError{
		Operation: op,
		UserId:    userId,
		ApiKey:    apiKey,
		Err:       err,
		timestamp: time.Now(),
	}
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("auth failed for user %q during %s at %s %v", e.UserId,
		e.Operation, e.timestamp.Format(time.RFC3339), e.Err.Error())
}

func (e *AuthError) Unwrap() error {
	return e.Err
}

func (e *AuthError) Timeout() bool {
	switch {
	case errors.Is(e.Err, context.DeadlineExceeded):
		return true
	}
	return false
}

func (e *AuthError) Temporary() bool {
	switch {
	case errors.Is(e.Err, context.DeadlineExceeded):
		return true
	}
	return false
}

type AuthService struct {
}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (a *AuthService) Authenticate(ctx context.Context, userId, apiKey string) error {
	if err := ctx.Err(); err != nil {
		return NewAuthError("Authenticate", userId, apiKey, err)
	}

	if apiKey != ValidApiKey {
		return NewAuthError("Authenticate", userId, apiKey, errors.New("invalid api key"))
	}
	if userId == InvalidUserId {
		return NewAuthError("Authenticate", userId, apiKey, errors.New("invalid user id"))
	}
	if userId == TimeoutUserId {
		return NewAuthError("Authenticate", userId, apiKey, context.DeadlineExceeded)
	}

	return nil
}
