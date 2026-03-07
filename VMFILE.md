# Vmfile - Construtor de Imágenes Personalizadas

Similar a **Dockerfile**, un **Vmfile** permite definir y construir imágenes QCOW2 personalizadas de forma reproducible y automatizada.

## Sintaxis

```
NAME "Nombre descriptivo"
FROM <imagen-base>
OUTPUT <archivo-salida.qcow2>
FORMAT qcow2
OS_VARIANT <variante-so>

LAYER <nombre-capa>
PACKAGES
paquete1
paquete2

RUN
comando1
comando2

COPY
archivo-local /ruta/destino
```

## Directivas

| Directiva | Descripción | Ejemplo |
|-----------|-------------|---------|
| **NAME** | Nombre descriptivo de la imagen | `NAME "Servidor Ubuntu personalizado"` |
| **FROM** | Imagen base a usar | `FROM ubuntu24` o `FROM https://url.../image.img` |
| **OUTPUT** | Archivo QCOW2 de salida | `OUTPUT mi-imagen.qcow2` |
| **FORMAT** | Formato de salida | `FORMAT qcow2` |
| **OS_VARIANT** | Variante del SO | `OS_VARIANT ubuntu24.04` |
| **LAYER** | Define una capa de customización | `LAYER development-tools` |
| **PACKAGES** | Paquetes a instalar | Sigue esta directiva |
| **RUN** | Comandos a ejecutar | Sigue esta directiva |
| **COPY** | Copiar archivos a la imagen | `COPY archivo.conf /etc/app.conf` |

## Ejemplos

### Ejemplo 1: Servidor Web Nginx

```vmfile
NAME "Nginx Ubuntu 24.04"
FROM ubuntu24
OUTPUT nginx-server.qcow2
FORMAT qcow2
OS_VARIANT ubuntu24.04

LAYER system
RUN
apt update && apt upgrade -y

LAYER webserver
PACKAGES
nginx
curl
wget

RUN
systemctl enable nginx
mkdir -p /var/www/html
echo "Servidor listo" > /var/www/html/index.html
```

### Ejemplo 2: Servidor Python/Django

```vmfile
NAME "Django Development Server"
FROM ubuntu24
OUTPUT django-server.qcow2
FORMAT qcow2
OS_VARIANT ubuntu24.04

LAYER base-system
RUN
apt update && apt upgrade -y
apt install -y python3-pip python3-venv git curl

LAYER django-dev
RUN
pip3 install --upgrade pip
pip3 install django djangorestframework psycopg2-binary

LAYER user-setup
RUN
useradd -m -s /bin/bash developer
mkdir -p /home/developer/projects
chown -R developer:developer /home/developer/projects
```

### Ejemplo 3: Stack LAMP

```vmfile
NAME "LAMP Stack - Apache MySQL PHP"
FROM ubuntu24
OUTPUT lamp-stack.qcow2
FORMAT qcow2
OS_VARIANT ubuntu24.04

LAYER webstack
PACKAGES
apache2
mysql-server
php
php-mysql
php-gd
curl

RUN
systemctl enable apache2 mysql
a2enmod rewrite
a2enmod php
usermod -d /var/www/html www-data

LAYER configuration
RUN
mysql_install_db
mkdir -p /var/www/html
chown -R www-data:www-data /var/www/html
```

## Flujo de trabajo

### 1. Crear un Vmfile

```bash
cat > mi-servidor.vmfile <<EOF
NAME "Mi Servidor"
FROM ubuntu24
OUTPUT mi-servidor-final.qcow2
FORMAT qcow2
OS_VARIANT ubuntu24.04

LAYER setup
PACKAGES
vim
curl
git

RUN
apt update && apt upgrade -y
EOF
```

### 2. Construir la imagen

```bash
vm build ./mi-servidor.vmfile
```

Output esperado:
```
🔨 Construyendo imagen 'Mi Servidor'...
✓ Imagen base copiada: /home/user/.vm-manager/images/ubuntu24.qcow2
  ⚙️  Aplicando capa 'setup'...
    - Instalando paquetes: vim,curl,git
    - Ejecutando: apt update && apt upgrade -y
    ✓ Capa 'setup' aplicada
✨ Imagen construida exitosamente en: /home/user/.vm-manager/images/mi-servidor-final.qcow2
```

### 3. Usar la imagen en VMs

Después de construir, la imagen está disponible como cualquier otra:

```bash
# Listar imágenes
vm image list

# Crear VM usando la imagen construida
vm create mi-servidor-final vm1 --ram 2048 --cpus 2
```

## Características

✅ **Capas múltiples** - Organiza la customización en pasos lógicos  
✅ **Reproducible** - Los mismos resultados cada vez  
✅ **Reusable** - Reutiliza imágenes base  
✅ **Control total** - Acceso a todos los comandos del sistema  
✅ **Versioning** - Versioná tus Vmfiles con git  

## Requisitos del sistema

```bash
# Ubuntu/Debian
sudo apt install libguestfs-tools qemu-utils wget

# Arch/CachyOS
sudo pacman -S libguestfs qemu wget
```

Verifica la instalación:
```bash
virt-customize --version
qemu-img --version
```

## Variables de entorno

Puedes usar variables dentro del Vmfile:

```vmfile
RUN
echo $HOME
export APP_ENV=production
```

## Consejos y mejores prácticas

1. **Minimiza las capas** - Combine comandos relacionados en una sola capa
2. **Limpia cachés** - Ejecute `apt clean` para reducir tamaño
3. **Usa capas lógicas** - Agrupa por funcionalidad (base, dev, apps)
4. **Documenta capas** - El nombre de LAYER debe ser descriptivo
5. **Ordena por dependencias** - Instala paquetes base antes

```vmfile
# ✓ BIEN - Una sola capa para operaciones relacionadas
LAYER development
PACKAGES
git
curl
build-essential

RUN
apt update && apt upgrade -y
apt clean

# ✗ MAL - Capas innecesarias
LAYER git
RUN apt install -y git

LAYER curl
RUN apt install -y curl
```

## Troubleshooting

### Error: virt-customize no encontrado
```bash
sudo apt install libguestfs-tools
```

### Error: Permisos insuficientes
```bash
sudo usermod -aG kvm $USER
sudo usermod -aG libvirt $USER
# Luego haz logout y login
```

### La imagen es muy grande
Limpia cachés en el Vmfile:
```vmfile
RUN
apt update
apt install -y paquetes
apt clean
rm -rf /var/lib/apt/lists/*
```

## Ejemplos incluidos

- `Vmfile.example` - Servidor web completo
- `Vmfile.manjaro.example` - Sistema Manjaro personalizado

Copy y modifica según necesites:
```bash
cp Vmfile.example miproyecto.vmfile
vm build ./miproyecto.vmfile
```
