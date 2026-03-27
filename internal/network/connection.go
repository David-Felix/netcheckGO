// Este arquivo implementa as ferramentas de diagnóstico de rede:
// traceroute (mapeamento do caminho dos pacotes) e resolução DNS.
// O traceroute delega ao binário do sistema via os/exec.
// A resolução DNS usa o resolver nativo do Go (net.LookupHost),
// que respeita os nameservers configurados no /etc/resolv.conf.
package network

import (
	"fmt"
	"net"
	"os/exec"
)

// ResultadoTraceroute armazena a saída de uma execução de traceroute.
type ResultadoTraceroute struct {
	Host string // Destino do traceroute
	Raw  string // Saída bruta do comando (exibida diretamente ao usuário)
}

// ResultadoDNS armazena o resultado de uma resolução de nome.
type ResultadoDNS struct {
	Host string   // Hostname consultado
	IPs  []string // Lista de IPs resolvidos para o hostname
}

// Traceroute mapeia o caminho dos pacotes até o host de destino.
// Tenta usar o binário "traceroute"; se não encontrado, tenta "tracepath".
// O host é validado antes de ser passado ao exec.Command para evitar
// injeção de comandos (o valor é passado como argumento, nunca interpolado em shell).
func Traceroute(host string) (ResultadoTraceroute, error) {
	if err := ValidarHost(host); err != nil {
		return ResultadoTraceroute{}, err
	}

	// exec.LookPath localiza o binário no PATH do sistema
	bin, err := exec.LookPath("traceroute")
	if err != nil {
		// Fallback para tracepath, presente em algumas distribuições por padrão
		bin, err = exec.LookPath("tracepath")
		if err != nil {
			return ResultadoTraceroute{}, fmt.Errorf(
				"traceroute não encontrado. Instale com: sudo apt install traceroute",
			)
		}
	}

	// Flags: -n evita resolução DNS dos hops (mais rápido), -m 20 limita a 20 saltos
	// host é passado como argumento separado — sem interpolação em string de shell
	cmd := exec.Command(bin, "-n", "-m", "20", host)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Retorna a saída mesmo em caso de erro para exibir hops parciais
		return ResultadoTraceroute{Host: host, Raw: string(out)},
			fmt.Errorf("traceroute falhou: %w", err)
	}

	return ResultadoTraceroute{Host: host, Raw: string(out)}, nil
}

// ResolverDNS resolve um hostname para seus endereços IP.
// Usa net.LookupHost, que consulta os nameservers configurados
// no /etc/resolv.conf — incluindo os adicionados via a função de DNS desta ferramenta.
func ResolverDNS(host string) (ResultadoDNS, error) {
	if err := ValidarHost(host); err != nil {
		return ResultadoDNS{}, err
	}

	ips, err := net.LookupHost(host)
	if err != nil {
		return ResultadoDNS{}, fmt.Errorf("falha na resolução DNS de '%s': %w", host, err)
	}

	return ResultadoDNS{Host: host, IPs: ips}, nil
}
