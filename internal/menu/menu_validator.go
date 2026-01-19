package menu

import (
	"errors"
	"path/filepath"
	"strings"
)

var allowedExt = map[string]bool{
	".pdf":  true,
	".txt":  true,
	".csv":  true,
	".json": true,
	".xml":  true,
}

func ValidateFileExtension(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))

	if ext == "" {
		return errors.New("file extension missing")
	}

	if !allowedExt[ext] {
		return errors.New("file type not allowed")
	}

	return nil
}
