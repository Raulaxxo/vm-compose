package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// ImageEntry representa una imagen en el catálogo
type ImageEntry struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	File      string `json:"file"`
	Format    string `json:"format"`
	OSVariant string `json:"os_variant"`
}

// Config contiene toda la configuración de la aplicación
type Config struct {
	BaseDir        string
	ImagesDir      string
	VMsDir         string
	ImagesCatalog  string
	DefaultRAM     int
	DefaultCPUs    int
	DefaultDisk    int
	Network        string
	Images         map[string]ImageEntry
}

// Load carga la configuración desde viper y el catálogo de imágenes
func Load() (*Config, error) {
	cfg := &Config{
		BaseDir:       viper.GetString("base_dir"),
		ImagesDir:     viper.GetString("images_dir"),
		VMsDir:        viper.GetString("vms_dir"),
		ImagesCatalog: viper.GetString("images_catalog"),
		DefaultRAM:    viper.GetInt("vm.default_ram"),
		DefaultCPUs:   viper.GetInt("vm.default_cpus"),
		DefaultDisk:   viper.GetInt("vm.default_disk"),
		Network:       viper.GetString("vm.network"),
	}

	// Crear directorios necesarios si no existen
	if err := cfg.ensureDirs(); err != nil {
		return nil, fmt.Errorf("error creando directorios: %w", err)
	}

	// Cargar catálogo de imágenes
	if err := cfg.loadImagesCatalog(); err != nil {
		return nil, fmt.Errorf("error cargando catálogo de imágenes: %w", err)
	}

	return cfg, nil
}

// ensureDirs crea los directorios base si no existen
func (c *Config) ensureDirs() error {
	dirs := []string{c.BaseDir, c.ImagesDir, c.VMsDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("no se pudo crear %s: %w", dir, err)
		}
	}
	return nil
}

// loadImagesCatalog lee el images.json y lo deserializa
func (c *Config) loadImagesCatalog() error {
	// Si no existe el catálogo en la ruta configurada,
	// intentar cargarlo desde el directorio del binario/proyecto
	catalogPath := c.ImagesCatalog
	if _, err := os.Stat(catalogPath); os.IsNotExist(err) {
		// Buscar en ./config/images.json como fallback
		fallback := "config/images.json"
		if _, err2 := os.Stat(fallback); err2 == nil {
			catalogPath = fallback
		} else {
			// Crear catálogo vacío si no existe ninguno
			c.Images = make(map[string]ImageEntry)
			return nil
		}
	}

	data, err := os.ReadFile(catalogPath)
	if err != nil {
		return fmt.Errorf("no se pudo leer %s: %w", catalogPath, err)
	}

	if err := json.Unmarshal(data, &c.Images); err != nil {
		return fmt.Errorf("images.json mal formado: %w", err)
	}

	return nil
}

// GetImage devuelve una imagen del catálogo por su clave
func (c *Config) GetImage(key string) (ImageEntry, error) {
	img, ok := c.Images[key]
	if !ok {
		return ImageEntry{}, fmt.Errorf("imagen '%s' no encontrada en el catálogo", key)
	}
	return img, nil
}

// ImageExists comprueba si una imagen ya está descargada localmente
func (c *Config) ImageExists(key string) bool {
	img, err := c.GetImage(key)
	if err != nil {
		return false
	}
	path := c.ImagesDir + "/" + img.File
	_, err = os.Stat(path)
	return err == nil
}

// ImagePath devuelve la ruta completa de una imagen descargada
func (c *Config) ImagePath(key string) (string, error) {
	img, err := c.GetImage(key)
	if err != nil {
		return "", err
	}
	return c.ImagesDir + "/" + img.File, nil
}

// VMDir devuelve el directorio de una VM específica
func (c *Config) VMDir(name string) string {
	return c.VMsDir + "/" + name
}

// VMDiskPath devuelve la ruta del disco de una VM
func (c *Config) VMDiskPath(name string) string {
	return c.VMDir(name) + "/disk.qcow2"
}

// VMCloudInitPath devuelve la ruta del ISO de cloud-init de una VM
func (c *Config) VMCloudInitPath(name string) string {
	return c.VMDir(name) + "/cloud-init.iso"
}
