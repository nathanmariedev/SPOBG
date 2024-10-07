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

type User struct {
	DisplayName  string       `json:"display_name"`
	ExternalURLs ExternalURLs `json:"external_urls"`
	Followers    Followers    `json:"followers"`
	Href         string       `json:"href"`
	ID           string       `json:"id"`
	Images       []Image      `json:"images"`
	Type         string       `json:"type"`
	URI          string       `json:"uri"`
}
type ExternalURLs struct {
	Spotify string `json:"spotify"`
}
type Followers struct {
	Href  *string `json:"href"` // Href is a pointer because it can be null
	Total int     `json:"total"`
}
type Image struct {
	Height int    `json:"height"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
}
type SpotifyData struct {
	Device               Device  `json:"device"`
	RepeatState          string  `json:"repeat_state"`
	ShuffleState         bool    `json:"shuffle_state"`
	Context              Context `json:"context"`
	Timestamp            int64   `json:"timestamp"`
	ProgressMs           int64   `json:"progress_ms"`
	IsPlaying            bool    `json:"is_playing"`
	Item                 Item    `json:"item"`
	CurrentlyPlayingType string  `json:"currently_playing_type"`
	Actions              Actions `json:"actions"`
}
type Device struct {
	ID               string `json:"id"`
	IsActive         bool   `json:"is_active"`
	IsPrivateSession bool   `json:"is_private_session"`
	IsRestricted     bool   `json:"is_restricted"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	VolumePercent    int    `json:"volume_percent"`
	SupportsVolume   bool   `json:"supports_volume"`
}
type Context struct {
	Type         string       `json:"type"`
	Href         string       `json:"href"`
	ExternalURLs ExternalURLs `json:"external_urls"`
	URI          string       `json:"uri"`
}
type Item struct {
	Album            Album        `json:"album"`
	Artists          []Artist     `json:"artists"`
	AvailableMarkets []string     `json:"available_markets"`
	DiscNumber       int          `json:"disc_number"`
	DurationMs       int          `json:"duration_ms"`
	Explicit         bool         `json:"explicit"`
	ExternalIDs      ExternalIDs  `json:"external_ids"`
	ExternalURLs     ExternalURLs `json:"external_urls"`
	Href             string       `json:"href"`
	ID               string       `json:"id"`
	IsPlayable       bool         `json:"is_playable"`
	LinkedFrom       LinkedFrom   `json:"linked_from"`
	Restrictions     Restrictions `json:"restrictions"`
	Name             string       `json:"name"`
	Popularity       int          `json:"popularity"`
	PreviewURL       string       `json:"preview_url"`
	TrackNumber      int          `json:"track_number"`
	Type             string       `json:"type"`
	URI              string       `json:"uri"`
	IsLocal          bool         `json:"is_local"`
}
type Album struct {
	AlbumType            string       `json:"album_type"`
	TotalTracks          int          `json:"total_tracks"`
	AvailableMarkets     []string     `json:"available_markets"`
	ExternalURLs         ExternalURLs `json:"external_urls"`
	Href                 string       `json:"href"`
	ID                   string       `json:"id"`
	Images               []Image      `json:"images"`
	Name                 string       `json:"name"`
	ReleaseDate          string       `json:"release_date"`
	ReleaseDatePrecision string       `json:"release_date_precision"`
	Restrictions         Restrictions `json:"restrictions"`
	Type                 string       `json:"type"`
	URI                  string       `json:"uri"`
	Artists              []Artist     `json:"artists"`
}
type Artist struct {
	ExternalURLs ExternalURLs `json:"external_urls"`
	Href         string       `json:"href"`
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Type         string       `json:"type"`
	URI          string       `json:"uri"`
}
type ExternalIDs struct {
	ISRC string `json:"isrc"`
	EAN  string `json:"ean"`
	UPC  string `json:"upc"`
}
type LinkedFrom struct{}
type Restrictions struct {
	Reason string `json:"reason"`
}
type Actions struct {
	InterruptingPlayback  bool `json:"interrupting_playback"`
	Pausing               bool `json:"pausing"`
	Resuming              bool `json:"resuming"`
	Seeking               bool `json:"seeking"`
	SkippingNext          bool `json:"skipping_next"`
	SkippingPrev          bool `json:"skipping_prev"`
	TogglingRepeatContext bool `json:"toggling_repeat_context"`
	TogglingShuffle       bool `json:"toggling_shuffle"`
	TogglingRepeatTrack   bool `json:"toggling_repeat_track"`
	TransferringPlayback  bool `json:"transferring_playback"`
}

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

func GetUser(token string) (*User, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}
	response, code, err := http_client.GetRequest(spotifyApi+"me", headers)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la requête GET: %v", err)
	}

	switch code {
	case 200:
		fmt.Println("200 SUCCESS")
	case 204:
		fmt.Println("204 ERROR WHILE GET PROFILE")
	case 404:
		fmt.Println("404 OOPS")
	}

	// Décoder la réponse JSON
	var result User
	err = json.Unmarshal(response, &result)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la désérialisation de la réponse: %v", err)
	}

	return &result, nil
}
func GetCurrentlyPlayedSong(token string) (*SpotifyData, error) {
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

	var result SpotifyData
	err = json.Unmarshal(response, &result)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la désérialisation de la réponse: %v", err)
	}

	return &result, nil
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
