package ioc

import (
	"gitee.com/geekbang/basic-go/webook/internal/service/oauth2/wechat"
	"gitee.com/geekbang/basic-go/webook/pkg/logger"
	"os"
)

func InitWechatService(logger logger.LoggerV1) wechat.Service {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		//panic("没有找到环境变量 WECHAT_APP_ID ")
		appId = "appid"
	}
	appKey, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		//panic("没有找到环境变量 WECHAT_APP_SECRET")
		appKey = "appkey"
	}
	return wechat.NewService(appId, appKey, logger)
}
