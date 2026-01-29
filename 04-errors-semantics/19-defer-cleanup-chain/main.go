package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
)

type DB interface {
	Begin(ctx context.Context) (TX, error)
	Close() error
}

type TX interface {
	Query() (Rows, error)
	Commit() error
	Rollback() error
}

type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
}

const (
	beginFail              = "begin-fail"             // 事务开启失败
	commitAndFileCloseFail = "commit-file-close-fail" // 提交失败 + 关闭失败
	commitFail             = "commit-fail"
)

func OpenDB(url string) (DB, error) {
	switch url {
	case beginFail:
		return &MockDB{
			beginFail: true,
		}, nil
	case commitAndFileCloseFail, commitFail:
		return &MockDB{
			commitFail: true,
		}, nil
	}
	return &MockDB{}, nil
}

func BackupDatabase(ctx context.Context, dbURL, filename string) (err error) {
	//打开输出文件；
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer func() {
		if dbURL == commitAndFileCloseFail {
			err = errors.Join(err, fmt.Errorf("file close failed"))
		} else {
			err = errors.Join(err, f.Close())
		}
	}()

	//连接数据库；
	db, err := OpenDB(dbURL)
	if err != nil {
		return fmt.Errorf("open db client: %w", err)
	}
	defer func() {
		err = errors.Join(err, db.Close())
	}()

	//开启事务；
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	var committed bool // 标记是否事务回滚
	defer func() {
		if !committed {
			err = errors.Join(err, tx.Rollback())
		}
	}()

	//将行数据流式写入文件；
	rows, err := tx.Query()
	if err != nil {
		return fmt.Errorf("query rows: %w", err)
	}
	defer func() {
		err = errors.Join(err, rows.Close())
	}()
	for rows.Next() {
		var data string
		scanErr := rows.Scan(&data)
		if scanErr != nil {
			return fmt.Errorf("scan row: %w", scanErr)
		}

		if _, writeErr := io.WriteString(f, data+"\n"); writeErr != nil {
			return fmt.Errorf("write data: %w", writeErr)
		}

		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf("context: %w", ctxErr)
		}
	}

	//成功时提交事务；
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	//若任一步失败，必须对已获取的资源按正确顺序关闭/回滚。
	committed = true
	return nil
}
