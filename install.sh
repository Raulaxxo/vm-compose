#!/bin/bash
# Script para instalar o actualizar vm-manager globalmente

set -e

BIN_SOURCE="./build/vm"
BIN_TARGET="/usr/local/bin/vm"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}VM-Manager Installation${NC}"
echo "=========================="

# Verificar que el binario existe
if [ ! -f "$BIN_SOURCE" ]; then
    echo -e "${RED}Error: $BIN_SOURCE no encontrado${NC}"
    echo "Ejecuta primero: make build"
    exit 1
fi

# Verificar permisos
if [ ! -w "$(dirname $BIN_TARGET)" ]; then
    echo -e "${YELLOW}Se requieren permisos de administrador para instalar en $BIN_TARGET${NC}"
    echo "Ejecuta: sudo bash install.sh"
    exit 1
fi

# Hacer backup de la versión anterior
if [ -f "$BIN_TARGET" ]; then
    BACKUP="${BIN_TARGET}.backup.$(date +%s)"
    echo "Backup de versión anterior: $BACKUP"
    cp "$BIN_TARGET" "$BACKUP"
fi

# Copiar el nuevo binario
cp "$BIN_SOURCE" "$BIN_TARGET"
chmod +x "$BIN_TARGET"

# Verificar instalación
echo -e "${GREEN}✓ Instalación completada${NC}"
echo ""
echo "Versión instalada:"
$BIN_TARGET --help | head -20

echo ""
echo "Prueba con: vm --help"
