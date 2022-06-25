package master

import (
	"log"

	"github.com/fufuok/utils"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/service"
	"github.com/fufuok/xy-data-router/tunnel"
)

// Watcher 监听程序二进制变化(重启)和配置文件(热加载)
func Watcher() {
	mainFile := utils.Executable(true)
	if mainFile == "" {
		log.Fatalln("Failed to initialize Watcher: miss executable", "\nbye.")
	}

	md5Main, _ := utils.MD5Sum(mainFile)
	md5Conf, _ := utils.MD5Sum(conf.ConfigFile)

	common.Log.Info().
		Str("main", mainFile).Str("config", conf.ConfigFile).
		Msg("Watching")

	go func() {
		ticker := common.TWm.NewTicker(conf.Config.SYSConf.WatcherIntervalDuration)
		defer ticker.Stop()

		for range ticker.C {
			// 程序二进制变化时重启
			md5New, _ := utils.MD5Sum(mainFile)
			if md5New != md5Main {
				md5Main = md5New
				common.Log.Warn().Msg(">>>>>>> restart main <<<<<<<")
				restartChan <- true
				continue
			}
			// 配置文件变化时热加载
			md5New, _ = utils.MD5Sum(conf.ConfigFile)
			if md5New != md5Conf {
				md5Conf = md5New
				if err := conf.LoadConf(); err != nil {
					common.Log.Error().Err(err).Msg("reload config")
					continue
				}

				// 重启程序指令
				if conf.Config.SYSConf.RestartMain {
					common.Log.Warn().Msg(">>>>>>> restart main(config) <<<<<<<")
					restartChan <- true
					continue
				}

				// 重新连接 ES
				if err := common.InitES(); err != nil {
					common.Log.Error().Err(err).Msg("Failed to update elasticsearch connection")
				}

				// 日志配置更新
				_ = common.InitLogger()

				// 更新 Tunnel 日志配置
				tunnel.InitLogger()

				// 更新 HTTP 客户端请求配置
				common.InitReq()

				// 同步数据分发器配置
				service.InitDataRouter()

				// 调节协程池
				service.TuneDataProcessorSize(conf.Config.DataConf.ProcessorSize)
				service.TuneESBulkWorkerSize(conf.Config.DataConf.ESBulkWorkerSize)

				// 更新配置文件监控周期
				ticker.Reset(conf.Config.SYSConf.WatcherIntervalDuration)

				common.Log.Warn().Msg(">>>>>>> reload config <<<<<<<")
				reloadChan <- true
			}
		}
	}()
}
