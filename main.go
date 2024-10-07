package main

import (
	"SPOBG/front/views"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Box Layout")
	myWindow.Resize(fyne.NewSize(400, 400))

	myWindow.SetContent(views.Home(myWindow))
	myWindow.ShowAndRun()
}
