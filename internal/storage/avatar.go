package storage

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AvatarStorage defines persistence behaviour for user avatar files.
type AvatarStorage interface {
	Save(userID uuid.UUID, data io.Reader, originalFilename string) (string, error)
}

// LocalAvatarStorage stores avatar images on the local filesystem and exposes them via a configurable URL prefix.
type LocalAvatarStorage struct {
	baseDir   string
	urlPrefix string
}

// NewLocalAvatarStorage constructs a LocalAvatarStorage ensuring the target directory exists.
func NewLocalAvatarStorage(baseDir, urlPrefix string) (*LocalAvatarStorage, error) {
	if baseDir == "" {
		return nil, fmt.Errorf("baseDir is required")
	}
	if urlPrefix == "" {
		urlPrefix = "/avatars"
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("create avatar storage dir: %w", err)
	}
	return &LocalAvatarStorage{baseDir: baseDir, urlPrefix: trimTrailingSlash(urlPrefix)}, nil
}

// Save writes the avatar file to disk under a user-specific directory and returns the public URL path.
func (s *LocalAvatarStorage) Save(userID uuid.UUID, data io.Reader, originalFilename string) (string, error) {
	if userID == uuid.Nil {
		return "", fmt.Errorf("userID is required")
	}

	userDir := filepath.Join(s.baseDir, userID.String())
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		return "", fmt.Errorf("create user avatar dir: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(originalFilename))
	if len(ext) > 10 {
		ext = ext[:10]
	}
	filename := fmt.Sprintf("%d-%s%s", time.Now().UTC().UnixNano(), uuid.NewString(), ext)
	fullPath := filepath.Join(userDir, filename)

	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create avatar file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, data); err != nil {
		return "", fmt.Errorf("write avatar file: %w", err)
	}

	publicPath := path.Join(s.urlPrefix, userID.String(), filename)
	if !strings.HasPrefix(publicPath, "/") {
		publicPath = "/" + publicPath
	}
	return publicPath, nil
}

func trimTrailingSlash(prefix string) string {
	if prefix == "/" {
		return prefix
	}
	return strings.TrimRight(prefix, "/")
}
