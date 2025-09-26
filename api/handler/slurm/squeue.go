package slurm

import (
	"bytes"
	"errors"
	"golang.org/x/crypto/ssh"
	"net/http"
	"star-dim/internal/models"
	"star-dim/internal/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// squeue to show slurm jobs
// @Summary 获取作业列表
// @Description 获取作业列表信息，包括作业ID、用户、状态、提交时间等信息
// @Tags 作业管理
// @Accept json
// @Produce json
// @Param request body models.SqueueRequest true "squeue请求参数"
// @Param sessionKey header string true "SSH会话密钥" example("tsh_a2e932b625c0d598db3800aa91b92016")
// @Success 200 {object} object{home_path=string,session_key=string} "查询成功，返回作业会计信息列表"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 401 {object} object{error=string} "认证失败，用户名或密码错误"
// @Failure 500 {object} object{error=string} "服务器内部错误或SSH连接失败"
// @Router /api/v1/slurm/jobs/ [post]
func (h *SlurmHandler) GetQueue(c *gin.Context) {
	key := c.GetHeader("sessionKey")
	if key == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if _, ok := h.Server.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	h.SSHClient = h.Server.Clients[key].SSHClient
	var req models.SqueueRequest
	// 从查询参数中获取过滤条件
	if accounts := c.Query("accounts"); accounts != "" {
		req.Accounts = strings.Split(accounts, ",")
	}
	if jobs := c.Query("jobs"); jobs != "" {
		req.Jobs = strings.Split(jobs, ",")
	}
	if partitions := c.Query("partitions"); partitions != "" {
		req.Partitions = strings.Split(partitions, ",")
	}
	if qos := c.Query("qos"); qos != "" {
		req.QOS = strings.Split(qos, ",")
	}
	if states := c.Query("states"); states != "" {
		req.States = strings.Split(states, ",")
	}
	if users := c.Query("users"); users != "" {
		req.Users = strings.Split(users, ",")
	}
	if names := c.Query("names"); names != "" {
		req.Names = strings.Split(names, ",")
	}
	if clusters := c.Query("clusters"); clusters != "" {
		req.Clusters = strings.Split(clusters, ",")
	}
	if licenses := c.Query("licenses"); licenses != "" {
		req.Licenses = strings.Split(licenses, ",")
	}
	if nodelist := c.Query("nodelist"); nodelist != "" {
		req.NodeList = strings.Split(nodelist, ",")
	}
	if steps := c.Query("steps"); steps != "" {
		req.Steps = strings.Split(steps, ",")
	}
	req.Reservation = c.Query("reservation")

	// 输出格式控制
	req.Format = c.Query("format")
	req.FormatLong = c.Query("format_long")
	req.NoHeader = c.Query("noheader") == "true"
	req.Long = c.Query("long") == "true"
	req.NoConvert = c.Query("noconvert") == "true"
	req.Array = c.Query("array") == "true"
	req.Start = c.Query("start") == "true"
	req.Verbose = c.Query("verbose") == "true"
	req.All = c.Query("all") == "true"
	req.Hide = c.Query("hide") == "true"
	req.Federation = c.Query("federation") == "true"
	req.Local = c.Query("local") == "true"
	req.Sibling = c.Query("sibling") == "true"
	req.OnlyJobState = c.Query("only_job_state") == "true"

	req.JSON = c.Query("json")
	req.YAML = c.Query("yaml")

	if sort := c.Query("sort"); sort != "" {
		req.Sort = strings.Split(sort, ",")
	}

	if iterateStr := c.Query("iterate"); iterateStr != "" {
		if iterate, err := strconv.Atoi(iterateStr); err == nil {
			req.Iterate = iterate
		}
	}

	// 处理请求
	processQueueRequest(c, req, h.SSHClient)
}

// GetJobQueue 获取指定作业的队列信息
func (h *SlurmHandler) GetJobQueue(c *gin.Context) {
	key := c.GetHeader("sessionKey")
	if key == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if _, ok := h.Server.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	h.SSHClient = h.Server.Clients[key].SSHClient
	jobid := c.Param("jobid")
	if jobid == "" {
		c.JSON(http.StatusBadRequest, models.SqueueResponse{
			Success: "no",
			Message: "Job ID is required",
		})
		return
	}

	var req models.SqueueRequest

	// 从查询参数中获取 SSH 连接信息
	req.Host = c.Query("host")
	req.Username = c.Query("username")
	req.Password = c.Query("password")
	req.PrivateKey = c.Query("privatekey")

	if portStr := c.Query("port"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			req.Port = port
		}
	}

	// 设置要查询的作业ID
	req.Jobs = []string{jobid}

	// 其他可选参数
	req.Format = c.Query("format")
	req.Long = c.Query("long") == "true"
	req.Verbose = c.Query("verbose") == "true"

	// 处理请求
	processQueueRequest(c, req, h.SSHClient)
}

