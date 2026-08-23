package main

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/scarypuppp/metrics-service/internal/agent"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	agentConfig, err := agent.GetConfig()
	if err != nil {
		logger.Fatal("read agent config failed", zap.Error(err))
	}

	client := resty.NewWithClient(&http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	certificate, err := readCert(agentConfig.CryptoKey)
	if err != nil {
		logger.Fatal("error reading certificate", zap.Error(err))
	}

	host, err := getOutboundIP()
	if err != nil {
		logger.Fatal("error getting outbound ip", zap.Error(err))
	}

	transport := agent.TransportType(strings.ToUpper(agentConfig.Transport))

	var grpcConn *grpc.ClientConn
	if transport == agent.TransportGRPC {
		grpcConn, err = grpc.NewClient(agentConfig.GRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logger.Fatal("error creating grpc client", zap.Error(err))
		}
		defer grpcConn.Close()
	}

	newAgent := agent.NewAgent(
		agent.NewCollector(),
		agent.NewSender(client, grpcConn, agentConfig.ServerAddr, agentConfig.Key, certificate, host, transport),
		*logger,
		agentConfig.PoolInterval,
		agentConfig.ReportInterval,
	)

	log.Println("Agent started")
	newAgent.RunCtx(ctx, agentConfig.RateLimit)
	log.Println("Agent stopped")
}

func readCert(path string) (*x509.Certificate, error) {
	certificateBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	certificatePemBlock, _ := pem.Decode(certificateBytes)
	if certificatePemBlock == nil {
		return nil, fmt.Errorf("error decoding certfificate bytes")
	}
	certificate, err := x509.ParseCertificate(certificatePemBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return certificate, nil
}

func getOutboundIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String(), nil
		}
	}
	return "", fmt.Errorf("no suitable IP address found")
}
