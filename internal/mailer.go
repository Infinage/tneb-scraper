package internal

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "embed"
)

//go:embed assets/bills.html
var billsTemplate string

// fetchLatestScreenshot lookups in the given path for .png files having the
// most recent modification time.
func fetchLatestScreenshot(fpath string) (string, error) {
	dir, err := os.ReadDir(fpath)
	if err != nil {
		return "", fmt.Errorf("read screenshot directory: %w", err)
	}

	// Fetch the latest screenshot
	var mxTime int64
	var mxFile os.DirEntry
	for _, file := range dir {
		if !file.Type().IsRegular() || !strings.HasSuffix(file.Name(), ".png") {
			continue
		}

		info, err := file.Info()
		if err != nil {
			return "", fmt.Errorf("get screenshot info: %w", err)
		}

		if t := info.ModTime().UnixMilli(); t > mxTime {
			mxTime = t
			mxFile = file
		}
	}

	if mxFile == nil {
		return "", fmt.Errorf("no screenshot found")
	}

	return mxFile.Name(), nil
}

// SendMail sends a summary email with bill details neatly printed or an error
// message. Always attaches the latest screenshot from './tmp/screenshots'.
func SendMail(username, password, target string, bills []EBBill, extractErr error) error {
	latestPNG, err := fetchLatestScreenshot("tmp/screenshots")
	if err != nil {
		return fmt.Errorf("latest screenshot: %w", err)
	}

	screenshot := filepath.Join("tmp/screenshots", latestPNG)
	img, err := os.ReadFile(screenshot)
	if err != nil {
		return fmt.Errorf("read screenshot: %w", err)
	}

	now := time.Now().Format("02 Jan 2006")
	boundary := "TNEB-SCRAPER-BOUNDARY"

	templFunc := template.FuncMap{"inc": func(i int) int { return i + 1 }}
	templ, err := template.New("bills").Funcs(templFunc).Parse(billsTemplate)
	if err != nil {
		return fmt.Errorf("template parse: %w", err)
	}

	// Construct the message in required format
	var body strings.Builder

	// Email header
	body.WriteString("From: " + username + "\r\n")
	body.WriteString("To: " + target + "\r\n")
	body.WriteString("Subject: TNEB Bill Summary (" + now + ")\r\n")
	body.WriteString("MIME-Version: 1.0\r\n")
	body.WriteString("Content-Type: multipart/mixed; boundary=\"" + boundary + "\"\r\n")
	body.WriteString("\r\n")

	// Email body
	body.WriteString("--" + boundary + "\r\n")
	body.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	body.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	body.WriteString("\r\n")
	if extractErr != nil {
		fmt.Fprintf(&body, "<p>Failed to extract bills:</p><pre>%v</pre>", extractErr)
	} else {
		if err = templ.ExecuteTemplate(&body, "bills", bills); err != nil {
			return fmt.Errorf("exec template: %w", err)
		}
	}

	// Email screenshot
	body.WriteString("\r\n--" + boundary + "\r\n")
	body.WriteString("Content-Type: image/png\r\n")
	body.WriteString("Content-Transfer-Encoding: base64\r\n")
	body.WriteString("Content-Disposition: attachment; filename=\"" + latestPNG + "\"\r\n")
	body.WriteString("\r\n")
	encoder := base64.NewEncoder(base64.StdEncoding, &body)
	_, err = encoder.Write(img)
	if err != nil {
		return fmt.Errorf("encode screenshot: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("close base64 encoder: %w", err)
	}

	body.WriteString("\r\n--" + boundary + "--\r\n")
	auth := smtp.PlainAuth("", username, password, "smtp.gmail.com")
	return smtp.SendMail("smtp.gmail.com:587", auth, username,
		[]string{target}, []byte(body.String()),
	)
}
