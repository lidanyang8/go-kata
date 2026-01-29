package main

import (
	"errors"
	"fmt"
	"testing"
)

func TestDoThing(t *testing.T) {
	type args struct {
		fail bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "DoThing err",
			args:    args{fail: true},
			wantErr: true,
		},
		{
			name:    "DoThing nil",
			args:    args{fail: false},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DoThing(tt.args.fail); (err != nil) != tt.wantErr {
				t.Errorf("DoThing() type = %T error = %v, wantErr %v", err, err, tt.wantErr)
			} else {
				t.Logf("DoThing() type = %T error = %v, wantErr %v", err, err, tt.wantErr)
			}
		})
	}
}

func TestDoThingFix(t *testing.T) {
	type args struct {
		fail bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "DoThingFix err",
			args:    args{fail: true},
			wantErr: true,
		},
		{
			name:    "DoThingFix nil",
			args:    args{fail: false},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DoThingFix(tt.args.fail); (err != nil) != tt.wantErr {
				t.Errorf("DoThingFix() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				t.Logf("DoThingFix() type = %T error = %v, wantErr %v", err, err, tt.wantErr)
			}
		})
	}
}

func TestDoThingSafe(t *testing.T) {
	// 1. 模拟包装错误（即使嵌套多层，As 也能递归找到）
	werr := fmt.Errorf("wrapped context: %w", DoThing(false))

	var me *MyError

	// 2. 这里的 errors.As 会返回 true，因为类型匹配成功 (*MyError)
	if errors.As(werr, &me) {
		// 3. 关键点：提取出来的 me 本身是 nil 指针
		if me == nil {
			t.Log("Successfully extracted nil *MyError from wrapped interface")
		} else {
			t.Logf("Extracted valid MyError: %v", me)
		}
		return
	}

	t.Errorf("Failed to extract MyError type, got: %T", werr)
}
