package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetDBFiles() ([]string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("unable to find user config directory: %w", err)
	}
	path := filepath.Join(configDir, "SentryVault", "users")
	err = os.MkdirAll(path, 0700)
	if err != nil {
		return nil, fmt.Errorf("error creating directory: %w", err)
	}

	var files []string

	dir, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file path: %w", err)
	}
	for _, file := range dir {
		fileName := file.Name()
		if !file.IsDir() && strings.HasSuffix(fileName, ".db") {
			files = append(files, fileName)
		}
	}

	return files, nil
}
