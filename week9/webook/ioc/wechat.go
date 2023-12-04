package ioc

import (
	"gitee.com/geekbang/basic-go/webook/internal/service/oauth2/wechat"
	"gitee.com/geekbang/basic-go/webook/pkg/logger"
	"os"
)

func InitWechatService(logger logger.LoggerV1) wechat.Service {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		appId = "123"
		//panic("没有找到环境变量 WECHAT_APP_ID ")
	}
	appKey, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		appKey = "123"
		//logger.Warn("没有找到环境变量 WECHAT_APP_SECRET")
	}

	return wechat.NewService(appId, appKey, logger)
}
