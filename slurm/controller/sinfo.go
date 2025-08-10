package controller

import (
	"net/http"
	"slurm-jobacct/models"
	"slurm-jobacct/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetClusterInfo 获取集群信息列表（GET 请求）
func GetClusterInfo(c *gin.Context) {
	var req models.SinfoRequest

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
	if nodes := c.Query("nodes"); nodes != "" {
		req.Nodes = strings.Split(nodes, ",")
	}
	if partitions := c.Query("partitions"); partitions != "" {
		req.Partitions = strings.Split(partitions, ",")
	}
	if states := c.Query("states"); states != "" {
		req.States = strings.Split(states, ",")
	}
	if clusters := c.Query("clusters"); clusters != "" {
		req.Clusters = strings.Split(clusters, ",")
	}

	// 显示选项
	req.All = c.Query("all") == "true"
	req.Dead = c.Query("dead") == "true"
	req.Exact = c.Query("exact") == "true"
	req.Future = c.Query("future") == "true"
	req.Hide = c.Query("hide") == "true"
	req.Long = c.Query("long") == "true"
	req.NodeCentric = c.Query("node_centric") == "true"
	req.Responding = c.Query("responding") == "true"
	req.ListReasons = c.Query("list_reasons") == "true"
	req.Summarize = c.Query("summarize") == "true"
	req.Reservation = c.Query("reservation") == "true"
	req.Verbose = c.Query("verbose") == "true"
	req.Federation = c.Query("federation") == "true"
	req.Local = c.Query("local") == "true"
	req.NoConvert = c.Query("noconvert") == "true"
	req.NoHeader = c.Query("noheader") == "true"

	// 输出格式控制
	req.Format = c.Query("format")
	req.FormatLong = c.Query("format_long")
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
	processClusterInfoRequest(c, req)
}

// GetNodeInfo 获取指定节点的信息
func GetNodeInfo(c *gin.Context) {
	nodename := c.Param("nodename")
	if nodename == "" {
		c.JSON(http.StatusBadRequest, models.SinfoResponse{
			Success: false,
			Message: "Node name is required",
		})
		return
	}

	var req models.SinfoRequest

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

	// 设置要查询的节点
	req.Nodes = []string{nodename}
	req.Long = c.Query("long") == "true"
	req.Verbose = c.Query("verbose") == "true"

	// 处理请求
	processClusterInfoRequest(c, req)
}

// GetPartitionInfo 获取指定分区的信息
func GetPartitionInfo(c *gin.Context) {
	partition := c.Param("partition")
	if partition == "" {
		c.JSON(http.StatusBadRequest, models.SinfoResponse{
			Success: false,
			Message: "Partition name is required",
		})
		return
	}

	var req models.SinfoRequest

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

	// 设置要查询的分区
	req.Partitions = []string{partition}

	// 其他可选参数
	if states := c.Query("states"); states != "" {
		req.States = strings.Split(states, ",")
	}
	req.Long = c.Query("long") == "true"
	req.All = c.Query("all") == "true"

	// 处理请求
	processClusterInfoRequest(c, req)
}

// GetReservationInfo 获取预留信息
func GetReservationInfo(c *gin.Context) {
	var req models.SinfoRequest

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

	// 设置预留模式
	req.Reservation = true
	req.Long = c.Query("long") == "true"

	// 处理请求
	processClusterInfoRequest(c, req)
}

// QueryClusterInfo 复杂集群信息查询（POST 请求）
func QueryClusterInfo(c *gin.Context) {
	var req models.SinfoRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.SinfoResponse{
			Success: false,
			Message: "Invalid request parameters: " + err.Error(),
		})
		return
	}

	// 处理请求
	processClusterInfoRequest(c, req)
}

// GetClusterSummary 获取集群摘要信息
func GetClusterSummary(c *gin.Context) {
	var req models.SinfoRequest

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

	// 设置摘要模式
	req.Summarize = true
	req.NoHeader = true

	// 处理请求
	processClusterInfoRequest(c, req)
}

// processClusterInfoRequest 处理集群信息查询请求的通用逻辑
func processClusterInfoRequest(c *gin.Context, req models.SinfoRequest) {
	// 验证请求参数
	parser := utils.NewSinfoParser()
	if err := parser.ValidateRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, models.SinfoResponse{
			Success: false,
			Message: "Validation error: " + err.Error(),
		})
		return
	}

	// 构建 sinfo 命令
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
		c.JSON(http.StatusInternalServerError, models.SinfoResponse{
			Success: false,
			Message: "Failed to connect to SSH server: " + err.Error(),
			Command: cmd,
		})
		return
	}
	defer client.Close()

	output, err := client.ExecuteCommand(cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.SinfoResponse{
			Success:   false,
			Message:   "Failed to execute sinfo command: " + err.Error(),
			Command:   cmd,
			RawOutput: output,
		})
		return
	}

	// 解析输出
	data, err := parser.ParseOutput(output, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.SinfoResponse{
			Success:   false,
			Message:   "Failed to parse sinfo output: " + err.Error(),
			Command:   cmd,
			RawOutput: output,
		})
		return
	}

	// 计算总数
	total := 0
	switch v := data.(type) {
	case []models.NodeInfo:
		total = len(v)
	case []models.NodeSummary:
		total = len(v)
	case []models.ReservationInfo:
		total = len(v)
	}

	// 返回结果
	c.JSON(http.StatusOK, models.SinfoResponse{
		Success: true,
		Message: "Success",
		Data:    data,
		Total:   total,
		Command: cmd,
	})
}
