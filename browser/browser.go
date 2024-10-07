package browser

import (
	"context"
	"errors"
	"fmt"
	"github.com/chromedp/chromedp"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

func ExtractBackgroundImageURL(styleContent string) string {
	// Regex pour trouver l'URL dans le background-image
	re := regexp.MustCompile(`background-image:\s*url\(["']?(.*?)["']?\);`)
	matches := re.FindStringSubmatch(styleContent)

	if len(matches) > 1 {
		return matches[1] // Retourne l'URL extraite
	}
	return ""
}
func ScrapeBackgroundImageDiv(url string) (string, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Créer un contexte avec un timeout plus long pour éviter l'expiration prématurée
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var styleContent string
	var childCount int

	// Exécuter les actions de scraping avec des étapes séparées
	err := chromedp.Run(ctx,
		// Naviguer vers l'URL spécifiée
		chromedp.Navigate(url),
		chromedp.WaitReady(`body`, chromedp.ByQuery),
		chromedp.WaitVisible(`.under-main-view`, chromedp.ByQuery),
		chromedp.Evaluate(`document.querySelector('.under-main-view').children.length`, &childCount),
		chromedp.Sleep(1*time.Second),
	)

	if err != nil {
		log.Fatalf("Erreur lors de l'exécution de Chromedp : %v", err)
	}

	if childCount == 0 {
		fmt.Println("ERROR IMAGE")
		return "", errors.New("No image found")
	}

	err = chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady(`body`, chromedp.ByQuery),
		chromedp.WaitVisible(`.under-main-view`, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		chromedp.AttributeValue(`[data-testid="background-image"]`, "style", &styleContent, nil),
	)

	if err != nil {
		log.Fatalf("Erreur lors de l'exécution de Chromedp : %v", err)
	}

	backgroundImageURL := ExtractBackgroundImageURL(styleContent)
	if backgroundImageURL != "" {
		return backgroundImageURL, nil
	} else {
		fmt.Println("Aucune URL trouvée dans le style.")
	}

	return "", errors.New("Something wrong append...")
}

func DownloadImage(url, filepath string) error {
	// Envoyer une requête GET pour télécharger l'image
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("erreur lors du téléchargement de l'image: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	// Vérifier que la requête est réussie
	if resp.StatusCode != 200 {
		return fmt.Errorf("échec du téléchargement de l'image, statut HTTP: %d", resp.StatusCode)
	}

	// Créer le fichier pour sauvegarder l'image
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("erreur lors de la création du fichier: %v", err)
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {

		}
	}(out)

	// Copier le contenu téléchargé dans le fichier
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("erreur lors de l'enregistrement de l'image: %v", err)
	}

	return nil
}
