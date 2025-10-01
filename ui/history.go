package ui

import (
	"fmt"
	"golem/storage"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type HistoryPanel struct {
	container     *fyne.Container
	historyList   *widget.List
	searchEntry   *widget.Entry
	db            *storage.DB
	history       []*storage.RequestHistory
	onRequestLoad func(url, method string)
	parentWindow  fyne.Window
}

func NewHistoryPanel(db *storage.DB, onRequestLoad func(url, method string), parentWindow fyne.Window) *HistoryPanel {
	hp := &HistoryPanel{
		db:            db,
		onRequestLoad: onRequestLoad,
		parentWindow:  parentWindow,
		history:       []*storage.RequestHistory{},
	}

	hp.createUI()
	hp.loadHistory()

	return hp
}

func (hp *HistoryPanel) createUI() {
	hp.searchEntry = widget.NewEntry()
	hp.searchEntry.SetPlaceHolder("Search history...")
	hp.searchEntry.OnChanged = func(text string) {
		hp.searchHistory(text)
	}

	searchBar := container.NewBorder(nil, nil, nil,
		widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
			hp.searchHistory(hp.searchEntry.Text)
		}),
		hp.searchEntry,
	)

	hp.historyList = widget.NewList(
		func() int {
			return len(hp.history)
		},
		func() fyne.CanvasObject {
			methodLabel := widget.NewLabel("METHOD")
			methodLabel.TextStyle = fyne.TextStyle{Bold: true}
			urlLabel := widget.NewLabel("https://example.com/api")
			timeLabel := widget.NewLabel("2 min ago")
			statusLabel := widget.NewLabel("200 OK")

			topRow := container.NewHBox(
				methodLabel,
				widget.NewSeparator(),
				statusLabel,
				widget.NewSeparator(),
				timeLabel,
			)

			return container.NewVBox(
				topRow,
				urlLabel,
				widget.NewSeparator(),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i >= len(hp.history) {
				return
			}

			item := hp.history[i]
			cont := o.(*fyne.Container)

			// The structure is: VBox containing [HBox, Label, Separator]
			hbox := cont.Objects[0].(*fyne.Container)
			urlLabel := cont.Objects[1].(*widget.Label)

			// HBox contains [Label, Separator, Label, Separator, Label]
			methodLabel := hbox.Objects[0].(*widget.Label)
			statusLabel := hbox.Objects[2].(*widget.Label)
			timeLabel := hbox.Objects[4].(*widget.Label)

			methodLabel.SetText(item.Method)
			methodLabel.TextStyle = fyne.TextStyle{Bold: true}

			urlLabel.SetText(item.URL)
			statusLabel.SetText(item.ResponseStatus)

			timeLabel.SetText(hp.formatTime(item.Timestamp))
		},
	)

	hp.historyList.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(hp.history) {
			item := hp.history[id]
			hp.onRequestLoad(item.URL, item.Method)
		}
	}

	clearButton := widget.NewButtonWithIcon("Clear History", theme.ContentClearIcon(), func() {
		dialog.ShowConfirm("Clear History",
			"Are you sure you want to clear all request history?",
			func(confirmed bool) {
				if confirmed {
					hp.clearHistory()
				}
			}, hp.parentWindow)
	})

	exportButton := widget.NewButtonWithIcon("Export", theme.DownloadIcon(), func() {
		dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, hp.parentWindow)
				return
			}
			if writer == nil {
				return
			}
			defer writer.Close()

			if err := hp.db.ExportHistory(writer.URI().Path()); err != nil {
				dialog.ShowError(err, hp.parentWindow)
			} else {
				dialog.ShowInformation("Success", "History exported successfully", hp.parentWindow)
			}
		}, hp.parentWindow)
	})

	buttonBar := container.NewHBox(
		clearButton,
		exportButton,
	)

	hp.container = container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Request History", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			searchBar,
		),
		buttonBar,
		nil,
		nil,
		hp.historyList,
	)
}

func (hp *HistoryPanel) loadHistory() {
	history, err := hp.db.GetRequestHistory(100, 0)
	if err != nil {
		dialog.ShowError(err, hp.parentWindow)
		return
	}

	hp.history = history
	hp.historyList.Refresh()
}

func (hp *HistoryPanel) searchHistory(searchTerm string) {
	if searchTerm == "" {
		hp.loadHistory()
		return
	}

	history, err := hp.db.SearchRequestHistory(searchTerm, 100)
	if err != nil {
		dialog.ShowError(err, hp.parentWindow)
		return
	}

	hp.history = history
	hp.historyList.Refresh()
}

func (hp *HistoryPanel) clearHistory() {
	if err := hp.db.ClearRequestHistory(); err != nil {
		dialog.ShowError(err, hp.parentWindow)
		return
	}

	hp.history = []*storage.RequestHistory{}
	hp.historyList.Refresh()
}

func (hp *HistoryPanel) formatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("Jan 2, 2006")
	}
}

func (hp *HistoryPanel) AddToHistory(req *storage.RequestHistory) {
	if err := hp.db.SaveRequestHistory(req); err != nil {
		fmt.Printf("Failed to save request to history: %v\n", err)
		return
	}

	hp.history = append([]*storage.RequestHistory{req}, hp.history...)
	if len(hp.history) > 100 {
		hp.history = hp.history[:100]
	}
	hp.historyList.Refresh()
}

func (hp *HistoryPanel) GetContainer() *fyne.Container {
	return hp.container
}

func (hp *HistoryPanel) Refresh() {
	hp.loadHistory()
}
