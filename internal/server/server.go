package server

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ItsMonish/StartPage/internal/config"
)

func StartServer(logger *log.Logger, conf config.Configuration) {

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./web/"))

	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templateObject := template.Must(template.ParseFiles("./template/startpage.html"))
		templateObject.Execute(w, nil)
	})

	clientServer := &http.Server{
		Addr:    ":" + strconv.Itoa(conf.Props.Port),
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

		/* Include actions to be performed before server closes */
	}()

	logger.Println("Server starting at: ", conf.Props.Port)

	if err := clientServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Server failed to start on port: ", conf.Props.Port)
	}
}
