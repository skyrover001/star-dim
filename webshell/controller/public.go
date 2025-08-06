package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"net/http"
	"path"
	"time"
	"webshell/utils"
)

type SSHInfo struct {
	Cluster    string `json:"cluster"`
	UserName   string `json:"username"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
	Host       string `json:"host"`
	Port       string `json:"port"`
	HomePath   string `json:"home_path"`
}

type JumpClient struct {
	SftpClient *sftp.Client
	SSHClient  *ssh.Client
	SSHInfo    *SSHInfo
}

type RequestInfo struct {
	Path           string `json:"path"`
	SrcPath        string `json:"src_path"`
	DstPath        string `json:"dst_path"`
	Type           string `json:"type"`
	Cluster        string `json:"cluster"`
	SystemUsername string `json:"system_username"`
	OldPath        string `json:"old_path"`
	NewPath        string `json:"new_path"`
	Content        string `json:"content"`
	Mode           string `json:"mode"`
	Owner          string `json:"owner"`
	Group          string `json:"group"`
}

type LogInfo struct {
	Log string
	Err error
}

type JumpController struct {
	Clients     map[string]*JumpClient
	Record      bool
	RecordPath  string
	Log         bool
	LogFilePath string
	Cache       *utils.Cache
}

func (jc *JumpClient) KeepAlive() {
	go func() {
		ticker := time.NewTicker(30 * time.Second) // 每 30 秒发送一次心跳
		defer ticker.Stop()

		for range ticker.C {
			session, err := jc.SSHClient.NewSession()
			if err != nil {
				fmt.Printf("心跳会话创建失败: %v（可能连接已断开）\n", err)
				return
			}
			_, err = session.CombinedOutput("echo -n")
			session.Close()
			if err != nil {
				fmt.Printf("心跳发送失败: %v（连接可能已断开）\n", err)
				return
			}
		}
	}()
}

func (jc *JumpClient) RepackPath(pathStr string) string {
	return path.Join(jc.SSHInfo.HomePath, pathStr)
}

func (jc *JumpController) GetKeyFromRequest(c *gin.Context) (string, *RequestInfo, error) {
	// ssh session key must be in header
	sessionKey := c.GetHeader("sessionKey")
	if sessionKey == "" {
		return "", nil, fmt.Errorf("sessionKey is required in header")
	}
	var requestInfo RequestInfo
	// set current cluster and username
	switch c.Request.Method {
	case http.MethodGet:
		systemUsername := c.Query("systemUsername")
		cluster := c.Query("cluster")
		if cluster != "" {
			jc.Clients[sessionKey].SSHInfo.Cluster = cluster
		}
		if systemUsername != "" {
			jc.Clients[sessionKey].SSHInfo.UserName = systemUsername
		}
		// if path is in query, set it to requestInfo
		requestInfo.Path = c.Query("path")
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		contentType := c.GetHeader("Content-Type")
		switch contentType {
		case "application/json":
			if err := c.ShouldBindJSON(&requestInfo); err != nil {
				return "", nil, fmt.Errorf("failed to bind JSON: %v", err)
			}
			systemUsername := c.Query("systemUsername")
			cluster := c.Query("cluster")
			if cluster != "" {
				jc.Clients[sessionKey].SSHInfo.Cluster = cluster
			}
			if systemUsername != "" {
				jc.Clients[sessionKey].SSHInfo.UserName = systemUsername
			}
		case "application/x-www-form-urlencoded", "multipart/form-data":
			cluster, _ := c.GetPostForm("cluster")
			systemUsername, _ := c.GetPostForm("systemUsername")
			if cluster != "" {
				jc.Clients[sessionKey].SSHInfo.Cluster = cluster
			}
			if systemUsername != "" {
				jc.Clients[sessionKey].SSHInfo.UserName = systemUsername
			}
		default:
			return "", nil, fmt.Errorf("unsupported content type: %s", contentType)
		}
	default:
		return "", nil, fmt.Errorf("unsupported method: %s", c.Request.Method)
	}
	return sessionKey, &requestInfo, nil
}
