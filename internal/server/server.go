package server

import (
	"html/template"
	"io"
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

	mux.HandleFunc("/rss", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, collector.GetRssFullFeed())
	})

	mux.HandleFunc("/rss/all", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, collector.GetRssFullFeed())
	})

	mux.HandleFunc("/rss/srcs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, collector.GetRssSources())
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

	mux.HandleFunc("/rss/{id}/read", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			logger.Println("Invalid ID received for read: ", id)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = markRssIdAsRead(id)
		if err != nil {
			logger.Println(err)
		}

		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/rss/{category}/viewed", func(w http.ResponseWriter, r *http.Request) {
		category := r.PathValue("category")
		var returnJson string
		if category == "all" {
			returnJson = database.GetRssViewed("", "")
		} else {
			returnJson = database.GetRssViewed(category, "")
		}

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, returnJson)
	})

	mux.HandleFunc("/rss/{category}/{source}/viewed", func(w http.ResponseWriter, r *http.Request) {
		category := r.PathValue("category")
		source := r.PathValue("source")
		returnJson := database.GetRssViewed(category, source)

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, returnJson)
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

func markRssIdAsRead(id int) error {
	rssItem, err := collector.GetFeedItemWithId(id)
	if err != nil {
		return err
	}

	err = database.AddToRssHistory(rssItem)
	if err != nil {
		return err
	}

	err = collector.RemoveFeedItemWithId(id)
	if err != nil {
		return err
	}

	return nil
}
