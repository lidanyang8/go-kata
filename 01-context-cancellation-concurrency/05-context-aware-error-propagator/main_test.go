package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestSensitiveDataLeak 测试场景 1: "The Sensitive Data Leak"
// 要求：fmt.Sprint(err) 不应包含 API key 字符串
func TestSensitiveDataLeak(t *testing.T) {
	// 使用包含敏感信息的 API key
	sensitiveAPIKey := "secret-api-key-xyz-12345-should-not-appear-in-logs"

	auth := NewAuthService()
	err := auth.Authenticate(context.Background(), "user123", sensitiveAPIKey)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// 关键测试：错误字符串不应包含 API key
	errStr := fmt.Sprint(err)
	if strings.Contains(errStr, sensitiveAPIKey) {
		t.Errorf("FAIL: error string contains sensitive API key\n"+
			"Error string: %q\n"+
			"API key: %q", errStr, sensitiveAPIKey)
	}

	// 验证错误类型可以正确提取
	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatal("expected AuthError, but errors.As failed")
	}

	// 验证 AuthError 内部确实包含 API key（但不应在 Error() 中暴露）
	if authErr.ApiKey != sensitiveAPIKey {
		t.Errorf("expected ApiKey to be %q, got %q", sensitiveAPIKey, authErr.ApiKey)
	}

	// 验证通过 uploadFile 包装后仍然不包含 API key
	upload := NewUploadFileService(WithAuth(), WithMeta(), WithStorage())
	wrappedErr := upload.uploadFile(context.Background(), "user123", sensitiveAPIKey, "file001")
	if wrappedErr == nil {
		t.Fatal("expected error from uploadFile, got nil")
	}

	wrappedErrStr := fmt.Sprint(wrappedErr)
	if strings.Contains(wrappedErrStr, sensitiveAPIKey) {
		t.Errorf("FAIL: wrapped error string contains sensitive API key\n"+
			"Wrapped error string: %q\n"+
			"API key: %q", wrappedErrStr, sensitiveAPIKey)
	}
}

// TestLostContext 测试场景 2: "The Lost Context"
// 要求：将 AuthError 包装三次后，errors.As(err, &AuthError{}) 仍应返回 true
func TestLostContext(t *testing.T) {
	auth := NewAuthService()

	// 触发认证错误
	err := auth.Authenticate(context.Background(), "user123", "invalid-key")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// 第一层包装：uploadFile 已经包装了一次
	upload := NewUploadFileService(WithAuth(), WithMeta(), WithStorage())
	wrapped1 := upload.uploadFile(context.Background(), "user123", "invalid-key", "file001")
	if wrapped1 == nil {
		t.Fatal("expected error from uploadFile, got nil")
	}

	// 手动再包装两次，模拟多层调用
	wrapped2 := fmt.Errorf("layer2: %w", wrapped1)
	wrapped3 := fmt.Errorf("layer3: %w", wrapped2)

	// 关键测试：即使包装了三次，errors.As 仍应能找到原始 AuthError
	var authErr *AuthError
	if !errors.As(wrapped3, &authErr) {
		t.Fatalf("FAIL: expected to extract AuthError after three layers of wrapping, but errors.As returned false\n"+
			"Error chain: %v", wrapped3)
	}

	// 验证提取的错误信息正确
	if authErr.UserId != "user123" {
		t.Errorf("expected UserId to be 'user123', got %q", authErr.UserId)
	}

	if authErr.Operation != "Authenticate" {
		t.Errorf("expected Operation to be 'Authenticate', got %q", authErr.Operation)
	}

	// 验证错误链的完整性
	if !errors.Is(wrapped3, authErr) {
		t.Error("errors.Is should return true for wrapped error chain")
	}
}

// TestTimeoutConfusion 测试场景 3: "The Timeout Confusion"
// 要求：存储层超时错误，errors.Is(err, context.DeadlineExceeded) 应返回 true
func TestTimeoutConfusion(t *testing.T) {
	upload := NewUploadFileService(WithAuth(), WithMeta(), WithStorage())

	// 使用会触发存储层超时的 userId
	err := upload.uploadFile(
		context.Background(),
		TimeoutStorageUserId,
		ValidApiKey,
		"file001",
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// 关键测试：errors.Is 应能识别 context.DeadlineExceeded
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("FAIL: expected error to be context.DeadlineExceeded, but errors.Is returned false\n"+
			"Error: %v", err)
	}

	// 验证错误类型
	var storageErr *StorageQuotaError
	if !errors.As(err, &storageErr) {
		t.Fatal("expected StorageQuotaError, but errors.As failed")
	}

	// 验证 Timeout() 方法
	if !storageErr.Timeout() {
		t.Error("StorageQuotaError.Timeout() should return true for timeout errors")
	}

	// 验证 Temporary() 方法（超时错误应该是临时的）
	if !storageErr.Temporary() {
		t.Error("StorageQuotaError.Temporary() should return true for timeout errors")
	}
}

