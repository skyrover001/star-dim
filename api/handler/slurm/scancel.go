package slurm

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"star-dim/internal/models"
	"strings"
)

// CancelJob cancels one or more Slurm jobs using scancel
// @Summary 取消Slurm作业
// @Description 使用scancel命令取消一个或多个Slurm作业，支持多种过滤条件
// @Tags 作业管理
// @Accept json
// @Produce json
// @Param session_key header string true "会话密钥"
// @Param request body models.ScancelRequest true "取消作业请求参数"
// @Success 200 {object} object{message=string,output=string} "操作成功"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 401 {object} object{error=string} "用户未认证"
// @Failure 500 {object} object{error=string} "服务器内部错误"
// @Router /api/v1/slurm/job [delete]
func (h *SlurmHandler) CancelJob(c *gin.Context) {
	sessionKey := c.GetHeader("session_key")
	if sessionKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_key is required"})
		return
	}

	client, exists := h.Server.Clients[sessionKey]
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	var req models.ScancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 构建scancel命令
	cmd := "scancel"

	// 添加各种参数
	if req.Account != "" {
		cmd += fmt.Sprintf(" --account=%s", req.Account)
	}
	if req.Batch {
		cmd += " --batch"
	}
	if req.Ctld {
		cmd += " --ctld"
	}
	if req.Cron {
		cmd += " --cron"
	}
	if req.Full {
		cmd += " --full"
	}
	if req.Hurry {
		cmd += " --hurry"
	}
	if req.Interactive {
		cmd += " --interactive"
	}
	if req.Clusters != "" {
		cmd += fmt.Sprintf(" --clusters=%s", req.Clusters)
	}
	if req.Name != "" {
		cmd += fmt.Sprintf(" --name=%s", req.Name)
	}
	if req.Partition != "" {
		cmd += fmt.Sprintf(" --partition=%s", req.Partition)
	}
	if req.Quiet {
		cmd += " --quiet"
	}
	if req.QOS != "" {
		cmd += fmt.Sprintf(" --qos=%s", req.QOS)
	}
	if req.Reservation != "" {
		cmd += fmt.Sprintf(" --reservation=%s", req.Reservation)
	}
	if req.Sibling != "" {
		cmd += fmt.Sprintf(" --sibling=%s", req.Sibling)
	}
	if req.Signal != "" {
		cmd += fmt.Sprintf(" --signal=%s", req.Signal)
	}
	if req.State != "" {
		cmd += fmt.Sprintf(" --state=%s", req.State)
	}
	if req.User != "" {
		cmd += fmt.Sprintf(" --user=%s", req.User)
	}
	if req.Verbose {
		cmd += " --verbose"
	}
	if req.NodeList != "" {
		cmd += fmt.Sprintf(" --nodelist=%s", req.NodeList)
	}
	if req.WCKey != "" {
		cmd += fmt.Sprintf(" --wckey=%s", req.WCKey)
	}

	// 添加作业ID（可选）
	if len(req.JobIDs) > 0 {
		cmd += " " + strings.Join(req.JobIDs, " ")
	}

	// 执行命令
	session, err := client.SSHClient.NewSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   fmt.Sprintf("scancel failed: %v", err),
			"output":  string(output),
			"command": cmd,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "scancel executed successfully",
		"output":  string(output),
		"command": cmd,
	})
}
