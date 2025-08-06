package utils

import (
	"github.com/gorilla/websocket"
	"io"
	"sync"
)

type IOWebsocket struct {
	Conn   *websocket.Conn
	RLock  sync.Mutex
	WLock  sync.Mutex
	Reader io.Reader
}

type IOWebsocketMethods interface {
	Read([]byte) (n int, err error)
	Write([]byte) (n int, err error)
	GetReader() (io.Reader, error)
}

func (ws *IOWebsocket) Read(b []byte) (n int, err error) {
	ws.RLock.Lock()
	defer ws.RLock.Unlock()
back:
	reader, err := ws.GetReader()
	if err != nil {
		return 0, err
	}
	len, err := reader.Read(b)
	if err == io.EOF {
		ws.Reader = nil
		goto back
	}
	return len, err
}

func (ws *IOWebsocket) Write(b []byte) (n int, err error) {
	ws.WLock.Lock()
	defer ws.WLock.Unlock()
	if err := ws.Conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (ws *IOWebsocket) GetReader() (io.Reader, error) {
	if ws.Reader != nil {
		return ws.Reader, nil
	}
	_, reader, err := ws.Conn.NextReader()
	if err != nil {
		return nil, err
	}
	ws.Reader = reader
	return reader, nil
}

func CopyConn(w io.Writer, r io.Reader, doneCh chan<- string) {
	_, err := io.Copy(w, r)
	if err != nil {
		doneCh <- err.Error()
	}
	doneCh <- ""
}
