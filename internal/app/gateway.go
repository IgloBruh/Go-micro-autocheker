package app

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"time"

	"autocheck-microservices/internal/contracts"
	"autocheck-microservices/internal/grpcjson"
	"autocheck-microservices/internal/httpapi"
	"autocheck-microservices/internal/minioinit"
	"autocheck-microservices/internal/queue"
	"autocheck-microservices/internal/redisstate"
	"autocheck-microservices/internal/service"
	"autocheck-microservices/internal/storage"
	"autocheck-microservices/pkg/config"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"
)

func RunGateway(ctx context.Context, cfg config.Config) error {
	encoding.RegisterCodec(grpcjson.Codec{})

	store, err := storage.New(cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey, cfg.MinIOUseSSL)
	if err != nil {
		return err
	}
	for _, bucket := range []string{cfg.SubmissionsBucket, cfg.ResultsBucket, cfg.TestsBucket} {
		if err = store.EnsureBucket(ctx, bucket); err != nil {
			return err
		}
	}
	if err = minioinit.BootstrapTests(ctx, store, cfg.TestsBucket, cfg.BootstrapTestsDir); err != nil {
		return err
	}

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer redisClient.Close()
	q := queue.New(redisClient, cfg.QueueName, cfg.EventsChannel)
	stateRepo := redisstate.New(redisClient, cfg.StatePrefix)
	gatewayService := &service.GatewayService{Store: store, Queue: q, State: stateRepo, SubBucket: cfg.SubmissionsBucket}

	grpcServer := grpc.NewServer(grpc.ForceServerCodec(grpcjson.Codec{}))
	service.RegisterGatewayServer(grpcServer, gatewayService)

	grpcLis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()
	go func() {
		if serveErr := grpcServer.Serve(grpcLis); serveErr != nil {
			log.Printf("grpc server stopped: %v", serveErr)
		}
	}()

	conn, err := grpc.NewClient("127.0.0.1"+cfg.GRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(grpcjson.Codec{})),
	)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := service.NewGatewayClient(conn)
	httpServer := &http.Server{Addr: cfg.HTTPAddr, Handler: (&httpapi.Handler{Client: client}).Routes()}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	go subscribeResults(ctx, q, stateRepo)

	log.Printf("gateway http=%s grpc=%s", cfg.HTTPAddr, cfg.GRPCAddr)
	if err = httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func subscribeResults(ctx context.Context, q *queue.RedisQueue, stateRepo *redisstate.Repository) {
	pubsub := q.Subscribe(ctx)
	defer pubsub.Close()
	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			var event contracts.RunCompletedEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				log.Printf("decode event: %v", err)
				continue
			}
			state := contracts.PersistedRunState{
				Id:         event.RunID,
				Status:     event.Status,
				Task:       event.Task,
				FinishedAt: event.FinishedAt,
				Summary:    &event.Summary,
				Results:    event.Results,
				ResultsKey: event.ResultsKey,
				Error:      event.Error,
			}
			old, err := stateRepo.Get(ctx, event.RunID)
			if err == nil {
				state.CreatedAt = old.CreatedAt
			}
			if err = stateRepo.Save(ctx, state); err != nil {
				log.Printf("save state: %v", err)
			}
		}
	}
}
