package slurm

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"star-dim/internal/models"
	"star-dim/internal/service"
)

// sacct to show job accounting information
// @Summary 获取作业会计信息
// @Description 获取作业会计信息，包括作业ID、用户、状态、提交时间等详细信息
// @Tags 作业管理
// @Accept json
// @Produce json
// @Param request body models.SacctRequest true "sacct请求参数"
// @Param sessionKey header string true "SSH会话密钥" example("tsh_a2e932b625c0d598db3800aa91b92016")
// @Success 200 {object} object{home_path=string,session_key=string} "查询成功，返回作业会计信息列表"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 401 {object} object{error=string} "认证失败，用户名或密码错误"
// @Failure 500 {object} object{error=string} "服务器内部错误或SSH连接失败"
// @Router /api/v1/slurm/sacct/jobs/ [post]
//func (h *SlurmHandler) GetJobs(c *gin.Context) {
//	key := c.GetHeader("sessionKey")
//	if key == "" {
//		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
//		return
//	}
//	if _, ok := h.Server.Clients[key]; !ok {
//		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
//		return
//	}
//
//	h.SSHClient = h.Server.Clients[key].SSHClient
//	req := &models.SacctRequest{}
//	sacctService := service.NewSlurmService(h.SSHClient, h.Parser)
//	if jobids := c.Query("jobids"); jobids != "" {
//		req.JobIDs = []string{jobids}
//	}
//
//	if users := c.Query("users"); users != "" {
//		req.Users = []string{users}
//	}
//
//	if accounts := c.Query("accounts"); accounts != "" {
//		req.Accounts = []string{accounts}
//	}
//
//	if partitions := c.Query("partitions"); partitions != "" {
//		req.Partitions = []string{partitions}
//	}
//
//	if states := c.Query("states"); states != "" {
//		req.States = []string{states}
//	}
//
//	req.StartTime = c.Query("starttime")
//	req.EndTime = c.Query("endtime")
//	req.Format = c.Query("format")
//
//	if brief := c.Query("brief"); brief == "true" {
//		req.Brief = true
//	}
//
//	if long := c.Query("long"); long == "true" {
//		req.Long = true
//	}
//
//	if parsable := c.Query("parsable"); parsable == "true" {
//		req.Parsable = true
//	}
//
//	if noheader := c.Query("noheader"); noheader == "true" {
//		req.NoHeader = true
//	}
//
//	if allusers := c.Query("allusers"); allusers == "true" {
//		req.AllUsers = true
//	}
//
//	response := sacctService.ExecuteSacct(req)
//	if response.Success == "yes" {
//		c.JSON(http.StatusOK, response)
//	} else {
//		c.JSON(http.StatusInternalServerError, response)
//	}
//}

// sacct to show job accounting information
// @Summary 获取指定作业会计信息
// @Description 获取作业会计信息，包括作业ID、用户、状态、提交时间等详细信息
// @Tags 作业管理
// @Accept json
// @Produce json
// @Param request body models.SacctRequest true "sacct请求参数"
// @Param sessionKey header string true "SSH会话密钥" example("tsh_a2e932b625c0d598db3800aa91b92016")
// @Success 200 {object} object{home_path=string,session_key=string} "查询成功，返回作业会计信息列表"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 401 {object} object{error=string} "认证失败，用户名或密码错误"
// @Failure 500 {object} object{error=string} "服务器内部错误或SSH连接失败"
// @Router /api/v1/slurm/sacct/jobs/{jobid}/ [post]
//func (h *SlurmHandler) GetJobDetail(c *gin.Context) {
//	key := c.GetHeader("sessionKey")
//	if key == "" {
//		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
//		return
//	}
//	if _, ok := h.Server.Clients[key]; !ok {
//		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
//		return
//	}
//
//	sacctService := service.NewSlurmService(h.SSHClient, h.Parser)
//	jobid := c.Param("jobid")
//
//	if jobid == "" {
//		c.JSON(http.StatusBadRequest, models.SacctResponse{
//			Success: "no",
//			Message: "Job ID is required",
//		})
//		return
//	}
//
//	req := &models.SacctRequest{
//		JobIDs: []string{jobid},
//		Long:   true, // 获取详细信息
//	}
//
//	response := sacctService.ExecuteSacct(req)
//	if response.Success == "yes" && len(response.Data) > 0 {
//		c.JSON(http.StatusOK, response)
//	} else if response.Success == "yes" && len(response.Data) == 0 {
//		c.JSON(http.StatusNotFound, models.SacctResponse{
//			Success: "no",
//			Message: "Job not found",
//		})
//	} else {
//		c.JSON(http.StatusInternalServerError, response)
//	}
//}

