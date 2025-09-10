package shell

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"star-dim/api/public"
	"star-dim/internal/models"
	"strings"
	"sync"
	"time"
)

const (
	MsgData   = '1'
	MsgResize = '2'
)

type RecType string

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// 允许所有来源的连接请求，在生产环境中应该设置更严格的策略
		return true
	},
}

type Transfer struct {
	SSHRemoteStdinPipe  io.WriteCloser
	SSHSession          *ssh.Session
	WebsocketConnection *websocket.Conn
	CastRecorder        *models.Recorder
	Cluster             string
	User                string
}

func (t *Transfer) Write(p []byte) (n int, err error) {
	writer, err := t.WebsocketConnection.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return 0, err
	}
	defer writer.Close()
	if t.CastRecorder != nil {
		t.CastRecorder.Lock()
		t.CastRecorder.WriteData(models.OutPutType, string(p))
		t.CastRecorder.Unlock()
	}
	return writer.Write(p)
}
func (t *Transfer) Close() error {
	if t.SSHSession != nil {
		t.SSHSession.Close()
	}
	return t.WebsocketConnection.Close()
}

func (t *Transfer) Read(p []byte) (n int, err error) {
	for {
		msgType, reader, err := t.WebsocketConnection.NextReader()
		if err != nil {
			return 0, err
		}
		if msgType != websocket.BinaryMessage {
			continue
		}
		return reader.Read(p)
	}
}

func (t *Transfer) SessionWait() error {
	if err := t.SSHSession.Wait(); err != nil {
		return err
	}
	return nil
}

type Resize struct {
	Columns int
	Rows    int
}

type LogInfo struct {
	Log string
	Err error
}

func (t *Transfer) ListenWebsocket(logBuff *bytes.Buffer, context context.Context, cmdLog chan LogInfo) error {
	cmdByes := make([]byte, 0)
	for {
		select {
		case <-context.Done():
			err := errors.New("LoopRead exit")
			cmdLog <- LogInfo{
				Log: "",
				Err: err,
			}
			return err
		default:
			_, wsData, err := t.WebsocketConnection.ReadMessage()
			if err != nil {
				cmdLog <- LogInfo{
					Log: "",
					Err: err,
				}
				return fmt.Errorf("reading webSocket message err:%s", err)
			}
			body := wsData[1:]
			switch wsData[0] {
			case MsgResize:
				var args Resize
				err := json.Unmarshal(body, &args)
				if err != nil {
					return fmt.Errorf("ssh pty resize windows err:%s", err)
				}
				if args.Columns > 0 && args.Rows > 0 {
					if err := t.SSHSession.WindowChange(args.Rows, args.Columns); err != nil {
						return fmt.Errorf("ssh pty resize windows err:%s", err)
					}
				}
			case MsgData:
				if _, err := t.SSHRemoteStdinPipe.Write(body); err != nil {
					return fmt.Errorf("StdinPipe write err:%s", err)
				}
				if body[0] == '\n' || body[0] == '\r' {
					cmdByes = append(cmdByes, body[0])
					now := time.Now()
					layout := "Jan 02 15:04:05"
					formattedTime := now.Format(layout)
					cmdStr := fmt.Sprintf("%s th_webshell %s %s input: %s", formattedTime, t.Cluster, t.User, string(cmdByes))
					cmdLog <- LogInfo{
						Log: cmdStr,
						Err: nil,
					}
					cmdByes = make([]byte, 0)
				} else {
					cmdByes = append(cmdByes, body...)
				}
				//fmt.Println("logBuff write body:", body)
				if _, err := logBuff.Write(body); err != nil {
					return fmt.Errorf("logBuff write err:%s", err)
				}
			}
		}
	}
}

func NewTransfer(wsConn *websocket.Conn, sshClient *ssh.Client, rec *models.Recorder, cluster string, user string) (*Transfer, error) {
	sess, err := sshClient.NewSession()
	if err != nil {
		return nil, err
	}

	stdinPipe, err := sess.StdinPipe()
	if err != nil {
		return nil, err
	}

	transfer := &Transfer{SSHRemoteStdinPipe: stdinPipe, SSHSession: sess, WebsocketConnection: wsConn, Cluster: cluster, User: user}
	sess.Stdout = transfer
	sess.Stderr = transfer

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // disable echo
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	if err := sess.RequestPty("xterm", 150, 30, modes); err != nil {
		return nil, err
	}
	if err := sess.Shell(); err != nil {
		return nil, err
	}

	if rec != nil {
		transfer.CastRecorder = rec
		transfer.CastRecorder.Lock()
		transfer.CastRecorder.WriteHeader(30, 150)
		transfer.CastRecorder.Unlock()
	}

	return transfer, nil
}

