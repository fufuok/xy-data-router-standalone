package common

import (
	"context"
	"time"

	"github.com/fufuok/utils/timewheel"
	"github.com/panjf2000/gnet/pool/goroutine"
	"github.com/tidwall/gjson"
)

var (
	IPv4Zero = "0.0.0.0"
	CtxBG    = context.Background()

	// 同步时间差值
	GTimeSub time.Duration

	// 协程池
	Pool = goroutine.Default()

	// 时间轮, 精度 50ms, 1s, 1m
	TWms *timewheel.TimeWheel
	TWs  *timewheel.TimeWheel
	TWm  *timewheel.TimeWheel

	// 当前时间
	Now3399UTC string
	Now3399    string
)

func InitCommon() {
	// 30 秒
	TWms, _ = timewheel.NewTimeWheel(50*time.Millisecond, 600)
	TWms.Start()

	// 30 分
	TWs, _ = timewheel.NewTimeWheel(time.Second, 1800)
	TWs.Start()

	// 24 时
	TWm, _ = timewheel.NewTimeWheel(time.Minute, 1440)
	TWm.Start()

	// 初始化日志环境
	initLogger()

	// 初始化 ES 连接
	initES()

	// 每秒格式化时间字符串
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			// TODO: GTimeSub
			now := GetGlobalTime()
			Now3399UTC = now.Format("2006-01-02T15:04:05Z")
			Now3399 = now.Format(time.RFC3339)
			<-ticker.C
		}
	}()
}

func TWStop() {
	TWms.Stop()
	TWs.Stop()
	TWm.Stop()
}

func PoolRelease() {
	Pool.Release()
}

// 统一时间
func GetGlobalTime() time.Time {
	return time.Now().Add(GTimeSub)
}

// 统一时间并格式化
func GetGlobalDataTime(layout string) string {
	if layout == "" {
		layout = time.RFC3339
	}

	return GetGlobalTime().Format(layout)
}

// 检查必有字段: 只要存在该字段即可, 值可为空
func CheckRequiredField(body *[]byte, fields []string) bool {
	for _, field := range fields {
		if !gjson.GetBytes(*body, field).Exists() {
			return false
		}
	}

	return true
}
