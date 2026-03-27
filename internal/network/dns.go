// Este arquivo gerencia os nameservers (servidores DNS) configurados no sistema.
// Todas as operações leem e escrevem no arquivo /etc/resolv.conf,
// que é o arquivo padrão do Linux para configuração de resolução de nomes.
package network

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// resolvConf é o caminho padrão do arquivo de configuração DNS no Linux.
const resolvConf = "/etc/resolv.conf"

// LerNameservers lê o arquivo /etc/resolv.conf e retorna uma lista
// com os IPs dos nameservers configurados (linhas "nameserver <ip>").
func LerNameservers() ([]string, error) {
	file, err := os.Open(resolvConf)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir %s: %w", resolvConf, err)
	}
	defer file.Close()

	var servers []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Considera apenas linhas que começam com "nameserver"
		if strings.HasPrefix(line, "nameserver") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				servers = append(servers, fields[1]) // campo [1] é o IP do DNS
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("erro ao ler %s: %w", resolvConf, err)
	}

	return servers, nil
}

// AdicionarNameserver valida e adiciona um novo IP de nameserver ao /etc/resolv.conf.
// Rejeita IPs inválidos e duplicatas antes de escrever no arquivo.
func AdicionarNameserver(ip string) error {
	if err := ValidarIP(ip); err != nil {
		return err
	}

	servers, err := LerNameservers()
	if err != nil {
		return err
	}

	// Verifica se o nameserver já está configurado para evitar duplicatas
	for _, s := range servers {
		if s == ip {
			return fmt.Errorf("nameserver %s já está configurado", ip)
		}
	}

	return atualizarResolvConf(append(servers, ip))
}

// RemoverNameserver remove o nameserver na posição indicada pelo índice (base 0).
// Lê a lista atual, remove o item e reescreve o arquivo.
func RemoverNameserver(index int) error {
	servers, err := LerNameservers()
	if err != nil {
		return err
	}

	if index < 0 || index >= len(servers) {
		return fmt.Errorf("índice inválido: %d", index)
	}

	// Remove o elemento do slice sem alterar a ordem dos demais
	return atualizarResolvConf(append(servers[:index], servers[index+1:]...))
}

// atualizarResolvConf reescreve o /etc/resolv.conf de forma atômica.
// Escreve primeiro em um arquivo temporário e depois usa os.Rename para
// substituir o original. Isso garante que o arquivo nunca fique corrompido
// caso o processo seja interrompido durante a escrita.
func atualizarResolvConf(servers []string) error {
	tmp := resolvConf + ".netcheck.tmp"

	file, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo temporário: %w", err)
	}

	for _, s := range servers {
		if _, err := fmt.Fprintf(file, "nameserver %s\n", s); err != nil {
			file.Close()
			return fmt.Errorf("erro ao escrever: %w", err)
		}
	}

	// Fecha antes do Rename para garantir que todos os dados foram gravados
	if err := file.Close(); err != nil {
		return err
	}

	// os.Rename é atômico no Linux quando origem e destino estão no mesmo filesystem
	return os.Rename(tmp, resolvConf)
}
