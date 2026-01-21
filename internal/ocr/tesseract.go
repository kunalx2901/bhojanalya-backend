package ocr

import (
	"fmt"
	"os/exec"
)

func ExtractText(filePath string) (string, error) {
	cmd := exec.Command("tesseract", filePath, "stdout", "-l", "eng")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("tesseract failed: %w - output: %s", err, string(out))
	}
	return string(out), nil
}
