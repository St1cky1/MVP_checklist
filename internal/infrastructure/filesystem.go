package infrastructure

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileSystemStorage struct {
	basePath string
}

func NewFileSystemStorage(basePath string) *FileSystemStorage {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		fmt.Printf("Warning: failed to create upload directory: %v\n", err)
	}
	return &FileSystemStorage{
		basePath: basePath,
	}
}

func (s *FileSystemStorage) Upload(ctx context.Context, bucket, key string, data []byte) (string, error) {
	// Игнорируем bucket для FS или используем как подпапку
	fullPath := filepath.Join(s.basePath, key)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create folder: %w", err)
	}

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Возвращаем ключ (относительный путь), который будет использоваться в GetURL
	return key, nil
}

func (s *FileSystemStorage) GetURL(ctx context.Context, bucket, key string) (string, error) {
	if strings.HasPrefix(key, "/uploads/") {
		return key, nil
	}
	return "/uploads/" + key, nil
}
