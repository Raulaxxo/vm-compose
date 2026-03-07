BINARY=vm
BUILD_DIR=./build
INSTALL_DIR=/usr/local/bin

.PHONY: all build install clean run

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

# Instalar dependencias
deps:
	go mod tidy
	go mod download
