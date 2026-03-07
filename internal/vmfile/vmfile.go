package vmfile

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Layer representa una capa de customización en la imagen
type Layer struct {
	Name     string
	Packages []string
	Commands []string
	Files    map[string]string // src -> dst
}

// Vmfile representa la estructura de un archivo Vmfile
type Vmfile struct {
	Name       string
	BaseImage  string
	OutputFile string
	Format     string
	OSVariant  string
	Layers     []Layer
}

// Parse lee y parsea un archivo Vmfile
func Parse(filePath string) (*Vmfile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir Vmfile: %w", err)
	}
	defer file.Close()

	vmfile := &Vmfile{
		Layers: []Layer{},
	}

	scanner := bufio.NewScanner(file)
	currentLayer := -1
	inPackages := false
	inCommands := false
	inFiles := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Saltar líneas vacías y comentarios
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Directivas principales
		if strings.HasPrefix(line, "NAME ") {
			vmfile.Name = strings.TrimSpace(strings.TrimPrefix(line, "NAME "))
		} else if strings.HasPrefix(line, "FROM ") {
			vmfile.BaseImage = strings.TrimSpace(strings.TrimPrefix(line, "FROM "))
		} else if strings.HasPrefix(line, "OUTPUT ") {
			vmfile.OutputFile = strings.TrimSpace(strings.TrimPrefix(line, "OUTPUT "))
		} else if strings.HasPrefix(line, "FORMAT ") {
			vmfile.Format = strings.TrimSpace(strings.TrimPrefix(line, "FORMAT "))
		} else if strings.HasPrefix(line, "OS_VARIANT ") {
			vmfile.OSVariant = strings.TrimSpace(strings.TrimPrefix(line, "OS_VARIANT "))
		} else if strings.HasPrefix(line, "LAYER ") {
			inPackages = false
			inCommands = false
			inFiles = false
			layerName := strings.TrimSpace(strings.TrimPrefix(line, "LAYER "))
			vmfile.Layers = append(vmfile.Layers, Layer{
				Name:     layerName,
				Packages: []string{},
				Commands: []string{},
				Files:    make(map[string]string),
			})
			currentLayer = len(vmfile.Layers) - 1
		} else if strings.HasPrefix(line, "PACKAGES") {
			inPackages = true
			inCommands = false
			inFiles = false
		} else if strings.HasPrefix(line, "RUN") {
			inCommands = true
			inPackages = false
			inFiles = false
		} else if strings.HasPrefix(line, "COPY") {
			inFiles = true
			inCommands = false
			inPackages = false
		} else if currentLayer != -1 {
			// Agregar línea al contexto correcto
			if inPackages && line != "" {
				vmfile.Layers[currentLayer].Packages = append(vmfile.Layers[currentLayer].Packages, line)
			} else if inCommands && line != "" {
				vmfile.Layers[currentLayer].Commands = append(vmfile.Layers[currentLayer].Commands, line)
			} else if inFiles && line != "" {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					src := parts[0]
					dst := parts[1]
					vmfile.Layers[currentLayer].Files[src] = dst
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error leyendo Vmfile: %w", err)
	}

	// Validar que existe información mínima
	if vmfile.BaseImage == "" {
		return nil, fmt.Errorf("Vmfile debe contener 'FROM' (imagen base)")
	}
	if vmfile.OutputFile == "" {
		return nil, fmt.Errorf("Vmfile debe contener 'OUTPUT' (archivo de salida)")
	}
	if vmfile.Format == "" {
		vmfile.Format = "qcow2"
	}

	return vmfile, nil
}

// Validate valida que el Vmfile tenga la información necesaria
func (v *Vmfile) Validate() error {
	if v.BaseImage == "" {
		return fmt.Errorf("imagen base no especificada (FROM)")
	}
	if v.OutputFile == "" {
		return fmt.Errorf("archivo de salida no especificado (OUTPUT)")
	}
	if v.Format == "" {
		return fmt.Errorf("formato no especificado (FORMAT)")
	}
	return nil
}

// String devuelve una representación legible del Vmfile
func (v *Vmfile) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Vmfile: %s\n", v.Name))
	sb.WriteString(fmt.Sprintf("  FROM: %s\n", v.BaseImage))
	sb.WriteString(fmt.Sprintf("  OUTPUT: %s\n", v.OutputFile))
	sb.WriteString(fmt.Sprintf("  FORMAT: %s\n", v.Format))
	sb.WriteString(fmt.Sprintf("  OS_VARIANT: %s\n", v.OSVariant))
	sb.WriteString(fmt.Sprintf("  LAYERS: %d\n", len(v.Layers)))
	for _, layer := range v.Layers {
		sb.WriteString(fmt.Sprintf("    - %s (%d paquetes, %d comandos)\n",
			layer.Name, len(layer.Packages), len(layer.Commands)))
	}
	return sb.String()
}
