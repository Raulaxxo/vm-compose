package cmd

import (
	"fmt"
	"os"

	"github.com/Raulaxxo/vm-compose/internal/kvm"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Listar todas las máquinas virtuales",
	RunE: func(cmd *cobra.Command, args []string) error {
		vms, err := kvm.ListVMs()
		if err != nil {
			return err
		}

		if len(vms) == 0 {
			fmt.Println("No hay VMs creadas.")
			fmt.Println("Crea una con: vm create <imagen> <nombre>")
			return nil
		}

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetStyle(table.StyleRounded)
		t.Style().Title.Align = text.AlignCenter

		t.SetTitle("Máquinas Virtuales")
		t.AppendHeader(table.Row{"NOMBRE", "ESTADO", "IP"})

		for _, vm := range vms {
			state := vm.State
			// Colorear estado
			switch vm.State {
			case "running":
				state = "▶ running"
			case "shut off":
				state = "■ detenida"
			case "paused":
				state = "⏸ pausada"
			}
			t.AppendRow(table.Row{vm.Name, state, vm.IP})
		}

		t.Render()
		return nil
	},
}

var startCmd = &cobra.Command{
	Use:     "start <nombre>",
	Short:   "Iniciar una VM detenida",
	Args:    cobra.ExactArgs(1),
	Example: `  vm start mi-vm`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		fmt.Printf("→ Iniciando VM '%s'...\n", name)
		if err := kvm.StartVM(name); err != nil {
			return err
		}
		fmt.Printf("✓ VM '%s' iniciada\n", name)
		return nil
	},
}

var stopCmd = &cobra.Command{
	Use:     "stop <nombre>",
	Short:   "Detener una VM (shutdown graceful)",
	Args:    cobra.ExactArgs(1),
	Example: `  vm stop mi-vm`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if force {
			fmt.Printf("→ Forzando apagado de VM '%s'...\n", name)
		} else {
			fmt.Printf("→ Apagando VM '%s' (graceful)...\n", name)
		}

		if err := kvm.StopVM(name, force); err != nil {
			return err
		}
		fmt.Printf("✓ VM '%s' detenida\n", name)
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:     "delete <nombre>",
	Short:   "Eliminar una VM y todos sus archivos",
	Args:    cobra.ExactArgs(1),
	Example: `  vm delete mi-vm`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		yes, _ := cmd.Flags().GetBool("yes")

		if !yes {
			fmt.Printf("¿Eliminar VM '%s' y todos sus datos? [s/N]: ", name)
			var resp string
			fmt.Scanln(&resp)
			if resp != "s" && resp != "S" {
				fmt.Println("Cancelado.")
				return nil
			}
		}

		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		return kvm.DeleteVM(cfg, name)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(deleteCmd)

	stopCmd.Flags().BoolP("force", "f", false, "Forzar apagado (destroy)")
	deleteCmd.Flags().BoolP("yes", "y", false, "Confirmar sin preguntar")
}
