// Este arquivo gerencia endereços IP de interfaces de rede usando a biblioteca
// github.com/vishvananda/netlink, que comunica com o kernel via Netlink socket.
// A principal funcionalidade é AlterarIP com rollback automático:
// se a nova configuração causar perda de conectividade, o IP original é restaurado.
package network

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

// InfoInterface representa uma interface de rede com seus endereços e estado.
type InfoInterface struct {
	Nome      string   // Nome da interface (ex: "ens33", "eth0")
	Enderecos []string // Lista de endereços IPv4 em CIDR (ex: ["192.168.1.10/24"])
	Ativo     bool     // true se a interface estiver com o flag UP
}

// ListarInterfaces retorna todas as interfaces de rede do sistema, exceto loopback (lo).
// Para cada interface, coleta os endereços IPv4 e verifica se está ativa.
func ListarInterfaces() ([]InfoInterface, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return nil, fmt.Errorf("erro ao listar interfaces: %w", err)
	}

	var interfaces []InfoInterface
	for _, link := range links {
		nome := link.Attrs().Name
		// Ignora a interface de loopback (127.0.0.1), não gerenciável pelo usuário
		if nome == "lo" {
			continue
		}

		addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
		if err != nil {
			continue
		}

		// Converte cada endereço para notação CIDR legível (ex: "192.168.1.10/24")
		var enderecos []string
		for _, addr := range addrs {
			ones, _ := addr.Mask.Size()
			enderecos = append(enderecos, fmt.Sprintf("%s/%d", addr.IP.String(), ones))
		}

		// net.FlagUp == 1; verifica bit 0 do campo Flags para saber se a interface está UP
		ativo := link.Attrs().Flags&1 != 0

		interfaces = append(interfaces, InfoInterface{
			Nome:      nome,
			Enderecos: enderecos,
			Ativo:     ativo,
		})
	}

	return interfaces, nil
}

// AlterarIP substitui todos os endereços IPv4 da interface pelo novo CIDR informado.
// Retorna a lista de endereços anteriores para permitir rollback pelo chamador.
//
// Fluxo:
//  1. Valida a interface e o CIDR
//  2. Salva os endereços atuais (snapshot para rollback)
//  3. Remove todos os endereços atuais via netlink
//  4. Adiciona o novo endereço via netlink
//
// Se a adição falhar, retorna os endereços anteriores junto com o erro,
// permitindo que o chamador tente restaurar o estado original com RollbackIP.
func AlterarIP(iface, novoCIDR string) ([]netlink.Addr, error) {
	if err := ValidarInterface(iface); err != nil {
		return nil, err
	}

	// netlink.ParseAddr aceita formato "IP/prefixo" e preserva o IP do host
	// (diferente de net.ParseCIDR, que retorna o endereço de rede)
	novoAddr, err := netlink.ParseAddr(novoCIDR)
	if err != nil {
		return nil, fmt.Errorf("CIDR inválido '%s': %w", novoCIDR, err)
	}

	link, err := netlink.LinkByName(iface)
	if err != nil {
		return nil, fmt.Errorf("interface '%s' não encontrada: %w", iface, err)
	}

	// Snapshot dos endereços atuais — necessário para rollback
	anteriores, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler endereços atuais: %w", err)
	}

	// Remove cada endereço atual da interface
	for _, addr := range anteriores {
		if err := netlink.AddrDel(link, &addr); err != nil {
			return nil, fmt.Errorf("erro ao remover endereço %s: %w", addr.IP, err)
		}
	}

	// Adiciona o novo endereço; retorna `anteriores` mesmo em erro para possibilitar rollback
	if err := netlink.AddrAdd(link, novoAddr); err != nil {
		return anteriores, fmt.Errorf("erro ao adicionar novo IP: %w", err)
	}

	return anteriores, nil
}

// RollbackIP restaura os endereços anteriores de uma interface após uma falha.
// Chamado pelo menu quando o teste de conectividade com o gateway falha após AlterarIP.
//
// Processo:
//  1. Remove o endereço que foi aplicado com falha
//  2. Readiciona todos os endereços do snapshot
func RollbackIP(iface string, anteriores []netlink.Addr) error {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		return fmt.Errorf("interface '%s' não encontrada: %w", iface, err)
	}

	// Remove o endereço atual (o que causou falha de conectividade)
	atuais, _ := netlink.AddrList(link, netlink.FAMILY_V4)
	for _, addr := range atuais {
		netlink.AddrDel(link, &addr) // ignora erros individuais; continua tentando restaurar
	}

	// Restaura cada endereço do snapshot original
	for _, addr := range anteriores {
		a := addr // copia local para evitar que o ponteiro aponte para o mesmo endereço no loop
		if err := netlink.AddrAdd(link, &a); err != nil {
			return fmt.Errorf("erro ao restaurar %s: %w", addr.IP, err)
		}
	}

	return nil
}
