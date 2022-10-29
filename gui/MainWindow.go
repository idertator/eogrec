package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

func CreateMainWindow(app fyne.App) fyne.Window {

	toolbar := CreateMainToolbar()
	content := container.NewBorder(&toolbar, nil, nil, nil)

	mainWindow := app.NewWindow("EogRec")
	mainWindow.Resize(fyne.NewSize(640, 480))
	mainWindow.SetContent(content)

	return mainWindow
}
