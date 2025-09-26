package slurm

import (
	"bytes"
	"errors"
	"golang.org/x/crypto/ssh"
	"net/http"
	"star-dim/internal/models"
	"star-dim/internal/service"
	"star-dim/internal/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// sinfo to show cluster and node info
// @Summary 获取集群、分区和节点信息
// @Description 获取集群信息
// @Tags 作业管理
// @Accept json
// @Produce json
// @Param request body models.SinfoRequest true "sinfo请求参数"
// @Param sessionKey header string true "SSH会话密钥" example("tsh_a2e932b625c0d598db3800aa91b92016")
// @Success 200 {object} object{home_path=string,session_key=string} "查询成功，返回作业会计信息列表"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 401 {object} object{error=string} "认证失败，用户名或密码错误"
// @Failure 500 {object} object{error=string} "服务器内部错误或SSH连接失败"
// @Router /api/v1/slurm/cluster/ [post]
func (h *SlurmHandler) GetClusterInfo(c *gin.Context) {
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
	slurmService := service.NewSlurmService(h.SSHClient, h.Parser)
	var req models.SinfoRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果无法从body中获取参数，则尝试从查询参数中获取
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
	}

	// 处理请求
	processClusterInfoRequest(c, req, slurmService.SSHClient())
}

// GetNodeInfo 获取指定节点的信息
func (h *SlurmHandler) GetNodeInfo(c *gin.Context) {
	key := c.GetHeader("sessionKey")
	if key == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if _, ok := h.Server.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	nodename := c.Param("nodename")
	h.SSHClient = h.Server.Clients[key].SSHClient
	slurmService := service.NewSlurmService(h.SSHClient, h.Parser)
	if nodename == "" {
		c.JSON(http.StatusBadRequest, models.SinfoResponse{
			Success: "no",
			Message: "Node name is required",
		})
		return
	}

	var req models.SinfoRequest

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
	processClusterInfoRequest(c, req, slurmService.SSHClient())
}

// GetPartitionInfo 获取指定分区的信息
func (h *SlurmHandler) GetPartitionInfo(c *gin.Context) {
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
	slurmService := service.NewSlurmService(h.SSHClient, h.Parser)
	partition := c.Param("partition")
	if partition == "" {
		c.JSON(http.StatusBadRequest, models.SinfoResponse{
			Success: "no",
			Message: "Partition name is required",
		})
		return
	}

	var req models.SinfoRequest

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
	processClusterInfoRequest(c, req, slurmService.SSHClient())
}

// GetReservationInfo 获取预留信息
func (h *SlurmHandler) GetReservationInfo(c *gin.Context) {
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
	slurmService := service.NewSlurmService(h.SSHClient, h.Parser)
	var req models.SinfoRequest

	if portStr := c.Query("port"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			req.Port = port
		}
	}

	// 设置预留模式
	req.Reservation = true
	req.Long = c.Query("long") == "true"

	// 处理请求
	processClusterInfoRequest(c, req, slurmService.SSHClient())
}

// QueryClusterInfo 复杂集群信息查询（POST 请求）
func (h *SlurmHandler) QueryClusterInfo(c *gin.Context) {
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
	slurmService := service.NewSlurmService(h.SSHClient, h.Parser)
	var req models.SinfoRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.SinfoResponse{
			Success: "no",
			Message: "Invalid request parameters: " + err.Error(),
		})
		return
	}

	// 处理请求
	processClusterInfoRequest(c, req, slurmService.SSHClient())
}

// GetClusterSummary 获取集群摘要信息
func (h *SlurmHandler) GetClusterSummary(c *gin.Context) {
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
	slurmService := service.NewSlurmService(h.SSHClient, h.Parser)
	var req models.SinfoRequest

	// 设置摘要模式
	req.Summarize = true
	req.NoHeader = true

	// 处理请求
	processClusterInfoRequest(c, req, slurmService.SSHClient())
}

// processClusterInfoRequest 处理集群信息查询请求的通用逻辑
func processClusterInfoRequest(c *gin.Context, req models.SinfoRequest, sshClient *ssh.Client) {
	// 验证请求参数
	parser := utils.NewSlurmParser(nil)
	if err := parser.ValidateSinfoRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, models.SinfoResponse{
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
	// 构建 sinfo 命令
	cmd := parser.BuildSinfoCommand(req)
	err = session.Run(cmd)
	var output = stdout.String()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.SinfoResponse{
			Success:   "no",
			Message:   "Failed to execute sinfo command: " + err.Error(),
			Command:   cmd,
			RawOutput: output,
		})
		return
	}

	// 解析输出
	data, err := parser.ParseSinfoOutput(output, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.SinfoResponse{
			Success:   "no",
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
		Success: "yes",
		Message: "",
		Data:    data,
		Total:   total,
		Command: cmd,
	})
}
