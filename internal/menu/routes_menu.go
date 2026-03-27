// Este arquivo implementa o submenu de gerenciamento de rotas.
// O fluxo de adição inclui teste automático e rollback:
// se a rota não passar no teste de conectividade, ela é removida imediatamente.
// Se passar, é persistida em /etc/netcheck-routes.conf.
package menu

import (
	"fmt"
	"net"
	"strings"

	"netcheck/internal/network"
)

// menuRotas exibe o submenu de rotas em loop até o usuário escolher "Voltar".
func menuRotas() {
	for {
		_, opcao, err := selecionarMenu("Gerenciar Rotas", []string{
			"Listar Rotas",
			"Adicionar Rota",
			"Remover Rota",
			"Voltar",
		})
		if err != nil {
			return
		}

		switch opcao {
		case "Listar Rotas":
			listarRotas()
		case "Adicionar Rota":
			adicionarRota()
		case "Remover Rota":
			removerRota()
		case "Voltar":
			return
		}
	}
}

// listarRotas exibe todas as rotas IPv4 ativas em formato tabular.
// A rota padrão (0.0.0.0/0) recebe o label "(padrão)" para melhor legibilidade.
func listarRotas() {
	rotas, err := network.ListarRotas()
	if err != nil {
		fmt.Println("Erro:", err)
		return
	}

	if len(rotas) == 0 {
		fmt.Println("Nenhuma rota encontrada.")
		return
	}

	fmt.Printf("\n%-35s %-20s %s\n", "Destino", "Gateway", "Interface")
	fmt.Println(strings.Repeat("-", 70))
	for _, r := range rotas {
		destino := r.Destino
		if destino == "0.0.0.0/0" {
			destino = "0.0.0.0/0  (padrão)"
		}
		fmt.Printf("%-35s %-20s %s\n", destino, r.Gateway, r.Interface)
	}
	fmt.Println()
}

// adicionarRota conduz o usuário pelo processo de adição de rota com teste e persistência.
//
// Fluxo completo:
//  1. Solicita o destino em CIDR (ex: 10.0.0.0/8)
//  2. Solicita o gateway
//  3. Solicita um IP de teste — validado para estar dentro da rede de destino
//  4. Adiciona a rota via network.AdicionarRota (temporária no kernel)
//  5. Testa conectividade via ping ao IP de teste
//     - Falha: remove a rota (rollback)
//     - Sucesso: persiste em /etc/netcheck-routes.conf
func adicionarRota() {
	destino, err := lerTexto("Destino (CIDR, ex: 10.0.0.0/8)", func(s string) error {
		return network.ValidarCIDR(s)
	})
	if err != nil {
		return
	}

	gateway, err := lerTexto("Gateway", func(s string) error {
		return network.ValidarIP(s)
	})
	if err != nil {
		return
	}

	// Extrai a rede de destino para validar o IP de teste em tempo real.
	// net.ParseCIDR retorna a rede normalizada (ex: "3.3.3.3/20" → "3.3.0.0/20"),
	// que é usada para verificar se o IP de teste pertence à rede correta.
	_, rede, _ := net.ParseCIDR(destino)

	hostTeste, err := lerTexto("IP para testar a rota (deve estar dentro da rede de destino)", func(s string) error {
		if err := network.ValidarIP(s); err != nil {
			return err
		}
		// Rejeita IPs fora da rede de destino para garantir que o teste seja relevante
		if !rede.Contains(net.ParseIP(s)) {
			return fmt.Errorf("%s não pertence à rede %s", s, rede.String())
		}
		return nil
	})
	if err != nil {
		return
	}

	fmt.Printf("\nAdicionando rota %s via %s...\n", destino, gateway)
	if err := network.AdicionarRota(destino, gateway); err != nil {
		fmt.Println("Erro:", err)
		return
	}

	fmt.Printf("Testando conectividade para %s...\n", hostTeste)
	if err := network.TestarConectividade(hostTeste); err != nil {
		// Teste falhou: remove a rota para não deixar configuração inconsistente
		fmt.Printf("Teste falhou: %v\n", err)
		fmt.Println("Removendo rota...")

		if err := network.RemoverRota(destino); err != nil {
			fmt.Println("Erro ao remover rota:", err)
		} else {
			fmt.Println("Rota removida.")
		}
		return
	}

	// Teste passou: persiste a rota para referência futura
	fmt.Println("Teste bem-sucedido! Persistindo rota...")
	if err := network.PersistirRota(destino, gateway); err != nil {
		fmt.Println("Aviso: rota ativa, mas não foi possível persistir:", err)
	} else {
		fmt.Printf("Rota %s via %s adicionada e persistida.\n\n", destino, gateway)
	}
}

// removerRota lista as rotas em um menu de seleção e remove a escolhida.
// Usar selecionarMenu evita que o usuário precise digitar o destino manualmente.
func removerRota() {
	rotas, err := network.ListarRotas()
	if err != nil {
		fmt.Println("Erro:", err)
		return
	}

	if len(rotas) == 0 {
		fmt.Println("Nenhuma rota para remover.")
		return
	}

	// Formata cada rota como string legível para exibição no menu de seleção
	items := make([]string, len(rotas))
	for i, r := range rotas {
		items[i] = fmt.Sprintf("%-30s via %-18s (%s)", r.Destino, r.Gateway, r.Interface)
	}

	idx, _, err := selecionarMenu("Selecione a rota para remover", items)
	if err != nil {
		return
	}

	rota := rotas[idx]
	if err := network.RemoverRota(rota.Destino); err != nil {
		fmt.Println("Erro:", err)
		return
	}

	fmt.Printf("Rota %s removida com sucesso.\n\n", rota.Destino)
}
