package slurm

import (
	"errors"
	"net/http"
	"star-dim/internal/service"
	"star-dim/internal/utils"

	"github.com/gin-gonic/gin"
	"star-dim/internal/models"
)

// submit job by sbatch command
// @Summary 使用sbatch命令提交作业
// @Description 使用sbatch命令提交作业
// @Tags 作业管理
// @Accept json
// @Produce json
// @Param request body models.SbatchRequest true "sbatch请求参数"
// @Param sessionKey header string true "SSH会话密钥" example("tsh_a2e932b625c0d598db3800aa91b92016")
// @Success 200 {object} object{home_path=string,session_key=string} "查询成功，返回作业会计信息列表"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 401 {object} object{error=string} "认证失败，用户名或密码错误"
// @Failure 500 {object} object{error=string} "服务器内部错误或SSH连接失败"
// @Router /api/v1/slurm/sbatch/job/ [post]
func (h *SlurmHandler) SubmitJob(c *gin.Context) {
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
	var req models.SbatchRequest
	slurmService := service.NewSlurmService(h.SSHClient, h.Parser)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.SbatchResponse{
			Success: "no",
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	// 验证请求参数
	parser := utils.NewSlurmParser()
	if err := parser.ValidateSbatchRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.SbatchResponse{
			Success: "no",
			Message: "Request validation failed: " + err.Error(),
		})
		return
	}

	// 执行作业提交
	response := slurmService.ExecuteSbatch(&req)

	if response.Success == "yes" {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusInternalServerError, response)
	}
}

// SubmitJobWithScript 通过上传脚本文件提交作业
func (h *SlurmHandler) SubmitJobWithScript(c *gin.Context) {
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
	var req models.SbatchRequest
	slurmService := service.NewSlurmService(h.SSHClient, h.Parser)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.SbatchResponse{
			Success: "no",
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	// 解析 multipart form
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		c.JSON(http.StatusBadRequest, models.SbatchResponse{
			Success: "no",
			Message: "Failed to parse multipart form: " + err.Error(),
		})
		return
	}

	// 获取上传的脚本文件
	file, header, err := c.Request.FormFile("script")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.SbatchResponse{
			Success: "no",
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
			Success: "no",
			Message: "Failed to read script file: " + err.Error(),
		})
		return
	}

	req = models.SbatchRequest{
		Script: string(scriptContent),
	}
	h.parseFormData(c, &req)

	// 验证请求参数
	parser := utils.NewSlurmParser()
	if err := parser.ValidateSbatchRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.SbatchResponse{
			Success: "no",
			Message: "Request validation failed: " + err.Error(),
		})
		return
	}

	// 执行作业提交
	response := slurmService.ExecuteSbatchWithUpload(&req, header.Filename)

	if response.Success == "yes" {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusInternalServerError, response)
	}
}

// QuickSubmit 快速提交作业（简化接口）
func (h *SlurmHandler) QuickSubmit(c *gin.Context) {
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
			Success: "no",
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
	parser := utils.NewSlurmParser()
	if err := parser.ValidateSbatchRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.SbatchResponse{
			Success: "no",
			Message: "Request validation failed: " + err.Error(),
		})
		return
	}

	// 执行作业提交
	var response *models.SbatchResponse
	if req.Script != "" {
		response = slurmService.ExecuteSbatchWithUpload(&req, "job_script.sh")
	} else {
		response = slurmService.ExecuteSbatch(&req)
	}

	if response.Success == "yes" {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusInternalServerError, response)
	}
}

// parseFormData 从表单数据解析参数
func (s *SlurmHandler) parseFormData(c *gin.Context, req *models.SbatchRequest) {
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
