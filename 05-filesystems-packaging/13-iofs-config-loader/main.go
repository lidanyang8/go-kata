package main

import (
	"io/fs"
	"strings"
)

func LoadConfigs(fsys fs.FS, root string) (map[string][]byte, error) {
	configs := make(map[string][]byte)

	// fs.WalkDir 是处理 io/fs 递归遍历的标准方式
	err := fs.WalkDir(fsys, root, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 只处理文件，且后缀名为 .conf
		if !d.IsDir() && strings.HasSuffix(filePath, ".conf") {
			// 使用 fs.ReadFile 从抽象文件系统中读取
			data, err := fs.ReadFile(fsys, filePath)
			if err != nil {
				return err
			}
			configs[filePath] = data
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return configs, nil
}
