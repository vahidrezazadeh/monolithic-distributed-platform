package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
)

type Services struct {
	Redis      *redis.Client
	Kafka      *kafka.Writer
	Cluster    *Cluster
	Config     Config
	jobQueue   chan Job
	workerWg   sync.WaitGroup
	ctx        context.Context
	cancelFunc context.CancelFunc
	isLeader   atomic.Bool
}

type Job struct {
	ID     string
	Data   []byte
	Result chan<- JobResult
}

type JobResult struct {
	ID      string
	Result  string
	Worker  string
	Success bool
}

func SetupServices(ctx context.Context, cfg Config, cluster *Cluster) (*Services, error) {
	svcCtx, cancel := context.WithCancel(ctx)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddress,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if _, err := redisClient.Ping(svcCtx).Result(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	kafkaWriter := &kafka.Writer{
		Addr:     kafka.TCP(cfg.KafkaAddress),
		Balancer: &kafka.LeastBytes{},
		Async:    true,
	}

	services := &Services{
		Redis:      redisClient,
		Kafka:      kafkaWriter,
		Cluster:    cluster,
		Config:     cfg,
		jobQueue:   make(chan Job, 1000),
		ctx:        svcCtx,
		cancelFunc: cancel,
	}

	workerCount := 10
	for i := 0; i < workerCount; i++ {
		services.workerWg.Add(1)
		go services.worker(i)
	}

	go services.electLeader()

	return services, nil
}

func (s *Services) worker(id int) {
	defer s.workerWg.Done()
	workerID := fmt.Sprintf("%s-worker-%d", s.Config.InstanceID, id)

	for {
		select {
		case job := <-s.jobQueue:
			result := s.processJob(job, workerID)
			job.Result <- result
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Services) processJob(job Job, workerID string) JobResult {
	processingTime := time.Duration(rand.Intn(500)) * time.Millisecond
	time.Sleep(processingTime)

	success := rand.Float32() > 0.2 // 80% Success

	fmt.Println("Processing job:", job.ID, "by", workerID, "with success:", success)

	return JobResult{
		ID:      job.ID,
		Result:  fmt.Sprintf("Processed by %s in %v", workerID, processingTime),
		Worker:  workerID,
		Success: success,
	}
}

func (s *Services) SubmitJob(job Job) {
	s.jobQueue <- job
}

func (s *Services) electLeader() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			nodes := s.Cluster.GetNodes()
			if len(nodes) == 0 {
				continue
			}

			smallestID := nodes[0].ID
			for _, node := range nodes[1:] {
				if node.ID < smallestID {
					smallestID = node.ID
				}
			}

			isLeader := smallestID == s.Config.InstanceID
			s.isLeader.Store(isLeader)

			if isLeader {
				log.Println("This node is now the leader")
			}
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Services) IsLeader() bool {
	return s.isLeader.Load()
}

func (s *Services) Shutdown(ctx context.Context) error {
	s.cancelFunc()

	close(s.jobQueue)

	done := make(chan struct{})
	go func() {
		s.workerWg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	}

	if err := s.Redis.Close(); err != nil {
		return fmt.Errorf("failed to close Redis: %v", err)
	}

	if err := s.Kafka.Close(); err != nil {
		return fmt.Errorf("failed to close Kafka writer: %v", err)
	}

	return nil
}
