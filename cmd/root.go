package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "vm",
	Short: "vm-manager — gestión de máquinas virtuales con KVM",
	Long: `
██╗   ██╗███╗   ███╗
██║   ██║████╗ ████║
██║   ██║██╔████╔██║
╚██╗ ██╔╝██║╚██╔╝██║
 ╚████╔╝ ██║ ╚═╝ ██║
  ╚═══╝  ╚═╝     ╚═╝

Herramienta para gestión automatizada de máquinas virtuales con KVM.

Uso básico:
  vm image download ubuntu22   → descargar imagen base
  vm create ubuntu22 mi-vm     → crear nueva VM
  vm list                      → listar todas las VMs
  vm start mi-vm               → iniciar VM
  vm stop  mi-vm               → detener VM
  vm delete mi-vm              → eliminar VM
`,
	// Si se llama solo "vm" sin subcomando, muestra el help
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// Execute es el punto de entrada llamado desde main.go
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Flag global para config file custom (opcional)
	rootCmd.PersistentFlags().StringVar(
		&cfgFile,
		"config",
		"",
		"archivo de config (por defecto: $HOME/.vm-manager/config.yaml)",
	)

	// Flag de verbose global
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "output detallado")
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

func initConfig() {
	if cfgFile != "" {
		// Usar config file especificado por el usuario
		viper.SetConfigFile(cfgFile)
	} else {
		// Buscar config en $HOME/.vm-manager/
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error obteniendo home dir:", err)
			os.Exit(1)
		}

		viper.AddConfigPath(home + "/.vm-manager")
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// Leer variables de entorno con prefijo VM_
	viper.SetEnvPrefix("VM")
	viper.AutomaticEnv()

	// Leer config file si existe (no es obligatorio)
	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Println("Usando config:", viper.ConfigFileUsed())
		}
	}

	// Valores por defecto
	setDefaults()
}

func setDefaults() {
	// Directorio base — por defecto en home del usuario
	home, _ := os.UserHomeDir()
	baseDir := home + "/.vm-manager"

	viper.SetDefault("base_dir", baseDir)
	viper.SetDefault("images_dir", baseDir+"/images")
	viper.SetDefault("vms_dir", baseDir+"/vms")
	viper.SetDefault("images_catalog", baseDir+"/images.json")

	// Valores por defecto para nuevas VMs
	viper.SetDefault("vm.default_ram", 2048)   // MB
	viper.SetDefault("vm.default_cpus", 2)
	viper.SetDefault("vm.default_disk", 20)    // GB
	viper.SetDefault("vm.network", "default")  // red NAT de libvirt
}
