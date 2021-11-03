package service

import (
	"bytes"
	"time"

	"github.com/fufuok/bytespool"
	"github.com/imroc/req"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

func apiWorker(dr *tDataRouter) {
	// 默认 1s 空转
	interval := 1
	if dr.apiConf.PostAPI.Interval > 0 {
		interval = dr.apiConf.PostAPI.Interval
		common.Log.Info().
			Int("interval", interval).Str("apiname", dr.apiConf.APIName).
			Msg("apiWorker start")
	}
	ticker := common.TWs.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	var buf bytes.Buffer
	for {
		select {
		case <-ticker.C:
			if dr.apiConf.PostAPI.Interval > 0 {
				// 校准推送周期
				if dr.apiConf.PostAPI.Interval != interval {
					interval = dr.apiConf.PostAPI.Interval
					ticker.Reset(time.Duration(interval) * time.Second)
					common.Log.Warn().
						Int("interval", interval).Str("apiname", dr.apiConf.APIName).
						Msg("apiWorker restart")
				}
				// 指定时间间隔推送数据
				postAPI(&buf, dr.apiConf.PostAPI.API)
			}
		case v, ok := <-dr.drOut.apiChan.Out:
			if !ok {
				// 消费所有数据
				postAPI(&buf, dr.apiConf.PostAPI.API)
				common.Log.Warn().Str("apiname", dr.apiConf.APIName).Msg("apiWorker exited")
				return
			}

			jsData := v.([]byte)
			buf.WriteByte(',')
			buf.Write(jsData)
			bytespool.Release(jsData)
		}
	}
}

// 推送数据到 API
func postAPI(buf *bytes.Buffer, api []string) {
	if buf.Len() == 0 {
		return
	}

	defer buf.Reset()

	apiBody := bytespool.Make(buf.Len() + 1)
	apiBody = append(apiBody, '[')
	apiBody = append(apiBody, buf.Bytes()[1:]...)
	apiBody = append(apiBody, ']')

	_ = common.Pool.Submit(func() {
		defer bytespool.Release(apiBody)
		for _, u := range api {
			if _, err := req.Post(u, req.BodyJSON(apiBody), conf.ReqUserAgent); err != nil {
				common.LogSampled.Error().Err(err).Str("url", u).Msg("apiWorker")
			}
		}
	})
}
