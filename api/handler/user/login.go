package user

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"log"
	"net/http"
	"star-dim/api/public"
	"star-dim/internal/models"
	"star-dim/internal/service"
)

type UserHandler struct {
	Server         *public.Server
	ClusterService *service.ClusterService
	UserService    *service.UserService
}

func NewUserHandler(server *public.Server) *UserHandler {
	return &UserHandler{
		Server: server,
	}
}

// Login creates a new SSH session
// @Summary 用户登录认证
// @Description 创建SSH连接会话，验证用户凭据并返回会话密钥用于后续API调用
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param request body models.LoginInfo true "登录请求参数"
// @Success 200 {object} object{home_path=string,session_key=string} "登录成功，返回用户主目录路径和会话密钥"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 401 {object} object{error=string} "认证失败，用户名或密码错误"
// @Failure 500 {object} object{error=string} "服务器内部错误或SSH连接失败"
// @Router /api/v1/user/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var loginInfo models.LoginInfo
	if err := c.ShouldBindJSON(&loginInfo); err != nil {
		fmt.Println("error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	config := &ssh.ClientConfig{
		User: loginInfo.User.Name,
		Auth: []ssh.AuthMethod{
			ssh.Password(loginInfo.User.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 仅用于测试
	}
	// choose login node
	var loginNode *models.LoginNode
	clusters := h.ClusterService.GetClusters()
	if loginInfo.LoginNode == nil {
		for _, cluster := range clusters {
			if cluster.Name == loginInfo.User.Cluster.Name {
				loginNode = cluster.LoginNodes[0]
			}
		}
	} else {
		for _, cluster := range clusters {
			if cluster.Name == loginInfo.User.Cluster.Name {
				for _, clusterNode := range cluster.LoginNodes {
					if loginInfo.LoginNode.Name == clusterNode.Name {
						loginNode = clusterNode
					}
				}
			}
		}
	}
	if loginNode == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "login node not found"})
		return
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", loginNode.Host, loginNode.Port), config)
	log.Println("Connecting to", loginNode.Host, "on port", loginNode.Port, "config:", config)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return

	}
	sftpClient, err := sftp.NewClient(conn)
	if err != nil {
		log.Fatal(err)
	}
	// create session key and save ssh session
	u, _ := uuid.NewUUID()
	sessionStr := fmt.Sprintf("%s_%s_%s", loginInfo.User.Cluster.Name, loginInfo.LoginNode.Name, u.String())
	// use md5 to hash sessionStr
	hash := md5.Sum([]byte(sessionStr))
	sessionKey := fmt.Sprintf("tsh_%s", hex.EncodeToString(hash[:]))
	client := &public.UserClient{
		SftpClient: sftpClient,
		SSHClient:  conn,
		UserInfo: &models.User{
			Cluster:    loginInfo.User.Cluster,
			Name:       loginInfo.User.Name,
			Password:   loginInfo.User.Password,
			PrivateKey: loginInfo.User.PrivateKey,
			HomePath:   "",
		},
	}
	h.Server.Clients[sessionKey] = client
	// get home path use sftp
	homePath, err := sftpClient.Getwd()
	if err == nil {
		h.Server.Clients[sessionKey].UserInfo.HomePath = homePath
		h.Server.Clients[sessionKey].KeepAlive()
	}
	c.JSON(201, map[string]string{"session_key": sessionKey, "home_path": homePath})
	// curl test :curl -X POST -H "Content-Type: application/json" -d "{\"name\":\"root\",\"password\":\"Ty83Hujy88\",\"host\":\"129.204.183.32\"}" http://localhost:8080/api/v2/document/login/
}

// Logout deletes the session
// @Summary 用户登出
// @Description 删除用户会话
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param session_key header string true "会话��钥"
// @Success 200 {object} object{message=string} "登出成功"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 500 {object} object{error=string} "服务器内部错误或用户未登录"
// @Router /api/v1/user/logout/ [get]
func (h *UserHandler) Logout(c *gin.Context) {
	// delete session
	sessionKey := c.GetHeader("session_key")
	if sessionKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_key is required"})
		return
	}
	if _, ok := h.Server.Clients[sessionKey]; !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not login"})
		return
	}
	delete(h.Server.Clients, sessionKey)
	c.JSON(http.StatusOK, gin.H{"message": "logout success"})
}
