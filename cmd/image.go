package cmd

import (
	"fmt"
	"os"

	"github.com/Raulaxxo/vm-compose/internal/config"
	"github.com/Raulaxxo/vm-compose/internal/kvm"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Gestión del catálogo de imágenes base",
	Long:  `Permite descargar, listar y gestionar imágenes base para crear VMs.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var imageListCmd = &cobra.Command{
	Use:   "list",
	Short: "Listar imágenes disponibles en el catálogo",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		images := kvm.ListImages(cfg)
		if len(images) == 0 {
			fmt.Println("No hay imágenes en el catálogo.")
			return nil
		}

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetStyle(table.StyleRounded)
		t.Style().Title.Align = text.AlignCenter

		t.SetTitle("Catálogo de imágenes")
		t.AppendHeader(table.Row{"CLAVE", "NOMBRE", "ARCHIVO", "ESTADO", "TAMAÑO"})

		for _, img := range images {
			status := "✗ no descargada"
			size := "-"
			if img.Available {
				status = "✓ disponible"
				size = fmt.Sprintf("%.1f GB", float64(img.Size)/1024/1024/1024)
			}
			t.AppendRow(table.Row{img.Key, img.Name, img.File, status, size})
		}

		t.Render()
		fmt.Printf("\nPara descargar: vm image download <clave>\n")
		return nil
	},
}

var imageDownloadCmd = &cobra.Command{
	Use:   "download <imagen>",
	Short: "Descargar una imagen base del catálogo",
	Args:  cobra.ExactArgs(1),
	Example: `  vm image download ubuntu22
  vm image download debian12
  vm image download rocky9`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		return kvm.DownloadImage(cfg, args[0])
	},
}

func init() {
	rootCmd.AddCommand(imageCmd)
	imageCmd.AddCommand(imageListCmd)
	imageCmd.AddCommand(imageDownloadCmd)
}
