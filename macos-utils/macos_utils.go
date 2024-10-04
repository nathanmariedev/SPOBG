package macos_utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

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
	if err != nil && !errors.Is(err, filepath.SkipDir) {
		return false, err
	}
	return found, nil
}
