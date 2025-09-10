package slurm

import (
	"golang.org/x/crypto/ssh"
	"star-dim/api/public"
	"star-dim/internal/utils"
)

type SlurmHandler struct {
	Server    *public.Server
	Parser    *utils.SlurmParser
	SSHClient *ssh.Client
}

func NewSlurmHandler(server *public.Server) *SlurmHandler {
	return &SlurmHandler{
		Parser: utils.NewSlurmParser(),
		Server: server,
	}
}
