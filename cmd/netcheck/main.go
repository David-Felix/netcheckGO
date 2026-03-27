// Pacote main é o ponto de entrada da aplicação netcheck.
// Verifica se o programa está sendo executado como root antes de iniciar,
// pois operações de rede (netlink, ICMP) exigem privilégios elevados.
package main

import (
	"fmt"
	"os"

	"netcheck/internal/menu"
)

func main() {
	// Operações como alterar IP, gerenciar rotas e enviar pacotes ICMP
	// requerem CAP_NET_ADMIN e CAP_NET_RAW, disponíveis apenas para root.
	if os.Getuid() != 0 {
		fmt.Fprintln(os.Stderr, "Erro: netcheck precisa ser executado como root.")
		fmt.Fprintln(os.Stderr, "Use: sudo ./netcheck")
		os.Exit(1)
	}

	menu.Iniciar()
}
