package main

import (
	"context"
	"crypto/rsa"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Viva-Fidel/metrics-and-alerting/internal/agent"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/config"
	pb "github.com/Viva-Fidel/metrics-and-alerting/internal/proto"
	"github.com/Viva-Fidel/metrics-and-alerting/internal/security"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"resty.dev/v3"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildInfo()

	// logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Загружаем параметры запуска
	flags, err := config.LoadAgentFlags()
	if err != nil {
		logger.Error("failed to load config", slog.Any("error", err))
	}

	var publicKey *rsa.PublicKey
	if flags.CryptoKey != "" {
		publicKey, err = security.LoadPublicKey(flags.CryptoKey)
		if err != nil {
			logger.Error("failed to load public key", slog.Any("error", err))
			os.Exit(1)
		}
	}

	// Инициализируем хранилище
	metrics := agent.NewMetrics()

	hostIP, err := agent.HostIP()
	if err != nil {
		logger.Error("failed to resolve host IP", slog.Any("error", err))
		os.Exit(1)
	}

	var httpClient *resty.Client
	var grpcConn *grpc.ClientConn
	var grpcClient pb.MetricsClient

	if grpcAddr := strings.TrimSpace(flags.GRPCAddress); grpcAddr != "" {
		grpcConn, err = grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logger.Error("failed to create gRPC client", slog.Any("error", err))
			os.Exit(1)
		}
		defer func() {
			if err := grpcConn.Close(); err != nil {
				logger.Error("failed to close gRPC client", slog.Any("error", err))
			}
		}()
		grpcClient = pb.NewMetricsClient(grpcConn)
	} else {
		httpClient = resty.New().
			SetBaseURL("http://" + flags.RunAddr).
			SetHeader(security.RealIPHeader, hostIP).
			SetRetryCount(3).
			SetRetryWaitTime(time.Second).
			SetRetryMaxWaitTime(5 * time.Second).
			AddRetryConditions(func(_ *resty.Response, err error) bool {
				return err != nil
			})
		defer func() {
			if err := httpClient.Close(); err != nil {
				logger.Error("failed to close http client", slog.Any("error", err))
			}
		}()
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	// Запускаем основной цикл сбора и отправки метрик
	agent.Run(ctx, logger, httpClient, grpcClient, hostIP, metrics, flags.PollInterval, flags.ReportInterval, flags.RateLimit, flags.Key, publicKey)
}
