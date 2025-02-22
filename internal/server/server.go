package server

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func StartServer(logger *log.Logger, port int) {

	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("./web")))

	clientServer := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: mux,
	}

	quitServer := make(chan os.Signal, 1)
	signal.Notify(quitServer, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGINT)

	go func() {
		<-quitServer

		logger.Println("Server closing due to signal interrupt")

		if err := clientServer.Close(); err != nil {
			logger.Println("Error closing server gracefully")
		}

		/* Include actions to be performed befreo server closes */
	}()

	logger.Println("Server starting")

	if err := clientServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Server failed to start on port " + strconv.Itoa(port))
	}
}
