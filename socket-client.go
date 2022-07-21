package main

import (
	"sync"

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
			if !sc.isConnected {
				log.Warn("socket is not connecting. try to reconnect")
				if err := sc.connect(); err != nil {
					log.Warnf("socket reconnecting failed: %s", err.Error())
					errCh <- err
					return
				}
			}

			_, message, err := sc.conn.ReadMessage()
			if err == nil {
				msgCh <- message
			}

			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				sc.Close()
				continue
			}

			if err != nil {
				errCh <- err
				return
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
