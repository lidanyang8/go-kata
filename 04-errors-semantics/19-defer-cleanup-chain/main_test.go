package main

import (
	"context"
	"os"
	"runtime"
	"testing"
)

func TestBackupDatabase(t *testing.T) {
	type args struct {
		ctx      context.Context
		dbURL    string
		filename string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "事务开启失败",
			args: args{
				ctx:      context.Background(),
				dbURL:    beginFail,
				filename: "./backup/begin-fail.txt",
			},
			wantErr: true,
		},
		{
			name: "提交失败 + 关闭失败",
			args: args{
				ctx:      context.Background(),
				dbURL:    commitAndFileCloseFail,
				filename: "./backup/commit-fail.txt",
			},
			wantErr: true,
		},
		{
			name: "备份成功",
			args: args{
				ctx:      context.Background(),
				dbURL:    "127.0.0.1:3306",
				filename: "./backup/backup.txt",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := BackupDatabase(tt.args.ctx, tt.args.dbURL, tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("BackupDatabase() error = %v, wantErr %v", err, tt.wantErr)
			}
			t.Logf("BackupDatabase() error = %v", err)
		})
	}
}

func countOpenFDs() (int, error) {
	if runtime.GOOS != "linux" {
		// 非 Linux 系统（如 macOS）可能路径不同，
		// 这里仅作演示，实际生产测试建议在容器中运行。
		// 在 macOS 上可以使用 /dev/fd
		fds, err := os.ReadDir("/dev/fd")
		if err != nil {
			return 0, err
		}
		return len(fds), nil
	}
	fds, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		return 0, err
	}
	return len(fds), nil
}

func TestFDs(t *testing.T) {
	count, err := countOpenFDs()
	t.Logf("countOpenFDs() count = %d, err = %v", count, err)
}

// TestBackupDatabaseNoFDLeak 无文件描述符泄漏
func TestBackupDatabaseNoFDLeak(t *testing.T) {
	initialFDs, err := countOpenFDs()
	if err != nil {
		t.Skip("Skipping FD leak test: cannot read FD count on this OS")
	}

	ctx := context.Background()
	file := "./backup/leak.bak"
	for i := 0; i < 1000; i++ {
		if i%3 == 0 {
			_ = BackupDatabase(ctx, commitFail, file)
		} else {
			_ = BackupDatabase(ctx, "1", file)
		}
	}

	finalFDs, err := countOpenFDs()
	if err != nil {
		t.Fatal(err)
	}

	delta := finalFDs - initialFDs
	if delta > 5 { // 阈值设为 5，允许极小的抖动
		t.Errorf("File descriptor leak detected! Initial: %d, Final: %d, Delta: %d",
			initialFDs, finalFDs, delta)
	} else {
		t.Logf("FD consistency check passed. Delta: %d", delta)
	}

}
