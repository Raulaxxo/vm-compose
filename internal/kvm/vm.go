package kvm

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/namishh/vm-manager/internal/config"
)

// VMOptions opciones para crear una nueva VM
type VMOptions struct {
	Name     string
	Image    string
	RAM      int
	CPUs     int
	DiskSize int
	Network  string
	User     string
	Password string
	SSHKey   string
}

// CreateVM crea una nueva VM completa con cloud-init
func CreateVM(cfg *config.Config, opts VMOptions) error {
	fmt.Printf("→ Creando VM '%s' desde imagen '%s'\n\n", opts.Name, opts.Image)

	if !cfg.ImageExists(opts.Image) {
		fmt.Printf("  Imagen '%s' no encontrada localmente, descargando...\n", opts.Image)
		if err := DownloadImage(cfg, opts.Image); err != nil {
			return fmt.Errorf("error descargando imagen: %w", err)
		}
	}

	basePath, err := cfg.ImagePath(opts.Image)
	if err != nil {
		return err
	}

	vmDir := cfg.VMDir(opts.Name)
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		return fmt.Errorf("no se pudo crear directorio de VM: %w", err)
	}

	diskPath := cfg.VMDiskPath(opts.Name)
	fmt.Printf("  [1/4] Creando disco (%dGB)...\n", opts.DiskSize)
	if err := createDisk(basePath, diskPath, opts.DiskSize); err != nil {
		return fmt.Errorf("error creando disco: %w", err)
	}

	fmt.Printf("  [2/4] Generando cloud-init...\n")
	cloudInitPath := cfg.VMCloudInitPath(opts.Name)
	if err := generateCloudInit(vmDir, cloudInitPath, opts); err != nil {
		return fmt.Errorf("error generando cloud-init: %w", err)
	}

	fmt.Printf("  [3/4] Registrando VM en libvirt...\n")
	if err := virtInstall(cfg, opts, diskPath, cloudInitPath); err != nil {
		os.RemoveAll(vmDir)
		return fmt.Errorf("error en virt-install: %w", err)
	}

	fmt.Printf("  [4/4] Iniciando VM...\n")
	if err := StartVM(opts.Name); err != nil {
		if !strings.Contains(err.Error(), "already active") {
			return fmt.Errorf("VM creada pero no se pudo iniciar: %w", err)
		}
	}

	fmt.Printf("\n✓ VM '%s' creada y corriendo\n", opts.Name)
	fmt.Printf("  RAM:  %dMB | CPUs: %d | Disco: %dGB\n", opts.RAM, opts.CPUs, opts.DiskSize)
	fmt.Printf("  Red:  %s\n", opts.Network)
	fmt.Printf("  User: %s\n\n", opts.User)
	fmt.Printf("  Espera ~30s y ejecuta:  virsh net-dhcp-leases default\n")
	fmt.Printf("  Para conectarse:        ssh %s@<IP>\n", opts.User)

	return nil
}

