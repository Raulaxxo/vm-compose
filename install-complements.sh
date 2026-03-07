#!/usr/bin/env bash

set -e

echo "Actualizando repositorios..."
sudo apt update

echo "Instalando stack de virtualización..."

sudo apt install -y qemu-kvm libvirt-daemon-system libvirt-clients virtinst bridge-utils cpu-checker cloud-image-utils

echo "Instalando herramientas utiles..."
sudo apt install -y curl wget git

echo "Instalando yq (lector YAML)..."
sudo snap install yq

echo "Agregando usuario a grupos de virtualización..."
sudo usermod -aG libvirt $USER
sudo usermod -aG kvm $USER

echo "Habilitando libvirt..."
sudo systemctl enable --now libvirtd

echo "Verificando soporte de virtualización..."
kvm-ok || true

echo "Activando red default..."
virsh net-start default 2>/dev/null || true
virsh net-autostart default

echo ""
echo "Instalación terminada."
echo "Cierra sesión o reinicia para aplicar permisos."
