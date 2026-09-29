package internal

import (
	"bytes"
	"fmt"
	"os/exec"
)

// extractCaptcha calls tesseract binary and returns the number only captcha output
func extractCaptcha(img []byte) (string, error) {
	// Read tesseract from stdin from buffer, output to stdout
	cmd := exec.Command(
		"tesseract",
		"stdin", "stdout",
		"--psm", "13",
		"--dpi", "300",
		"-c", "tessedit_char_whitelist=0123456789",
	)
	cmd.Stdin = bytes.NewBuffer(img)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("tesseract exec call fail: %w", err)
	}

	return string(bytes.TrimSpace(out)), nil
}