func createDisk(basePath, diskPath string, sizeGB int) error {
	cmd := exec.Command("qemu-img", "create",
		"-f", "qcow2",
		"-b", basePath,
		"-F", "qcow2",
		diskPath,
		fmt.Sprintf("%dG", sizeGB),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func generateCloudInit(vmDir, isoPath string, opts VMOptions) error {
	metaData := fmt.Sprintf("instance-id: %s\nlocal-hostname: %s\n", opts.Name, opts.Name)
	metaDataPath := vmDir + "/meta-data"
	if err := os.WriteFile(metaDataPath, []byte(metaData), 0644); err != nil {
		return err
	}

	userData := buildUserData(opts)
	userDataPath := vmDir + "/user-data"
	if err := os.WriteFile(userDataPath, []byte(userData), 0644); err != nil {
		return err
	}

	return buildCloudInitISO(isoPath, userDataPath, metaDataPath)
}

func buildUserData(opts VMOptions) string {
	var sb strings.Builder
	sb.WriteString("#cloud-config\n")
	sb.WriteString(fmt.Sprintf("hostname: %s\n", opts.Name))
	sb.WriteString("manage_etc_hosts: true\n\n")

	// Usuario
	sb.WriteString("users:\n")
	sb.WriteString(fmt.Sprintf("  - name: %s\n", opts.User))
	sb.WriteString("    sudo: ALL=(ALL) NOPASSWD:ALL\n")
	sb.WriteString("    shell: /bin/bash\n")
	sb.WriteString("    groups: [sudo, adm]\n")
	sb.WriteString("    lock_passwd: false\n")

	// SSH key si se proporcionó
	if opts.SSHKey != "" {
		sb.WriteString("    ssh_authorized_keys:\n")
		sb.WriteString(fmt.Sprintf("      - %s\n", opts.SSHKey))
	}

	// Password via chpasswd formato compatible con todas las versiones
	if opts.Password != "" {
		sb.WriteString("\nchpasswd:\n")
		sb.WriteString("  expire: false\n")
		sb.WriteString("  list: |\n")
		sb.WriteString(fmt.Sprintf("    %s:%s\n", opts.User, opts.Password))
		sb.WriteString("\nssh_pwauth: true\n")
	}

	// Paquetes básicos
	sb.WriteString("\npackages:\n")
	sb.WriteString("  - qemu-guest-agent\n")
	sb.WriteString("\nruncmd:\n")
	sb.WriteString("  - systemctl enable --now qemu-guest-agent\n")

	return sb.String()
}

func buildCloudInitISO(isoPath, userDataPath, metaDataPath string) error {
	if path, err := exec.LookPath("cloud-localds"); err == nil {
		cmd := exec.Command(path, isoPath, userDataPath, metaDataPath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("cloud-localds: %s: %w", strings.TrimSpace(string(out)), err)
		}
		return nil
	}

	if path, err := exec.LookPath("genisoimage"); err == nil {
		cmd := exec.Command(path,
			"-output", isoPath,
			"-volid", "cidata",
			"-joliet", "-rock",
			userDataPath, metaDataPath,
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("genisoimage: %s: %w", strings.TrimSpace(string(out)), err)
		}
		return nil
	}

	return fmt.Errorf("no se encontró cloud-localds ni genisoimage — instala cloud-image-utils o genisoimage")
}

func virtInstall(cfg *config.Config, opts VMOptions, diskPath, cloudInitPath string) error {
	osVariant := "generic"
	if img, err := cfg.GetImage(opts.Image); err == nil && img.OSVariant != "" {
		osVariant = img.OSVariant
	}

	args := []string{
		"--name", opts.Name,
		"--ram", fmt.Sprintf("%d", opts.RAM),
		"--vcpus", fmt.Sprintf("%d", opts.CPUs),
		"--disk", fmt.Sprintf("path=%s,format=qcow2", diskPath),
		"--disk", fmt.Sprintf("path=%s,device=cdrom", cloudInitPath),
		"--os-variant", osVariant,
		"--network", fmt.Sprintf("network=%s", opts.Network),
		"--graphics", "none",
		"--console", "pty,target_type=serial",
		"--noautoconsole",
		"--import",
	}

	cmd := exec.Command("virt-install", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func StartVM(name string) error {
	out, err := exec.Command("virsh", "start", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func StopVM(name string, force bool) error {
	action := "shutdown"
	if force {
		action = "destroy"
	}
	out, err := exec.Command("virsh", action, name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func DeleteVM(cfg *config.Config, name string) error {
	_ = StopVM(name, true)

	out, err := exec.Command("virsh", "undefine", name, "--remove-all-storage").CombinedOutput()
	if err != nil {
		out, err = exec.Command("virsh", "undefine", name).CombinedOutput()
		if err != nil {
			return fmt.Errorf("virsh undefine: %s: %w", strings.TrimSpace(string(out)), err)
		}
	}

	vmDir := cfg.VMDir(name)
	if err := os.RemoveAll(vmDir); err != nil {
		return fmt.Errorf("error eliminando archivos de VM: %w", err)
	}

	fmt.Printf("✓ VM '%s' eliminada\n", name)
	return nil
}

type VMInfo struct {
	Name  string
	State string
	IP    string
}

func ListVMs() ([]VMInfo, error) {
	out, err := exec.Command("virsh", "list", "--all", "--name").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error listando VMs: %w", err)
	}

	var vms []VMInfo
	for _, line := range strings.Split(string(out), "\n") {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		vms = append(vms, VMInfo{
			Name:  name,
			State: getVMState(name),
			IP:    getVMIP(name),
		})
	}

	return vms, nil
}

func getVMState(name string) string {
	out, err := exec.Command("virsh", "domstate", name).CombinedOutput()
	if err != nil {
		return "desconocido"
	}
	return strings.TrimSpace(string(out))
}

func getVMIP(name string) string {
	out, err := exec.Command("virsh", "domifaddr", name, "--source", "agent").CombinedOutput()
	if err != nil {
		return "-"
	}

	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "ipv4") {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				ip := strings.Split(fields[3], "/")[0]
				if !strings.HasPrefix(ip, "127.") {
					return ip
				}
			}
		}
	}
	return "-"
}
