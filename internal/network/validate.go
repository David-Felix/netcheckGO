// Pacote network contém toda a lógica de rede da aplicação.
// Este arquivo centraliza as funções de validação de entrada do usuário,
// garantindo que dados inválidos ou maliciosos nunca cheguem às funções
// que interagem com o sistema operacional.
package network

import (
	"fmt"
	"net"
	"regexp"
)

// hostnameRegex define os caracteres permitidos em um hostname.
// Restringe a apenas letras, números, pontos e hífens para prevenir
// injeção de comandos em chamadas ao os/exec (ex: traceroute).
var hostnameRegex = regexp.MustCompile(`^[a-zA-Z0-9.\-]+$`)

// ValidarHost verifica se o valor informado é um IP ou hostname válido.
// Primeiro tenta interpretar como IP; se não for, valida como hostname.
// Retorna erro se o host for vazio, contiver caracteres inválidos ou for muito longo.
func ValidarHost(host string) error {
	if host == "" {
		return fmt.Errorf("host não pode ser vazio")
	}
	// Se for um IP válido (IPv4 ou IPv6), aceita diretamente
	if net.ParseIP(host) != nil {
		return nil
	}
	// Valida como hostname usando a regex restritiva
	if !hostnameRegex.MatchString(host) {
		return fmt.Errorf("hostname inválido: use apenas letras, números, pontos e hífens")
	}
	// RFC 1035 limita hostnames a 253 caracteres
	if len(host) > 253 {
		return fmt.Errorf("hostname muito longo (máx. 253 caracteres)")
	}
	return nil
}

// ValidarIP verifica se o valor informado é um endereço IP válido (v4 ou v6).
func ValidarIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("IP inválido: %s", ip)
	}
	return nil
}

// ValidarCIDR verifica se o valor informado está no formato CIDR correto.
// Exemplos válidos: "192.168.1.0/24", "10.0.0.0/8".
// Nota: net.ParseCIDR normaliza o IP para o endereço de rede (ex: 192.168.1.100/24 → 192.168.1.0/24).
func ValidarCIDR(cidr string) error {
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("CIDR inválido: %s (exemplo: 192.168.1.0/24)", cidr)
	}
	return nil
}

// ValidarInterface verifica se a interface de rede informada existe no sistema.
// Usa net.InterfaceByName que consulta diretamente o kernel, sem chamadas shell.
func ValidarInterface(nome string) error {
	_, err := net.InterfaceByName(nome)
	if err != nil {
		return fmt.Errorf("interface '%s' não encontrada", nome)
	}
	return nil
}
