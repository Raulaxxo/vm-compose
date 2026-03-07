BINARY=vm
BUILD_DIR=./build
INSTALL_DIR=/usr/local/bin
DEB_BUILD_DIR=../deb-build
DEB_PKG_NAME=vm-manager_1.0.0_amd64.deb

.PHONY: all build install clean run deb

all: build

build:
	@echo "→ Compilando $(BINARY)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY) .
	@echo "✓ Binario en $(BUILD_DIR)/$(BINARY)"

install: build
	@echo "→ Instalando en $(INSTALL_DIR)/$(BINARY)..."
	sudo cp $(BUILD_DIR)/$(BINARY) $(INSTALL_DIR)/$(BINARY)
	@echo "✓ Instalado. Prueba: vm --help"

clean:
	@rm -rf $(BUILD_DIR)
	@echo "✓ Limpiado"

run:
	go run . $(ARGS)

# Construir paquete .deb
deb: build
	@echo "→ Construyendo paquete .deb..."
	@rm -rf $(DEB_BUILD_DIR)
	@mkdir -p $(DEB_BUILD_DIR)/vm-manager-1.0.0/DEBIAN
	@mkdir -p $(DEB_BUILD_DIR)/vm-manager-1.0.0/usr/local/bin
	@cp $(BUILD_DIR)/$(BINARY) $(DEB_BUILD_DIR)/vm-manager-1.0.0/usr/local/bin/vm
	@chmod 755 $(DEB_BUILD_DIR)/vm-manager-1.0.0/usr/local/bin/vm
	@echo 'Package: vm-manager\nVersion: 1.0.0\nSection: utils\nPriority: optional\nArchitecture: amd64\nDepends: qemu-kvm, libvirt-daemon-system, virtinst, cloud-image-utils\nMaintainer: namishh <your-email@example.com>\nHomepage: https://github.com/namishh/vm-manager\nDescription: CLI tool for automated KVM virtual machine management\n vm-manager is a command-line tool for managing virtual machines with KVM/libvirt.\n It provides an easy interface to create, build, list and manage VMs with predefined\n configurations using Vmfiles.' > $(DEB_BUILD_DIR)/vm-manager-1.0.0/DEBIAN/control
	@cd $(DEB_BUILD_DIR) && fakeroot dpkg-deb --build vm-manager-1.0.0 $(DEB_PKG_NAME)
	@cp $(DEB_BUILD_DIR)/$(DEB_PKG_NAME) $(BUILD_DIR)/$(DEB_PKG_NAME)
	@echo "✓ Paquete .deb generado: $(BUILD_DIR)/$(DEB_PKG_NAME)"

# Instalar dependencias
deps:
	go mod tidy
	go mod download
