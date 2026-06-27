package internal

import (
	"context"
	"errors"
	"fmt"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/grpccreds"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/grpcserver"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/audit"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/config"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/handler"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/repository"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/scheduler"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

// App — точка входа HTTP-приложения: роутер и конфигурация.
type App struct {
	handler         handler.MetricsHandler
	router          http.Handler
	config          *config.Config
	auditPub        *audit.Publisher
	httpServer      *http.Server
	grpcServer      *grpc.Server
	grpcListener    net.Listener
	fileService     service.FileService
	schedulerCancel context.CancelFunc
}

// NewApp инициализирует хранилище, сервисы, handlers и HTTP-роутер.
func NewApp(cfg *config.Config, pool *pgxpool.Pool, logger *zap.Logger) (*App, error) {
	var metricsRepo repository.MetricsStorage
	if cfg.Server.DatabaseDSN != "" {
		metricsRepo = repository.NewPostgresStorage(pool)
	} else {
		metricsRepo = repository.NewMemStorage()
	}
	auditPub := audit.NewFromConfig(cfg.Server)

	metricsFile := repository.NewFileStorage(cfg.Server.FileStoragePath)
	metricsService := service.NewMetricsService(metricsRepo, metricsFile, cfg, auditPub)
	fileService := service.NewMetricsFileService(metricsRepo, metricsFile, cfg)

	ctxLoad, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := fileService.LoadMeticsFromFile(ctxLoad); err != nil {
		log.Printf("load metrics from file: %v", err)
	}
	metricsScheduler := scheduler.NewMetricsScheduler(cfg, fileService)
	schedCtx, schedulerCancel := context.WithCancel(context.Background())
	go metricsScheduler.RunWriteMetricsFile(schedCtx)

	metricsHandler := handler.NewMetricsHandler(metricsService)
	metricsJSONHandler := handler.NewMetricsJSONHandler(metricsService, cfg)

	metricsGRPCServer := grpcserver.NewMetricsGRPCServer(metricsService)
	var (
		grpcServer   *grpc.Server
		grpcListener net.Listener
	)
	if cfg.Server.GRPCAddress != "" {
		lis, err := net.Listen("tcp", cfg.Server.GRPCAddress)
		if err != nil {
			schedulerCancel()
			return nil, fmt.Errorf("listen gRPC on %s: %w", cfg.Server.GRPCAddress, err)
		}

		creds, err := grpccreds.ServerCredentials(cfg.Server.GRPCCert, cfg.Server.GRPCKey)
		if err != nil {
			_ = lis.Close()
			schedulerCancel()
			return nil, fmt.Errorf("grpc tls: %w", err)
		}

		grpcServer = grpc.NewServer(
			grpc.Creds(creds),
			grpc.ChainUnaryInterceptor(
				grpcserver.UnaryInterceptors(logger, cfg.Server.TrustedSubnet)...,
			),
		)
		proto.RegisterMetricsServer(grpcServer, metricsGRPCServer)
		grpcListener = lis
	}

	router := SetupRouter(metricsHandler, metricsJSONHandler, pool, cfg)
	return &App{
		handler:         metricsHandler,
		router:          router,
		config:          cfg,
		auditPub:        auditPub,
		fileService:     fileService,
		schedulerCancel: schedulerCancel,
		httpServer: &http.Server{
			Addr:    cfg.Server.Address,
			Handler: router,
		},
		grpcServer:   grpcServer,
		grpcListener: grpcListener,
	}, nil
}

// Run запускает HTTP-сервер на адресе из конфигурации.
func (a *App) Run() error {
	log.Printf("HTTP server starting on %s", a.config.Server.Address)
	if a.grpcServer != nil {
		go func() {
			log.Printf("gRPC server starting on %s", a.config.Server.GRPCAddress)
			if err := a.grpcServer.Serve(a.grpcListener); err != nil {
				log.Printf("gRPC server stopped: %v", err)
			}
		}()
	}
	err := a.httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (a *App) Shutdown(ctx context.Context) error {
	if a.schedulerCancel != nil {
		a.schedulerCancel()
	}
	if err := a.httpServer.Shutdown(ctx); err != nil {
		return err
	}
	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}
	if a.fileService != nil {
		if err := a.fileService.WriteMetricsFile(ctx); err != nil {
			return err
		}
	}
	return a.Close()
}

func (a *App) Close() error {
	if a == nil {
		return nil
	}
	return a.auditPub.Close()
}
