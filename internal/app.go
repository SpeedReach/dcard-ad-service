package internal

import (
	"advertise_service/internal/handlers"
	"advertise_service/internal/infra"
	"advertise_service/internal/infra/logging"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func NewServer(config infra.Config) *http.ServeMux {
	//initializing resources
	resourceMiddleware := infra.NewResourceMiddleware(config)
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	loggerMiddleware := logging.LoggerMiddleware{Logger: logger}

	//inject resources into handlers
	var middlewares = func(next http.Handler) http.Handler {
		return resourceMiddleware.Middleware(loggerMiddleware.Middleware(next))
	}

	//initializing server
	mux := http.NewServeMux()
	//registering handlers
	mux.Handle("/api/v1/ad", middlewares(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodPost:
			handlers.PostAdHandler(writer, request)
			break
		case http.MethodGet:
			handlers.GetAdsHandler(writer, request)
		default:
			http.NotFound(writer, request)
		}
	})))
	return mux
}

func ServerUp() {
	log.Print("Starting advertise service")

	mux := NewServer(infra.LoadConfig())

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
