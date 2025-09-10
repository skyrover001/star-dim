package public

import (
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"path"
	models2 "star-dim/internal/models"
	"time"
)

type Config struct {
	Clusters []*models2.Cluster `yaml:"clusters"`
}

type LogInfo struct {
	Log string
	Err error
}

type UserClient struct {
	SftpClient *sftp.Client
	SSHClient  *ssh.Client
	UserInfo   *models2.User
}

type Server struct {
	Clients     map[string]*UserClient
	Record      bool
	RecordPath  string
	Log         bool
	LogFilePath string
}

func (uc *UserClient) RepackPath(pathStr string) string {
	return path.Join(uc.UserInfo.HomePath, pathStr)
}

func (uc *UserClient) KeepAlive() {
	go func() {
		ticker := time.NewTicker(30 * time.Second) // 每 30 秒发送一次心跳
		defer ticker.Stop()

		for range ticker.C {
			session, err := uc.SSHClient.NewSession()
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
