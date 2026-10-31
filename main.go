package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

const cgroupRoot = "/sys/fs/cgroup/meu-container"
const rootfs = "/opt/container/rootfs"

func main() {
	if len(os.Args) < 3 {
		fmt.Println("uso: sudo go run main.go run <comando> [args]")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("comando inválido")
	}
}

func setupCgroups(pid int) (func(), error) {
	// Cria a subpasta do cgroup
	if err := os.MkdirAll(cgroupRoot, 0755); err != nil {
		return nil, fmt.Errorf("erro ao criar pasta do cgroup: %w", err)
	}

	// Deleta o cgroup quando o container encerrar
	cleanup := func() {
		_ = os.Remove(cgroupRoot)
	}

	// Limita a quantidade maxima de processos, aqui está sendo limiteado a 20 processos
	if err := os.WriteFile(filepath.Join(cgroupRoot, "pids.max"), []byte("20"), 0644); err != nil {
		cleanup()
		return nil, fmt.Errorf("erro ao configurar pids.max: %w", err)
	}

	// Limita a memoria RAM, aqui está sendo limitado a 100MB (104857600 bytes)
	if err := os.WriteFile(filepath.Join(cgroupRoot, "memory.max"), []byte("104857600"), 0644); err != nil {
		cleanup()
		return nil, fmt.Errorf("erro ao configurar memory.max: %w", err)
	}

	// Limita o uso de CPU (20000 microsegundos de CPU a cada 100000 microsegundos, ou seja, 20% do tempo de CPU)
	if err := os.WriteFile(filepath.Join(cgroupRoot, "cpu.max"), []byte("20000 100000"), 0644); err != nil {
		cleanup()
		return nil, fmt.Errorf("erro ao configurar cpu.max: %w", err)
	}

	// Associa o PID do container ao cgroup
	pidStr := strconv.Itoa(pid)
	if err := os.WriteFile(filepath.Join(cgroupRoot, "cgroup.procs"), []byte(pidStr), 0644); err != nil {
		cleanup()
		return nil, fmt.Errorf("erro ao associar PID no cgroup.procs: %w", err)
	}

	return cleanup, nil
}

func run() {
	fmt.Printf("[run] iniciando %v (pid %d)\n", os.Args[2:], os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWUSER,
		Unshareflags: syscall.CLONE_NEWNS,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getgid(), Size: 1},
		},
	}

	// Inicia o processo filho para obtermos o PID no host
	must(cmd.Start())

	// Aplica restrições de cgroup no host usando o PID do filho
	cleanup, err := setupCgroups(cmd.Process.Pid)
	if err != nil {
		panic(err)
	} else {
		defer cleanup()
	}

	// Aguarda o termino do container
	must(cmd.Wait())
}

func child() {
	fmt.Printf("[child] rodando %v (pid %d)", os.Args[2:], os.Getpid())

	must(syscall.Sethostname([]byte("meu-container")))
	must(syscall.Chroot(rootfs))
	must(os.Chdir("/"))
	must(syscall.Mount("proc", "/proc", "proc", 0, ""))

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Env = []string{"PATH=/bin:/usr/bin:/sbin:/usr/sbin", "PS1=[container] # "}

	must(cmd.Run())
	must(syscall.Unmount("/proc", 0))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
