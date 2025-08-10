package controller

import (
	"net/http"
	"slurm-jobacct/models"
	"slurm-jobacct/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SacctService SLURM sacct 服务
type SacctService struct {
	parser *utils.SacctParser
}

// NewSacctService 创建新的 sacct 服务
func NewSacctService() *SacctService {
	return &SacctService{
		parser: utils.NewSacctParser(),
	}
}

// GetJobs 获取作业列表
func GetJobs(c *gin.Context) {
	service := NewSacctService()

	// 解析请求参数
	req := &models.SacctRequest{}

	// 从查询参数解析
	if jobids := c.Query("jobids"); jobids != "" {
		req.JobIDs = []string{jobids}
	}

	if users := c.Query("users"); users != "" {
		req.Users = []string{users}
	}

	if accounts := c.Query("accounts"); accounts != "" {
		req.Accounts = []string{accounts}
	}

	if partitions := c.Query("partitions"); partitions != "" {
		req.Partitions = []string{partitions}
	}

	if states := c.Query("states"); states != "" {
		req.States = []string{states}
	}

	req.StartTime = c.Query("starttime")
	req.EndTime = c.Query("endtime")
	req.Format = c.Query("format")

	if brief := c.Query("brief"); brief == "true" {
		req.Brief = true
	}

	if long := c.Query("long"); long == "true" {
		req.Long = true
	}

	if parsable := c.Query("parsable"); parsable == "true" {
		req.Parsable = true
	}

	if noheader := c.Query("noheader"); noheader == "true" {
		req.NoHeader = true
	}

	if allusers := c.Query("allusers"); allusers == "true" {
		req.AllUsers = true
	}

	// SSH 连接信息
	req.Host = c.Query("host")
	req.Username = c.Query("username")
	req.Password = c.Query("password")
	req.PrivateKey = c.Query("privatekey")

	if portStr := c.Query("port"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			req.Port = port
		}
	}

	// 验证必要参数
	if req.Host == "" || req.Username == "" {
		c.JSON(http.StatusBadRequest, models.SacctResponse{
			Success: false,
			Message: "Host and username are required",
		})
		return
	}

	// 如果没有提供密码和私钥，返回错误
	if req.Password == "" && req.PrivateKey == "" {
		c.JSON(http.StatusBadRequest, models.SacctResponse{
			Success: false,
			Message: "Password or private key is required",
		})
		return
	}

	// 执行 sacct 命令
	response := service.ExecuteSacct(req)

	if response.Success {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusInternalServerError, response)
	}
}

// GetJobDetail 获取单个作业详情
func GetJobDetail(c *gin.Context) {
	service := NewSacctService()
	jobid := c.Param("jobid")

	if jobid == "" {
		c.JSON(http.StatusBadRequest, models.SacctResponse{
			Success: false,
			Message: "Job ID is required",
		})
		return
	}

	// 构建请求
	req := &models.SacctRequest{
		JobIDs: []string{jobid},
		Long:   true, // 获取详细信息
	}

	// SSH 连接信息（从查询参数或请求体获取）
	req.Host = c.Query("host")
	req.Username = c.Query("username")
	req.Password = c.Query("password")
	req.PrivateKey = c.Query("privatekey")

	if portStr := c.Query("port"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			req.Port = port
		}
	}

	// 验证必要参数
	if req.Host == "" || req.Username == "" {
		c.JSON(http.StatusBadRequest, models.SacctResponse{
			Success: false,
			Message: "Host and username are required",
		})
		return
	}

	if req.Password == "" && req.PrivateKey == "" {
		c.JSON(http.StatusBadRequest, models.SacctResponse{
			Success: false,
			Message: "Password or private key is required",
		})
		return
	}

	// 执行命令
	response := service.ExecuteSacct(req)

	if response.Success && len(response.Data) > 0 {
		c.JSON(http.StatusOK, response)
	} else if response.Success && len(response.Data) == 0 {
		c.JSON(http.StatusNotFound, models.SacctResponse{
			Success: false,
			Message: "Job not found",
		})
	} else {
		c.JSON(http.StatusInternalServerError, response)
	}
}

// GetUserJobs 获取用户作业
func GetUserJobs(c *gin.Context) {
	service := NewSacctService()
	user := c.Param("user")

	if user == "" {
		c.JSON(http.StatusBadRequest, models.SacctResponse{
			Success: false,
			Message: "User is required",
		})
		return
	}

	// 构建请求
	req := &models.SacctRequest{
		Users: []string{user},
	}

	// 从查询参数解析其他选项
	if states := c.Query("states"); states != "" {
		req.States = []string{states}
	}

	req.StartTime = c.Query("starttime")
	req.EndTime = c.Query("endtime")

	// SSH 连接信息
	req.Host = c.Query("host")
	req.Username = c.Query("username")
	req.Password = c.Query("password")
	req.PrivateKey = c.Query("privatekey")

	if portStr := c.Query("port"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			req.Port = port
		}
	}

	// 验证必要参数
	if req.Host == "" || req.Username == "" {
		c.JSON(http.StatusBadRequest, models.SacctResponse{
			Success: false,
			Message: "Host and username are required",
		})
		return
	}

	if req.Password == "" && req.PrivateKey == "" {
		c.JSON(http.StatusBadRequest, models.SacctResponse{
			Success: false,
			Message: "Password or private key is required",
		})
		return
	}

	// 执行命令
	response := service.ExecuteSacct(req)

	if response.Success {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusInternalServerError, response)
	}
}

// GetAccounting 获取记账信息（POST 方法，支持更复杂的查询）
func GetAccounting(c *gin.Context) {
	service := NewSacctService()

	var req models.SacctRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.SacctResponse{
			Success: false,
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	// 验证必要参数
	if req.Host == "" || req.Username == "" {
		c.JSON(http.StatusBadRequest, models.SacctResponse{
			Success: false,
			Message: "Host and username are required",
		})
		return
	}

	if req.Password == "" && req.PrivateKey == "" {
		c.JSON(http.StatusBadRequest, models.SacctResponse{
			Success: false,
			Message: "Password or private key is required",
		})
		return
	}

	// 执行命令
	response := service.ExecuteSacct(&req)

	if response.Success {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusInternalServerError, response)
	}
}

// ExecuteSacct 执行 sacct 命令
func (s *SacctService) ExecuteSacct(req *models.SacctRequest) models.SacctResponse {
	// 构建命令
	command := s.parser.BuildCommand(req)

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
		return models.SacctResponse{
			Success: false,
			Message: "Failed to connect to SSH server: " + err.Error(),
			Command: command,
		}
	}
	defer sshClient.Close()

	// 执行命令
	output, err := sshClient.ExecuteCommand(command)
	if err != nil {
		return models.SacctResponse{
			Success:   false,
			Message:   "Failed to execute sacct command: " + err.Error(),
			Command:   command,
			RawOutput: output,
		}
	}

	// 解析输出
	jobs, err := s.parser.ParseOutput(output, req.Format)
	if err != nil {
		return models.SacctResponse{
			Success:   false,
			Message:   "Failed to parse sacct output: " + err.Error(),
			Command:   command,
			RawOutput: output,
		}
	}

	return models.SacctResponse{
		Success: true,
		Message: "Success",
		Data:    jobs,
		Total:   len(jobs),
		Command: command,
	}
}
