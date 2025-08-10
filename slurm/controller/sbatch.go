package controller

import (
	"fmt"
	"net/http"
	"path/filepath"
	"slurm-jobacct/models"
	"slurm-jobacct/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// SbatchService SLURM sbatch 服务
type SbatchService struct {
	parser *utils.SbatchParser
}

// NewSbatchService 创建新的 sbatch 服务
func NewSbatchService() *SbatchService {
	return &SbatchService{
		parser: utils.NewSbatchParser(),
	}
}

// SubmitJob 提交作业
func SubmitJob(c *gin.Context) {
	service := NewSbatchService()

	var req models.SbatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.SbatchResponse{
			Success: false,
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	// 验证请求参数
	if err := service.parser.ValidateRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.SbatchResponse{
			Success: false,
			Message: "Request validation failed: " + err.Error(),
		})
		return
	}

	// 执行作业提交
	response := service.ExecuteSbatch(&req)

	if response.Success {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusInternalServerError, response)
	}
}

// SubmitJobWithScript 通过上传脚本文件提交作业
func SubmitJobWithScript(c *gin.Context) {
	service := NewSbatchService()

	// 解析 multipart form
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		c.JSON(http.StatusBadRequest, models.SbatchResponse{
			Success: false,
			Message: "Failed to parse multipart form: " + err.Error(),
		})
		return
	}

	// 获取上传的脚本文件
	file, header, err := c.Request.FormFile("script")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.SbatchResponse{
			Success: false,
			Message: "Script file is required: " + err.Error(),
		})
		return
	}
	defer file.Close()

	// 读取脚本内容
	scriptContent := make([]byte, header.Size)
	_, err = file.Read(scriptContent)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.SbatchResponse{
			Success: false,
			Message: "Failed to read script file: " + err.Error(),
		})
		return
	}

	// 构建请求对象
	req := models.SbatchRequest{
		Script: string(scriptContent),
	}

	// 从表单数据解析其他参数
	service.parseFormData(c, &req)

	// 验证请求参数
	if err := service.parser.ValidateRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.SbatchResponse{
			Success: false,
			Message: "Request validation failed: " + err.Error(),
		})
		return
	}

	// 执行作业提交
	response := service.ExecuteSbatchWithUpload(&req, header.Filename)

	if response.Success {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusInternalServerError, response)
	}
}

// QuickSubmit 快速提交作业（简化接口）
func QuickSubmit(c *gin.Context) {
	service := NewSbatchService()

	// 简化的请求结构
	type QuickSubmitRequest struct {
		Host       string `json:"host" binding:"required"`
		Username   string `json:"username" binding:"required"`
		Password   string `json:"password,omitempty"`
		PrivateKey string `json:"privatekey,omitempty"`

		Script string `json:"script,omitempty"`
		Wrap   string `json:"wrap,omitempty"`

		JobName     string `json:"job_name,omitempty"`
		Partition   string `json:"partition,omitempty"`
		Time        string `json:"time,omitempty"`
		NTasks      int    `json:"ntasks,omitempty"`
		CPUsPerTask int    `json:"cpus_per_task,omitempty"`
		Memory      string `json:"memory,omitempty"`
		Output      string `json:"output,omitempty"`
		Error       string `json:"error,omitempty"`
	}

	var quickReq QuickSubmitRequest
	if err := c.ShouldBindJSON(&quickReq); err != nil {
		c.JSON(http.StatusBadRequest, models.SbatchResponse{
			Success: false,
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	// 转换为完整的 SbatchRequest
	req := models.SbatchRequest{
		Host:        quickReq.Host,
		Username:    quickReq.Username,
		Password:    quickReq.Password,
		PrivateKey:  quickReq.PrivateKey,
		Script:      quickReq.Script,
		Wrap:        quickReq.Wrap,
		JobName:     quickReq.JobName,
		Partition:   quickReq.Partition,
		Time:        quickReq.Time,
		NTasks:      quickReq.NTasks,
		CPUsPerTask: quickReq.CPUsPerTask,
		Memory:      quickReq.Memory,
		Output:      quickReq.Output,
		Error:       quickReq.Error,
	}

	// 验证请求参数
	if err := service.parser.ValidateRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.SbatchResponse{
			Success: false,
			Message: "Request validation failed: " + err.Error(),
		})
		return
	}

	// 执行作业提交
	var response *models.SbatchResponse
	if req.Script != "" {
		response = service.ExecuteSbatchWithUpload(&req, "job_script.sh")
	} else {
		response = service.ExecuteSbatch(&req)
	}

	if response.Success {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusInternalServerError, response)
	}
}

// parseFormData 从表单数据解析参数
func (s *SbatchService) parseFormData(c *gin.Context, req *models.SbatchRequest) {
	// SSH 连接信息
	req.Host = c.PostForm("host")
	req.Username = c.PostForm("username")
	req.Password = c.PostForm("password")
	req.PrivateKey = c.PostForm("privatekey")

	// 基本作业参数
	req.JobName = c.PostForm("job_name")
	req.Partition = c.PostForm("partition")
	req.Time = c.PostForm("time")
	req.Memory = c.PostForm("memory")
	req.Output = c.PostForm("output")
	req.Error = c.PostForm("error")
	req.Account = c.PostForm("account")
	req.QOS = c.PostForm("qos")

	// 布尔值参数
	req.Hold = c.PostForm("hold") == "true"
	req.Wait = c.PostForm("wait") == "true"
	req.Parsable = c.PostForm("parsable") == "true"

	// 更多参数可以根据需要添加...
}

