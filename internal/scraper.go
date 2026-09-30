package internal

import (
	"errors"
	"fmt"
	"log"
	"net/url"
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

// scraperOnExit recovers the stack if it had panicked and screenshots
// the browser state into './tmp/screenshots'.
func scraperOnExit(page *rod.Page, err *error) {
	r := recover()
	if r != nil {
		*err = fmt.Errorf("scrape fail: %v", r)
	}

	// Wrapping the screenshot error with the error passed in
	var scErr error
	defer func() {
		if scErr == nil {
			return
		}

		if *err != nil {
			*err = errors.Join(*err, scErr)
		} else {
			*err = scErr
		}
	}()

	// Screenshot latest status
	var img []byte
	img, scErr = page.CancelTimeout().Screenshot(true, nil)
	if scErr != nil {
		scErr = fmt.Errorf("screenshot failed: %w", scErr)
		return
	}

	// Attempt to create a temp directory
	scErr = os.MkdirAll("tmp/screenshots/", 0755)
	if scErr != nil {
		scErr = fmt.Errorf("screen dir create: %w", scErr)
		return
	}

	// Save into 'tmp/screenshots/<timestamp>.png'
	now := time.Now().UTC().Format("20060102T150405.000000000Z")
	scErr = os.WriteFile("tmp/screenshots/"+now+".png", img, 0644)
	if scErr != nil {
		scErr = fmt.Errorf("screen create: %w", scErr)
		return
	}
}

func attemptLogin(page *rod.Page, username, password string) (success bool, err error) {
	// Wait for the page network activity to settle
	page.MustWaitLoad().MustWaitIdle()

	captchaElem := page.MustElement("img#CaptchaImgID")
	box := captchaElem.MustShape().Box()
	clip := &proto.PageViewport{
		X: box.X + 3, Y: box.Y + 3,
		Width:  box.Width - 6,
		Height: box.Height - 6,
		Scale:  1,
	}
	img, err := page.Screenshot(false, &proto.PageCaptureScreenshot{Clip: clip})
	if err != nil {
		return false, fmt.Errorf("screenshot captcha fail: %w", err)
	}

	captcha, err := extractCaptcha(img)
	if err != nil {
		return false, fmt.Errorf("extractCaptcha fail: %w", err)
	} else if captcha == "" {
		return false, fmt.Errorf("extractCaptcha returned empty")
	}

	// Enter credentials and captcha
	page.MustElement("input#userName").MustSelectAllText().MustInput(username)
	page.MustElement("input#password").MustSelectAllText().MustInput(password)
	page.MustElement("input#CaptchaID").MustSelectAllText().MustInput(captcha)

	// Login and wait for navigation
	wait := page.MustWaitNavigation()
	page.MustElement("input[type='submit']").MustClick()
	wait()

	urlStr := page.MustInfo().URL
	url, err := url.Parse(urlStr)
	if err != nil {
		return false, fmt.Errorf("failed to parse page url: %s", urlStr)
	}

	return !url.Query().Has("login_error"), nil
}

// extractBills fetches EB bill from tnebnet and maps the consumer number against
// provided mapping. Always captures the final page screenshot into './tmp/screenshots'
func extractBills(rc runConfig) (
	bills []EBBill, err error) {

	// Connect to standalone 'rod' container if env var 'ROD_URL' is set
	var browser *rod.Browser
	rodURL := os.Getenv("ROD_URL")
	if rodURL != "" {
		browser = rod.New().ControlURL(rodURL).MustConnect()
	} else {
		browser = rod.New().MustConnect()
	}

	page := browser.MustPage(rc.ebURL).Timeout(time.Second * 120)
	defer browser.MustClose()

	// Always screenshot the final state to './tmp/screenshots/<timestamp>.png'
	defer scraperOnExit(page, &err)

	// Reattempt captcha utmost 3 times on failure
	var loggedIn bool
	for range rc.retryAttempts {
		loggedIn, err = attemptLogin(page, rc.ebUser, rc.ebPass)
		if err != nil {
			return nil, fmt.Errorf("fatal login error: %w", err)
		} else if loggedIn {
			break
		}

		// refreshes captcha
		page.Reload()
	}

	if !loggedIn {
		return nil, fmt.Errorf("login fail after %d attempts", rc.retryAttempts)
	}

	// Logged in, proceed to extract the table of interest
	log.Println("Login successful")
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
		billAmtStr = strings.ReplaceAll(strings.TrimPrefix(billAmtStr, "Rs."), ",", "")
		billAmt, err := strconv.ParseFloat(strings.TrimSpace(billAmtStr), 32)
		if err != nil {
			return nil, fmt.Errorf("invalid bill amount %q: %w", billAmtStr, err)
		}

		dueStr := rows[headers[string(fieldDueDate)]].MustText()
		due, err := time.Parse("02/01/2006", strings.TrimSpace(dueStr))
		if err != nil {
			return nil, fmt.Errorf("invalid due date %q: %w", dueStr, err)
		}

		bills = append(bills, EBBill{ConsumerNo: consumerNo, BillAmt: float32(billAmt), Due: due})
	}

	// Map consumer name from the provided mapping, blank if no mapping found
	for idx := range bills {
		bills[idx].ConsumerName = rc.mapping[bills[idx].ConsumerNo]
	}

	return bills, nil
}
