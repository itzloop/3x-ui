package global

import (
	"context"
	"github.com/robfig/cron/v3"
	_ "unsafe"
	"x-ui/database/model"
)

var (
	webServer       WebServer
	inboundsService InboundsService
)

type WebServer interface {
	GetCron() *cron.Cron
	GetCtx() context.Context
}

func SetWebServer(s WebServer) {
	webServer = s
}

func GetWebServer() WebServer {
	return webServer
}

type InboundsService interface {
	AddInbounds(inbounds []*model.Inbound) error
}

func SetInbounds(inboundsInterface InboundsService) {
	inboundsService = inboundsInterface
}

func GetInbounds() InboundsService {
	return inboundsService
}
