package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	sshUser string
	sshPort string
)

var sshCmd = &cobra.Command{
	Use:   "ssh <vm-name>",
	Short: "Conectarse por SSH a una máquina virtual",
	Long: `Conectarse automáticamente a una máquina virtual por SSH.

Obtiene la IP de la VM y se conecta directamente sin necesidad de saber la dirección.

Espera a que la VM esté lista si acaba de iniciarse.`,
	Args: cobra.ExactArgs(1),
	Example: `  vm ssh mi-vm
  vm ssh test1 --user ubuntu
  vm ssh web-server --user root --port 2222
  vm ssh db-server --user admin`,
	RunE: func(cmd *cobra.Command, args []string) error {
		vmName := args[0]

		// Obtener IP de la VM
		ip, err := getVMIP(vmName, 60) // Esperar max 60 segundos
		if err != nil {
			return err
		}

		fmt.Printf("✓ Conectando a %s (%s)...\n\n", vmName, ip)

		// Construir comando SSH
		sshTarget := fmt.Sprintf("%s@%s", sshUser, ip)
		if sshPort != "" {
			sshTarget = fmt.Sprintf("%s -p %s", sshTarget, sshPort)
		}

		// Ejecutar ssh de forma interactiva
		sshCmd := exec.Command("ssh", strings.Fields(sshTarget)...)
		sshCmd.Stdin = os.Stdin
		sshCmd.Stdout = os.Stdout
		sshCmd.Stderr = os.Stderr

		if err := sshCmd.Run(); err != nil {
			// No es error si el usuario cierra la conexión normalmente
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
			}
			return fmt.Errorf("error en conexión SSH: %w", err)
		}

		return nil
	},
}

// getVMIP obtiene la IP de una VM esperando si es necesario
func getVMIP(vmName string, maxWaitSeconds int) (string, error) {
	startTime := time.Now()
	timeout := time.Duration(maxWaitSeconds) * time.Second

	for {
		ip, err := fetchVMIP(vmName)
		if err == nil && ip != "" {
			return ip, nil
		}

		if time.Since(startTime) > timeout {
			return "", fmt.Errorf("VM '%s' no obtuvo IP después de %d segundos. Usa 'virsh net-dhcp-leases default' para verificar", vmName, maxWaitSeconds)
		}

		fmt.Printf("  ⏳ Esperando IP de %s... (%ds)\r", vmName, maxWaitSeconds-int(time.Since(startTime).Seconds()))
		time.Sleep(2 * time.Second)
	}
}

// fetchVMIP obtiene la IP actual de una VM
func fetchVMIP(vmName string) (string, error) {
	// Ejecutar: virsh net-dhcp-leases default
	cmd := exec.Command("virsh", "net-dhcp-leases", "default")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error ejecutando virsh: %w", err)
	}

	// Parsear salida buscando la VM por nombre
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// Formato: NETWORK          MAC ADDRESS           EXPIRY TIME              HOSTNAME             IP ADDRESS
		// default          52:54:00:12:34:56    2026-03-07 12:34:56 -0400   mi-vm                192.168.122.50/24

		if strings.Contains(line, vmName) {
			// Extraer IP del formato "192.168.122.50/24"
			fields := strings.Fields(line)
			for _, field := range fields {
				if strings.Contains(field, ".") && strings.Contains(field, "/") {
					ip := strings.Split(field, "/")[0]
					if isValidIP(ip) {
						return ip, nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("no se encontró IP para %s", vmName)
}

// isValidIP verifica que sea una IP válida (simple check)
func isValidIP(ip string) bool {
	parts := strings.Split(ip, ".")
	return len(parts) == 4
}

func init() {
	rootCmd.AddCommand(sshCmd)

	sshCmd.Flags().StringVarP(&sshUser, "user", "u", "ubuntu", "Usuario SSH de la VM")
	sshCmd.Flags().StringVarP(&sshPort, "port", "p", "", "Puerto SSH (opcional)")
}
