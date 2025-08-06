package main

import (
	"golang.org/x/crypto/ssh"
	"log"
	"os"
)

func main() {
	config := &ssh.ClientConfig{
		User: "your-username",
		Auth: []ssh.AuthMethod{
			ssh.Password("your-password"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 仅用于测试
	}

	client, err := ssh.Dial("tcp", "your-ssh-server:22", config)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}

	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if err := session.Run("ls -l"); err != nil {
		log.Fatalf("Failed to run: %v", err)
	}
}
