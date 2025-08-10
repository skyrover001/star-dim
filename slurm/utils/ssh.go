package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"slurm-jobacct/models"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHClient SSH 客户端
type SSHClient struct {
	client *ssh.Client
	config *models.SSHConfig
}

// NewSSHClient 创建新的 SSH 客户端
func NewSSHClient(config *models.SSHConfig) *SSHClient {
	return &SSHClient{
		config: config,
	}
}

// Connect 连接到 SSH 服务器
func (c *SSHClient) Connect() error {
	var auth []ssh.AuthMethod

	// 添加密码认证
	if c.config.Password != "" {
		auth = append(auth, ssh.Password(c.config.Password))
	}

	// 添加私钥认证
	if c.config.PrivateKey != "" {
		key, err := ioutil.ReadFile(c.config.PrivateKey)
		if err != nil {
			return fmt.Errorf("unable to read private key: %v", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return fmt.Errorf("unable to parse private key: %v", err)
		}

		auth = append(auth, ssh.PublicKeys(signer))
	}

	// SSH 客户端配置
	config := &ssh.ClientConfig{
		User:            c.config.Username,
		Auth:            auth,
		Timeout:         time.Duration(c.config.Timeout) * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 注意：生产环境应该验证主机密钥
	}

	// 连接
	port := c.config.Port
	if port == 0 {
		port = 22
	}

	addr := fmt.Sprintf("%s:%d", c.config.Host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("failed to dial: %v", err)
	}

	c.client = client
	return nil
}

// ExecuteCommand 执行命令
func (c *SSHClient) ExecuteCommand(command string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("SSH client not connected")
	}

	// 创建会话
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	// 执行命令
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(command)
	if err != nil {
		return "", fmt.Errorf("command execution failed: %v, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// Close 关闭连接
func (c *SSHClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// TestConnection 测试 SSH 连接
func TestConnection(config *models.SSHConfig) error {
	client := NewSSHClient(config)

	if err := client.Connect(); err != nil {
		return err
	}
	defer client.Close()

	// 执行简单的测试命令
	_, err := client.ExecuteCommand("echo 'connection test'")
	return err
}

// IsPortOpen 检查端口是否开放
func IsPortOpen(host string, port int) bool {
	timeout := time.Second * 3
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
