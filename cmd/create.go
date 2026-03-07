package cmd

import (
	"fmt"

	"github.com/namishh/vm-manager/internal/config"
	"github.com/namishh/vm-manager/internal/kvm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var createCmd = &cobra.Command{
	Use:   "create <imagen> <nombre>",
	Short: "Crear una nueva máquina virtual",
	Args:  cobra.ExactArgs(2),
	Example: `  vm create ubuntu22 mi-vm
  vm create debian12 servidor --ram 4096 --cpus 4
  vm create ubuntu22 dev --disk 40 --user ubuntu --ssh-key "ssh-rsa AAAA..."`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		image := args[0]
		name := args[1]

		// Verificar que la imagen existe en el catálogo
		if _, err := cfg.GetImage(image); err != nil {
			fmt.Printf("✗ Imagen '%s' no encontrada.\n", image)
			fmt.Println("  Ejecuta 'vm image list' para ver imágenes disponibles.")
			return nil
		}

		// Leer flags con fallback a valores por defecto de config
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

		user, _ := cmd.Flags().GetString("user")
		if user == "" {
			user = "admin"
		}

		password, _ := cmd.Flags().GetString("password")
		sshKey, _ := cmd.Flags().GetString("ssh-key")

		if password == "" && sshKey == "" {
			return fmt.Errorf("se requiere --password o --ssh-key para poder acceder a la VM")
		}

		opts := kvm.VMOptions{
			Name:     name,
			Image:    image,
			RAM:      ram,
			CPUs:     cpus,
			DiskSize: disk,
			Network:  network,
			User:     user,
			Password: password,
			SSHKey:   sshKey,
		}

		return kvm.CreateVM(cfg, opts)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().Int("ram", 0, "RAM en MB (default: 2048)")
	createCmd.Flags().Int("cpus", 0, "Número de CPUs (default: 2)")
	createCmd.Flags().Int("disk", 0, "Tamaño de disco en GB (default: 20)")
	createCmd.Flags().String("network", "", "Red de libvirt (default: default)")
	createCmd.Flags().String("user", "admin", "Usuario a crear en la VM")
	createCmd.Flags().String("password", "", "Password del usuario")
	createCmd.Flags().String("ssh-key", "", "Clave SSH pública para acceso")
}
