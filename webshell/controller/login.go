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

// Login and save session
func (s *JumpController) Login(c *gin.Context) {
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

func (s *JumpController) Logout(c *gin.Context) {
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
