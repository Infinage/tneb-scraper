package internal

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type fieldKey string

const (
	fieldConsumerNo fieldKey = "Consumer No"
	fieldBillAmt             = "Bill Amt (Rs)"
	fieldDueDate             = "Due Date"
)

type EBBill struct {
	ConsumerNo   string
	ConsumerName string
	BillAmt      float32
	Due          time.Time
}

// extractBills fetches EB bill from tnebnet and maps the consumer number against
// provided mapping. Always captures the final page screenshot into './tmp/screenshots'
func ExtractBills(login, username, password string, mapping map[string]string) (
	bills []EBBill, err error) {

	browser := rod.New().MustConnect()
	page := browser.MustPage(login).Timeout(10 * time.Second)
	defer browser.MustClose()

	// Always screenshot the final state to './tmp/screenshots/<timestamp>.png'
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("scrape fail: %v", r)	
		}

		img, scErr := page.Screenshot(true, nil)
		if scErr != nil {
			err = fmt.Errorf("%w (screenshot failed: %w)", err, scErr)
			return
		}

		now := time.Now().UTC().Format(time.RFC3339)
		scErr = os.WriteFile("tmp/" + now + ".png", img, 0644)
		if scErr != nil {
			err = fmt.Errorf("%w (screenshot create: %w)", err, scErr)
			return
		}
	}()

	captchaElem := page.MustElement("img#CaptchaImgID")
	img, err := captchaElem.Screenshot(proto.PageCaptureScreenshotFormatPng, 0)
	if err != nil {
		return nil, fmt.Errorf("screenshot captcha fail: %w", err)
	}

	captcha, err := extactCaptcha(img)
	if err != nil || captcha == "" {
		return nil, fmt.Errorf("extactCaptcha fail: %w", err)
	}

	// Enter credentials and login
	page.MustElement("input#userName").MustInput(username)
	page.MustElement("input#password").MustInput(password)
	page.MustElement("input#CaptchaID").MustInput(captcha)
	page.MustElement("input[type='submit']").MustClick()

	// Extact the table of interest
	legend := page.MustElementR("legend", "^Bill Payments$")
	table := legend.MustParent().MustElement("table")

	// Figure out indices for 'Consumer No', 'Bill Amt (Rs)', 'Due Date'
	headers := make(map[string]int)
	headerTr := table.MustElement("thead").MustElement("tr[role='row']")
	for idx, header := range headerTr.MustElements("th") {
		headerText := strings.TrimSpace(header.MustText())
		headers[headerText] = idx
	}

	// Ensure all required headers are present
	for _, required := range []fieldKey{fieldConsumerNo, fieldBillAmt, fieldDueDate} {
		if _, ok := headers[string(required)]; !ok {
			return nil, fmt.Errorf("missing mandatory header: %q", required)
		}
	}

	for _, row := range table.MustElement("tbody").MustElements("tr") {
		rows := row.MustElements("td")
		if len(rows) == 1 && strings.Contains(rows[0].MustText(), "No records found") {
			break // No pending bills found
		}

		if want, got := len(rows), len(headers); want != got {
			return nil, fmt.Errorf("expected %d columns, got %d from table body", want, got)
		}

		consumerNo := rows[headers[string(fieldConsumerNo)]].MustText()

		billAmtStr := rows[headers[string(fieldBillAmt)]].MustText()
		billAmt, err := strconv.ParseFloat(billAmtStr, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid bill amount %q: %w", billAmtStr, err)
		}

		dueStr := rows[headers[string(fieldDueDate)]].MustText()
		due, err := time.Parse("02-01-06", dueStr)
		if err != nil {
			return nil, fmt.Errorf("invalid due date %q: %w", dueStr, err)
		}

		bills = append(bills, EBBill{ConsumerNo: consumerNo, BillAmt: float32(billAmt), Due: due})
	}

	// Map consumer name from the provided mapping, blank if no mapping found
	for idx := range bills {
		bills[idx].ConsumerName = mapping[bills[idx].ConsumerNo]
	}

	return bills, nil
}
