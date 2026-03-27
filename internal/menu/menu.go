// Pacote menu contém toda a interface com o usuário da aplicação netcheck.
// Este pacote é responsável exclusivamente por exibir menus, ler entradas
// e mostrar resultados — não contém nenhuma lógica de rede.
// Toda operação de rede é delegada ao pacote internal/network.
package menu

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

// Iniciar é o ponto de entrada do menu. Exibe o menu principal em loop
// até que o usuário escolha "Sair" ou pressione Ctrl+C.
func Iniciar() {
	for {
		_, opcao, err := selecionarMenu("NETCHECK - Menu Principal", []string{
			"Rede/IP",
			"DNS",
			"Teste de Conexão",
			"Sair",
		})
		if err != nil {
			// err != nil ocorre quando o usuário pressiona Ctrl+C
			return
		}

		switch opcao {
		case "Rede/IP":
			menuRedeIP()
		case "DNS":
			menuDNS()
		case "Teste de Conexão":
			menuTesteConexao()
		case "Sair":
			fmt.Println("Encerrando...")
			return
		}
	}
}

// selecionarMenu exibe uma lista navegável com setas do teclado e retorna
// o índice (base 0) e o texto do item selecionado.
// Retorna erro se o usuário pressionar Ctrl+C ou Esc.
func selecionarMenu(label string, itens []string) (int, string, error) {
	m := promptui.Select{
		Label: label,
		Items: itens,
		Size:  10, // número máximo de itens visíveis antes de rolar
	}
	return m.Run()
}

// lerTexto exibe um prompt de texto e retorna o valor digitado pelo usuário.
// O parâmetro validate é chamado a cada tecla digitada — se retornar erro,
// o promptui exibe a mensagem de erro em tempo real e impede o envio.
// Passar nil em validate desabilita a validação.
func lerTexto(label string, validate promptui.ValidateFunc) (string, error) {
	p := promptui.Prompt{
		Label:    label,
		Validate: validate,
	}
	return p.Run()
}
