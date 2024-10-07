package components

import (
	"SPOBG/browser"
	"SPOBG/front"
	macos_utils "SPOBG/macos-utils"
	"SPOBG/spoAPI"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

var layoutFinal = &front.CustomLayout{
	MarginTop:    20,
	MarginBottom: 20,
	MarginLeft:   20,
	MarginRight:  20,
	Gap:          25,
	Direction:    "horizontal",
	MinHeight:    50,
}

func PlayingSong(playingSong *spoAPI.SpotifyData) (*fyne.Container, *canvas.Text, *canvas.Text, *canvas.Image, *widget.Button, *spoAPI.SpotifyData) {
	rightComponent := &front.CustomLayout{
		MarginTop:    0,
		MarginBottom: 0,
		MarginLeft:   0,
		MarginRight:  0,
		Gap:          30,
		Direction:    "vertical",
		MinWidth:     0,
		MinHeight:    0,
		MaxWidth:     0,
		MaxHeight:    0,
	}
	textComponent := &front.CustomLayout{
		MarginTop:    0,
		MarginBottom: 0,
		MarginLeft:   0,
		MarginRight:  0,
		Gap:          0,
		Direction:    "vertical",
	}

	// SONG NAME
	songName := canvas.NewText(front.TruncateText(playingSong.Item.Name, 22, (layoutFinal.MaxWidth)), color.White)
	songName.TextSize = 22

	// SONG ARTIST
	artists := ""
	if len(playingSong.Item.Artists) > 1 {
		for _, artist := range playingSong.Item.Artists[1:] {
			artists += ", " + artist.Name
		}
	}
	songArtist := canvas.NewText(playingSong.Item.Artists[0].Name, color.RGBA{R: 255, G: 255, B: 255, A: 150})
	songArtist.TextSize = 15

	// SONG COVER
	exists, err := macos_utils.FileExistsInDirectory("/Users/nathan/Documents/WORK/SPOBG/images/", playingSong.Item.Album.Name+".jpeg")
	if err != nil {
		fmt.Println(err)
	}

	if !exists {
		browser.DownloadImage(playingSong.Item.Album.Images[0].URL, "images/"+playingSong.Item.Album.Name+".jpeg")
	}
	songCover := canvas.NewImageFromFile("images/" + playingSong.Item.Album.Name + ".jpeg")
	songCover.FillMode = canvas.ImageFillContain
	songCover.SetMinSize(fyne.NewSize(110, 110))

	backgroundButton := widget.NewButton("SET BACKGROUND", func() {
		exists, err := macos_utils.FileExistsInDirectory("/Users/nathan/Documents/WORK/SPOBG/images/", playingSong.Item.Artists[0].ID+".jpeg")
		if err != nil {
			panic(err)
		}
		if exists {
			fmt.Println("EXISTS")
			_ = macos_utils.SetWallpaperMacOS("/Users/nathan/Documents/WORK/SPOBG/images/" + playingSong.Item.Artists[0].ID + ".jpeg")
			return
		}
		url := "https://open.spotify.com/intl-pt/artist/" + playingSong.Item.Artists[0].ID

		img, err := browser.ScrapeBackgroundImageDiv(url)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = browser.DownloadImage(img, "images/"+playingSong.Item.Artists[0].ID+".jpeg")
		if err != nil {
			return
		}
		fmt.Println("DOWNLOADING")
	})

	actualWallpaper := macos_utils.GetWallpaperMacOS()

	if actualWallpaper == playingSong.Item.Artists[0].ID+".jpeg" {
		backgroundButton.Disable()
	}

	// TEXT CONTAINER
	textContainer := container.New(textComponent, songName, songArtist)

	// FINAL CONTAINER
	background := canvas.NewRectangle(color.RGBA{R: 70, G: 70, B: 70, A: 255})
	background.CornerRadius = 10

	rightContainer := container.New(rightComponent, textContainer, backgroundButton)
	finalContainer := container.New(layoutFinal, songCover, rightContainer)
	finalContainer = container.NewMax(background, finalContainer)

	backgroundButton.Resize(fyne.NewSize(backgroundButton.MinSize().Width, backgroundButton.MinSize().Height))
	return finalContainer, songName, songArtist, songCover, backgroundButton, playingSong
}
func UpdatePlayingSong(songName *canvas.Text, songArtist *canvas.Text, songCover *canvas.Image, button *widget.Button, previousSong *spoAPI.SpotifyData, actualSong *spoAPI.SpotifyData) {
	// MAJ SONG
	*previousSong = *actualSong

	// MAJ SONG NAME
	songName = canvas.NewText(front.TruncateText(actualSong.Item.Name, 22, (layoutFinal.MaxWidth)), color.White)
	songName.Refresh()

	// MAJ ARTISTS
	artists := actualSong.Item.Artists[0].Name
	if len(actualSong.Item.Artists) > 1 {
		for _, artist := range actualSong.Item.Artists[1:] {
			artists += ", " + artist.Name
		}
	}
	songArtist.Text = artists
	songArtist.Refresh()

	// // MAJ COVER
	browser.DownloadImage(actualSong.Item.Album.Images[0].URL, "images/"+actualSong.Item.Album.Name+".jpeg")
	songCover.File = "images/" + actualSong.Item.Album.Name + ".jpeg"
	songCover.Refresh()

	// MAJ BUTTON
	actualWallpaper := macos_utils.GetWallpaperMacOS()

	if actualWallpaper == actualSong.Item.Artists[0].ID+".jpeg" {
		button.Disable()
	} else {
		button.Enable()
	}
}
