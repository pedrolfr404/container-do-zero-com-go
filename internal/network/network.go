package network

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

const (
	hostIfName      = "veth-host"
	containerIfName = "veth-cont"
	hostCIDR        = "10.200.0.1/24"
	containerCIDR   = "10.200.0.2/24"
	hostIP          = "10.200.0.1"
)

// Setup configura a rede entre o host e o container que tem o PID informado.
func Setup(pid int) error {
	// 1. Cria o par veth (as duas pontas nascem no host)
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{Name: hostIfName},
		PeerName:  containerIfName,
	}
	if err := netlink.LinkAdd(veth); err != nil {
		return fmt.Errorf("criando par veth: %w", err)
	}

	// 2. Configura a ponta do host
	host, err := netlink.LinkByName(hostIfName)
	if err != nil {
		return fmt.Errorf("buscando %s: %w", hostIfName, err)
	}
	hostAddr, err := netlink.ParseAddr(hostCIDR)
	if err != nil {
		return err
	}
	if err := netlink.AddrAdd(host, hostAddr); err != nil {
		return fmt.Errorf("atribuindo IP ao host: %w", err)
	}
	if err := netlink.LinkSetUp(host); err != nil {
		return fmt.Errorf("subindo %s: %w", hostIfName, err)
	}

	// 3. Move a outra ponta para o namespace do container
	peer, err := netlink.LinkByName(containerIfName)
	if err != nil {
		return fmt.Errorf("buscando %s: %w", containerIfName, err)
	}
	if err := netlink.LinkSetNsPid(peer, pid); err != nil {
		return fmt.Errorf("movendo veth para o namespace do container: %w", err)
	}

	// 4. Configura a ponta que está dentro do container
	ns, err := netns.GetFromPid(pid)
	if err != nil {
		return fmt.Errorf("abrindo namespace do pid %d: %w", pid, err)
	}
	defer ns.Close()

	h, err := netlink.NewHandleAt(ns) // executa comandos de rede DENTRO do namespace
	if err != nil {
		return err
	}
	defer h.Close()

	// loopback
	lo, err := h.LinkByName("lo")
	if err != nil {
		return err
	}
	if err := h.LinkSetUp(lo); err != nil {
		return err
	}

	// interface veth: renomeia para eth0, atribui IP e sobe
	cont, err := h.LinkByName(containerIfName)
	if err != nil {
		return err
	}
	if err := h.LinkSetName(cont, "eth0"); err != nil {
		return err
	}
	contAddr, err := netlink.ParseAddr(containerCIDR)
	if err != nil {
		return err
	}
	if err := h.AddrAdd(cont, contAddr); err != nil {
		return fmt.Errorf("atribuindo IP ao container: %w", err)
	}
	if err := h.LinkSetUp(cont); err != nil {
		return err
	}

	// rota padrão: tudo que sai do container passa pelo host
	route := &netlink.Route{
		LinkIndex: cont.Attrs().Index,
		Gw:        net.ParseIP(hostIP),
	}
	if err := h.RouteAdd(route); err != nil {
		return fmt.Errorf("adicionando rota padrão: %w", err)
	}

	return nil
}

// Cleanup remove o par veth. Apagar uma ponta remove as duas.
func Cleanup() error {
	link, err := netlink.LinkByName(hostIfName)
	if err != nil {
		return nil // já não existe
	}
	return netlink.LinkDel(link)
}