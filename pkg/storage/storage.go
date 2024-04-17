package storage

import (
	"io"
	"log/slog"
	"os"
)

type Storage struct {
	filePath string
	content  string
}

func New(filePath string) *Storage {
	return &Storage{
		filePath: filePath,
	}
}

func (s *Storage) Initialize(filePath string) {
	f, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		slog.Error("failed to open file", slog.String("path", filePath), slog.Any("error", err))
	}

	content, err := io.ReadAll(f)
	if err != nil {
		slog.Error("failed to read file", slog.String("path", filePath), slog.Any("error", err))
	}

	s.content = string(content)
}

func (s *Storage) Store(key string, value any) string {
	return ""
}
