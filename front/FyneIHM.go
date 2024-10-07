package front

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"image/color"
)

type CustomLayout struct {
	MarginTop    float32
	MarginBottom float32
	MarginLeft   float32
	MarginRight  float32
	Gap          float32
	Direction    string // "vertical" ou "horizontal"
	MinWidth     float32
	MinHeight    float32
	MaxWidth     float32
	MaxHeight    float32
}

func (c *CustomLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	// Calculer la largeur et la hauteur disponibles après application des marges
	availableWidth := containerSize.Width - c.MarginLeft - c.MarginRight
	availableHeight := containerSize.Height - c.MarginTop - c.MarginBottom

	// Appliquer les limites maximales définies pour la taille totale du conteneur
	if c.MaxWidth > 0 {
		availableWidth = fyne.Min(availableWidth, c.MaxWidth)
	}
	if c.MaxHeight > 0 {
		availableHeight = fyne.Min(availableHeight, c.MaxHeight)
	}

	xOffset := c.MarginLeft
	yOffset := c.MarginTop

	// Disposer les objets enfants en fonction de la direction (verticale ou horizontale)
	for _, obj := range objects {
		if c.Direction == "vertical" {
			// Définir la largeur de l'objet en fonction de la largeur disponible (et ne pas dépasser MaxWidth)
			objWidth := fyne.Min(availableWidth, obj.MinSize().Width)
			objHeight := obj.MinSize().Height

			// Redimensionner l'objet en tenant compte de la largeur limitée
			obj.Resize(fyne.NewSize(objWidth, objHeight))
			obj.Move(fyne.NewPos(xOffset, yOffset))
			yOffset += objHeight + c.Gap
		} else if c.Direction == "horizontal" {
			// Définir la hauteur de l'objet en fonction de la hauteur disponible (et ne pas dépasser MaxHeight)
			objWidth := obj.MinSize().Width
			objHeight := fyne.Min(availableHeight, obj.MinSize().Height)

			// Redimensionner l'objet en tenant compte de la hauteur limitée
			obj.Resize(fyne.NewSize(objWidth, objHeight))
			obj.Move(fyne.NewPos(xOffset, yOffset))
			xOffset += objWidth + c.Gap
		}
	}
}

func (c *CustomLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	totalWidth := c.MarginLeft + c.MarginRight
	totalHeight := c.MarginTop + c.MarginBottom

	// Calculer la taille minimale du conteneur en fonction des enfants
	for _, obj := range objects {
		if c.Direction == "vertical" {
			totalHeight += obj.MinSize().Height + c.Gap
			totalWidth = fyne.Max(totalWidth, obj.MinSize().Width+c.MarginLeft+c.MarginRight)
		} else if c.Direction == "horizontal" {
			totalWidth += obj.MinSize().Width + c.Gap
			totalHeight = fyne.Max(totalHeight, obj.MinSize().Height+c.MarginTop+c.MarginBottom)
		}
	}

	// Supprimer le dernier gap ajouté
	if len(objects) > 0 {
		if c.Direction == "vertical" {
			totalHeight -= c.Gap
		} else if c.Direction == "horizontal" {
			totalWidth -= c.Gap
		}
	}

	// Assurer que le conteneur respecte les tailles minimales définies
	finalWidth := fyne.Max(totalWidth, c.MinWidth)
	finalHeight := fyne.Max(totalHeight, c.MinHeight)

	// Assurer que le conteneur ne dépasse pas les tailles maximales définies
	if c.MaxWidth > 0 {
		finalWidth = fyne.Min(finalWidth, c.MaxWidth)
	}
	if c.MaxHeight > 0 {
		finalHeight = fyne.Min(finalHeight, c.MaxHeight)
	}

	return fyne.NewSize(finalWidth, finalHeight)
}

func TruncateText(text string, textSize float32, availableWidth float32) string {
	// Créer un canvas.Text temporaire pour calculer la largeur du texte
	tempText := canvas.NewText(text, color.Black)
	tempText.TextSize = textSize

	// Calculer la largeur du texte complet
	textWidth := tempText.MinSize().Width

	// Vérifier si le texte rentre dans la largeur disponible
	if textWidth <= availableWidth {
		return text
	}

	// Commencer à tronquer le texte et ajouter des "..."
	truncated := text
	ellipsisWidth := canvas.NewText("...", color.Black).MinSize().Width

	// Retirer des caractères jusqu'à ce que le texte rentre dans la largeur disponible
	for len(truncated) > 0 {
		truncated = truncated[:len(truncated)-1]
		tempText.Text = truncated + "..."
		textWidth = tempText.MinSize().Width

		if textWidth+ellipsisWidth <= availableWidth {
			return truncated + "..."
		}
	}

	// Si même un seul caractère avec "..." ne tient pas, retourner seulement "..."
	return "..."
}
