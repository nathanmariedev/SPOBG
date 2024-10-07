package views

import (
	"SPOBG/front"
	"SPOBG/front/components"
	"SPOBG/spoAPI"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"time"
)

func Home(window fyne.Window) *fyne.Container {
	// TITLE
	title := canvas.NewText("SPOBG", color.White)
	title.Alignment = fyne.TextAlignCenter
	title.TextSize = 60

	// BUTTON SPOTIFY
	buttonSpotify := widget.NewButton("Spotify", func() {
		window.SetContent(SpotifyPage())
	})

	customLayout := &front.CustomLayout{
		MarginTop:    20,
		MarginBottom: 20,
		MarginLeft:   40,
		MarginRight:  20,
		Gap:          15,
		Direction:    "vertical",
		MinHeight:    400,
	}

	return container.New(customLayout, title, buttonSpotify)
}

func SpotifyPage() *fyne.Container {
	accessToken, err := spoAPI.GetAccessTokenFromRefreshToken()
	if err != nil {
		fmt.Println("Get access token fail")
	}

	user, err := spoAPI.GetUser(accessToken)
	if err != nil {
		fmt.Println("Get user fail")
	}

	fmt.Println("user:", user.DisplayName)

	// Welcome
	welcome := canvas.NewText("Welcome, "+user.DisplayName, color.White)
	welcome.TextSize = 30
	welcome.TextStyle.Bold = true
	customLayout := &front.CustomLayout{
		MarginTop:    20,
		MarginBottom: 20,
		MarginLeft:   40,
		MarginRight:  20,
		Gap:          15,
		Direction:    "vertical",
		MinHeight:    400,
		MaxHeight:    400,
		MaxWidth:     600,
	}
	songPlayingData, _ := spoAPI.GetCurrentlyPlayedSong(accessToken)
	songPlayingComponent, songName, songArtists, songImage, backgroundButton, playingSong := components.PlayingSong(songPlayingData)

	// Créer un ticker qui émet un événement toutes les 10 secondes pour actualiser la chanson en cours
	ticker := time.NewTicker(3 * time.Second)
	quit := make(chan struct{})

	// Lancer le polling dans une goroutine
	go func() {
		for {
			select {
			case <-ticker.C:
				// Effectuer un appel à l'API pour vérifier la chanson en cours
				song, err := spoAPI.GetCurrentlyPlayedSong(accessToken)
				if err != nil {
					fmt.Println("Erreur lors de l'appel API :", err)
					continue
				}

				components.UpdatePlayingSong(songName, songArtists, songImage, backgroundButton, playingSong, song)

			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	return container.New(customLayout, welcome, songPlayingComponent)
}
