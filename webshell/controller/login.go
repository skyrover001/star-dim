package controller

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"log"
	"net/http"
)

// Login creates a new SSH session
// @Summary 用户登录认证
// @Description 创建SSH连接会话，验证用户凭据并返回会话密钥用于后续API调用
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param request body object{cluster=string,username=string,password=string,host=string} true "登录请求参数" Example({"cluster":"hpc1","username":"root","password":"password","host":"192.168.1.2"})
// @Success 200 {object} object{home_path=string,session_key=string} "登录成功，返回用户主目录路径和会话密钥"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 401 {object} object{error=string} "认证失败，用户名或密码错误"
// @Failure 500 {object} object{error=string} "服务器内部错误或SSH连接失败"
// @Router /api/v2/login/ [post]
func (s *JumpService) Login(c *gin.Context) {
	// get ssh auth info
	var sshInfo SSHInfo
	// 尝试将请求中的 JSON 数据绑定到 user 结构体
	if err := c.ShouldBindJSON(&sshInfo); err != nil {
		fmt.Println("error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if sshInfo.Cluster == "" {
		sshInfo.Cluster = "default"
	}

	config := &ssh.ClientConfig{
		User: sshInfo.UserName,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshInfo.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 仅用于测试
	}
	if sshInfo.Port == "" {
		sshInfo.Port = "22"
	}
	log.Println(sshInfo)
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", sshInfo.Host, sshInfo.Port), config)
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
	//u, _ := uuid2.NewUUID()
	//sessionKey := u.String()
	sessionStr := fmt.Sprintf("%s_%s", sshInfo.Cluster, sshInfo.UserName)
	// use md5 to hash sessionStr
	hash := md5.Sum([]byte(sessionStr))
	sessionKey := fmt.Sprintf("tsh_%s", hex.EncodeToString(hash[:]))
	s.Clients[sessionKey] = &JumpClient{SSHClient: conn, SftpClient: sftpClient, SSHInfo: &sshInfo}
	// get home path use sftp
	homePath, err := sftpClient.Getwd()
	if err == nil {
		s.Clients[sessionKey].SSHInfo.HomePath = homePath
		s.Clients[sessionKey].KeepAlive()
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
// @Param session_key header string true "会话密钥"
// @Success 200 {object} object{message=string} "登出成功"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 500 {object} object{error=string} "服务器内部错误或用户未登录"
// @Router /api/v2/logout/ [get]
func (s *JumpService) Logout(c *gin.Context) {
	// delete session
	sessionKey := c.GetHeader("session_key")
	if sessionKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_key is required"})
		return
	}
	if _, ok := s.Clients[sessionKey]; !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not login"})
		return
	}
	delete(s.Clients, sessionKey)
	c.JSON(http.StatusOK, gin.H{"message": "logout success"})
}
