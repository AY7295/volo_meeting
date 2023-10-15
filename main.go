package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"volo_meeting/api"
	"volo_meeting/config"
	"volo_meeting/internal/model"
	"volo_meeting/lib/db"
	"volo_meeting/lib/log"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	Init()

	srv := &http.Server{
		Addr:    viper.GetString("server.addr"),
		Handler: api.Init(),
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			zap.L().Error("Server ListenAndServe", zap.Error(err))
			panic(err)
		}
	}()

	closeServer(srv)

}

func Init() {
	config.Init()
	log.Init()

	db.Init()
	model.Init()
}

func closeServer(srv *http.Server) {
	defer func(l *zap.Logger) {
		err := l.Sync()
		if err != nil {
			panic(err)
		}
	}(zap.L())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Error("Server Shutdown", zap.Error(err))
	}
	zap.L().Info("Server exited")

}
