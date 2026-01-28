package main

import "testing"

func TestNormalizeHeaderKey(t *testing.T) {

	tests := []struct {
		name    string
		args    string
		want    string
		wantErr bool
	}{
		{"standard", "content-type", "Content-Type", false},
		{"mixed case", "x-REquest-ID", "X-Request-Id", false},
		{"with numbers", "user-agent-2", "User-Agent-2", false},
		{"invalid underscore", "content_type", "", true},
		{"invalid space", "content type", "", true},
		{"invalid symbol", "header@", "", true},
		{"empty", "", "", true},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NormalizeHeaderKey(tc.args)
			if (err != nil) != tc.wantErr {
				t.Errorf("NormalizeHeaderKey() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if got != tc.want {
				t.Errorf("NormalizeHeaderKey() got = %v, want %v", got, tc.want)
			}
		})
	}
}

func FuzzNormalizeHeaderKey(f *testing.F) {
	// 种子语料库
	seeds := []string{"Content-Type", "x-api-key", "Invalid_Key"}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		got, err := NormalizeHeaderKey(input)
		if err != nil {
			return // 预期内的错误，正常跳过
		}

		// 1. 验证安全性：输出不应包含非法字符
		for i := 0; i < len(got); i++ {
			c := got[i]
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-') {
				t.Errorf("Normalized key contains invalid character: %q", c)
			}
		}

		// 2. 验证幂等性：Normalize(Normalize(s)) == Normalize(s)
		gotAgain, errAgain := NormalizeHeaderKey(got)
		if errAgain != nil {
			t.Fatalf("Second normalization failed: %v", errAgain)
		}
		if got != gotAgain {
			t.Errorf("Idempotency failed: first=%q, second=%q", got, gotAgain)
		}
	})
}
