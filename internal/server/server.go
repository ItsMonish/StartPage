package server

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ItsMonish/StartPage/internal/collector"
	"github.com/ItsMonish/StartPage/internal/config"
)

var jsonRssFeed string

func StartServer(logger *log.Logger, conf config.Configuration) {

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./web/"))

	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templateObject := template.Must(template.ParseFiles("./template/startpage.html"))
		templateObject.Execute(w, conf.Links)
	})

	mux.HandleFunc("/rss", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, collector.CollectRssAsJson())
	})

	mux.HandleFunc("/rss/all", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, collector.CollectRssAsJson())
	})

	mux.HandleFunc("/rss/srcs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, collector.GetSourcesAsStr())
	})

	mux.HandleFunc("/rss/{category}", func(w http.ResponseWriter, r *http.Request) {
		category := r.PathValue("category")
		content, err := collector.GetCategoryFeed(category)
		if err != nil {
			logger.Println("Error in getting feed of category", category)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, content)
	})

	mux.HandleFunc("/rss/{category}/{source}", func(w http.ResponseWriter, r *http.Request) {
		source := r.PathValue("source")
		content, err := collector.GetSourceFeed(source)
		if err != nil {
			logger.Println("Error in getting feed of source", source)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, content)
	})

	clientServer := &http.Server{
		Addr:    ":" + strconv.Itoa(conf.Props.Port),
		Handler: mux,
	}

	quitServer := make(chan os.Signal, 1)
	stopRoutine := make(chan bool, 1)
	signal.Notify(quitServer, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGINT)

	go func() {
		<-quitServer

		stopRoutine <- true

		logger.Println("Server closing due to signal interrupt")

		if err := clientServer.Close(); err != nil {
			logger.Println("Error closing server gracefully")
		}

		/* Include actions to be performed before server closes */
	}()

	go startServerRoutine(logger, stopRoutine, conf)

	logger.Println("Server starting at: ", conf.Props.Port)

	if err := clientServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Server failed to start on port: ", conf.Props.Port)
	}
}

func updateWithInterval(interval int) time.Time {
	return time.Now().Add(time.Duration(interval) * time.Minute)
}

func startServerRoutine(logger *log.Logger, stopRoutine chan bool, conf config.Configuration) {
	nextRefresh := updateWithInterval(conf.Props.RefreshInterval)

	collector.RefreshRssFeed(logger, conf.Rss)
	logger.Println("Collecting from RSS sources")

	for {
		select {
		case <-stopRoutine:
			logger.Println("Stopping server routine")
			return
		default:
			if time.Now().After(nextRefresh) {
				logger.Println("Collecting from RSS sources")
				collector.RefreshRssFeed(logger, conf.Rss)

				nextRefresh = updateWithInterval(conf.Props.RefreshInterval)
			} else {
				time.Sleep(time.Minute)
			}
		}
	}
}
