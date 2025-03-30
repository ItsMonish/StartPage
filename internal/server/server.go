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
	"github.com/ItsMonish/StartPage/internal/config"
	"github.com/ItsMonish/StartPage/internal/database"
)

var jsonRssFeed string

func StartServer(logger *log.Logger, conf config.Configuration, configPath string) {

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

	mux.HandleFunc("/rss/{id}/read", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			logger.Println("Cannot retrieve item with ID:", id)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		item, err := collector.GetRSSItem(id)
		if err != nil {
			logger.Println("Cannot retrieve item with ID:", id)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		dbItem := convertToDBItem(item)

		err = database.AddToHistory(dbItem)
		if err != nil {
			logger.Println(err)
			logger.Println("Error in adding to history")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = collector.RemoveFromList(id)

		if err != nil {
			logger.Println(err)
		}

		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/rss/{category}/viewed", func(w http.ResponseWriter, r *http.Request) {
		category := r.PathValue("category")

		var returnJson string
		var err error

		if category == "all" {
			returnJson, err = database.GetReadItemsAsJson("", "")
		} else {
			returnJson, err = database.GetReadItemsAsJson(category, "")
		}

		if err != nil {
			logger.Println(err)
		}

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, returnJson)
	})

	mux.HandleFunc("/rss/{category}/{source}/viewed", func(w http.ResponseWriter, r *http.Request) {
		category := r.PathValue("category")
		source := r.PathValue("source")

		returnJson, err := database.GetReadItemsAsJson(category, source)

		if err != nil {
			logger.Println(err)
		}

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, returnJson)
	})

	mux.HandleFunc("/rss/{category}/readAll", func(w http.ResponseWriter, r *http.Request) {
		category := r.PathValue("category")

		if category == "all" {

			categoryMap := collector.GetSourcesAsObj()
			for category := range categoryMap {
				categoryListOrg, err := collector.GetCategorySlice(category)

				categoryList := make([]collector.JsonFeedItem, len(categoryListOrg))
				copy(categoryList, categoryListOrg)

				if err != nil {
					logger.Println(err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				for _, item := range categoryList {
					dbItem := convertToDBItem(item)
					database.AddToHistory(dbItem)
				}
				for _, item := range categoryList {
					collector.RemoveFromList(item.ID)
				}
			}

			w.WriteHeader(http.StatusOK)
			return
		}

		categoryListOrg, err := collector.GetCategorySlice(category)

		categoryList := make([]collector.JsonFeedItem, len(categoryListOrg))
		copy(categoryList, categoryListOrg)

		if err != nil {
			logger.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		for _, item := range categoryList {
			dbItem := convertToDBItem(item)
			database.AddToHistory(dbItem)
		}
		for _, item := range categoryList {
			collector.RemoveFromList(item.ID)
		}

		w.WriteHeader(http.StatusOK)

	})

	mux.HandleFunc("/rss/{category}/{source}/readAll", func(w http.ResponseWriter, r *http.Request) {
		source := r.PathValue("source")

		sourceListOrg, err := collector.GetSourceSlice(source)

		sourceList := make([]collector.JsonFeedItem, len(sourceListOrg))
		copy(sourceList, sourceListOrg)

		if err != nil {
			logger.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		for _, item := range sourceList {
			dbItem := convertToDBItem(item)
			database.AddToHistory(dbItem)
		}
		for _, item := range sourceList {
			collector.RemoveFromList(item.ID)
		}

		w.WriteHeader(http.StatusOK)

	})

	mux.HandleFunc("POST /rss/item/favourite", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)

		if err != nil {
			logger.Println("Error parsing POST body")
		}

		link := string(body)

		err = database.AddToFavourites(link)

		if err != nil {
			logger.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("POST /rss/item/unfavourite", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)

		if err != nil {
			logger.Println("Error parsing POST body")
		}

		link := string(body)

		err = database.RemoveFromFavourites(link)

		if err != nil {
			logger.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/rss/{category}/favourites", func(w http.ResponseWriter, r *http.Request) {
		category := r.PathValue("category")

		returnList, err := database.GetFavourties(category, "")

		if err != nil {
			logger.Println(err)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, returnList)
	})

	mux.HandleFunc("/rss/{category}/{source}/favourites", func(w http.ResponseWriter, r *http.Request) {
		category := r.PathValue("category")
		source := r.PathValue("source")

		returnList, err := database.GetFavourties(category, source)

		if err != nil {
			logger.Println(err)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, returnList)
	})

	mux.HandleFunc("/yt/srcs", func(w http.ResponseWriter, r *http.Request) {
		channelList, err := collector.GetYtChannelList()

		if err != nil {
			logger.Println(err)
		}

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, channelList)
	})

	mux.HandleFunc("/yt/{channel}", func(w http.ResponseWriter, r *http.Request) {
		channel := r.PathValue("channel")

		if channel == "all" || channel == "" {
			w.Header().Add("Content-Type", "application/json")
			io.WriteString(w, collector.GetFullYtFeed())
		} else {
			w.Header().Add("Content-Type", "application/json")
			content, err := collector.GetChannelFeed(channel)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			io.WriteString(w, content)
		}
	})

	mux.HandleFunc("/yt/{id}/seen", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(r.PathValue("id"))

		item, err := collector.GetYTItem(id)
		if err != nil {
			logger.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		dbItem := convertYTtoDBItem(item)

		err = database.AddYtItemToHistory(dbItem)
		if err != nil {
			logger.Println("Error adding item to history")
		}

		err = collector.DeleteYTItem(id)
		if err != nil {
			logger.Println("Error removing item from memory")
		}
	})

	mux.HandleFunc("/yt/{channel}/viewed", func(w http.ResponseWriter, r *http.Request) {
		channel := r.PathValue("channel")

		content, err := database.GetYTReadItemsAsJson(channel)
		if err != nil {
			logger.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		io.WriteString(w, content)
	})

	mux.HandleFunc("POST /yt/item/favourite", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)

		if err != nil {
			logger.Println("Error parsing POST body")
		}

		link := string(body)

		err = database.AddToYtFavourites(link)

		if err != nil {
			logger.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("POST /yt/item/unfavourite", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)

		if err != nil {
			logger.Println("Error parsing POST body")
		}

		link := string(body)

		err = database.RemoveFromYtFavourites(link)

		if err != nil {
			logger.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/yt/{channel}/favourites", func(w http.ResponseWriter, r *http.Request) {
		channel := r.PathValue("channel")

		returnList, err := database.GetYtFavourites(channel)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Println(err)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, returnList)
	})

	mux.HandleFunc("/yt/{channel}/markAll", func(w http.ResponseWriter, r *http.Request) {
		channel := r.PathValue("channel")

		sourceListOrg, err := collector.GetYTFilterSlice(channel)

		sourceList := make([]collector.JsonYtItem, len(sourceListOrg))
		copy(sourceList, sourceListOrg)

		if err != nil {
			logger.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		for _, item := range sourceList {
			dbItem := convertYTtoDBItem(item)
			database.AddYtItemToHistory(dbItem)
		}
		for _, item := range sourceList {
			collector.DeleteYTItem(item.ID)
		}

		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/refreshPage", func(w http.ResponseWriter, r *http.Request) {
		logger.Println("Request for refresh recieved. Initiating refresh")
		conf = config.GetConfig(logger, configPath)

		e1Flag := collector.RefreshRssFeed(logger, conf.Rss)
		if e1Flag {
			logger.Println("There was some error in collecting RSS feed. Retrying in", conf.Props.RetryInterval, "minutes")
		} else {
			logger.Println("Collected from RSS sources successfully")
		}
		e2Flag := collector.RefreshYtFeed(logger, conf.Yt)
		if e2Flag {
			logger.Println("There was some error in collecting YT feed. Retrying in", conf.Props.RetryInterval, "minutes")
		} else {
			logger.Println("Collected from YT sources successfully")
		}

		w.WriteHeader(http.StatusOK)
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

		database.CloseDBInstance()

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
