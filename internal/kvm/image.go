package kvm

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/namishh/vm-manager/internal/config"
)

// progressWriter trackea el progreso de la descarga
type progressWriter struct {
	total      int64
	downloaded int64
	writer     io.Writer
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	pw.downloaded += int64(n)

	if pw.total > 0 {
		pct := float64(pw.downloaded) / float64(pw.total) * 100
		downloaded := float64(pw.downloaded) / 1024 / 1024
		total := float64(pw.total) / 1024 / 1024
		bar := renderBar(pct, 30)
		fmt.Printf("\r  %s %.1f/%.1fMB (%.0f%%)", bar, downloaded, total, pct)
	} else {
		downloaded := float64(pw.downloaded) / 1024 / 1024
		fmt.Printf("\r  Descargando... %.1fMB", downloaded)
	}

	return n, err
}

func renderBar(pct float64, width int) string {
	filled := int(pct / 100 * float64(width))
	bar := "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
	return bar
}

// DownloadImage descarga una imagen del catálogo al directorio de imágenes
func DownloadImage(cfg *config.Config, key string) error {
	img, err := cfg.GetImage(key)
	if err != nil {
		return err
	}

	destPath := cfg.ImagesDir + "/" + img.File

	// Si ya existe, no descargar de nuevo
	if _, err := os.Stat(destPath); err == nil {
		fmt.Printf("✓ La imagen '%s' ya existe en %s\n", key, destPath)
		return nil
	}

	fmt.Printf("↓ Descargando %s\n", img.Name)
	fmt.Printf("  URL: %s\n", img.URL)
	fmt.Printf("  Destino: %s\n\n", destPath)

	// Iniciar request HTTP
	resp, err := http.Get(img.URL)
	if err != nil {
		return fmt.Errorf("error iniciando descarga: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error HTTP %d al descargar imagen", resp.StatusCode)
	}

	// Crear archivo temporal para descarga segura
	tmpPath := destPath + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("no se pudo crear archivo temporal: %w", err)
	}

	// Cleanup del temp si algo falla
	defer func() {
		file.Close()
		if _, err := os.Stat(tmpPath); err == nil {
			os.Remove(tmpPath)
		}
	}()

	// Copiar con progress bar
	pw := &progressWriter{
		total:  resp.ContentLength,
		writer: file,
	}

	if _, err := io.Copy(pw, resp.Body); err != nil {
		return fmt.Errorf("error durante la descarga: %w", err)
	}

	fmt.Println() // nueva línea después de la progress bar

	// Mover temporal a destino final
	if err := os.Rename(tmpPath, destPath); err != nil {
		return fmt.Errorf("error guardando imagen: %w", err)
	}

	fmt.Printf("\n✓ Imagen '%s' descargada correctamente\n", key)
	fmt.Printf("  Guardada en: %s\n", destPath)
	return nil
}

// ListImages muestra todas las imágenes del catálogo y su estado local
func ListImages(cfg *config.Config) []ImageStatus {
	var result []ImageStatus

	for key, img := range cfg.Images {
		path := cfg.ImagesDir + "/" + img.File
		info, err := os.Stat(path)

		status := ImageStatus{
			Key:       key,
			Name:      img.Name,
			File:      img.File,
			Available: err == nil,
		}

		if err == nil {
			status.Size = info.Size()
		}

		result = append(result, status)
	}

	return result
}

// ImageStatus representa el estado local de una imagen
type ImageStatus struct {
	Key       string
	Name      string
	File      string
	Available bool
	Size      int64
}
