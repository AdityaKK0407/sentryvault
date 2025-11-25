package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetDBFiles() ([]string, error) {
	path := filepath.Join(".", "users")
	dir, err := os.ReadDir(path)

	if os.IsNotExist(err) {
		if err = os.Mkdir("users", 0700); err != nil {
			return nil, fmt.Errorf("error creating file: %w", err)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("error reading file path: %w", err)
	}

	var files []string

	for _, file := range dir {
		fileName := file.Name()
		if !file.IsDir() && strings.HasSuffix(fileName, ".db") {
			files = append(files, fileName)
		}
	}

	return files, nil
}
