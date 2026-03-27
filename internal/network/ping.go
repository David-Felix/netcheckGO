// Este arquivo implementa o envio de pacotes ICMP (ping) usando a biblioteca
// github.com/go-ping/ping, que oferece estatísticas detalhadas como latência
// e perda de pacotes. Requer execução como root ou CAP_NET_RAW.
package network

import (
	"fmt"
	"time"

	"github.com/go-ping/ping"
)

// ResultadoPing agrupa as estatísticas retornadas após uma sessão de ping.
type ResultadoPing struct {
	Host      string        // Endereço ou hostname testado
	Enviados  int           // Total de pacotes enviados
	Recebidos int           // Total de pacotes recebidos com resposta
	PerdaPct  float64       // Percentual de pacotes perdidos (0.0 a 100.0)
	MinRTT    time.Duration // Menor tempo de resposta registrado
	AvgRTT    time.Duration // Tempo médio de resposta
	MaxRTT    time.Duration // Maior tempo de resposta registrado
}

// PingHost envia `count` pacotes ICMP para o host e retorna as estatísticas.
// Usa modo privilegiado (raw socket) que requer root ou CAP_NET_RAW.
// O timeout total da sessão é de 10 segundos.
func PingHost(host string, count int) (ResultadoPing, error) {
	if err := ValidarHost(host); err != nil {
		return ResultadoPing{}, err
	}

	pinger, err := ping.NewPinger(host)
	if err != nil {
		return ResultadoPing{}, fmt.Errorf("erro ao criar pinger: %w", err)
	}

	pinger.Count = count
	pinger.Timeout = 10 * time.Second
	// SetPrivileged(true) usa ICMP raw socket (necessário no Linux sem configuração extra)
	pinger.SetPrivileged(true)

	if err := pinger.Run(); err != nil {
		return ResultadoPing{}, fmt.Errorf(
			"erro ao executar ping: %w\nDica: execute como root ou: sudo setcap cap_net_raw=+ep ./netcheck",
			err,
		)
	}

	s := pinger.Statistics()
	return ResultadoPing{
		Host:      host,
		Enviados:  s.PacketsSent,
		Recebidos: s.PacketsRecv,
		PerdaPct:  s.PacketLoss,
		MinRTT:    s.MinRtt,
		AvgRTT:    s.AvgRtt,
		MaxRTT:    s.MaxRtt,
	}, nil
}

// TestarConectividade faz um ping rápido (3 pacotes) e retorna erro se o host
// estiver inacessível (100% de perda). Usada internamente para validar
// rotas e gateways após alterações de configuração.
func TestarConectividade(host string) error {
	resultado, err := PingHost(host, 3)
	if err != nil {
		return err
	}
	if resultado.PerdaPct == 100 {
		return fmt.Errorf("host %s inacessível: 100%% de perda de pacotes", host)
	}
	return nil
}
