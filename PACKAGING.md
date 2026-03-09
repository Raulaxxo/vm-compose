# Empaquetamiento Debian

Este proyecto incluye configuración para crear un paquete Debian (.deb) compatible con sistemas Ubuntu y Debian.

## Archivos de configuración

- **debian/** - Directorio con archivos de control de Debian
- **debian/control** - Metadatos del paquete (dependencias, descripción, etc.)
- **debian/copyright** - Información de licencia
- **debian/changelog** - Historial de cambios
- **debian/rules** - Reglas de compilación

## Compilar el paquete .deb

### Opción 1: Usando make (recomendado)

```bash
make deb
```

Esto compilará el binario y creará el paquete .deb automáticamente en `build/vm-manager_1.0.0_amd64.deb`.

### Opción 2: Paso a paso

```bash
# Compilar el binario
make build

# Crear la estructura del paquete
mkdir -p ../deb-build/vm-manager-1.0.0/{DEBIAN,usr/local/bin}
cp build/vm ../deb-build/vm-manager-1.0.0/usr/local/bin/
chmod 755 ../deb-build/vm-manager-1.0.0/usr/local/bin/vm

# Crear el archivo de control
cat > ../deb-build/vm-manager-1.0.0/DEBIAN/control << 'EOF'
Package: vm-manager
Version: 1.0.0
Section: utils
Priority: optional
Architecture: amd64
Depends: qemu-kvm, libvirt-daemon-system, virtinst, cloud-image-utils
Maintainer: Raulaxxo <your-email@example.com>
Homepage: https://github.com/Raulaxxo/vm-compose
Description: CLI tool for automated KVM virtual machine management
 vm-manager is a command-line tool for managing virtual machines with KVM/libvirt.
 It provides an easy interface to create, build, list and manage VMs with predefined
 configurations using Vmfiles.
EOF

# Construir el .deb
cd ../deb-build
fakeroot dpkg-deb --build vm-manager-1.0.0 vm-manager_1.0.0_amd64.deb
```

## Instalar el paquete

```bash
# Desde el archivo .deb
sudo dpkg -i vm-manager_1.0.0_amd64.deb

# Si faltan dependencias, instalarlas con apt
sudo apt-get install -f
```

## Verificar el paquete

```bash
# Ver contenidos del .deb
dpkg -c vm-manager_1.0.0_amd64.deb

# Ver metainformación
dpkg -I vm-manager_1.0.0_amd64.deb

# Verificar que se instaló correctamente
vm --help
```

## Requisitos del sistema

Para compilar el paquete necesitas:
- Go 1.22 o superior
- `build-essential`
- `fakeroot`
- `dpkg-dev`

En Ubuntu/Debian:
```bash
sudo apt-get install golang-go build-essential dpkg-dev fakeroot


wget https://github.com/Raulaxxo/vm-compose/releases/download/latest/vm-manager_1.0.0_amd64.deb 

dpkg -i vm-manager_1.0.0_amd64.deb 
```
