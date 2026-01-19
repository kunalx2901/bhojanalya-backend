package ocr

import "os/exec"

func ExtractText(filePath string) (string, error) {
	cmd := exec.Command("tesseract", filePath, "stdout")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
