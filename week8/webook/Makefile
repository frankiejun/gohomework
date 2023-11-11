.PHONY: docker
docker:
	# 把上次编译的东西删掉
	@rm webook || true
	@docker rmi -f flycash/webook:v0.0.1
	# 运行一下 go mod tidy，防止 go.sum 文件不对，编译失败
	@go mod tidy
	# 指定编译成在 ARM 架构的 linux 操作系统上运行的可执行文件，
	# 名字叫做 webook
	@GOOS=linux GOARCH=arm go build -tags=k8s -o webook .
	# 这里你可以随便改这个标签，记得对应的 k8s 部署里面也要改
	@docker build -t flycash/webook:v0.0.1 .

.PHONY:mock
mock:
	@mockgen -source=internal/service/user.go -package=svcmocks -destination=internal/service/mocks/user.mock.go
	@mockgen -source=internal/service/code.go -package=svcmocks -destination=internal/service/mocks/code.mock.go
	@mockgen -source=internal/service/interactive.go -package=svcmocks -destination=internal/service/mocks/interactive.mock.go
	@mockgen -source=internal/repository/interactive.go -package=repomocks -destination=internal/repository/mocks/interactive.mock.go

	@go mod tidy