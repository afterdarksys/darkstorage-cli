package main

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/darkstorage/cli/internal/config"
	"github.com/darkstorage/cli/internal/ipc"
)

type App struct {
	app       fyne.App
	window    fyne.Window
	ipcClient *ipc.Client

	// UI components
	statusLabel     *widget.Label
	queueSizeLabel  *widget.Label
	activityList    *widget.List
	foldersList     *widget.List

	// State
	status          *ipc.StatusResponse
	activities      []ipc.ActivityEntry
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Dark Storage")
	myWindow.Resize(fyne.NewSize(900, 600))

	dataDir, err := config.GetDefaultDataDir()
	if err != nil {
		log.Fatalf("Failed to get data directory: %v", err)
	}

	socketPath := filepath.Join(dataDir, "daemon.sock")
	client := ipc.NewClient(socketPath)

	application := &App{
		app:       myApp,
		window:    myWindow,
		ipcClient: client,
		activities: []ipc.ActivityEntry{},
	}

	application.makeUI()
	application.refreshStatus()

	// Auto-refresh every 5 seconds
	go application.autoRefresh()

	myWindow.ShowAndRun()
}

func (a *App) makeUI() {
	// Left navigation
	nav := container.NewVBox(
		widget.NewButton("Dashboard", func() {
			a.window.SetContent(a.makeDashboard())
		}),
		widget.NewButton("Sync Folders", func() {
			a.window.SetContent(a.makeSyncFolders())
		}),
		widget.NewButton("Activity", func() {
			a.window.SetContent(a.makeActivity())
		}),
		widget.NewButton("Settings", func() {
			a.window.SetContent(a.makeSettings())
		}),
	)

	content := a.makeDashboard()

	split := container.NewHSplit(
		container.NewVBox(nav),
		content,
	)
	split.SetOffset(0.15)

	a.window.SetContent(split)
}

func (a *App) makeDashboard() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Dark Storage", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.TextStyle.Bold = true

	a.statusLabel = widget.NewLabel("Checking status...")
	a.queueSizeLabel = widget.NewLabel("Queue: 0")

	statusCard := widget.NewCard("Status", "", container.NewVBox(
		a.statusLabel,
		a.queueSizeLabel,
	))

	recentActivity := widget.NewLabel("Recent Activity:")
	a.activityList = widget.NewList(
		func() int {
			return len(a.activities)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(a.activities) {
				activity := a.activities[id]
				label := obj.(*widget.Label)
				label.SetText(fmt.Sprintf("%s: %s (%s)",
					activity.Operation, activity.Path, activity.Status))
			}
		},
	)

	activityCard := widget.NewCard("Recent Activity", "", container.NewVBox(
		recentActivity,
		container.NewMax(a.activityList),
	))

	refreshBtn := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		a.refreshStatus()
		a.refreshActivity()
	})

	content := container.NewBorder(
		container.NewVBox(title, widget.NewSeparator()),
		refreshBtn,
		nil,
		nil,
		container.NewVBox(
			statusCard,
			activityCard,
		),
	)

	return content
}

func (a *App) makeSyncFolders() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Sync Folders", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	var folders []ipc.SyncFolderStatus
	if a.status != nil {
		folders = a.status.SyncFolders
	}

	a.foldersList = widget.NewList(
		func() int {
			return len(folders)
		},
		func() fyne.CanvasObject {
			return container.NewVBox(
				widget.NewLabel(""),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(folders) {
				folder := folders[id]
				box := obj.(*fyne.Container)
				nameLabel := box.Objects[0].(*widget.Label)
				pathLabel := box.Objects[1].(*widget.Label)
				nameLabel.SetText(fmt.Sprintf("%s [%s]", folder.Name, folder.Status))
				pathLabel.SetText(fmt.Sprintf("%s → %s", folder.LocalPath, folder.RemotePath))
			}
		},
	)

	addBtn := widget.NewButton("Add Folder", func() {
		a.showAddFolderDialog()
	})

	content := container.NewBorder(
		container.NewVBox(title, widget.NewSeparator()),
		addBtn,
		nil,
		nil,
		a.foldersList,
	)

	return content
}

