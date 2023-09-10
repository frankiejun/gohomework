//go:build wireinject

package main

import (
	"gitee.com/geekbang/basic-go/webook/internal/repository"
	"gitee.com/geekbang/basic-go/webook/internal/repository/cache"
	"gitee.com/geekbang/basic-go/webook/internal/repository/dao"
	"gitee.com/geekbang/basic-go/webook/internal/repository/localcache"
	"gitee.com/geekbang/basic-go/webook/internal/service"
	"gitee.com/geekbang/basic-go/webook/internal/web"
	"gitee.com/geekbang/basic-go/webook/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		//基础部分
		//wire.Bind(new(redis.Cmdable), new(*redis.Client)),
		ioc.InitRedis, ioc.InitDB,

		// DAO 部分
		dao.NewUserDAO,

		// Cache 部分
		//cache.NewCodeCache,
		cache.NewUserCache,
		localcache.NewCodeCache,

		// repository 部分
		repository.NewUserRepository,
		repository.NewCodeRepository,

		// service 部分
		ioc.InitSmsService,
		service.NewCodeService,
		service.NewUserService,

		// handler 部分
		web.NewUserHandler,

		// gin 的中间件
		ioc.GinMiddlewares,

		// Web 服务器
		ioc.InitWebServer,
	)
	// 随便返回一个
	return gin.Default()
}
