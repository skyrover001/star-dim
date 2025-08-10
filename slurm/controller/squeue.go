package controller

import (
	"net/http"
	"slurm-jobacct/models"
	"slurm-jobacct/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetQueue 获取作业队列列表（GET 请求）
func GetQueue(c *gin.Context) {
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
	processQueueRequest(c, req)
}

// GetJobQueue 获取指定作业的队列信息
func GetJobQueue(c *gin.Context) {
	jobid := c.Param("jobid")
	if jobid == "" {
		c.JSON(http.StatusBadRequest, models.SqueueResponse{
			Success: false,
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
	processQueueRequest(c, req)
}

// GetUserQueue 获取指定用户的作业队列
func GetUserQueue(c *gin.Context) {
	user := c.Param("user")
	if user == "" {
		c.JSON(http.StatusBadRequest, models.SqueueResponse{
			Success: false,
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
	processQueueRequest(c, req)
}

// QueryQueue 复杂队列查询（POST 请求）
func QueryQueue(c *gin.Context) {
	var req models.SqueueRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.SqueueResponse{
			Success: false,
			Message: "Invalid request parameters: " + err.Error(),
		})
		return
	}

	// 处理请求
	processQueueRequest(c, req)
}

// processQueueRequest 处理队列查询请求的通用逻辑
func processQueueRequest(c *gin.Context, req models.SqueueRequest) {
	// 验证请求参数
	parser := utils.NewSqueueParser()
	if err := parser.ValidateRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, models.SqueueResponse{
			Success: false,
			Message: "Validation error: " + err.Error(),
		})
		return
	}

	// 构建 squeue 命令
	cmd := parser.BuildCommand(req)

	// 创建 SSH 客户端配置
	sshConfig := &models.SSHConfig{
		Host:       req.Host,
		Username:   req.Username,
		Password:   req.Password,
		PrivateKey: req.PrivateKey,
		Port:       req.Port,
		Timeout:    30, // 30秒超时
	}

	// 执行命令
	client := utils.NewSSHClient(sshConfig)
	err := client.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.SqueueResponse{
			Success: false,
			Message: "Failed to connect to SSH server: " + err.Error(),
			Command: cmd,
		})
		return
	}
	defer client.Close()

	output, err := client.ExecuteCommand(cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.SqueueResponse{
			Success:   false,
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
		jobs, err = parser.ParseDetailedOutput(output)
	} else {
		// 标准输出解析
		hasHeader := !req.NoHeader
		jobs, err = parser.ParseOutput(output, hasHeader)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.SqueueResponse{
			Success:   false,
			Message:   "Failed to parse squeue output: " + err.Error(),
			Command:   cmd,
			RawOutput: output,
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, models.SqueueResponse{
		Success: true,
		Message: "Success",
		Data:    jobs,
		Total:   len(jobs),
		Command: cmd,
	})
}

// GetQueueStats 获取队列统计信息
func GetQueueStats(c *gin.Context) {
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

	// 设置获取所有状态的作业
	req.States = []string{"all"}
	req.Format = "%.10i %.2t %.9P %.8u"
	req.NoHeader = true

	// 验证请求参数
	parser := utils.NewSqueueParser()
	if err := parser.ValidateRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Validation error: " + err.Error(),
		})
		return
	}

	// 构建 squeue 命令
	cmd := parser.BuildCommand(req)

	// 创建 SSH 客户端配置
	sshConfig := &models.SSHConfig{
		Host:       req.Host,
		Username:   req.Username,
		Password:   req.Password,
		PrivateKey: req.PrivateKey,
		Port:       req.Port,
		Timeout:    30, // 30秒超时
	}

	// 执行命令
	client := utils.NewSSHClient(sshConfig)
	err := client.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to connect to SSH server: " + err.Error(),
			"command": cmd,
		})
		return
	}
	defer client.Close()

	output, err := client.ExecuteCommand(cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"message":    "Failed to execute squeue command: " + err.Error(),
			"command":    cmd,
			"raw_output": output,
		})
		return
	}

	// 解析输出并统计
	jobs, err := parser.ParseOutput(output, false)
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
