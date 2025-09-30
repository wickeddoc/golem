package main

import (
	"encoding/json"
	"fmt"
	"golem/storage"
	"golem/ui"
	"io"
	"net/http"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type AppPreferences struct {
	WindowWidth  float32
	WindowHeight float32
	LastURL      string
	LastMethod   string
}

type ResponseHeader struct {
	Key   string
	Value string
}

type ResponseInfo struct {
	Body         string
	Headers      []ResponseHeader
	Status       string
	Size         int
	ResponseTime time.Duration
}

func loadPreferencesFromDB(db *storage.DB) *AppPreferences {
	prefs := &AppPreferences{
		WindowWidth:  800,
		WindowHeight: 600,
		LastMethod:   "GET",
	}

	allPrefs, err := db.GetAllPreferences()
	if err != nil {
		fmt.Printf("Error loading preferences: %v\n", err)
		return prefs
	}

	if width, ok := allPrefs["window_width"]; ok {
		if w, err := strconv.ParseFloat(width, 32); err == nil {
			prefs.WindowWidth = float32(w)
		}
	}

	if height, ok := allPrefs["window_height"]; ok {
		if h, err := strconv.ParseFloat(height, 32); err == nil {
			prefs.WindowHeight = float32(h)
		}
	}

	if url, ok := allPrefs["last_url"]; ok {
		prefs.LastURL = url
	}

	if method, ok := allPrefs["last_method"]; ok {
		prefs.LastMethod = method
	}

	return prefs
}

func savePreferencesToDB(db *storage.DB, prefs *AppPreferences) {
	db.SetPreference("window_width", fmt.Sprintf("%f", prefs.WindowWidth))
	db.SetPreference("window_height", fmt.Sprintf("%f", prefs.WindowHeight))
	db.SetPreference("last_url", prefs.LastURL)
	db.SetPreference("last_method", prefs.LastMethod)
}

func executeRequest(method, url string) (*ResponseInfo, error) {
	startTime := time.Now()

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	responseTime := time.Since(startTime)

	headers := make([]ResponseHeader, 0)
	for key, values := range resp.Header {
		for _, value := range values {
			headers = append(headers, ResponseHeader{key, value})
		}
	}

	return &ResponseInfo{
		Body:         string(body),
		Headers:      headers,
		Status:       resp.Status,
		Size:         len(body),
		ResponseTime: responseTime,
	}, nil
}

func main() {
	a := app.New()
	w := a.NewWindow("Golem - API Tester")

	// Initialize database
	db, err := storage.GetDB()
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		dialog.ShowError(err, w)
		return
	}
	defer db.Close()

	// Load preferences from database
	prefs := loadPreferencesFromDB(db)

	// Set a window close handler to save preferences
	w.SetCloseIntercept(func() {
		size := w.Canvas().Size()
		prefs.WindowWidth = size.Width
		prefs.WindowHeight = size.Height

		savePreferencesToDB(db, prefs)
		w.Close()
	})

	methodDropdown := widget.NewSelect(
		[]string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		func(value string) {
			prefs.LastMethod = value
			savePreferencesToDB(db, prefs)
		},
	)
	if prefs.LastMethod != "" {
		methodDropdown.SetSelected(prefs.LastMethod)
	} else {
		methodDropdown.SetSelected("GET")
	}

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("Enter URL...")
	if prefs.LastURL != "" {
		urlEntry.SetText(prefs.LastURL)
	}
	urlEntry.OnChanged = func(text string) {
		prefs.LastURL = text
		savePreferencesToDB(db, prefs)
	}

	statusLabel := widget.NewLabel("Status: -")
	sizeLabel := widget.NewLabel("Size: -")
	timeLabel := widget.NewLabel("Time: -")

	statsRow := container.NewGridWithColumns(3,
		statusLabel,
		sizeLabel,
		timeLabel,
	)

	responseArea := widget.NewMultiLineEntry()
	responseArea.Disable()
	responseArea.SetText("Response will appear here...")

	responseScroll := container.NewScroll(responseArea)
	responseScroll.SetMinSize(fyne.NewSize(600, 400))

	// Create history panel
	var historyPanel *ui.HistoryPanel
	onRequestLoad := func(url, method string) {
		urlEntry.SetText(url)
		methodDropdown.SetSelected(method)
	}
	historyPanel = ui.NewHistoryPanel(db, onRequestLoad, w)

	// Extract submit logic into a function for reuse
	submitRequest := func() {
		url := urlEntry.Text
		method := methodDropdown.Selected

		if url == "" {
			responseArea.SetText("Error: Please enter a URL")
			statusLabel.SetText("Status: Error")
			sizeLabel.SetText("Size: -")
			timeLabel.SetText("Time: -")
			return
		}

		responseArea.SetText("Loading...")
		statusLabel.SetText("Status: Loading...")
		sizeLabel.SetText("Size: -")
		timeLabel.SetText("Time: -")

		go func() {
			response, err := executeRequest(method, url)

			// Create history entry
			historyEntry := &storage.RequestHistory{
				URL:       url,
				Method:    method,
				Timestamp: time.Now(),
			}

			if err != nil {
				historyEntry.ResponseStatus = "Error"
				responseText := fmt.Sprintf("Error: %v", err)

				// Use main thread for UI updates
				responseArea.SetText(responseText)
				statusLabel.SetText("Status: Error")
				sizeLabel.SetText("Size: -")
				timeLabel.SetText("Time: -")
			} else {
				// Update history entry with response data
				historyEntry.ResponseStatus = response.Status
				historyEntry.ResponseBody = response.Body
				historyEntry.ResponseTimeMs = int(response.ResponseTime.Milliseconds())
				historyEntry.ResponseSize = response.Size

				headersJSON, _ := json.Marshal(response.Headers)
				historyEntry.ResponseHeaders = string(headersJSON)

				// Use main thread for UI updates
				responseArea.SetText(response.Body)
				statusLabel.SetText(fmt.Sprintf("Status: %s", response.Status))
				sizeLabel.SetText(fmt.Sprintf("Size: %d bytes", response.Size))
				timeLabel.SetText(fmt.Sprintf("Time: %.2f ms", float64(response.ResponseTime.Milliseconds())))
			}

			// Add to history
			historyPanel.AddToHistory(historyEntry)
		}()
	}

	submitButton := widget.NewButton("Submit", submitRequest)

	topBar := container.NewBorder(
		nil,
		nil,
		methodDropdown,
		submitButton,
		urlEntry,
	)

	topSection := container.NewVBox(
		topBar,
		statsRow,
	)

	// Create main content with split view
	mainContent := container.NewBorder(
		topSection,
		nil,
		nil,
		nil,
		responseScroll,
	)

	// Create split container with history panel on the left
	content := container.NewHSplit(
		historyPanel.GetContainer(),
		mainContent,
	)
	content.SetOffset(0.3) // History panel takes 30% of the width

	w.SetContent(content)

	// Set up keyboard shortcuts using desktop.CustomShortcut
	ctrlEnterShortcut := &desktop.CustomShortcut{
		KeyName:  fyne.KeyReturn,
		Modifier: fyne.KeyModifierControl,
	}
	w.Canvas().AddShortcut(ctrlEnterShortcut, func(shortcut fyne.Shortcut) {
		submitRequest()
	})

	// Also support Ctrl+Enter with the Enter key (numpad)
	ctrlEnterNumpad := &desktop.CustomShortcut{
		KeyName:  fyne.KeyEnter,
		Modifier: fyne.KeyModifierControl,
	}
	w.Canvas().AddShortcut(ctrlEnterNumpad, func(shortcut fyne.Shortcut) {
		submitRequest()
	})

	// F6: Focus URL field (like browsers)
	w.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		if key.Name == fyne.KeyF6 {
			// Move cursor to end of text first
			urlEntry.CursorColumn = len([]rune(urlEntry.Text))
			urlEntry.Refresh()
			// Then focus the entry
			w.Canvas().Focus(urlEntry)
		}
	})

	w.Resize(fyne.NewSize(prefs.WindowWidth, prefs.WindowHeight))
	w.ShowAndRun()
}
