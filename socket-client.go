package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// SocketClient -
type SocketClient struct {
	urlAddr     string
	conn        *websocket.Conn
	isConnected bool
	mu          sync.RWMutex
}

func (sc *SocketClient) markConnected(isConnected bool) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.isConnected = isConnected
}

func (sc *SocketClient) connect() (err error) {
	c, _, err := websocket.DefaultDialer.Dial(sc.urlAddr, nil)
	if err != nil {
		sc.markConnected(false)
		return err
	}
	sc.mu.Lock()
	sc.conn = c
	sc.mu.Unlock()
	sc.markConnected(true)
	return nil
}

func (sc *SocketClient) reconnect() (err error) {
	log.Warn("socket is not connecting. try to reconnect")
	var interval = 5 * time.Second // default interval is 5 seconds
	var startTime = time.Now()
	var maxElapsedTime = 10 * time.Minute // default max elapsed time is 10 minutes

	for {
		err = sc.connect()
		if err == nil {
			return nil // connected successfully
		}

		if time.Since(startTime) > maxElapsedTime {
			return fmt.Errorf("could not connect to gotify: %s", err.Error())
		}

		log.Warnf("unable to connect to gotify socket: %s, retrying in %s", err.Error(), interval.String())
		interval += 5 * time.Second // increase interval by 5 seconds every time
		<-time.After(interval)      // wait interval seconds before reconnect again
	}
}

// Close -
func (sc *SocketClient) Close() {
	sc.conn.Close()
	sc.markConnected(false)
}

func (sc *SocketClient) getMessages() (msgCh chan []byte, errCh chan error) {
	msgCh = make(chan []byte)
	errCh = make(chan error, 1)
	go func(msgCh chan []byte, errCh chan error) {
		for {
			_, message, err := sc.conn.ReadMessage()
			switch {
			case err != nil && (websocket.IsCloseError(err, websocket.CloseNormalClosure) ||
				websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure)):
				// try to reconnect
				err = sc.reconnect() // try to reconnect
				if err != nil {
					errCh <- err // error occurred
				}
			case err != nil:
				errCh <- err
				continue
			default:
				// write message bytes to channel
				msgCh <- message
			}
		}
	}(msgCh, errCh)
	return
}

// NewSocketClient -
func NewSocketClient(urlAddr string) (sc *SocketClient, err error) {
	sc = &SocketClient{
		urlAddr: urlAddr,
	}

	err = sc.connect()
	if err != nil {
		return
	}

	return
}
