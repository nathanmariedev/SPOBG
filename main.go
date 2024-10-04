package main

import (
	"SPOBG/browser"
	macosUtils "SPOBG/macos-utils"
	"SPOBG/spoAPI"
	"fmt"
)

func main() {
	// Récupération d'un accessToken grâce au refreshToken
	accessToken, err := spoAPI.GetAccessTokenFromRefreshToken()
	if err != nil {
		// Si le refreshToken ne fonctionne pas, demander la reconnection
		fmt.Println("Échec de la récupération du token d'accès. Veuillez vous reconnecter.")

		code, err := spoAPI.ConnectSpotifyAccount()

		if err != nil {
			fmt.Println("Erreur lors de la connexion à Spotify :", err)
			return
		}

		// Échange du code contre un accessToken et un refreshToken
		accessToken, err = spoAPI.GetAccessToken(code)
		if err != nil {
			fmt.Println("Erreur lors de la récupération du token d'accès :", err)
			return
		}
	}

	// Récupération du morceau joué
	song, err := spoAPI.GetCurrentlyPlayedSong(accessToken)
	if err != nil {
		fmt.Println("Erreur lors de la récupération de la chanson en cours :", err)
	}

	// Exploitation de la réponse
	if item, ok := song["item"].(map[string]interface{}); ok {
		if name, ok := item["name"].(string); ok {
			fmt.Println("Song Name:", name)
		} else {
			fmt.Println("Song name not found or is not a string")
		}
	} else {
		fmt.Println("Item not found or is not a map")
	}

	// Récupération de l'id de l'artiste
	id, _ := spoAPI.GetArtistIdFromCurrent(song)

	// On vérifie si l'image est déja en local
	exists, err := macosUtils.FileExistsInDirectory("/Users/nathan/Documents/WORK/SPOBG/images/", id+".jpeg")
	if err != nil {
		fmt.Println(err)
	}

	if exists {
		fmt.Println("EXISTS")
		err = macosUtils.SetWallpaperMacOS("/Users/nathan/Documents/WORK/SPOBG/images/" + id + ".jpeg")
		return // FIN DU PROGRAMME
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
		return // FIN DU PROGRAMME
	}
}
