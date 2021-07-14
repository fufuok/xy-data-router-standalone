package tunnel

import (
	"log"
	"net"
	"sync/atomic"

	"github.com/lesismal/arpc"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/service"
)

func InitTunClient() {
	if conf.ForwardTunnel == "" {
		return
	}

	common.Log.Info().Str("addr", conf.ForwardTunnel).Msg("Start Tunnel Client")
	client, err := arpc.NewClient(func() (net.Conn, error) {
		return net.DialTimeout("tcp", conf.ForwardTunnel, conf.TunDialTimeout)
	})
	if err != nil {
		log.Fatalln("Failed to start Tunnel Client:", err, "\nbye.")
	}

	defer client.Stop()
	client.Codec = &genCodec{}

	// 接收数据转发到通道 (支持创建多个 client, 每 client 支持多协程并发处理数据)
	for item := range service.TunChan.Out {
		err = client.Notify(tunMethod, item, conf.TunSendTimeout)
		if err != nil {
			common.LogSampled.Info().Err(err).Msg("write to tunnel failed")
			atomic.AddUint64(&service.TunSendBadCounters, 1)
			continue
		}

		atomic.AddUint64(&service.TunSendCounters, 1)
	}
}