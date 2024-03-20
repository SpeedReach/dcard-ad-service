package internal

import (
	"advertise_service/internal/handlers"
	"advertise_service/internal/infra"
	"advertise_service/internal/infra/cache"
	"advertise_service/internal/infra/logging"
	"advertise_service/internal/infra/persistent"
	"context"
	"go.uber.org/zap"
	"log"
	"net/http"
)

type Server struct {
	mux *http.ServeMux
}

func NewServer(storage persistent.Storage, cache cache.Service, logger *zap.Logger) Server {
	mux := http.NewServeMux()

	loggerMiddleware := logging.LoggerMiddleware{Logger: logger}
	resourceMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, handlers.StorageContextKey{}, storage)
			ctx = context.WithValue(ctx, handlers.CacheContextKey{}, cache)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	//registering handlers
	mux.Handle("/api/v1/ad", loggerMiddleware.Middleware(resourceMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodPost:
			handlers.PostAdHandler(writer, request)
			break
		case http.MethodGet:
			handlers.GetAdsHandler(writer, request)
		default:
			http.NotFound(writer, request)
		}
	}))))

	return Server{mux: mux}
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func newProductionServer(config infra.Config) Server {
	//initializing resources
	storage, cache := infra.ProductionSetup(config)
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	//initializing server
	return NewServer(storage, cache, logger)
}

func ProductionServerUp() {
	log.Print("Starting advertise service")

	mux := newProductionServer(infra.LoadConfig())

	var port = "8080"

	done := make(chan bool)
	go func() {
		err := http.ListenAndServe(":"+port, mux)
		if err != nil {
			log.Fatalf("Server failed to start: %v", err)
		} else {
			done <- true
		}
	}()
	log.Printf("Server started at port %v", port)
	<-done
}