// TestErrorWrapping 测试错误包装使用 %w
func TestErrorWrapping(t *testing.T) {
	upload := NewUploadFileService(WithAuth(), WithMeta(), WithStorage())

	// 测试认证层错误
	t.Run("AuthError wrapping", func(t *testing.T) {
		err := upload.uploadFile(context.Background(), "user123", "invalid-key", "file001")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var authErr *AuthError
		if !errors.As(err, &authErr) {
			t.Error("errors.As should extract AuthError from wrapped error")
		}
	})

	// 测试元数据层错误
	t.Run("MetadataError wrapping", func(t *testing.T) {
		err := upload.uploadFile(context.Background(), "user123", ValidApiKey, InvalidFileId)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var metaErr *MetadataError
		if !errors.As(err, &metaErr) {
			t.Error("errors.As should extract MetadataError from wrapped error")
		}
	})

	// 测试存储层错误
	t.Run("StorageQuotaError wrapping", func(t *testing.T) {
		err := upload.uploadFile(context.Background(), InvalidStorageUserId, ValidApiKey, "file001")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var storageErr *StorageQuotaError
		if !errors.As(err, &storageErr) {
			t.Error("errors.As should extract StorageQuotaError from wrapped error")
		}
	})
}

// TestContextAwareErrors 测试 Context-Aware 错误方法
func TestContextAwareErrors(t *testing.T) {
	t.Run("AuthError Timeout and Temporary", func(t *testing.T) {
		auth := NewAuthService()

		// 测试超时错误
		err := auth.Authenticate(context.Background(), TimeoutUserId, ValidApiKey)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var authErr *AuthError
		if !errors.As(err, &authErr) {
			t.Fatal("expected AuthError")
		}

		if !authErr.Timeout() {
			t.Error("AuthError.Timeout() should return true for timeout errors")
		}

		if !authErr.Temporary() {
			t.Error("AuthError.Temporary() should return true for timeout errors")
		}

		// 测试非超时错误
		err2 := auth.Authenticate(context.Background(), InvalidUserId, ValidApiKey)
		if err2 == nil {
			t.Fatal("expected error, got nil")
		}

		var authErr2 *AuthError
		if !errors.As(err2, &authErr2) {
			t.Fatal("expected AuthError")
		}

		if authErr2.Timeout() {
			t.Error("AuthError.Timeout() should return false for non-timeout errors")
		}

		if authErr2.Temporary() {
			t.Error("AuthError.Temporary() should return false for non-timeout errors")
		}
	})

	t.Run("MetadataError Timeout and Temporary", func(t *testing.T) {
		meta := NewMetadataService()

		// 测试超时错误
		err := meta.SaveMetadata(context.Background(), "user123", TimeoutFileId)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var metaErr *MetadataError
		if !errors.As(err, &metaErr) {
			t.Fatal("expected MetadataError")
		}

		if !metaErr.Timeout() {
			t.Error("MetadataError.Timeout() should return true for timeout errors")
		}

		if !metaErr.Temporary() {
			t.Error("MetadataError.Temporary() should return true for timeout errors")
		}

		// 测试死锁错误（应该是临时的）
		err2 := meta.SaveMetadata(context.Background(), "user123", DeadlockFileId)
		if err2 == nil {
			t.Fatal("expected error, got nil")
		}

		var metaErr2 *MetadataError
		if !errors.As(err2, &metaErr2) {
			t.Fatal("expected MetadataError")
		}

		if metaErr2.Timeout() {
			t.Error("MetadataError.Timeout() should return false for deadlock errors")
		}

		if !metaErr2.Temporary() {
			t.Error("MetadataError.Temporary() should return true for deadlock errors")
		}
	})

	t.Run("StorageQuotaError Timeout and Temporary", func(t *testing.T) {
		storage := &StorageService{}

		// 测试超时错误
		err := storage.UploadFile(context.Background(), TimeoutStorageUserId, "file001")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var storageErr *StorageQuotaError
		if !errors.As(err, &storageErr) {
			t.Fatal("expected StorageQuotaError")
		}

		if !storageErr.Timeout() {
			t.Error("StorageQuotaError.Timeout() should return true for timeout errors")
		}

		if !storageErr.Temporary() {
			t.Error("StorageQuotaError.Temporary() should return true for timeout errors")
		}

		// 测试非超时错误（配额错误不是临时的）
		err2 := storage.UploadFile(context.Background(), InvalidStorageUserId, "file001")
		if err2 == nil {
			t.Fatal("expected error, got nil")
		}

		var storageErr2 *StorageQuotaError
		if !errors.As(err2, &storageErr2) {
			t.Fatal("expected StorageQuotaError")
		}

		if storageErr2.Timeout() {
			t.Error("StorageQuotaError.Timeout() should return false for non-timeout errors")
		}

		if storageErr2.Temporary() {
			t.Error("StorageQuotaError.Temporary() should return false for quota errors (not temporary)")
		}
	})
}

