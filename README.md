# vm-manager

Herramienta CLI para gestión automatizada de máquinas virtuales con KVM.

## Requisitos

```bash
# Arch / CachyOS
sudo pacman -S qemu-full libvirt virt-install cloud-image-utils

# Ubuntu / Debian
sudo apt install qemu-kvm libvirt-daemon-system virtinst cloud-image-utils

# Habilitar libvirt
sudo systemctl enable --now libvirtd
sudo usermod -aG libvirt $USER
```

## Instalación

```bash
git clone https://github.com/namishh/vm-manager
cd vm-manager

make deps
make install
```

## Uso

### Gestión de Imágenes

```bash
# Ver imágenes disponibles
vm image list

# Descargar imagen base
vm image download ubuntu22

# Construir imagen personalizada desde Vmfile (¡NUEVO!)
vm build ./Vmfile
vm build ./mi-servidor.vmfile
```

### Crear y gestionar VMs

```bash
# Crear VM con SSH key
vm create ubuntu22 mi-vm --ssh-key "$(cat ~/.ssh/id_rsa.pub)"

# Crear VM con password
vm create ubuntu22 mi-vm --password mipassword

# Crear VM con recursos custom
vm create debian12 servidor --ram 4096 --cpus 4 --disk 40

# Listar VMs
vm list

# Conectarse por SSH directamente (¡NUEVO!)
vm ssh mi-vm                           # Conectar como usuario 'ubuntu'
vm ssh servidor --user root            # Conectar como otro usuario
vm ssh test1 --user admin --port 2222  # Con puerto personalizado

# Iniciar / detener
vm start mi-vm
vm stop mi-vm
vm stop mi-vm --force   # forzar apagado

# Eliminar
vm delete mi-vm
vm delete mi-vm --yes   # sin confirmación
```

### Vmfiles: Construir imágenes personalizadas

Similar a **Dockerfile**, ahora puedes crear imágenes personalizadas:

```bash
# Ver ejemplos
cat Vmfile.example

# Crear tu propio Vmfile
cat > mi-servidor.vmfile <<EOF
NAME "Servidor Personalizado"
FROM ubuntu24
OUTPUT mi-servidor.qcow2
FORMAT qcow2
OS_VARIANT ubuntu24.04

LAYER tools
PACKAGES
vim
curl
git
nginx

RUN
systemctl enable nginx
EOF

# Construir
vm build ./mi-servidor.vmfile

# Usar la imagen construida
vm create mi-servidor vm1 --ram 2048
```

Para más detalles, ver [VMFILE.md](./VMFILE.md)

## Estructura

```
~/.vm-manager/
├── images/          # imágenes base descargadas
├── vms/             # discos y cloud-init de cada VM
│   └── mi-vm/
│       ├── disk.qcow2
│       └── cloud-init.iso
└── images.json      # catálogo de imágenes
```

## Configuración

`~/.vm-manager/config.yaml` (opcional):

```yaml
vm:
  default_ram: 2048
  default_cpus: 2
  default_disk: 20
  network: default
```
