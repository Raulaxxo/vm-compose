package vmfile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Builder maneja la construcción de imágenes desde Vmfiles
type Builder struct {
	Vmfile    *Vmfile
	BaseDir   string
	ImagesDir string
	TempDir   string
}

// NewBuilder crea un nuevo constructor de imágenes
func NewBuilder(vmfile *Vmfile, baseDir, imagesDir string) *Builder {
	return &Builder{
		Vmfile:    vmfile,
		BaseDir:   baseDir,
		ImagesDir: imagesDir,
		TempDir:   "/tmp/vm-builder",
	}
}

// Build construye la imagen según el Vmfile
func (b *Builder) Build() error {
	fmt.Printf("🔨 Construyendo imagen '%s'...\n", b.Vmfile.Name)

	// Validar Vmfile
	if err := b.Vmfile.Validate(); err != nil {
		return fmt.Errorf("Vmfile inválido: %w", err)
	}

	// Crear directorio temporal
	if err := os.MkdirAll(b.TempDir, 0755); err != nil {
		return fmt.Errorf("no se pudo crear directorio temporal: %w", err)
	}
	defer os.RemoveAll(b.TempDir)

	// Obtener imagen base
	baseImagePath, err := b.getBaseImage()
	if err != nil {
		return err
	}

	// Preparar imagen de trabajo
	workingImage := filepath.Join(b.TempDir, "working.img")
	if err := exec.Command("cp", baseImagePath, workingImage).Run(); err != nil {
		return fmt.Errorf("error copiando imagen base: %w", err)
	}
	fmt.Printf("✓ Imagen base copiada: %s\n", baseImagePath)

	// Aplicar cada capa
	for _, layer := range b.Vmfile.Layers {
		if err := b.applyLayer(workingImage, layer); err != nil {
			return fmt.Errorf("error aplicando capa '%s': %w", layer.Name, err)
		}
	}

	// Mover imagen final al directorio de salida
	outputPath := filepath.Join(b.ImagesDir, b.Vmfile.OutputFile)
	if err := exec.Command("mv", workingImage, outputPath).Run(); err != nil {
		return fmt.Errorf("error moviendo imagen final: %w", err)
	}

	fmt.Printf("✓ Imagen construida exitosamente: %s\n", outputPath)
	return nil
}

// getBaseImage obtiene la imagen base (descargando si es necesario)
func (b *Builder) getBaseImage() (string, error) {
	// Si es una URL, descargarla
	if strings.HasPrefix(b.Vmfile.BaseImage, "http://") || strings.HasPrefix(b.Vmfile.BaseImage, "https://") {
		return b.downloadImage(b.Vmfile.BaseImage)
	}

	// Si es una ruta local, usar directamente
	if _, err := os.Stat(b.Vmfile.BaseImage); err == nil {
		return b.Vmfile.BaseImage, nil
	}

	// Si es una clave de catálogo, buscarla en ImagesDir
	catalogPath := filepath.Join(b.ImagesDir, b.Vmfile.BaseImage+".qcow2")
	if _, err := os.Stat(catalogPath); err == nil {
		return catalogPath, nil
	}

	return "", fmt.Errorf("imagen base no encontrada: %s", b.Vmfile.BaseImage)
}

// downloadImage descarga una imagen desde URL
func (b *Builder) downloadImage(url string) (string, error) {
	fileName := filepath.Base(url)
	localPath := filepath.Join(b.TempDir, fileName)

	fmt.Printf("📥 Descargando imagen base: %s\n", url)
	cmd := exec.Command("wget", "-q", "--show-progress", url, "-O", localPath)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error descargando imagen: %w", err)
	}

	fmt.Printf("✓ Descarga completada\n")
	return localPath, nil
}

// applyLayer aplica una capa de customización a la imagen
func (b *Builder) applyLayer(imagePath string, layer Layer) error {
	if len(layer.Packages) == 0 && len(layer.Commands) == 0 && len(layer.Files) == 0 {
		fmt.Printf("  ⊝ Capa '%s' vacía (saltada)\n", layer.Name)
		return nil
	}

	fmt.Printf("  ⚙️  Aplicando capa '%s'...\n", layer.Name)

	// Compilar comando virt-customize
	args := []string{"-a", imagePath}

	// Agregar paquetes si es necesario
	if len(layer.Packages) > 0 {
		pkgsStr := strings.Join(layer.Packages, ",")
		args = append(args, "--install", pkgsStr)
		fmt.Printf("    - Instalando paquetes: %s\n", pkgsStr)
	}

	// Agregar comandos RUN
	for _, cmd := range layer.Commands {
		args = append(args, "--run-command", cmd)
		fmt.Printf("    - Ejecutando: %s\n", cmd)
	}

	// Copiar archivos
	for src, dst := range layer.Files {
		args = append(args, "--copy-in", fmt.Sprintf("%s:%s", src, dst))
		fmt.Printf("    - Copiando: %s → %s\n", src, dst)
	}

	// Ejecutar virt-customize con backend directo (evita problemas con supermin)
	// Necesita sudo para acceso a /boot
	cmdName := "virt-customize"
	cmdArgs := args

	// Intentar sin sudo primero
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Configurar variables de entorno para libguestfs
	cmd.Env = append(os.Environ(),
		"LIBGUESTFS_BACKEND=direct",
	)

	err := cmd.Run()

	// Si falla, reintentar con sudo
	if err != nil {
		fmt.Printf("    ⚠️  Reintentando con sudo (se requiere acceso a /boot)...\n")
		cmd = exec.Command("sudo", append([]string{cmdName}, cmdArgs...)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = append(os.Environ(),
			"LIBGUESTFS_BACKEND=direct",
		)
		err = cmd.Run()
	}

	if err != nil {
		return fmt.Errorf("virt-customize falló: %w", err)
	}

	fmt.Printf("    ✓ Capa '%s' aplicada\n", layer.Name)
	return nil
}

// CheckDependencies verifica que las herramientas necesarias estén instaladas
func CheckDependencies() error {
	tools := []string{"virt-customize", "wget", "qemu-img"}

	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("herramienta requerida no encontrada: %s\nInstala: sudo apt install libguestfs-tools qemu-utils", tool)
		}
	}

	return nil
}