type WebshellHandler struct {
	Server *public.Server
}

func NewWebshellHandler(server *public.Server) *WebshellHandler {
	return &WebshellHandler{
		Server: server,
	}
}

func (h *WebshellHandler) GetKeyFromRequest(c *gin.Context) (string, *models.RequestInfo, error) {
	// ssh session key must be in header
	sessionKey := c.GetHeader("sessionKey")
	if sessionKey == "" {
		return "", nil, fmt.Errorf("sessionKey is required in header")
	}
	var requestInfo models.RequestInfo
	return sessionKey, &requestInfo, nil
}

// webssh
func (h *WebshellHandler) SSH(c *gin.Context) {
	key, _, err := h.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	}
	if _, ok := h.Server.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	ip := c.ClientIP()
	ws, err := upgrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	//iows := IOWebsocket{
	//	Conn:   ws,
	//	RLock:  sync.Mutex{},
	//	WLock:  sync.Mutex{},
	//	Reader: nil,
	//}
	//defer ws.Close()
	//
	//client := s.Clients[username].SSHClient
	//session, err := client.NewSession()
	//ch := make(chan string)
	//go CopyConn(&iows, session.Stdin, ch)
	//go CopyConn(session.Stdout, &iows, ch)
	//<-ch
	//<-ch
	var recorder *models.Recorder
	if h.Server.Record {
		filename := fmt.Sprintf("%s_%s_%s_%s.cast", h.Server.Clients[key].UserInfo.Cluster.Name,
			h.Server.Clients[key].UserInfo.Name, strings.Replace(ip, ":", "", -1), time.Now().Format("20060102_150405"))
		var recordFilePath = ""
		var recordFileDir = ""
		switch runtime.GOOS {
		case "windows":
			recordFileDir = filepath.Join(h.Server.RecordPath, h.Server.Clients[key].UserInfo.Cluster.Name, h.Server.Clients[key].UserInfo.Name)
			recordFilePath = filepath.Join(recordFileDir, filename)
			break
		case "linux":
			recordFileDir = path.Join(h.Server.RecordPath, h.Server.Clients[key].UserInfo.Cluster.Name, h.Server.Clients[key].UserInfo.Name)
			recordFilePath = path.Join(recordFileDir, filename)
			break
		default:
			break
		}
		err = os.MkdirAll(recordFileDir, 0766)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
		}
		fmt.Println("recordFilePath=", recordFilePath)
		f, err := os.OpenFile(recordFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0766)
		if err != nil {
			c.AbortWithStatusJSON(200, gin.H{"ok": false, "msg": err.Error()})
		}
		defer f.Close()
		recorder = models.NewRecorder(f)
	}

	t, err := NewTransfer(ws, h.Server.Clients[key].SSHClient, recorder, h.Server.Clients[key].UserInfo.Cluster.Name, h.Server.Clients[key].UserInfo.Name)

	if err != nil {
		_ = ws.WriteControl(websocket.CloseMessage,
			[]byte(err.Error()), time.Now().Add(time.Second))
		return
	}
	defer t.Close()

	var logBuff = bufPool.Get().(*bytes.Buffer)
	logBuff.Reset()
	defer bufPool.Put(logBuff)

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(2)
	cmdLog := make(chan LogInfo, 1)
	logFileDir := "logs"
	logFilePath := fmt.Sprintf("logs/%s_%s_%s.log", h.Server.Clients[key].UserInfo.Cluster.Name, h.Server.Clients[key].UserInfo.Name, key)
	err = os.MkdirAll(logFileDir, 0766)
	l, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0766)
	defer l.Close()
	go func() {
		defer wg.Done()
		err := t.ListenWebsocket(logBuff, ctx, cmdLog)
		if err != nil {
			log.Printf("%#v", err)
		}
	}()
	go func() {
		defer wg.Done()
		err := t.SessionWait()
		if err != nil {
			log.Printf("%#v", err)
		}
		cancel()
	}()
	go func(f *os.File, cmdLog chan LogInfo) {
		for {
			cmd := <-cmdLog
			if cmd.Err != nil {
				break
			}
			_, err := f.Write([]byte(cmd.Log))
			if err != nil {
				return
			}
		}
	}(l, cmdLog)
	wg.Wait()
}
