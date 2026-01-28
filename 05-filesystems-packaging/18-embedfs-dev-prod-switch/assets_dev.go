//go:build dev

package main

import (
	"io/fs"
	"os"
)

func Assets() (templates fs.FS, static fs.FS, err error) {
	// 直接映射本地目录
	// 注意：os.DirFS 是相对于运行命令时的当前目录
	templates = os.DirFS("templates")
	static = os.DirFS("static")
	return templates, static, nil
}
