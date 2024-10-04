package spoAPI

import (
	"SPOBG/http_client"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"github.com/chromedp/chromedp"
	"log"
	"net/http"
	"net/url"
	"os"
)

const refreshTokenFile = "refresh_token.json"
const CLIENT_ID = "81a807a412c14e34bf6dbd81633c4ef6"
const CLIENT_SECRET = "0d68154389b24ce4aea35541e6cfbdd4"
const SPO_API = "https://api.spotify.com/v1/"
const SPO_AUTH = "https://accounts.spotify.com/api/token"
const SPO_AUTH_CMD = `
	curl -X POST "https://accounts.spotify.com/api/token" \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -d "grant_type=client_credentials&client_id=81a807a412c14e34bf6dbd81633c4ef6&client_secret=0d68154389b24ce4aea35541e6cfbdd4&code=AQBbpm3eDdFRK1MHpNpNZUIEbmxd12ItmdpTPyHd6l6Vddx95zIignOfKLpCHYG4fboByg9pMZEzG0q6WPZFmzzplzZ_Rnyl8_zGUmsasw01CpXCCcVxe6bmtTxcVlf8G6KwQSGIwarZrM--rblc4ytRw2imJIi21f-99tm3w7KcBjD1cIBNJtek6VPR2YNImrz2pHZl2W-zlUcL"

`

var serverStarted bool = false

func storeRefreshToken(refreshToken string) error {
	tokenData := map[string]string{"refresh_token": refreshToken}
	file, err := os.Create(refreshTokenFile)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(tokenData)
}
func loadRefreshToken() (string, error) {
	file, err := os.Open(refreshTokenFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var tokenData map[string]string
	err = json.NewDecoder(file).Decode(&tokenData)
	if err != nil {
		return "", err
	}

	return tokenData["refresh_token"], nil
}

func GetAccessToken(code string) (string, error) {
	data := map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     CLIENT_ID,
		"code":          code,
		"client_secret": CLIENT_SECRET,
		"redirect_uri":  "http://localhost:8080/callback",
	}

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	response, err := http_client.PostRequest(SPO_AUTH, data, headers)
	if err != nil {
		return "", fmt.Errorf("erreur lors de la requête POST: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return "", fmt.Errorf("erreur lors de la désérialisation de la réponse: %v", err)
	}

	token, ok := result["access_token"].(string)

	if !ok {
		return "", fmt.Errorf("token d'accès non trouvé dans la réponse")
	}

	refreshToken, _ := result["refresh_token"].(string)

	err = storeRefreshToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("erreur lors de la sauvegarde du refresh token: %v", err)
	}

	return token, nil
}
func GetAccessTokenFromRefreshToken() (string, error) {
	// Charger le refresh token depuis le fichier
	refreshToken, err := loadRefreshToken()
	if err != nil {
		return "", fmt.Errorf("erreur lors du chargement du refresh token: %v", err)
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", CLIENT_ID)
	data.Set("client_secret", CLIENT_SECRET)

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("token d'accès non trouvé dans la réponse")
	}

	return accessToken, nil
}
func ConnectSpotifyAccount() (string, error) {
	codeChan := make(chan string)

	// URL pour autoriser l'utilisateur
	authURL := fmt.Sprintf("https://accounts.spotify.com/authorize?client_id=%s&response_type=code&redirect_uri=%s&scope=user-read-playback-state", CLIENT_ID, "http://localhost:8080/callback")
	fmt.Println("Visitez cette URL pour autoriser l'application :", authURL)

	// Enregistrer le gestionnaire du callback UNE SEULE FOIS
	if !serverStarted {
		http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
			code := r.URL.Query().Get("code")
			if code == "" {
				http.Error(w, "Code d'autorisation manquant", http.StatusBadRequest)
				return
			}

			// Envoyer le code dans le canal
			codeChan <- code
			w.Write([]byte("Compte connecté avec succès !"))
		})

		// Démarrer le serveur HTTP UNE SEULE FOIS
		go func() {
			err := http.ListenAndServe(":8080", nil)
			if err != nil {
				fmt.Println("Erreur lors du démarrage du serveur HTTP:", err)
			}
		}()

		// Indiquer que le serveur est démarré pour éviter de le redémarrer
		serverStarted = true
	}

	// Attendre la réception du code d'autorisation
	code := <-codeChan

	// Retourner le code une fois qu'il est reçu
	return code, nil
}

