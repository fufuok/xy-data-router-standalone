package service

import (
	"runtime"
	"sync/atomic"
	"time"

	"github.com/fufuok/utils"
	"github.com/fufuok/utils/myip"
	"github.com/rs/zerolog"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

type tDataRouterStats struct {
	DataRouterQueue tChanLen
	APIQueue        tChanLen
}

type tChanLen struct {
	AllLen int
	BufLen int
}

var (
	// 系统启动时间
	start = time.Now()

	// 服务器 IP
	InternalIPv4 string
	ExternalIPv4 string
)

func initRuntime() {
	go func() {
		InternalIPv4 = myip.InternalIPv4()
	}()
	go func() {
		ExternalIPv4 = myip.ExternalIPAny(10)
	}()
}

// 运行状态
func RunningStatus() map[string]interface{} {
	return map[string]interface{}{
		"DATA": dataStats(),
		"SYS":  sysStatus(),
		"MEM":  memStats(),
	}
}

// 队列状态
func RunningQueueStatus() map[string]interface{} {
	return chanStats()
}

// 系统信息
func sysStatus() map[string]interface{} {
	return map[string]interface{}{
		"APPName":      conf.WebAPPName,
		"Version":      conf.CurrentVersion,
		"Update":       conf.LastChange,
		"Uptime":       time.Since(start).String(),
		"Debug":        conf.Config.SYSConf.Debug,
		"LogLevel":     zerolog.Level(conf.Config.SYSConf.Log.Level).String(),
		"ConfigVer":    conf.Config.SYSConf.MainConfig.ConfigVer,
		"ConfigMD5":    conf.Config.SYSConf.MainConfig.ConfigMD5,
		"GoVersion":    runtime.Version(),
		"NumCpus":      runtime.NumCPU(),
		"NumGoroutine": runtime.NumGoroutine(),
		"OS":           runtime.GOOS,
		"NumCgoCall":   runtime.NumCgoCall(),
		"InternalIPv4": InternalIPv4,
		"ExternalIPv4": ExternalIPv4,
	}
}

// 内存信息
func memStats() map[string]interface{} {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return map[string]interface{}{
		// 程序启动后累计申请的字节数
		"TotalAlloc":  ms.TotalAlloc,
		"TotalAlloc_": utils.HumanBytes(ms.TotalAlloc),
		// 虚拟占用, 总共向系统申请的字节数
		"HeapSys":  ms.HeapSys,
		"HeapSys_": utils.HumanBytes(ms.HeapSys),
		// 使用中或未使用, 但未被 GC 释放的对象的字节数
		"HeapAlloc":  ms.HeapAlloc,
		"HeapAlloc_": utils.HumanBytes(ms.HeapAlloc),
		// 使用中的对象的字节数
		"HeapInuse":  ms.HeapInuse,
		"HeapInuse_": utils.HumanBytes(ms.HeapInuse),
		// 已释放的内存, 还没被堆再次申请的内存
		"HeapReleased":  ms.HeapReleased,
		"HeapReleased_": utils.HumanBytes(ms.HeapReleased),
		// 没被使用的内存, 包含了 HeapReleased, 可被再次申请和使用
		"HeapIdle":  ms.HeapIdle,
		"HeapIdle_": utils.HumanBytes(ms.HeapIdle),
		// 下次 GC 的阈值, 当 HeapAlloc 达到该值触发 GC
		"NextGC":  ms.NextGC,
		"NextGC_": utils.HumanBytes(ms.NextGC),
		// 上次 GC 时间
		"LastGC": time.Unix(0, int64(ms.LastGC)).Format(time.RFC3339Nano),
		// GC 次数
		"NumGC": utils.Commau(ms.NextGC),
		// 被强制 GC 的次数
		"NumForcedGC": ms.NumForcedGC,
	}
}

// 数据处理信息
func dataStats() map[string]interface{} {
	return map[string]interface{}{
		"ESDataQueueAll": esChan.Len(),
		"ESDataQueueBuf": esChan.BufLen(),

		"CounterStartTime": counterStartTime,

		"ESDataCounters":             utils.Commau(atomic.LoadUint64(&esDataCounters)),
		"ESBulkWorkerCounters":       utils.Comma(atomic.LoadInt64(&esBulkWorkerCounters)),
		"ESBulkErrors______________": utils.Commau(atomic.LoadUint64(&esBulkErrors)),
		"ESBulkDropWorkerCounters__": utils.Commau(atomic.LoadUint64(&esBulkDropWorkerCounters)),

		"DataProcessorWorkerRunning": dataProcessorPool.Running(),
		"DataProcessorWorkerFree___": dataProcessorPool.Free(),
		"ESBulkWorkerRunning":        esBulkPool.Running(),
		"ESBulkWorkerFree__________": esBulkPool.Free(),
		"CommonPoolFree":             common.Pool.Free(),
		"CommonPoolRunning":          common.Pool.Running(),

		"HTTPRequestCounters": utils.Commau(atomic.LoadUint64(&HTTPRequestCounters)),
		"UDPRequestCounters":  utils.Commau(atomic.LoadUint64(&UDPRequestCounters)),
	}
}

// 数据队列信息
func chanStats() map[string]interface{} {
	stats := map[string]interface{}{}
	for item := range dataRouters.IterBuffered() {
		dr := item.Val.(*tDataRouter)
		stats[item.Key] = tDataRouterStats{
			DataRouterQueue: tChanLen{
				AllLen: dr.drChan.Len(),
				BufLen: dr.drChan.BufLen(),
			},
			APIQueue: tChanLen{
				AllLen: dr.drOut.apiChan.Len(),
				BufLen: dr.drOut.apiChan.BufLen(),
			},
		}
	}

	return stats
}
