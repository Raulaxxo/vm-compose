package kvm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/namishh/vm-manager/internal/config"
)

// ISOVMOptions opciones para crear una VM desde ISO local
type ISOVMOptions struct {
	Name        string
	ISOPath     string
	RAM         int
	CPUs        int
	DiskSize    int
	Network     string
	Graphics    string // spice | vnc | none
	OSVariant   string
	VNCPort     int
	BridgeIface string // solo si Network == "bridge"
	NoStart     bool
}

// CreateVMFromISO crea una VM con disco vacío y bootea desde ISO local.
// No usa cloud-init porque la instalación es manual (como Proxmox/VirtualBox).
func CreateVMFromISO(cfg *config.Config, opts ISOVMOptions) error {
	fmt.Printf("→ Creando VM '%s' desde ISO local\n", opts.Name)
	fmt.Printf("  ISO: %s\n\n", opts.ISOPath)

	// 1. Validar que no existe ya una VM con ese nombre
	if vmAlreadyExists(opts.Name) {
		return fmt.Errorf("ya existe una VM llamada '%s' — usa un nombre distinto o elimínala primero", opts.Name)
	}

	// 2. Crear directorio de la VM
	vmDir := cfg.VMDir(opts.Name)
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		return fmt.Errorf("no se pudo crear directorio de VM: %w", err)
	}

	diskPath := cfg.VMDiskPath(opts.Name)

	// 3. Crear disco vacío (sin backing file — instalación desde cero)
	fmt.Printf("  [1/3] Creando disco vacío (%dGB)...\n", opts.DiskSize)
	if err := createEmptyDisk(diskPath, opts.DiskSize); err != nil {
		os.RemoveAll(vmDir)
		return fmt.Errorf("error creando disco: %w", err)
	}

	// 4. Registrar VM en libvirt con virt-install
	fmt.Printf("  [2/3] Registrando VM en libvirt...\n")
	if err := virtInstallISO(cfg, opts, diskPath); err != nil {
		os.RemoveAll(vmDir)
		return fmt.Errorf("error en virt-install: %w", err)
	}

	// 5. Mostrar info de conexión
	fmt.Printf("  [3/3] VM lista\n\n")
	printISOConnectionInfo(opts)

	return nil
}

// createEmptyDisk crea un disco qcow2 vacío sin backing file
func createEmptyDisk(diskPath string, sizeGB int) error {
	cmd := exec.Command("qemu-img", "create",
		"-f", "qcow2",
		diskPath,
		fmt.Sprintf("%dG", sizeGB),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// virtInstallISO construye y ejecuta virt-install para boot desde ISO
func virtInstallISO(cfg *config.Config, opts ISOVMOptions, diskPath string) error {
	// Resolver ruta absoluta de la ISO (virt-install la necesita)
	isoAbs, err := filepath.Abs(opts.ISOPath)
	if err != nil {
		return fmt.Errorf("no se pudo resolver ruta de ISO: %w", err)
	}

	args := []string{
		"--name", opts.Name,
		"--ram", fmt.Sprintf("%d", opts.RAM),
		"--vcpus", fmt.Sprintf("%d", opts.CPUs),
		"--disk", fmt.Sprintf("path=%s,format=qcow2,bus=virtio", diskPath),
		"--cdrom", isoAbs,
		"--os-variant", opts.OSVariant,
		"--boot", "cdrom,hd",  // bootea primero desde cdrom, luego disco
		"--noautoconsole",
	}

	// --- Red ---
	switch opts.Network {
	case "bridge":
		if opts.BridgeIface == "" {
			opts.BridgeIface = "br0"
		}
		args = append(args, "--network", fmt.Sprintf("bridge=%s,model=virtio", opts.BridgeIface))
	default:
		// NAT con la red 'default' de libvirt
		args = append(args, "--network", fmt.Sprintf("network=%s,model=virtio", opts.Network))
	}

	// --- Gráficos / consola ---
	// A diferencia de cloud images, una ISO requiere consola gráfica
	// para que el usuario pueda interactuar con el instalador.
	switch opts.Graphics {
	case "vnc":
		port := opts.VNCPort
		if port == 0 {
			port = 5900
		}
		args = append(args,
			"--graphics", fmt.Sprintf("vnc,listen=127.0.0.1,port=%d", port),
		)
	case "none":
		// Útil si la ISO tiene instalador serial (ej: Alpine, CoreOS)
		args = append(args,
			"--graphics", "none",
			"--console", "pty,target_type=serial",
			"--extra-args", "console=ttyS0,115200n8",
		)
	default:
		// SPICE es la mejor opción por defecto en Linux:
		// - mejor rendimiento que VNC
		// - soporte para portapapeles, audio, USB
		// - accesible con virt-viewer o remote-viewer
		args = append(args,
			"--graphics", "spice,listen=127.0.0.1",
			"--channel", "spice,target.type=virtio",
			"--video", "qxl",
		)
	}

	// Si no se quiere arrancar ahora, solo definir
	if opts.NoStart {
		args = append(args, "--noreboot")
	}

	cmd := exec.Command("virt-install", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// vmAlreadyExists verifica si ya hay una VM con ese nombre en libvirt
func vmAlreadyExists(name string) bool {
	out, err := exec.Command("virsh", "dominfo", name).CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "Id:")
}

// printISOConnectionInfo muestra cómo conectarse a la VM según el backend gráfico
func printISOConnectionInfo(opts ISOVMOptions) {
	fmt.Printf("✓ VM '%s' creada\n", opts.Name)
	fmt.Printf("  RAM: %dMB | CPUs: %d | Disco: %dGB\n\n", opts.RAM, opts.CPUs, opts.DiskSize)

	switch opts.Graphics {
	case "vnc":
		port := opts.VNCPort
		if port == 0 {
			port = 5900
		}
		fmt.Printf("  Conéctate con cualquier cliente VNC:\n")
		fmt.Printf("    vncviewer 127.0.0.1:%d\n", port)
		fmt.Printf("    # o desde remoto con tunel SSH:\n")
		fmt.Printf("    ssh -L %d:127.0.0.1:%d usuario@host-kvm\n\n", port, port)

	case "none":
		fmt.Printf("  Consola serial activa. Conéctate con:\n")
		fmt.Printf("    virsh console %s\n\n", opts.Name)

	default: // spice
		fmt.Printf("  Conéctate con virt-viewer o remote-viewer:\n")
		fmt.Printf("    virt-viewer %s\n", opts.Name)
		fmt.Printf("    # o:\n")
		fmt.Printf("    remote-viewer spice://127.0.0.1\n\n")
	}

	if opts.NoStart {
		fmt.Printf("  VM definida pero no iniciada.\n")
		fmt.Printf("  Iníciala con:  vm start %s\n\n", opts.Name)
	} else {
		fmt.Printf("  La VM está corriendo el instalador de la ISO.\n")
		fmt.Printf("  Completa la instalación y luego:\n")
		fmt.Printf("    vm stop %s      # apagar\n", opts.Name)
		fmt.Printf("    vm start %s     # arrancar desde disco instalado\n\n", opts.Name)
	}

	fmt.Printf("  Otros comandos útiles:\n")
	fmt.Printf("    vm list             # ver estado\n")
	fmt.Printf("    vm stop %s --force  # forzar apagado\n", opts.Name)
	fmt.Printf("    vm delete %s        # eliminar VM y disco\n", opts.Name)
}