func GetCurrentlyPlayedSong(token string) (map[string]interface{}, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}
	response, code, err := http_client.GetRequest(SPO_API+"me/player/currently-playing", headers)

	if err != nil {
		return nil, fmt.Errorf("erreur lors de la requête GET: %v", err)
	}

	switch code {
	case 200:
		fmt.Println("200 SUCCESS")
	case 204:
		fmt.Println("204 NO SONG PLAYING RN")
	case 404:
		fmt.Println("404 OOPS")
	}
	fmt.Println(code)

	// Décoder la réponse JSON
	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la désérialisation de la réponse: %v", err)
	}

	return result, nil
}
func GetArtistIdFromCurrent(currentSong map[string]interface{}) (string, error) {
	// Accéder aux informations de la chanson
	item, ok := currentSong["item"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("aucune chanson en cours de lecture ou structure inattendue")
	}

	// Récupérer la liste des artistes associés à la chanson
	artists, ok := item["artists"].([]interface{})
	if !ok || len(artists) == 0 {
		return "", fmt.Errorf("aucun artiste trouvé pour la chanson en cours de lecture")
	}

	// Extraire l'ID du premier artiste (l'artiste principal)
	firstArtist, ok := artists[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("structure d'artiste inattendue")
	}

	artistID, ok := firstArtist["id"].(string)
	if !ok {
		return "", fmt.Errorf("ID de l'artiste non trouvé")
	}

	return artistID, nil
}
func GetArtistDetails(artistID string, accessToken string) (map[string]interface{}, error) {
	// Construire l'URL pour obtenir les informations de l'artiste
	url := fmt.Sprintf("https://api.spotify.com/v1/artists/%s", artistID)

	// Ajouter le token d'accès dans l'en-tête Authorization
	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
	}

	// Envoyer la requête GET pour obtenir les détails de l'artiste
	response, _, err := http_client.GetRequest(url, headers)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la requête GET pour l'artiste: %v", err)
	}

	// Décoder la réponse JSON
	var result map[string]interface{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la désérialisation de la réponse: %v", err)
	}

	return result, nil
}

func extractBackgroundImageURL(styleContent string) string {
	// Regex pour trouver l'URL dans le background-image
	re := regexp.MustCompile(`background-image:\s*url\(["']?(.*?)["']?\);`)
	matches := re.FindStringSubmatch(styleContent)

	if len(matches) > 1 {
		return matches[1] // Retourne l'URL extraite
	}
	return ""
}
func ScrapeBackgroundImageDiv(url string) string {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Créer un contexte avec un timeout plus long pour éviter l'expiration prématurée
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var styleContent string

	// Exécuter les actions de scraping avec des étapes séparées
	err := chromedp.Run(ctx,
		// Naviguer vers l'URL spécifiée
		chromedp.Navigate(url),

		// Attendre que l'élément principal de la page soit prêt
		chromedp.WaitReady(`body`, chromedp.ByQuery),

		// Attendre que l'élément avec la classe "under-main-view" soit visible
		chromedp.WaitVisible(`.under-main-view`, chromedp.ByQuery),

		// Ajouter un délai pour garantir le chargement du style
		chromedp.Sleep(1*time.Second),

		// Récupérer l'attribut "style" de l'élément avec `data-testid="background-image"`
		chromedp.AttributeValue(`[data-testid="background-image"]`, "style", &styleContent, nil),
	)

	if err != nil {
		log.Fatalf("Erreur lors de l'exécution de Chromedp : %v", err)
	}

	backgroundImageURL := extractBackgroundImageURL(styleContent)
	if backgroundImageURL != "" {
		return backgroundImageURL
	} else {
		fmt.Println("Aucune URL trouvée dans le style.")
	}

	return ""
}

func DownloadImage(url, filepath string) error {
	// Envoyer une requête GET pour télécharger l'image
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("erreur lors du téléchargement de l'image: %v", err)
	}
	defer resp.Body.Close()

	// Vérifier que la requête est réussie
	if resp.StatusCode != 200 {
		return fmt.Errorf("échec du téléchargement de l'image, statut HTTP: %d", resp.StatusCode)
	}

	// Créer le fichier pour sauvegarder l'image
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("erreur lors de la création du fichier: %v", err)
	}
	defer out.Close()

	// Copier le contenu téléchargé dans le fichier
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("erreur lors de l'enregistrement de l'image: %v", err)
	}

	return nil
}

// Fonction pour changer le fond d'écran sur macOS avec AppleScript
func SetWallpaperMacOS(filepath string) error {
	// Vérifier si l'image existe
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return fmt.Errorf("le fichier d'image n'existe pas: %s", filepath)
	}
	// Définir le chemin vers le fichier .plist
	//plistPath := "/Users/nathan/Library/Application Support/com.apple.wallpaper/Store/Index.plist"

	//cmd := exec.Command("killall", "WallpaperAgent")

	//cmd := exec.Command("wallpaper", filepath)
	//cmd := exec.Command("wal", "-i", filepath)
	script := fmt.Sprintf(`tell application "System Events" to tell every desktop to set picture to POSIX file "%s"`, filepath)
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Erreur : %s\n", err)
		fmt.Printf("Sortie : %s\n", string(output))
	}
	fmt.Println("CHANGED")
	return nil
}
func FileExistsInDirectory(directory string, filename string) (bool, error) {
	var found bool
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Vérifier si le fichier actuel est celui recherché
		if !info.IsDir() && info.Name() == filename {
			found = true
			// Retourner filepath.SkipDir pour arrêter la recherche après avoir trouvé le fichier
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil && err != filepath.SkipDir {
		return false, err
	}
	return found, nil
}
