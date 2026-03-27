// Este arquivo gerencia as rotas de rede do sistema usando a biblioteca
// github.com/vishvananda/netlink, que se comunica diretamente com o kernel
// Linux via Netlink socket — sem necessidade de chamar o comando "ip route".
// Requer CAP_NET_ADMIN (root) para adicionar e remover rotas.
package network

import (
	"fmt"
	"net"
	"os"

	"github.com/vishvananda/netlink"
)

// InfoRota representa uma rota de forma legível para exibição no menu.
type InfoRota struct {
	Destino   string // Rede de destino em CIDR (ex: "10.0.0.0/8" ou "0.0.0.0/0")
	Gateway   string // IP do próximo salto, ou "direto" para redes locais
	Interface string // Nome da interface de saída (ex: "ens33")
}

// ListarRotas retorna todas as rotas IPv4 ativas na tabela de roteamento do kernel.
// Rotas com Dst == nil são rotas padrão (default gateway), representadas como "0.0.0.0/0".
func ListarRotas() ([]InfoRota, error) {
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar rotas: %w", err)
	}

	var rotas []InfoRota
	for _, r := range routes {
		rota := InfoRota{}

		// Rota padrão (default gateway) tem Dst == nil no netlink
		if r.Dst == nil {
			rota.Destino = "0.0.0.0/0"
		} else {
			rota.Destino = r.Dst.String()
		}

		// Gw == nil indica rota direta (connected route), sem próximo salto
		if r.Gw != nil {
			rota.Gateway = r.Gw.String()
		} else {
			rota.Gateway = "direto"
		}

		// Resolve o nome da interface pelo índice numérico retornado pelo kernel
		if link, err := netlink.LinkByIndex(r.LinkIndex); err == nil {
			rota.Interface = link.Attrs().Name
		}

		rotas = append(rotas, rota)
	}

	return rotas, nil
}

// AdicionarRota insere uma nova rota na tabela de roteamento do kernel via netlink.
// A rota é adicionada apenas na sessão atual — não sobrevive a reinicializações.
// Para persistência, use PersistirRota após confirmar que a rota funciona.
func AdicionarRota(destino, gateway string) error {
	if err := ValidarCIDR(destino); err != nil {
		return err
	}
	if err := ValidarIP(gateway); err != nil {
		return err
	}

	// net.ParseCIDR normaliza o destino para o endereço de rede
	// (ex: "10.1.2.3/8" → dst = "10.0.0.0/8"), que é o comportamento padrão do kernel
	_, dst, _ := net.ParseCIDR(destino)
	gw := net.ParseIP(gateway)

	if err := netlink.RouteAdd(&netlink.Route{Dst: dst, Gw: gw}); err != nil {
		return fmt.Errorf("erro ao adicionar rota: %w", err)
	}

	return nil
}

// RemoverRota remove da tabela do kernel a primeira rota cujo destino corresponde ao CIDR informado.
// Usa comparação por string para localizar a rota, tratando nil como "0.0.0.0/0".
func RemoverRota(destino string) error {
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return fmt.Errorf("erro ao buscar rotas: %w", err)
	}

	for _, r := range routes {
		// Normaliza o destino da rota do kernel para comparação por string
		var dst string
		if r.Dst == nil {
			dst = "0.0.0.0/0"
		} else {
			dst = r.Dst.String()
		}

		if dst == destino {
			if err := netlink.RouteDel(&r); err != nil {
				return fmt.Errorf("erro ao remover rota: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("rota para %s não encontrada", destino)
}

// PersistirRota registra o comando equivalente em /etc/netcheck-routes.conf.
// Este arquivo serve como histórico das rotas adicionadas pela ferramenta.
// Para aplicar automaticamente no boot, os comandos deste arquivo devem ser
// adicionados ao /etc/rc.local ou equivalente da distribuição.
func PersistirRota(destino, gateway string) error {
	const arquivo = "/etc/netcheck-routes.conf"

	// O_APPEND garante que rotas anteriores não sejam apagadas
	f, err := os.OpenFile(arquivo, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("erro ao abrir %s: %w", arquivo, err)
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "ip route add %s via %s\n", destino, gateway)
	return err
}
