package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IRSHIT033/Atlas/backend"
	"github.com/IRSHIT033/Atlas/frontend"
	"github.com/IRSHIT033/Atlas/serverpool"
	"github.com/IRSHIT033/Atlas/utils"
	"go.uber.org/zap"
)

func main() {
	logger := utils.InitLogger()
	defer logger.Sync()

	config, err := utils.GetLBConfig()
	if err != nil {
		utils.Logger.Fatal(err.Error())
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverPool, err := serverpool.NewServerPool(utils.GetLoadBalanceStrategy(config.Strategy))
	if err != nil {
		utils.Logger.Fatal(err.Error())
	}

	loadBalancer := frontend.NewLoadBalancer(serverPool)

	for _, u := range config.Backends {
		endpoint, err := url.Parse(u.Url)
		if err != nil {
			logger.Fatal(err.Error(), zap.String("URL", u.Url))

		}
		reverseProxy := httputil.NewSingleHostReverseProxy(endpoint)
		backendServer := backend.NewBackend(endpoint, reverseProxy)
		reverseProxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
			logger.Error("error handling the request",
				zap.String("host", endpoint.Host),
				zap.Error(e),
			)
			backendServer.SetAlive(false)
			if !frontend.AllowRetry(request) {
				utils.Logger.Info(
					"Max retry attempts reached, terminating",
					zap.String("address", request.RemoteAddr),
					zap.String("path", request.URL.Path),
				)
				http.Error(writer, "Service not available", http.StatusServiceUnavailable)
				return
			}

			logger.Info(
				"Attempting retry",
				zap.String("address", request.RemoteAddr),
				zap.String("URL", request.URL.Path),
				zap.Bool("retry", true),
			)
			loadBalancer.Serve(
				writer,
				request.WithContext(
					context.WithValue(request.Context(), frontend.RETRY_ATTEMPTED, true),
				),
			)
		}
		serverPool.AddBackend(backendServer)

	}
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: http.HandlerFunc(loadBalancer.Serve),
	}

	go serverpool.LaunchHealthCheck(ctx, serverPool)

	go func() {
		<-ctx.Done()
		shutdownCtx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatal(err)
		}
	}()

	logger.Info(
		"Load Balancer started",
		zap.Int("port", config.Port),
	)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Fatal("ListenAndServe() error", zap.Error(err))
	}

}
