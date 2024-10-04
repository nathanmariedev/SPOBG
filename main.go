package main

import (
	"SPOBG/browser"
	macosUtils "SPOBG/macos-utils"
	"SPOBG/spoAPI"
	"fmt"
)

func main() {
	accessToken, err := spoAPI.GetAccessTokenFromRefreshToken()
	if err != nil {
		// Si le refresh token échoue, l'utilisateur doit se reconnecter
		fmt.Println("Échec de la récupération du token d'accès. Veuillez vous reconnecter.")

		// Si l'utilisateur doit se reconnecter, on utilise ConnectSpotifyAccount pour obtenir un nouveau code
		code, err := spoAPI.ConnectSpotifyAccount()
		if err != nil {
			fmt.Println("Erreur lors de la connexion à Spotify :", err)
			return
		}

		// Échanger le code contre un token d'accès et un refresh token
		accessToken, err = spoAPI.GetAccessToken(code)
		if err != nil {
			fmt.Println("Erreur lors de la récupération du token d'accès :", err)
			return
		}
	}

	// Utiliser le token d'accès pour récupérer la chanson actuellement en lecture
	song, err := spoAPI.GetCurrentlyPlayedSong(accessToken)
	if err != nil {
		fmt.Println("Erreur lors de la récupération de la chanson en cours :", err)
	}
	if item, ok := song["item"].(map[string]interface{}); ok {
		// Access the "name" field inside "item"
		if name, ok := item["name"].(string); ok {
			fmt.Println("Song Name:", name)
		} else {
			fmt.Println("Song name not found or is not a string")
		}
	} else {
		fmt.Println("Item not found or is not a map")
	}

	id, _ := spoAPI.GetArtistIdFromCurrent(song)

	exists, err := macosUtils.FileExistsInDirectory("/Users/nathan/Documents/WORK/SPOBG/images/", id+".jpeg")
	if err != nil {
		fmt.Println(err)
	}

	if exists {
		fmt.Println("EXISTS")
		err = macosUtils.SetWallpaperMacOS("/Users/nathan/Documents/WORK/SPOBG/images/" + id + ".jpeg")
		return
	}

	url := "https://open.spotify.com/intl-pt/artist/" + id

	img := browser.ScrapeBackgroundImageDiv(url)

	err = browser.DownloadImage(img, "images/"+id+".jpeg")
	if err != nil {
		return
	}
	fmt.Println("DOWNLOADING")

	// Changer le fond d'écran sur macOS
	err = macosUtils.SetWallpaperMacOS("/Users/nathan/Documents/WORK/SPOBG/images/" + id + ".jpeg")
	if err != nil {
		fmt.Printf("Erreur lors du changement du fond d'écran: %v\n", err)
		return
	}
}
