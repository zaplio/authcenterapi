package app

import (
	"authcenterapi/internal/handler/rest"
	"authcenterapi/internal/provider"
	"authcenterapi/internal/repository"
	"authcenterapi/internal/service"
	"authcenterapi/model/constant"
	"authcenterapi/util"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Run(cfg *util.Config) {
	ctx := context.WithValue(context.Background(), constant.CtxReqIDKey, "MAIN")

	logger := provider.NewLogger()

	db, err := provider.NewPostgresConnection(ctx)
	if err != nil {
		logger.Errorfctx(provider.AppLog, ctx, false, "Failed connect to PostgreSQL: %v", err)
		return
	}

	redis, err := provider.NewRedisConnection(ctx)
	if err != nil {
		logger.Errorfctx(provider.AppLog, ctx, false, "Failed connect to Redis: %v", err)
		return
	}
	repo := repository.NewRepository(logger, db)
	svc := service.NewService(logger, redis, repo)

	server := &http.Server{}
	go func(logger provider.ILogger, svc service.IService) {
		app := rest.NewRest(logger, svc)
		addr := fmt.Sprintf(":%v", util.Configuration.Server.Port)
		server, err = app.CreateServer(addr)
		if err != nil {
			logger.Errorfctx(provider.AppLog, ctx, false, "Failed to create server: %v", err)
		}

		logger.Infofctx(provider.AppLog, ctx, "Server running at: %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorfctx(provider.AppLog, ctx, false, "Server error: %v", err)
		}

	}(logger, svc)

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)

	sig := <-shutdownCh
	logger.Infofctx(provider.AppLog, ctx, "Receiving signal: %s", sig)

	func() {
		shutdownCtx, cancel := context.WithTimeout(ctx, util.Configuration.Server.ShutdownTimeout)
		defer cancel()
		server.Shutdown(shutdownCtx)

		logger.Infofctx(provider.AppLog, ctx, "Successfully stop Application.")
	}()

}
