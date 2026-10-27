# container-do-zero-com-go
Projeto Sistemas Operacionais II
## Executar
sudo mkdir -p /opt/container/rootfs
wget https://dl-cdn.alpinelinux.org/alpine/v3.20/releases/x86_64/alpine-minirootfs-3.20.3-x86_64.tar.gz
sudo tar -xzf alpine-minirootfs-*.tar.gz -C /opt/container/rootfs
ls /opt/container/rootfs
sudo go run main.go run /bin/sh
