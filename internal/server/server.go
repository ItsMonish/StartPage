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
	"time"

	embeds "github.com/ItsMonish/StartPage"
	"github.com/ItsMonish/StartPage/internal/collector"
	"github.com/ItsMonish/StartPage/internal/database"
	"github.com/ItsMonish/StartPage/internal/types"
)

var (
	isServerRoutineLive bool = true
)

func StartServer(logger *log.Logger, conf types.RootConfiguration) {
	quitServerChan := make(chan os.Signal, 1)
	stopRoutineChan := make(chan bool, 1)
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

	mux.HandleFunc("/rss/{category}/readAll", func(w http.ResponseWriter, r *http.Request) {
		category := r.PathValue("category")

		var err error
		if category == "all" {
			err = markRssListAsRead("", "")
		} else {
			err = markRssListAsRead(category, "")
		}
		if err != nil {
			logger.Println("Error marking category", category, "as read")
			logger.Println(err)
			w.WriteHeader(http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/rss/{category}/{source}/readAll", func(w http.ResponseWriter, r *http.Request) {
		category := r.PathValue("category")
		source := r.PathValue("source")
		err := markRssListAsRead(category, source)
		if err != nil {
			logger.Println("Error marking source", source, "as read")
			logger.Println(err)
			w.WriteHeader(http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("POST /rss/item/favourite", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Println("Error parsing POST body")
		}

		link := string(body)
		err = database.FavouriteRssItem(link)
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
		err = database.UnFavouriteRssItem(link)
		if err != nil {
			logger.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/rss/{category}/favourites", func(w http.ResponseWriter, r *http.Request) {
		category := r.PathValue("category")
		returnList, err := database.GetRssFavourites(category, "")
		if err != nil {
			logger.Println(err)
			w.WriteHeader(http.StatusBadRequest)
		}

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, returnList)
	})

	mux.HandleFunc("/rss/{category}/{source}/favourites", func(w http.ResponseWriter, r *http.Request) {
		category := r.PathValue("category")
		source := r.PathValue("source")

		returnList, err := database.GetRssFavourites(category, source)
		if err != nil {
			logger.Println(err)
			w.WriteHeader(http.StatusBadRequest)
		}

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, returnList)
	})

	mux.HandleFunc("/yt/srcs", func(w http.ResponseWriter, r *http.Request) {
		channelList := collector.GetYtSources()

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, channelList)
	})

	mux.HandleFunc("/yt/{channel}", func(w http.ResponseWriter, r *http.Request) {
		channel := r.PathValue("channel")

		if channel == "all" || channel == "" {
			io.WriteString(w, collector.GetYtFullFeed())
		} else {
			content, err := collector.GetYtChannelFeed(channel)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			io.WriteString(w, content)
		}
		w.Header().Add("Content-Type", "application/json")
	})

	mux.HandleFunc("/yt/{id}/seen", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(r.PathValue("id"))

		err := markYtIdAsRead(id)
		if err != nil {
			logger.Println("Error in marking youtube feed item as read")
			logger.Println(err)
			w.WriteHeader(http.StatusBadRequest)
		}

		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/yt/{channel}/viewed", func(w http.ResponseWriter, r *http.Request) {
		channel := r.PathValue("channel")

		returnJson := database.GetYtSeenItems(channel)

		io.WriteString(w, returnJson)
	})

	mux.HandleFunc("/yt/item/favourite", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Println("Error parsing POST body")
		}
		url := string(body)
		err = database.FavouriteYtItem(url)
		if err != nil {
			logger.Println("Error in marking", url, "as favourite")
			w.WriteHeader(http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/yt/item/unfavourite", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Println("Error parsing POST body")
		}
		url := string(body)
		err = database.UnFavouriteYtItem(url)
		if err != nil {
			logger.Println("Error in marking", url, "as favourite")
			w.WriteHeader(http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/yt/{channel}/favourites", func(w http.ResponseWriter, r *http.Request) {
		channel := r.PathValue("channel")
		retList, err := database.GetFavouritedYtItems(channel)
		if err != nil {
			logger.Println("Error collecting YT favourited items")
			logger.Println(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, retList)
	})

	mux.HandleFunc("/yt/{channel}/markAll", func(w http.ResponseWriter, r *http.Request) {
		channel := r.PathValue("channel")
		err := markYtListAsRead(channel)
		if err != nil {
			logger.Println("Error marking youtube feed list as read")
			logger.Println(err.Error())
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/refreshPage", func(w http.ResponseWriter, r *http.Request) {
		logger.Println("Refreshing feeds upon manual request")

		stopRoutineChan <- true
		go func() {
			for isServerRoutineLive {
				time.Sleep(2 * time.Second)
			}
			go startAndMaintainCollectors(logger, stopRoutineChan, conf)
			isServerRoutineLive = true
		}()

		w.WriteHeader(http.StatusOK)
	})

	clientServer := &http.Server{
		Addr:    ":" + strconv.Itoa(conf.Props.Port),
		Handler: mux,
	}

	go func() {
		<-quitServerChan
		stopRoutineChan <- true

		database.CloseDbInstance()

		logger.Println("Closing server...")
		if err := clientServer.Close(); err != nil {
			logger.Println("Error closing server gracefully")
		}
	}()

	logger.Println("Starting server at port: ", conf.Props.Port)

	go startAndMaintainCollectors(logger, stopRoutineChan, conf)

	if err := clientServer.ListenAndServe(); err != nil {
		logger.Println("Error starting server at port", conf.Props.Port)
		os.Exit(1)
	}
}

func startAndMaintainCollectors(logger *log.Logger, signalChan chan bool, conf types.RootConfiguration) {
	nextRssRefresh := updateWithInterval(conf.Props.RefreshInterval)
	nextYtRefresh := updateWithInterval(conf.Props.RefreshInterval)
	collector.StartCollectors(logger, conf)

	for {
		select {
		case <-signalChan:
			isServerRoutineLive = false
			logger.Println("Stopping server routine")
			return
		default:
			if time.Now().After(nextRssRefresh) {
				nextRssRefresh = updateWithInterval(conf.Props.RefreshInterval)
				collector.RefreshRssFeed(logger, conf.Rss)
				if collector.RssErrFlag {
					logger.Println("There was some error in collecting RSS feed. Retrying in", conf.Props.RefreshInterval, "minutes")
					nextRssRefresh = updateWithInterval(conf.Props.RefreshInterval)
				}
				collector.RssErrFlag = false
			}
			if time.Now().After(nextYtRefresh) {
				nextYtRefresh = updateWithInterval(conf.Props.RefreshInterval)
				collector.RefreshYtFeed(logger, conf.Yt)
				if collector.YtErrFlag {
					logger.Println("There was some error in collecting YT feed. Retrying in", conf.Props.RefreshInterval, "minutes")
					nextYtRefresh = updateWithInterval(conf.Props.RefreshInterval)
				}
				collector.YtErrFlag = false
			}
			time.Sleep(20 * time.Second)
		}

	}
}

func updateWithInterval(interval int) time.Time {
	return time.Now().Add(time.Duration(interval) * time.Minute)
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

func markRssListAsRead(category string, source string) error {
	targetList, err := collector.GetAndRemoveRssItems(category, source)
	if err != nil {
		return err
	}

	for _, item := range targetList {
		database.AddToRssHistory(item)
	}
	return nil
}

func markYtIdAsRead(id int) error {
	ytItem, err := collector.GetYtItemWithId(id)
	if err != nil {
		return err
	}

	err = database.AddToYtHistory(ytItem)
	if err != nil {
		return err
	}

	err = collector.RemoveYtItemWithId(id)
	if err != nil {
		return err
	}
	return nil
}

func markYtListAsRead(channel string) error {
	targetList, err := collector.GetAndRemoveYtItems(channel)
	if err != nil {
		return err
	}

	for _, item := range targetList {
		database.AddToYtHistory(item)
	}
	return nil
}
