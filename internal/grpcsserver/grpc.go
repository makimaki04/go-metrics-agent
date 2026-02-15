package grpcserver

import (
	"context"
	"net"
	"strings"

	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
	"github.com/makimaki04/go-metrics-agent.git/internal/observer"
	pb "github.com/makimaki04/go-metrics-agent.git/internal/proto"
	"github.com/makimaki04/go-metrics-agent.git/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type MetricsServer struct {
	pb.UnimplementedMetricsServer

	service service.MetricsService
	logger  *zap.Logger
}

func NewMetricsServer(service service.MetricsService, logger *zap.Logger) *MetricsServer {

	return &MetricsServer{
		service: service,
		logger:  logger,
	}
}

func (m *MetricsServer) UpdateMetrics(ctx context.Context, in *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	var response pb.UpdateMetricsResponse
	metrics, err := m.convertToMetric(in.GetMetrics())
	if err != nil {
		m.logger.Error("grpc server: failed to update metric batch", zap.Error(err))
		return nil, err
	}

	if err := m.service.UpdateMetricBatch(ctx, metrics); err != nil {
		m.logger.Error("grpc server: failed to update metric batch", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update metric batch")
	}

	return &response, nil
}

func (m *MetricsServer) convertToMetric(pbMetrics []*pb.Metric) ([]models.Metrics, error) {
	metrics := make([]models.Metrics, 0, len(pbMetrics))

	for _, metric := range pbMetrics {
		t := metric.GetType()
		switch t {
		case pb.Metric_GAUGE:
			v := metric.GetValue()
			metrics = append(metrics, models.Metrics{
				ID:    metric.GetId(),
				MType: models.Gauge,
				Value: &v,
			})
		case pb.Metric_COUNTER:
			d := metric.GetDelta()
			metrics = append(metrics, models.Metrics{
				ID:    metric.GetId(),
				MType: models.Counter,
				Delta: &d,
			})
		default:
			m.logger.Error("grpc server: couldn't conver metric: unknown metric type")
			return nil, status.Error(codes.InvalidArgument, "unknown metric type")
		}
	}

	return metrics, nil
}

func RunGRPC(url string, service service.MetricsService, ipNet *net.IPNet, logger *zap.Logger) (*grpc.Server, net.Listener, error) {
	listen, err := net.Listen("tcp", url)
	if err != nil {
		logger.Error("grpc server: listener initialization error", zap.Error(err))
		return nil, nil, err
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(TrustedSubnetInterceptor(ipNet)))
	pb.RegisterMetricsServer(s, NewMetricsServer(service, logger))
	logger.Info("grpc server successfully started")

	return s, listen, nil
}

func TrustedSubnetInterceptor(ipNet *net.IPNet) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if ipNet == nil {
			return handler(ctx, req)
		}

		md, _ := metadata.FromIncomingContext(ctx)
		vals := md.Get("x-real-ip")
		if len(vals) == 0 || strings.TrimSpace(vals[0]) == "" {
			return nil, status.Error(codes.PermissionDenied, "missing x-real-ip")
		}
		realIP := strings.TrimSpace(vals[0])
		ip := net.ParseIP(realIP)
		if ip == nil {
			return nil, status.Error(codes.PermissionDenied, "wrong x-real-ip format")
		}

		ok := ipNet.Contains(ip)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "x-real-ip not in trusted subnet")
		}

		ctx = context.WithValue(ctx, observer.ReqIDKey, ip.String())

		return handler(ctx, req)
	}
}
