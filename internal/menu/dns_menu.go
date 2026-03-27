// Este arquivo implementa o submenu de gerenciamento de nameservers (DNS).
// As operações de leitura e escrita são delegadas ao pacote network,
// que manipula diretamente o /etc/resolv.conf.
package menu

import (
	"fmt"

	"netcheck/internal/network"
)

// menuDNS exibe o submenu de DNS em loop até que o usuário escolha "Voltar".
func menuDNS() {
	for {
		_, opcao, err := selecionarMenu("DNS - Nameservers", []string{
			"Listar Nameservers",
			"Adicionar Nameserver",
			"Remover Nameserver",
			"Voltar",
		})
		if err != nil {
			return
		}

		switch opcao {
		case "Listar Nameservers":
			listarNameservers()
		case "Adicionar Nameserver":
			adicionarNameserver()
		case "Remover Nameserver":
			removerNameserver()
		case "Voltar":
			return
		}
	}
}

// listarNameservers lê e exibe os nameservers configurados no /etc/resolv.conf.
func listarNameservers() {
	servers, err := network.LerNameservers()
	if err != nil {
		fmt.Println("Erro:", err)
		return
	}

	if len(servers) == 0 {
		fmt.Println("Nenhum nameserver configurado.")
		return
	}

	fmt.Println("\nNameservers configurados:")
	for i, s := range servers {
		fmt.Printf("  %d. %s\n", i+1, s)
	}
	fmt.Println()
}

// adicionarNameserver solicita um IP e o adiciona ao /etc/resolv.conf.
// A validação do IP ocorre em tempo real no prompt (antes do envio).
func adicionarNameserver() {
	ip, err := lerTexto("IP do Nameserver", func(s string) error {
		return network.ValidarIP(s)
	})
	if err != nil {
		return
	}

	if err := network.AdicionarNameserver(ip); err != nil {
		fmt.Println("Erro:", err)
		return
	}

	fmt.Printf("Nameserver %s adicionado com sucesso.\n", ip)
}

// removerNameserver exibe os nameservers em um menu de seleção e remove o escolhido.
// Usa selecionarMenu para evitar que o usuário precise digitar um índice numérico.
func removerNameserver() {
	servers, err := network.LerNameservers()
	if err != nil {
		fmt.Println("Erro:", err)
		return
	}

	if len(servers) == 0 {
		fmt.Println("Nenhum nameserver para remover.")
		return
	}

	// O índice retornado por selecionarMenu corresponde diretamente ao índice do slice
	idx, escolhido, err := selecionarMenu("Selecione o nameserver para remover", servers)
	if err != nil {
		return
	}

	if err := network.RemoverNameserver(idx); err != nil {
		fmt.Println("Erro:", err)
		return
	}

	fmt.Printf("Nameserver %s removido com sucesso.\n", escolhido)
}
