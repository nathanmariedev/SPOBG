package spoAPI

import (
	"SPOBG/http-client"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

const refreshTokenFile = "refresh_token.json"
const clientId = "81a807a412c14e34bf6dbd81633c4ef6"
const clientSecret = "0d68154389b24ce4aea35541e6cfbdd4"
const spotifyApi = "https://api.spotify.com/v1/"
const spotifyAuthentification = "https://accounts.spotify.com/api/token"

var serverStarted = false

func storeRefreshToken(refreshToken string) error {
	tokenData := map[string]string{"refresh_token": refreshToken}
	file, err := os.Create(refreshTokenFile)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	return json.NewEncoder(file).Encode(tokenData)
}
func loadRefreshToken() (string, error) {
	file, err := os.Open(refreshTokenFile)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

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
		"client_id":     clientId,
		"code":          code,
		"client_secret": clientSecret,
		"redirect_uri":  "http://localhost:8080/callback",
	}

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	response, err := http_client.PostRequest(spotifyAuthentification, data, headers)
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
	data.Set("client_id", clientId)
	data.Set("client_secret", clientSecret)

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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

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
	authURL := fmt.Sprintf("https://accounts.spotify.com/authorize?client_id=%s&response_type=code&redirect_uri=%s&scope=user-read-playback-state", clientId, "http://localhost:8080/callback")
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
			_, err := w.Write([]byte("Compte connecté avec succès !"))
			if err != nil {
				return
			}
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
	response, code, err := http_client.GetRequest(spotifyApi+"me/player/currently-playing", headers)

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
	artistUrl := fmt.Sprintf("https://api.spotify.com/v1/artists/%s", artistID)

	// Ajouter le token d'accès dans l'en-tête Authorization
	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
	}

	// Envoyer la requête GET pour obtenir les détails de l'artiste
	response, _, err := http_client.GetRequest(artistUrl, headers)
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
