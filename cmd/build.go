package cmd

import (
	"fmt"
	"os"

	"github.com/Raulaxxo/vm-compose/internal/config"
	"github.com/Raulaxxo/vm-compose/internal/vmfile"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build <Vmfile>",
	Short: "Construir una imagen personalizada desde un Vmfile",
	Long: `Construye una imagen QCOW2 personalizada a partir de un Vmfile.

Un Vmfile es similar a un Dockerfile pero para máquinas virtuales.
Permite definir capas de customización que se aplican a una imagen base.

Ejemplo de Vmfile:
  NAME "Mi servidor personalizado"
  FROM ubuntu24
  OUTPUT mi-servidor.qcow2
  FORMAT qcow2
  OS_VARIANT ubuntu24.04

  LAYER system-update
  RUN
  apt update && apt upgrade -y

  LAYER development
  PACKAGES
  vim
  curl
  git
  docker.io

  LAYER config
  RUN
  useradd -m -s /bin/bash developer
  systemctl enable ssh
`,
	Args: cobra.ExactArgs(1),
	Example: `  vm build ./Vmfile
  vm build ./vm-images/manjaro.vmfile
  vm build /ruta/completa/custom.vmfile`,
	RunE: func(cmd *cobra.Command, args []string) error {
		vmfilePath := args[0]

		// Verificar que el archivo existe
		if _, err := os.Stat(vmfilePath); os.IsNotExist(err) {
			return fmt.Errorf("Vmfile no encontrado: %s", vmfilePath)
		}

		// Verificar dependencias
		if err := vmfile.CheckDependencies(); err != nil {
			return err
		}

		// Cargar configuración
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Parsear Vmfile
		vmf, err := vmfile.Parse(vmfilePath)
		if err != nil {
			return err
		}

		fmt.Println(vmf.String())

		// Crear builder
		builder := vmfile.NewBuilder(vmf, cfg.BaseDir, cfg.ImagesDir)

		// Construir imagen
		if err := builder.Build(); err != nil {
			return err
		}

		fmt.Printf("\n✨ Imagen construida exitosamente en: %s/%s\n", cfg.ImagesDir, vmf.OutputFile)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
