package cmd

import (
	"fmt"
	"os"

	"github.com/namishh/vm-manager/internal/config"
	"github.com/namishh/vm-manager/internal/kvm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var createISOCmd = &cobra.Command{
	Use:   "create-iso <ruta-iso> <nombre>",
	Short: "Crear una VM desde un archivo ISO local (instalación manual)",
	Long: `Crea una VM usando una ISO local como medio de instalación.

A diferencia de 'vm create' (que usa cloud images pre-configuradas),
este comando crea un disco vacío y arranca desde la ISO para que
puedas instalar el sistema operativo manualmente via VNC o SPICE.

Útil para:
  - ISOs de Windows
  - Distribuciones sin cloud image
  - Instalaciones personalizadas
  - ISOs propias / custom`,
	Args: cobra.ExactArgs(2),
	Example: `  vm create-iso ~/isos/windows11.iso win11 --ram 4096 --cpus 4 --disk 60
  vm create-iso /tmp/ubuntu-22.04.iso test-server --ram 2048 --disk 30
  vm create-iso archlinux.iso arch-custom --cpus 2 --graphics vnc --vnc-port 5910
  vm create-iso debian.iso prod-vm --network bridge --bridge-iface br0`,
	RunE: func(cmd *cobra.Command, args []string) error {
		isoPath := args[0]
		name := args[1]

		// Validar que la ISO existe
		if _, err := os.Stat(isoPath); os.IsNotExist(err) {
			return fmt.Errorf("ISO no encontrada: %s", isoPath)
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Leer flags con fallback a defaults de config
		ram, _ := cmd.Flags().GetInt("ram")
		if ram == 0 {
			ram = viper.GetInt("vm.default_ram")
		}

		cpus, _ := cmd.Flags().GetInt("cpus")
		if cpus == 0 {
			cpus = viper.GetInt("vm.default_cpus")
		}

		disk, _ := cmd.Flags().GetInt("disk")
		if disk == 0 {
			disk = viper.GetInt("vm.default_disk")
		}

		network, _ := cmd.Flags().GetString("network")
		if network == "" {
			network = viper.GetString("vm.network")
		}

		graphics, _ := cmd.Flags().GetString("graphics")
		osVariant, _ := cmd.Flags().GetString("os-variant")
		vnc_port, _ := cmd.Flags().GetInt("vnc-port")
		bridgeIface, _ := cmd.Flags().GetString("bridge-iface")
		noStart, _ := cmd.Flags().GetBool("no-start")

		opts := kvm.ISOVMOptions{
			Name:        name,
			ISOPath:     isoPath,
			RAM:         ram,
			CPUs:        cpus,
			DiskSize:    disk,
			Network:     network,
			Graphics:    graphics,
			OSVariant:   osVariant,
			VNCPort:     vnc_port,
			BridgeIface: bridgeIface,
			NoStart:     noStart,
		}

		return kvm.CreateVMFromISO(cfg, opts)
	},
}

func init() {
	rootCmd.AddCommand(createISOCmd)

	// Recursos
	createISOCmd.Flags().Int("ram", 0, "RAM en MB (default: 2048)")
	createISOCmd.Flags().Int("cpus", 0, "Número de CPUs (default: 2)")
	createISOCmd.Flags().Int("disk", 0, "Tamaño de disco en GB (default: 20)")

	// Red
	createISOCmd.Flags().String("network", "", "Tipo de red: 'default' (NAT) o 'bridge' (default: default)")
	createISOCmd.Flags().String("bridge-iface", "br0", "Interfaz bridge del host (solo si --network bridge)")

	// Display / consola
	createISOCmd.Flags().String("graphics", "spice", "Backend gráfico: spice, vnc, none (default: spice)")
	createISOCmd.Flags().Int("vnc-port", 5900, "Puerto VNC (solo si --graphics vnc, default: 5900)")

	// Sistema operativo
	createISOCmd.Flags().String("os-variant", "generic", "Variante OS para optimizaciones de libvirt (ej: win11, ubuntu22.04)")

	// Control de arranque
	createISOCmd.Flags().Bool("no-start", false, "Definir la VM pero no iniciarla automáticamente")
}
