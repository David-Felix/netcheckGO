// Este arquivo implementa o submenu Rede/IP e a funcionalidade de alteração de IP.
// O fluxo de alteração inclui rollback automático: se o novo IP causar perda
// de conectividade com o gateway, o endereço original é restaurado automaticamente.
package menu

import (
	"fmt"
	"time"

	"netcheck/internal/network"
)

// menuRedeIP exibe o submenu Rede/IP em loop até o usuário escolher "Voltar".
func menuRedeIP() {
	for {
		_, opcao, err := selecionarMenu("Rede/IP", []string{
			"Alterar IP",
			"Gerenciar Rotas",
			"Voltar",
		})
		if err != nil {
			return
		}

		switch opcao {
		case "Alterar IP":
			alterarIP()
		case "Gerenciar Rotas":
			menuRotas()
		case "Voltar":
			return
		}
	}
}

// alterarIP conduz o usuário pelo processo de alteração de IP com rollback automático.
//
// Fluxo completo:
//  1. Lista interfaces disponíveis para seleção (exceto loopback)
//  2. Solicita o novo IP em notação CIDR
//  3. Solicita o gateway para testar conectividade após a mudança
//  4. Aplica o novo IP via network.AlterarIP
//  5. Aguarda 2 segundos para o kernel processar a mudança
//  6. Testa ping ao gateway
//     - Falha: restaura o IP original com network.RollbackIP
//     - Sucesso: mantém a nova configuração
//
// Nota: os endereços anteriores são mantidos em memória apenas durante esta função.
// Se o processo for encerrado entre AlterarIP e RollbackIP, a restauração é manual.
func alterarIP() {
	interfaces, err := network.ListarInterfaces()
	if err != nil {
		fmt.Println("Erro ao listar interfaces:", err)
		return
	}

	if len(interfaces) == 0 {
		fmt.Println("Nenhuma interface de rede encontrada.")
		return
	}

	// Monta labels descritivos combinando nome, IP atual e status da interface
	labels := make([]string, len(interfaces))
	for i, iface := range interfaces {
		status := "inativo"
		if iface.Ativo {
			status = "ativo"
		}
		ip := "sem IP"
		if len(iface.Enderecos) > 0 {
			ip = iface.Enderecos[0]
		}
		labels[i] = fmt.Sprintf("%-10s  %-20s [%s]", iface.Nome, ip, status)
	}

	// O índice retornado pelo menu corresponde à posição na slice de interfaces
	idx, _, err := selecionarMenu("Selecione a interface", labels)
	if err != nil {
		return
	}
	ifaceNome := interfaces[idx].Nome

	// Solicita o novo IP; ValidarCIDR garante formato correto antes de aplicar
	novoCIDR, err := lerTexto("Novo IP com máscara (ex: 192.168.1.100/24)", func(s string) error {
		return network.ValidarCIDR(s)
	})
	if err != nil {
		return
	}

	// O gateway é usado apenas para testar conectividade — não é configurado como rota
	gateway, err := lerTexto("Gateway para teste de conectividade", func(s string) error {
		return network.ValidarIP(s)
	})
	if err != nil {
		return
	}

	fmt.Printf("\nAlterando IP de %s para %s...\n", ifaceNome, novoCIDR)

	// AlterarIP retorna os endereços anteriores para possibilitar rollback
	anteriores, err := network.AlterarIP(ifaceNome, novoCIDR)
	if err != nil {
		fmt.Println("Erro ao alterar IP:", err)
		return
	}

	// Aguarda o kernel processar a nova configuração antes de testar
	time.Sleep(2 * time.Second)

	fmt.Printf("Testando conectividade com o gateway %s...\n", gateway)
	if err := network.TestarConectividade(gateway); err != nil {
		fmt.Printf("Falha! %v\n", err)
		fmt.Println("Restaurando IP original...")

		if rbErr := network.RollbackIP(ifaceNome, anteriores); rbErr != nil {
			// Rollback falhou — situação crítica, usuário precisa intervir manualmente
			fmt.Println("Erro no rollback:", rbErr)
			fmt.Println("ATENÇÃO: restaure o IP manualmente!")
		} else {
			fmt.Println("IP restaurado com sucesso.")
		}
		return
	}

	fmt.Printf("Conectividade confirmada! IP de %s alterado para %s.\n\n", ifaceNome, novoCIDR)
}
