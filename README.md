# container-do-zero-com-go

Projeto Sistemas Operacionais II

## Executar

sudo mkdir -p /opt/container/rootfs
wget https://dl-cdn.alpinelinux.org/alpine/v3.20/releases/x86_64/alpine-minirootfs-3.20.3-x86_64.tar.gz
sudo tar -xzf alpine-minirootfs-\*.tar.gz -C /opt/container/rootfs
ls /opt/container/rootfs
sudo go run main.go run /bin/sh

## Maneiras de comprovar que o ambiente está isolado

## Maneiras de comprovar que o cgroup funciona

#### 1. Teste de limite de processos (pids.max = 20)

Dentro do container, tente criar 30 processos em segundo plano com o comando:

> for i in $(seq 1 30); do sleep 60 & done

Mensagem esperada: sh: can't fork: Resource temporarily unavailable

No Host: Você pode verificar a contagem atual, com o comando:

> cat /sys/fs/cgroup/meu-container/pids.current

#### 2. Teste Limite de CPU (cpu.max = 20%)

Dentro do container, execute:

> yes > /dev/null &

No Host execute o monitor de processos com o comando:

> top

Resultado esperado: o processo yes ficará limitado a no máximo 20% de CPU e não conseguirá usar 100% do núcleo

dentro do container pare o teste utilizando:

> killall yes

#### 3. Teste Limite de memória (memory.max = 100MB)

#### 4. Inspeção Rápida no Host

Enquanto o container estiver em execução, você pode conferir se o PID do container realmente foi anexado ao grupo inspecionando:

Verifica se o PID do processo filho está no cgroup

> cat /sys/fs/cgroup/meu-container/cgroup.procs

Uso atual de memória

> cat /sys/fs/cgroup/meu-container/memory.current
