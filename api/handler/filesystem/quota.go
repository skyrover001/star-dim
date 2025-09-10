package filesystem

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func ParseQuotaOutput(output string) (*QuotaInfo, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return nil, errors.New("invalid quota output format")
	}

	// 查找数据行（非标题行）
	var dataLine string
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && strings.Contains(line, "Filesystem") && (i+1) < len(lines) {
			dataLine = lines[i+1]
			break
		}
	}

	if dataLine == "" {
		return nil, errors.New("no data found in quota output")
	}

	// 分割字段
	fields := strings.Fields(dataLine)
	if len(fields) < 9 {
		return nil, errors.New("insufficient fields in quota output")
	}

	quota := &QuotaInfo{
		Filesystem:  fields[0],
		KBytesGrace: fields[4],
		FilesGrace:  fields[8],
	}

	// 解析数值字段
	var err error
	if quota.KBytes, err = strconv.ParseInt(fields[1], 10, 64); err != nil {
		return nil, fmt.Errorf("failed to parse kbytes: %v", err)
	}
	if quota.KBytesQuota, err = strconv.ParseInt(fields[2], 10, 64); err != nil {
		return nil, fmt.Errorf("failed to parse kbytes quota: %v", err)
	}
	if quota.KBytesLimit, err = strconv.ParseInt(fields[3], 10, 64); err != nil {
		return nil, fmt.Errorf("failed to parse kbytes limit: %v", err)
	}
	if quota.Files, err = strconv.ParseInt(fields[5], 10, 64); err != nil {
		return nil, fmt.Errorf("failed to parse files: %v", err)
	}
	if quota.FilesQuota, err = strconv.ParseInt(fields[6], 10, 64); err != nil {
		return nil, fmt.Errorf("failed to parse files quota: %v", err)
	}
	if quota.FilesLimit, err = strconv.ParseInt(fields[7], 10, 64); err != nil {
		return nil, fmt.Errorf("failed to parse files limit: %v", err)
	}

	return quota, nil
}

// GetQuota gets disk quota information
// @Summary 获取磁盘配额信息
// @Description 获取指定用户或文件系统的磁盘配额信息，包括已使用空间、配额限制、文件数量等详细信息
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param sessionKey header string true "SSH会话密钥" example("tsh_a2e932b625c0d598db3800aa91b92016")
// @Param cluster query string true "集群名称" example("hpc1")
// @Param path query string false "查询路径" example("/home/user")
// @Success 200 {object} QuotaInfo "获取配额信息成功" example({"filesystem":"/dev/sda1","kbytes":1024000,"kbytes_quota":2048000,"kbytes_limit":2097152,"kbytes_grace":"none","files":1000,"files_quota":5000,"files_limit":10000,"files_grace":"none"})
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 404 {object} object{error=string} "路径不存在或配额信息不可用"
// @Failure 500 {object} object{error=string} "服务器内部错误或用户未登录"
// @Router /api/v1/filesystem/quota/ [get]
func (h *FilesHandler) Quota(c *gin.Context) {
	key, _, err := h.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if _, ok := h.Server.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	sshClient := h.Server.Clients[key].SSHClient

	session, err := sshClient.NewSession()
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	defer session.Close()

	cmd := "lfs quota -u " + h.Server.Clients[key].UserInfo.Name + " /"
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	quotaInfo, err := ParseQuotaOutput(string(output))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"quota":   quotaInfo,
		"success": "yes",
	})
}