// sacct to show job accounting information
// @Summary 获取指定用户作业会计信息
// @Description 获取作业会计信息，包括作业ID、用户、状态、提交时间等详细信息
// @Tags 作业管理
// @Accept json
// @Produce json
// @Param request body models.SacctRequest true "sacct请求参数"
// @Param sessionKey header string true "SSH会话密钥" example("tsh_a2e932b625c0d598db3800aa91b92016")
// @Success 200 {object} object{home_path=string,session_key=string} "查询成功，返回作业会计信息列表"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 401 {object} object{error=string} "认证失败，用户名或密码错误"
// @Failure 500 {object} object{error=string} "服务器内部错误或SSH连接失败"
// @Router /api/v1/slurm/sacct/job/user/{user}/ [post]
//func (h *SlurmHandler) GetUserJobs(c *gin.Context) {
//	key := c.GetHeader("sessionKey")
//	if key == "" {
//		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
//		return
//	}
//	if _, ok := h.Server.Clients[key]; !ok {
//		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
//		return
//	}
//
//	sacctService := service.NewSlurmService(h.SSHClient, h.Parser)
//	user := c.Param("user")
//
//	if user == "" {
//		c.JSON(http.StatusBadRequest, models.SacctResponse{
//			Success: "no",
//			Message: "User is required",
//		})
//		return
//	}
//
//	// 构建请求
//	req := &models.SacctRequest{
//		Users: []string{user},
//	}
//
//	// 从查询参数解析其他选项
//	if states := c.Query("states"); states != "" {
//		req.States = []string{states}
//	}
//
//	req.StartTime = c.Query("starttime")
//	req.EndTime = c.Query("endtime")
//
//	response := sacctService.ExecuteSacct(req)
//
//	if response.Success == "yes" {
//		c.JSON(http.StatusOK, response)
//	} else {
//		c.JSON(http.StatusInternalServerError, response)
//	}
//}

// sacct to show job accounting information
// @Summary 获取作业会计信息
// @Description 获取作业会计信息，包括作业ID、用户、状态、提交时间等详细信息
// @Tags 作业管理
// @Accept json
// @Produce json
// @Param request body models.SacctRequest true "sacct请求参数"
// @Param sessionKey header string true "SSH会话密钥" example("tsh_a2e932b625c0d598db3800aa91b92016")
// @Success 200 {object} object{home_path=string,session_key=string} "查询成功，返回作业会计信息列表"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 401 {object} object{error=string} "认证失败，用户名或密码错误"
// @Failure 500 {object} object{error=string} "服务器内部错误或SSH连接失败"
// @Router /api/v1/slurm/account/ [post]
func (h *SlurmHandler) GetAccounting(c *gin.Context) {
	key := c.GetHeader("sessionKey")
	if key == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if _, ok := h.Server.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}

	var req models.SacctRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.SacctResponse{
			Success: "no",
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}
	// 执行命令
	h.SSHClient = h.Server.Clients[key].SSHClient
	sacctService := service.NewSlurmService(h.SSHClient, h.Parser)
	response := sacctService.ExecuteSacct(&req)

	if response.Success == "yes" {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusInternalServerError, response)
	}
}