// GetUserQueue 获取指定用户的作业队列
func (h *SlurmHandler) GetUserQueue(c *gin.Context) {
	key := c.GetHeader("sessionKey")
	if key == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if _, ok := h.Server.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	h.SSHClient = h.Server.Clients[key].SSHClient
	user := c.Param("user")
	if user == "" {
		c.JSON(http.StatusBadRequest, models.SqueueResponse{
			Success: "no",
			Message: "Username is required",
		})
		return
	}

	var req models.SqueueRequest

	// 从查询参数中获取 SSH 连接信息
	req.Host = c.Query("host")
	req.Username = c.Query("username")
	req.Password = c.Query("password")
	req.PrivateKey = c.Query("privatekey")

	if portStr := c.Query("port"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			req.Port = port
		}
	}

	// 设置要查询的用户
	req.Users = []string{user}

	// 其他过滤条件
	if states := c.Query("states"); states != "" {
		req.States = strings.Split(states, ",")
	}
	if partitions := c.Query("partitions"); partitions != "" {
		req.Partitions = strings.Split(partitions, ",")
	}

	// 输出格式控制
	req.Format = c.Query("format")
	req.Long = c.Query("long") == "true"
	req.Start = c.Query("start") == "true"

	// 处理请求
	processQueueRequest(c, req, h.SSHClient)
}

// QueryQueue 复杂队列查询（POST 请求）
func (h *SlurmHandler) QueryQueue(c *gin.Context) {
	var req models.SqueueRequest
	key := c.GetHeader("sessionKey")
	if key == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if _, ok := h.Server.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	h.SSHClient = h.Server.Clients[key].SSHClient
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.SqueueResponse{
			Success: "no",
			Message: "Invalid request parameters: " + err.Error(),
		})
		return
	}

	// 处理请求
	processQueueRequest(c, req, h.SSHClient)
}

// GetQueueStats 获取队列统计信息
func (h *SlurmHandler) GetQueueStats(c *gin.Context) {
	key := c.GetHeader("sessionKey")
	if key == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if _, ok := h.Server.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	h.SSHClient = h.Server.Clients[key].SSHClient
	var req models.SqueueRequest
	// 设置获取所有状态的作业
	req.States = []string{"all"}
	req.Format = "%.10i %.2t %.9P %.8u"
	req.NoHeader = true

	// 验证请求参数
	parser := utils.NewSlurmParser(nil)
	if err := parser.ValidateSqueueRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Validation error: " + err.Error(),
		})
		return
	}

	session, err := h.SSHClient.NewSession()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	// 构建 squeue 命令
	cmd := parser.BuildSqueueCommand(req)
	output := stdout.String()
	// 解析输出并统计
	jobs, err := parser.ParseSqueueOutput(output, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"message":    "Failed to parse squeue output: " + err.Error(),
			"command":    cmd,
			"raw_output": output,
		})
		return
	}

	// 统计各种状态的作业数量
	stats := map[string]int{
		"PENDING":    0,
		"RUNNING":    0,
		"SUSPENDED":  0,
		"COMPLETING": 0,
		"COMPLETED":  0,
		"CANCELLED":  0,
		"FAILED":     0,
		"TIMEOUT":    0,
		"PREEMPTED":  0,
		"NODE_FAIL":  0,
		"OTHER":      0,
	}

	totalJobs := len(jobs)
	for _, job := range jobs {
		state := strings.ToUpper(job.State)
		if _, exists := stats[state]; exists {
			stats[state]++
		} else {
			stats["OTHER"]++
		}
	}

	// 返回统计结果
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Success",
		"total_jobs": totalJobs,
		"statistics": stats,
		"command":    cmd,
	})
}

// processQueueRequest 处理队列查询请求的通用逻辑
func processQueueRequest(c *gin.Context, req models.SqueueRequest, sshClient *ssh.Client) {
	// 验证请求参数
	parser := utils.NewSlurmParser(nil)
	if err := parser.ValidateSqueueRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, models.SqueueResponse{
			Success: "no",
			Message: "Validation error: " + err.Error(),
		})
		return
	}

	session, err := sshClient.NewSession()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	// 构建 squeue 命令
	cmd := parser.BuildSqueueCommand(req)
	err = session.Run(cmd)
	output := stdout.String()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.SqueueResponse{
			Success:   "no",
			Message:   "Failed to execute squeue command: " + err.Error(),
			Command:   cmd,
			RawOutput: output,
		})
		return
	}

	// 解析输出
	var jobs []models.QueueJobInfo
	if req.Long {
		// 详细输出解析
		jobs, err = parser.ParseSqueueDetailedOutput(output)
	} else {
		// 标准输出解析
		hasHeader := !req.NoHeader
		jobs, err = parser.ParseSqueueOutput(output, hasHeader)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.SqueueResponse{
			Success:   "no",
			Message:   "Failed to parse squeue output: " + err.Error(),
			Command:   cmd,
			RawOutput: output,
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, models.SqueueResponse{
		Success: "yes",
		Message: "",
		Data:    jobs,
		Total:   len(jobs),
		Command: cmd,
	})
}
