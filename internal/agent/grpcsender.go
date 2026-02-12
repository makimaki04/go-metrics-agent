package agent

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
	pb "github.com/makimaki04/go-metrics-agent.git/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type GRPCSender struct {
	client  pb.MetricsClient
	IP      string
	storage SenderStorageIntreface
	conn    *grpc.ClientConn
}

func NewGRPCSender(address string, storage SenderStorageIntreface) (*GRPCSender, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("couldn't deal with %s: %v", address, err)
	}
	c := pb.NewMetricsClient(conn)

	udpConn, err := net.Dial("udp", address)
	if err != nil {
		return nil, fmt.Errorf("couldn't deal with %s: %v", address, err)
	}
	defer udpConn.Close()
	localIP := udpConn.LocalAddr().(*net.UDPAddr).IP.String()

	return &GRPCSender{
		client:  c,
		IP:      localIP,
		storage: storage,
		conn:    conn,
	}, nil
}

func(s *GRPCSender) Close() {
	s.conn.Close()
}

// SendMetricsBatch - method for sending metrics batch to the server
// send the metrics batch to the server
// if error, return error
// if success, return nil
func (s *GRPCSender) SendMetricsBatch(batch []models.Metrics) error {
	metrics := s.storage.GetAll()
	batchCopy := batch

	for _, m := range metrics {
		batchCopy = append(batchCopy, m)
		if len(batchCopy) == 100 {
			if err := s.sendBatch(batchCopy); err != nil {
				return err
			}
			batchCopy = batchCopy[:0]
		}
	}

	if len(batchCopy) > 0 {
		if err := s.sendBatch(batchCopy); err != nil {
			return err
		}
	}

	return nil
}

// sendBatch - method for sending a batch of metrics to the server
// send the batch of metrics to the server
// if error, return error
// if success, return nil
func (s *GRPCSender) sendBatch(batch []models.Metrics) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	ctx = metadata.AppendToOutgoingContext(ctx, "x-real-ip", s.IP)
	pbMetrics := make([]*pb.Metric, 0, len(batch))

	for _, m := range batch {
		pbM, err := toPBMetric(m)
		if err != nil {
			return err
		}

		pbMetrics = append(pbMetrics, pbM)
	}

	req := &pb.UpdateMetricsRequest{
		Metrics: pbMetrics,
	}

	resp, err := s.client.UpdateMetrics(ctx, req)
	if err != nil {
		return err
	}

	if resp == nil {
		return fmt.Errorf("nil response")
	}

	log.Printf("grpc update ok, batch=%d", len(batch))
	return nil
}

func toPBMetric(m models.Metrics) (*pb.Metric, error) {
	switch m.MType {
	case models.Gauge:
		if m.Value == nil {
			return nil, fmt.Errorf("gauge %q: value is nil", m.ID)
		}

		return &pb.Metric{
			Id:    m.ID,
			Type:  pb.Metric_GAUGE,
			Value: *m.Value,
		}, nil
	case models.Counter:
		if m.Delta == nil {
			return nil, fmt.Errorf("counter %q: delta is nil", m.ID)
		}

		return &pb.Metric{
			Id:    m.ID,
			Type:  pb.Metric_COUNTER,
			Delta: *m.Delta,
		}, nil
	default:
		return nil, fmt.Errorf("unknown metric type %q for %q", m.MType, m.ID)
	}
}