func (a *App) makeActivity() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Activity Log", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	activityList := widget.NewList(
		func() int {
			return len(a.activities)
		},
		func() fyne.CanvasObject {
			return container.NewVBox(
				widget.NewLabel(""),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(a.activities) {
				activity := a.activities[id]
				box := obj.(*fyne.Container)
				mainLabel := box.Objects[0].(*widget.Label)
				detailLabel := box.Objects[1].(*widget.Label)
				mainLabel.SetText(fmt.Sprintf("%s: %s", activity.Operation, activity.Path))
				detailLabel.SetText(fmt.Sprintf("Status: %s | %s", activity.Status, activity.Timestamp.Format(time.RFC822)))
			}
		},
	)

	refreshBtn := widget.NewButton("Refresh", func() {
		a.refreshActivity()
	})

	content := container.NewBorder(
		container.NewVBox(title, widget.NewSeparator()),
		refreshBtn,
		nil,
		nil,
		activityList,
	)

	return content
}

func (a *App) makeSettings() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Settings", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	daemonCard := widget.NewCard("Daemon", "", container.NewVBox(
		widget.NewButton("Check Status", func() {
			a.refreshStatus()
		}),
		widget.NewButton("View Logs", func() {
			// TODO: Implement log viewer
		}),
	))

	configCard := widget.NewCard("Configuration", "", container.NewVBox(
		widget.NewButton("Edit Config", func() {
			a.showConfigDialog()
		}),
	))

	content := container.NewBorder(
		container.NewVBox(title, widget.NewSeparator()),
		nil,
		nil,
		nil,
		container.NewVBox(
			daemonCard,
			configCard,
		),
	)

	return content
}

func (a *App) showAddFolderDialog() {
	localEntry := widget.NewEntry()
	localEntry.SetPlaceHolder("/path/to/local/folder")

	remoteEntry := widget.NewEntry()
	remoteEntry.SetPlaceHolder("bucket/remote/path")

	directionSelect := widget.NewSelect([]string{"bidirectional", "upload_only", "download_only"}, nil)
	directionSelect.SetSelected("bidirectional")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Local Path", Widget: localEntry},
			{Text: "Remote Path", Widget: remoteEntry},
			{Text: "Direction", Widget: directionSelect},
		},
		OnSubmit: func() {
			req := &ipc.AddSyncFolderRequest{
				LocalPath:          localEntry.Text,
				RemotePath:         remoteEntry.Text,
				Direction:          directionSelect.Selected,
				ConflictResolution: "keep_local",
			}

			_, err := a.ipcClient.AddSyncFolder(req)
			if err != nil {
				widget.ShowPopUp(widget.NewLabel(fmt.Sprintf("Error: %v", err)), a.window.Canvas())
			} else {
				a.refreshStatus()
			}
		},
	}

	dialog := widget.NewModalPopUp(
		container.NewVBox(
			widget.NewLabel("Add Sync Folder"),
			form,
		),
		a.window.Canvas(),
	)
	dialog.Resize(fyne.NewSize(400, 300))
	dialog.Show()
}

func (a *App) showConfigDialog() {
	configResp, err := a.ipcClient.GetConfig()
	if err != nil {
		widget.ShowPopUp(widget.NewLabel(fmt.Sprintf("Error loading config: %v", err)), a.window.Canvas())
		return
	}

	configText := fmt.Sprintf("%+v", configResp.Config)
	textWidget := widget.NewMultiLineEntry()
	textWidget.SetText(configText)

	dialog := widget.NewModalPopUp(
		container.NewVBox(
			widget.NewLabel("Configuration"),
			textWidget,
			widget.NewButton("Close", func() {}),
		),
		a.window.Canvas(),
	)
	dialog.Resize(fyne.NewSize(500, 400))
	dialog.Show()
}

func (a *App) refreshStatus() {
	status, err := a.ipcClient.GetStatus()
	if err != nil {
		a.statusLabel.SetText("Daemon: Not Connected")
		return
	}

	a.status = status
	a.statusLabel.SetText(fmt.Sprintf("Daemon: Running (%s)", status.Uptime))
	a.queueSizeLabel.SetText(fmt.Sprintf("Queue: %d pending", status.QueueSize))
}

func (a *App) refreshActivity() {
	req := &ipc.Command{
		Type: "get_activity",
		Data: []byte(`{"limit": 20}`),
	}

	resp, err := a.ipcClient.SendCommand(req)
	if err != nil {
		return
	}

	if resp.Success {
		var activityResp ipc.GetActivityResponse
		if err := json.Unmarshal(resp.Data, &activityResp); err == nil {
			a.activities = activityResp.Activities
			if a.activityList != nil {
				a.activityList.Refresh()
			}
		}
	}
}

func (a *App) autoRefresh() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		a.refreshStatus()
		a.refreshActivity()
	}
}
