package server

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	embeds "github.com/ItsMonish/StartPage"
	"github.com/ItsMonish/StartPage/internal/collector"
	"github.com/ItsMonish/StartPage/internal/database"
	"github.com/ItsMonish/StartPage/internal/types"
)

func StartServer(logger *log.Logger, conf types.RootConfiguration) {
	quitServerChan := make(chan os.Signal, 1)
	signal.Notify(quitServerChan, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGINT)

	mux := http.NewServeMux()

	content, _ := fs.Sub(embeds.StaticAssets, "web")
	fileSystem := http.FileServer(http.FS(content))

	mux.Handle("/assets/", http.StripPrefix("/assets/", fileSystem))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templateObject := template.Must(template.New("root").Parse(embeds.TemplateHTML))
		templateObject.Execute(w, conf.Links)
	})

	clientServer := &http.Server{
		Addr:    ":" + strconv.Itoa(conf.Props.Port),
		Handler: mux,
	}

	go func() {
		<-quitServerChan

		database.CloseDbInstance()

		logger.Println("Closing server...")
		if err := clientServer.Close(); err != nil {
			logger.Println("Error closing server gracefully")
		}
	}()

	logger.Println("Starting server at port: ", conf.Props.Port)

	go collector.InitRssCollector(logger, conf.Rss)

	if err := clientServer.ListenAndServe(); err != nil {
		logger.Println("Error starting server at port", conf.Props.Port)
		os.Exit(1)
	}
}
