package main

import (
	"encoding/json"
	"fmt"
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

	log.Info("Connecting to socket")
	sc, err := NewSocketClient(streamURLStr)
	if err != nil {
		log.Fatalf("Cannot connect to socket: %s", err.Error())
	}

	log.Info("Connected to socket. Start to read messages")
	msgChan, errChan := sc.getMessages()

	for {
		select {
		case err = <-errChan:
			log.Fatalf("Get message error: %s", err.Error())
			return // error occurred, exiting application

		case msgBytes := <-msgChan:
			err = processMessage(msgBytes)
			if err != nil {
				log.Errorf("Cannot process message: %s", err.Error())
			}

		case <-interrupt:
			log.Println("interrupt")
			sc.Close()
			return
		}
	}
}

func processMessage(msgBytes []byte) (err error) {
	var msg = &gotifymodel.MessageExternal{}
	err = json.Unmarshal(msgBytes, &msg)
	if err != nil {
		return fmt.Errorf("cannot parse message: %s", err.Error())
	}

	if msg.Priority < config.Gotify.MinPriority {
		return nil // skip message so message priority is too low
	}

	app, err := getApplicationInfo(msg.ApplicationID)
	if err != nil {
		return fmt.Errorf("cannot get application #%d info: %s", msg.ApplicationID, err.Error())
	}

	err = Notify(msg.Title, msg.Message, app.ImagePath, msg.Priority)
	if err != nil {
		return fmt.Errorf("notification failed: %s", err.Error())
	}

	return nil // message is processed successfully
}
