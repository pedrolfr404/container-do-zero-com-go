package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

const rootfs = "/home/so/container-go/rootfs"

func main(){
	if len(os.Args) < 3{
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

func run(){
	fmt.Printf("[run] iniciando %v (pid %d)\n", os.Args[2:], os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
		syscall.CLONE_NEWPID |
		syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWUTS,
	}

	must(cmd.Run())

}

func child(){
	fmt.Printf("[child] rodando %v (pid %d)", os.Args[2:], os.Getpid())

	must(syscall.Sethostname([]byte("meu-container")))
	must(syscall.Chroot(rootfs))
	must(os.Chdir("/"))
	must(syscall.Mount("proc","/proc", "proc", 0, ""))

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Env = []string{"PATH=/bin:/usr/bin:/sbin:/usr/sbin", "PS1=[container] # "}

	must(cmd.Run())
	must(syscall.Unmount("/proc", 0))
}

func must(err error){
	if err != nil{
		panic(err)
	}
}