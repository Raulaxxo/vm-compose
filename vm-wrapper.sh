#!/bin/bash
# Wrapper para ejecutar vm build con los permisos necesarios

# Si el comando es "build", ejecutar con sudo
if [ "$1" = "build" ]; then
    exec sudo /usr/local/bin/vm "$@"
fi

# Para otros comandos, ejecutar normalmente
exec /usr/local/bin/vm "$@"
