package service

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"path/filepath"
	"star-dim/internal/models"
	"star-dim/internal/utils"
	"time"
)

type SlurmService struct {
	sshClient *ssh.Client
	parser    *utils.SlurmParser
}

func NewSlurmService(sshClient *ssh.Client, parser *utils.SlurmParser) *SlurmService {
	return &SlurmService{
		sshClient: sshClient,
		parser:    parser,
	}
}

func (s *SlurmService) SSHClient() *ssh.Client {
	return s.sshClient
}

// ExecuteSacct 执行 sacct 命令
func (s *SlurmService) ExecuteSacct(req *models.SacctRequest) models.SacctResponse {
	command := s.parser.BuildSacctCommand(req)
	session, err := s.sshClient.NewSession()
	if err != nil {
		return models.SacctResponse{
			Success:   "no",
			Message:   err.Error(),
			Data:      nil,
			Total:     0,
			Command:   "",
			RawOutput: "",
		}
	}
	defer session.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	err = session.Run(command)

	if err != nil {
		return models.SacctResponse{
			Success:   "no",
			Message:   "Failed to execute sacct command: " + err.Error(),
			Command:   command,
			RawOutput: stdout.String(),
		}
	}
	output := stdout.String()
	jobs, err := s.parser.ParseSacctOutput(output, req.Format)
	if err != nil {
		return models.SacctResponse{
			Success:   "no",
			Message:   "Failed to parse sacct output: " + err.Error(),
			Command:   command,
			RawOutput: output,
		}
	}

	return models.SacctResponse{
		Success: "yes",
		Message: "",
		Data:    jobs,
		Total:   len(jobs),
		Command: command,
	}
}

func (s *SlurmService) ExecuteSbatch(req *models.SbatchRequest) *models.SbatchResponse {
	command, err := s.parser.BuildSbatchCommand(req)
	session, err := s.sshClient.NewSession()
	if err != nil {
		return &models.SbatchResponse{
			Success: "no",
			Message: "Failed to execute sbatch command: " + err.Error(),
		}
	}
	defer session.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	err = session.Run(command)
	// 解析输出
	output := stdout.String()
	response, err := s.parser.ParseSbatchOutput(output)
	if err != nil {
		return &models.SbatchResponse{
			Success:   "no",
			Message:   "Failed to parse sbatch output: " + err.Error(),
			Command:   command,
			RawOutput: output,
		}
	}

	response.Command = command
	return response
}

// ExecuteSbatchWithUpload 执行带脚本上传的 sbatch 命令
func (s *SlurmService) ExecuteSbatchWithUpload(req *models.SbatchRequest, filename string) *models.SbatchResponse {
	session, err := s.sshClient.NewSession()
	if err != nil {
		return &models.SbatchResponse{
			Success: "no",
			Message: "Failed to execute sbatch command: " + err.Error(),
		}
	}
	// 生成临时脚本文件路径
	timestamp := time.Now().Format("20060102_150405")
	scriptPath := filepath.Join("/tmp", fmt.Sprintf("sbatch_script_%s_%s", timestamp, filename))

	// 上传脚本文件
	uploadCommand := fmt.Sprintf("cat > %s << 'EOF'\n%s\nEOF", scriptPath, req.Script)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	err = session.Run(uploadCommand)
	if err != nil {
		return &models.SbatchResponse{
			Success: "no",
			Message: "Failed to upload script file: " + err.Error(),
		}
	}

	chmodCommand := fmt.Sprintf("chmod +x %s", scriptPath)
	err = session.Run(chmodCommand)
	if err != nil {
		return &models.SbatchResponse{
			Success: "no",
			Message: "Failed to set script permissions: " + err.Error(),
		}
	}

	req.ScriptFile = scriptPath
	req.Script = ""
	command, err := s.parser.BuildSbatchCommand(req)
	if err != nil {
		// 清理临时文件
		err = session.Run(fmt.Sprintf("rm -f %s", scriptPath))
		return &models.SbatchResponse{
			Success: "no",
			Message: "Failed to build sbatch command: " + err.Error(),
		}
	}

	err = session.Run(command)
	// 清理临时文件
	cleanupCommand := fmt.Sprintf("rm -f %s", scriptPath)
	session.Run(cleanupCommand)

	var output = stdout.String()
	if err != nil {
		return &models.SbatchResponse{
			Success:   "no",
			Message:   "Failed to execute sbatch command: " + err.Error(),
			Command:   command,
			RawOutput: output,
		}
	}

	// 解析输出
	response, err := s.parser.ParseSbatchOutput(output)
	if err != nil {
		return &models.SbatchResponse{
			Success:   "no",
			Message:   "Failed to parse sbatch output: " + err.Error(),
			Command:   command,
			RawOutput: output,
		}
	}

	response.Command = command
	return response
}
