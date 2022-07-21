package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"time"

	gotifymodel "github.com/gotify/server/v2/model"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

var (
	cacheStorage = cache.New(15*time.Minute, 5*time.Minute)
	config       Config
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func main() {
	var err error
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	config, err = loadConfig()
	if err != nil {
		log.Fatalf("Cannot load configuration: %s", err.Error())
	}

	log.SetLevel(config.Log.get())

	streamURLStr, err := getStreamURL()
	if err != nil {
		log.Fatal(err)
	}

	sc, err := NewSocketClient(streamURLStr)
	if err != nil {
		log.Fatalf("Cannot connect to socket: %s", err.Error())
	}

	msgChan, errChan := sc.getMessages()

	for {
		select {
		case err = <-errChan:
			log.Fatalf("Get message error: %s", err.Error())
			return

		case msgBytes := <-msgChan:
			msg := &gotifymodel.MessageExternal{}
			err = json.Unmarshal(msgBytes, &msg)
			if err != nil {
				log.Errorf("Cannot marse message: %s", err.Error())
				return
			}

			if msg.Priority < config.Gotify.MinPriority {
				continue
			}

			app, err := getApplicationInfo(msg.ApplicationID)
			if err != nil {
				log.Errorf("Cannot get application #%d info: %s", msg.ApplicationID, err.Error())
				return
			}

			err = Notify(msg.Title, msg.Message, app.ImagePath, msg.Priority)
			if err != nil {
				log.Errorf("Notification failed: %s", err.Error())
				return
			}

		case <-interrupt:
			log.Println("interrupt")
			sc.Close()
			return
		}
	}
}
