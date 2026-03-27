// Este arquivo implementa o submenu de teste de conexão.
// Agrupa quatro ferramentas de diagnóstico: ping, traceroute,
// análise de perda de pacotes e resolução DNS.
// Toda lógica de rede é delegada ao pacote network.
package menu

import (
	"fmt"

	"netcheck/internal/network"
)

// menuTesteConexao exibe o submenu de diagnóstico em loop até o usuário escolher "Voltar".
func menuTesteConexao() {
	for {
		_, opcao, err := selecionarMenu("Teste de Conexão", []string{
			"Ping / Latência",
			"Traceroute",
			"Perda de Pacotes",
			"Resolução DNS",
			"Voltar",
		})
		if err != nil {
			return
		}

		switch opcao {
		case "Ping / Latência":
			testarPing()
		case "Traceroute":
			testarTraceroute()
		case "Perda de Pacotes":
			testarPerdaPacotes()
		case "Resolução DNS":
			testarDNS()
		case "Voltar":
			return
		}
	}
}

// pedirHost solicita um host ou IP ao usuário com validação em tempo real.
// Reutilizada por todas as funções de teste que precisam de um destino.
func pedirHost() (string, error) {
	return lerTexto("Host ou IP", func(s string) error {
		return network.ValidarHost(s)
	})
}

// testarPing envia 4 pacotes ICMP e exibe latência mínima, média e máxima.
func testarPing() {
	host, err := pedirHost()
	if err != nil {
		return
	}

	fmt.Printf("\nExecutando ping para %s...\n", host)

	resultado, err := network.PingHost(host, 4)
	if err != nil {
		fmt.Println("Erro:", err)
		return
	}

	fmt.Println("\n--- Resultado do Ping ---")
	fmt.Printf("Host:      %s\n", resultado.Host)
	fmt.Printf("Enviados:  %d | Recebidos: %d\n", resultado.Enviados, resultado.Recebidos)
	fmt.Printf("Latência:  min=%.2fms  avg=%.2fms  max=%.2fms\n",
		ms(resultado.MinRTT), ms(resultado.AvgRTT), ms(resultado.MaxRTT))
	fmt.Println()
}

// testarTraceroute mapeia os saltos (hops) até o destino e exibe a saída bruta do comando.
// Se o traceroute retornar erro mas tiver saída parcial, ela é exibida mesmo assim.
func testarTraceroute() {
	host, err := pedirHost()
	if err != nil {
		return
	}

	fmt.Printf("\nExecutando traceroute para %s...\n\n", host)

	resultado, err := network.Traceroute(host)
	if err != nil {
		fmt.Println("Erro:", err)
		// Exibe saída parcial caso o traceroute tenha conseguido alguns hops antes de falhar
		if resultado.Raw != "" {
			fmt.Println(resultado.Raw)
		}
		return
	}

	fmt.Println("--- Traceroute ---")
	fmt.Println(resultado.Raw)
}

// testarPerdaPacotes envia 10 pacotes e classifica a perda em três níveis:
// OK (0%), ATENÇÃO (1-49%) e CRÍTICO (50-100%).
func testarPerdaPacotes() {
	host, err := pedirHost()
	if err != nil {
		return
	}

	fmt.Printf("\nTestando perda de pacotes para %s (10 pacotes)...\n", host)

	resultado, err := network.PingHost(host, 10)
	if err != nil {
		fmt.Println("Erro:", err)
		return
	}

	// Classificação da qualidade do link com base na perda de pacotes
	status := "OK - Sem perda"
	if resultado.PerdaPct > 0 && resultado.PerdaPct < 50 {
		status = "ATENÇÃO - Perda parcial"
	} else if resultado.PerdaPct >= 50 {
		status = "CRÍTICO - Alta perda"
	}

	fmt.Println("\n--- Perda de Pacotes ---")
	fmt.Printf("Host:     %s\n", resultado.Host)
	fmt.Printf("Enviados: %d | Recebidos: %d\n", resultado.Enviados, resultado.Recebidos)
	fmt.Printf("Perda:    %.1f%%\n", resultado.PerdaPct)
	fmt.Printf("Status:   %s\n\n", status)
}

// testarDNS resolve um hostname para seus IPs usando os nameservers do sistema.
// Os nameservers consultados são os configurados no /etc/resolv.conf,
// que podem ter sido alterados via o submenu DNS desta ferramenta.
func testarDNS() {
	host, err := lerTexto("Hostname para resolver", func(s string) error {
		return network.ValidarHost(s)
	})
	if err != nil {
		return
	}

	resultado, err := network.ResolverDNS(host)
	if err != nil {
		fmt.Println("Erro:", err)
		return
	}

	fmt.Println("\n--- Resolução DNS ---")
	fmt.Printf("Host: %s\n", resultado.Host)
	fmt.Println("IPs resolvidos:")
	for _, ip := range resultado.IPs {
		fmt.Printf("  - %s\n", ip)
	}
	fmt.Println()
}

// ms converte time.Duration para milissegundos (float64).
// Usa Microseconds() para maior precisão antes de dividir por 1000.
func ms(d interface{ Microseconds() int64 }) float64 {
	return float64(d.Microseconds()) / 1000
}