// TestContextCancellation 测试上下文取消场景
func TestContextCancellation(t *testing.T) {
	upload := NewUploadFileService(WithAuth(), WithMeta(), WithStorage())

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 等待上下文超时
	time.Sleep(150 * time.Millisecond)

	// 测试认证层上下文取消
	err := upload.uploadFile(ctx, "user123", ValidApiKey, "file001")
	if err == nil {
		t.Fatal("expected error due to context timeout, got nil")
	}

	// 应该能识别为超时错误
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, but errors.Is returned false\n"+
			"Error: %v", err)
	}
}

// TestSuccessfulUpload 测试成功上传场景
func TestSuccessfulUpload(t *testing.T) {
	upload := NewUploadFileService(WithAuth(), WithMeta(), WithStorage())

	err := upload.uploadFile(
		context.Background(),
		"user123",
		ValidApiKey,
		"file001",
	)

	if err != nil {
		t.Errorf("expected no error for successful upload, got: %v", err)
	}
}

// TestMetadataDeadlock 测试元数据死锁场景
func TestMetadataDeadlock(t *testing.T) {
	upload := NewUploadFileService(WithAuth(), WithMeta(), WithStorage())

	err := upload.uploadFile(
		context.Background(),
		"user123",
		ValidApiKey,
		DeadlockFileId,
	)

	if err == nil {
		t.Fatal("expected error for deadlock, got nil")
	}

	var metaErr *MetadataError
	if !errors.As(err, &metaErr) {
		t.Fatal("expected MetadataError, but errors.As failed")
	}

	// 死锁错误应该是临时的
	if !metaErr.Temporary() {
		t.Error("MetadataError.Temporary() should return true for deadlock errors")
	}

	// 验证死锁错误类型
	var deadlockErr *DeadlockError
	if !errors.As(metaErr.Err, &deadlockErr) {
		t.Error("expected DeadlockError in error chain")
	}
}

// TestErrorUnwrapping 测试错误解包支持 errors.Is 和 errors.As
func TestErrorUnwrapping(t *testing.T) {
	upload := NewUploadFileService(WithAuth(), WithMeta(), WithStorage())

	// 测试 errors.Is 通过错误链
	t.Run("errors.Is through error chain", func(t *testing.T) {
		err := upload.uploadFile(
			context.Background(),
			TimeoutStorageUserId,
			ValidApiKey,
			"file001",
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		// errors.Is 应该能找到 context.DeadlineExceeded
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Error("errors.Is should find context.DeadlineExceeded through error chain")
		}
	})

	// 测试 errors.As 提取具体错误类型
	t.Run("errors.As extracts specific error types", func(t *testing.T) {
		testCases := []struct {
			name            string
			userId          string
			apiKey          string
			fileId          string
			expectedErrType interface{}
		}{
			{
				name:            "AuthError",
				userId:          "user123",
				apiKey:          "invalid-key",
				fileId:          "file001",
				expectedErrType: (*AuthError)(nil),
			},
			{
				name:            "MetadataError",
				userId:          "user123",
				apiKey:          ValidApiKey,
				fileId:          InvalidFileId,
				expectedErrType: (*MetadataError)(nil),
			},
			{
				name:            "StorageQuotaError",
				userId:          InvalidStorageUserId,
				apiKey:          ValidApiKey,
				fileId:          "file001",
				expectedErrType: (*StorageQuotaError)(nil),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := upload.uploadFile(
					context.Background(),
					tc.userId,
					tc.apiKey,
					tc.fileId,
				)

				if err == nil {
					t.Fatal("expected error, got nil")
				}

				// errors.As 应该能提取具体错误类型
				switch expected := tc.expectedErrType.(type) {
				case *AuthError:
					var authErr *AuthError
					if !errors.As(err, &authErr) {
						t.Error("errors.As should extract AuthError")
					}
				case *MetadataError:
					var metaErr *MetadataError
					if !errors.As(err, &metaErr) {
						t.Error("errors.As should extract MetadataError")
					}
				case *StorageQuotaError:
					var storageErr *StorageQuotaError
					if !errors.As(err, &storageErr) {
						t.Error("errors.As should extract StorageQuotaError")
					}
				default:
					t.Fatalf("unexpected error type: %T", expected)
				}
			})
		}
	})
}

// TestWithStorageOption 测试 WithStorage 选项
func TestWithStorageOption(t *testing.T) {
	upload := NewUploadFileService(WithAuth(), WithMeta(), WithStorage())

	if upload.storage == nil {
		t.Fatal("WithStorage() option should initialize storage service")
	}

	err := upload.uploadFile(
		context.Background(),
		"user123",
		ValidApiKey,
		"file001",
	)

	if err != nil {
		t.Errorf("expected no error for valid upload, got: %v", err)
	}
}