// ExecuteSbatch 执行 sbatch 命令
func (s *SbatchService) ExecuteSbatch(req *models.SbatchRequest) *models.SbatchResponse {
	// 构建命令
	command, err := s.parser.BuildCommand(req)
	if err != nil {
		return &models.SbatchResponse{
			Success: false,
			Message: "Failed to build sbatch command: " + err.Error(),
			Command: command,
		}
	}

	// 创建 SSH 配置
	sshConfig := &models.SSHConfig{
		Host:       req.Host,
		Port:       req.Port,
		Username:   req.Username,
		Password:   req.Password,
		PrivateKey: req.PrivateKey,
		Timeout:    30, // 30 秒超时
	}

	// 创建 SSH 客户端
	sshClient := utils.NewSSHClient(sshConfig)

	// 连接
	if err := sshClient.Connect(); err != nil {
		return &models.SbatchResponse{
			Success: false,
			Message: "Failed to connect to SSH server: " + err.Error(),
			Command: command,
		}
	}
	defer sshClient.Close()

	// 执行命令
	output, err := sshClient.ExecuteCommand(command)
	if err != nil {
		return &models.SbatchResponse{
			Success:   false,
			Message:   "Failed to execute sbatch command: " + err.Error(),
			Command:   command,
			RawOutput: output,
		}
	}

	// 解析输出
	response, err := s.parser.ParseOutput(output)
	if err != nil {
		return &models.SbatchResponse{
			Success:   false,
			Message:   "Failed to parse sbatch output: " + err.Error(),
			Command:   command,
			RawOutput: output,
		}
	}

	response.Command = command
	return response
}

// ExecuteSbatchWithUpload 执行带脚本上传的 sbatch 命令
func (s *SbatchService) ExecuteSbatchWithUpload(req *models.SbatchRequest, filename string) *models.SbatchResponse {
	// 创建 SSH 配置
	sshConfig := &models.SSHConfig{
		Host:       req.Host,
		Port:       req.Port,
		Username:   req.Username,
		Password:   req.Password,
		PrivateKey: req.PrivateKey,
		Timeout:    30, // 30 秒超时
	}

	// 创建 SSH 客户端
	sshClient := utils.NewSSHClient(sshConfig)

	// 连接
	if err := sshClient.Connect(); err != nil {
		return &models.SbatchResponse{
			Success: false,
			Message: "Failed to connect to SSH server: " + err.Error(),
		}
	}
	defer sshClient.Close()

	// 生成临时脚本文件路径
	timestamp := time.Now().Format("20060102_150405")
	scriptPath := filepath.Join("/tmp", fmt.Sprintf("sbatch_script_%s_%s", timestamp, filename))

	// 上传脚本文件
	uploadCommand := fmt.Sprintf("cat > %s << 'EOF'\n%s\nEOF", scriptPath, req.Script)
	_, err := sshClient.ExecuteCommand(uploadCommand)
	if err != nil {
		return &models.SbatchResponse{
			Success: false,
			Message: "Failed to upload script file: " + err.Error(),
		}
	}

	// 设置脚本文件权限
	chmodCommand := fmt.Sprintf("chmod +x %s", scriptPath)
	_, err = sshClient.ExecuteCommand(chmodCommand)
	if err != nil {
		return &models.SbatchResponse{
			Success: false,
			Message: "Failed to set script permissions: " + err.Error(),
		}
	}

	// 更新请求，使用上传的脚本文件
	req.ScriptFile = scriptPath
	req.Script = "" // 清空脚本内容，避免冲突

	// 构建 sbatch 命令
	command, err := s.parser.BuildCommand(req)
	if err != nil {
		// 清理临时文件
		sshClient.ExecuteCommand(fmt.Sprintf("rm -f %s", scriptPath))
		return &models.SbatchResponse{
			Success: false,
			Message: "Failed to build sbatch command: " + err.Error(),
		}
	}

	// 执行 sbatch 命令
	output, err := sshClient.ExecuteCommand(command)

	// 清理临时文件
	cleanupCommand := fmt.Sprintf("rm -f %s", scriptPath)
	sshClient.ExecuteCommand(cleanupCommand)

	if err != nil {
		return &models.SbatchResponse{
			Success:   false,
			Message:   "Failed to execute sbatch command: " + err.Error(),
			Command:   command,
			RawOutput: output,
		}
	}

	// 解析输出
	response, err := s.parser.ParseOutput(output)
	if err != nil {
		return &models.SbatchResponse{
			Success:   false,
			Message:   "Failed to parse sbatch output: " + err.Error(),
			Command:   command,
			RawOutput: output,
		}
	}

	response.Command = command
	return response
}
