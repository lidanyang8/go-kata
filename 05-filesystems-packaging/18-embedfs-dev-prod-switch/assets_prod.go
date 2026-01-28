//go:build !dev

package main

import (
	"embed"
	"io/fs"
)

//go:embed static templates
var embedFiles embed.FS

func Assets() (templates fs.FS, static fs.FS, err error) {
	// 使用 fs.Sub 剥离前缀，使得返回的 FS 根目录就是内容本身
	templates, err = fs.Sub(embedFiles, "templates")
	if err != nil {
		return nil, nil, err
	}

	static, err = fs.Sub(embedFiles, "static")
	if err != nil {
		return nil, nil, err
	}

	return templates, static, nil
}
