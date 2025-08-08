package controller

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/sftp"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type FileInfoJSON struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	Mode   string `json:"mode"`
	Modify string `json:"modify"`
	IsDir  bool   `json:"isDir"`
}

type QuotaInfo struct {
	Filesystem  string `json:"filesystem"`
	KBytes      int64  `json:"kbytes"`
	KBytesQuota int64  `json:"kbytes_quota"`
	KBytesLimit int64  `json:"kbytes_limit"`
	KBytesGrace string `json:"kbytes_grace"`
	Files       int64  `json:"files"`
	FilesQuota  int64  `json:"files_quota"`
	FilesLimit  int64  `json:"files_limit"`
	FilesGrace  string `json:"files_grace"`
}

func ParseQuotaOutput(output string) (*QuotaInfo, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return nil, errors.New("invalid quota output format")
	}

	// 查找数据行（非标题行）
	var dataLine string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.Contains(line, "Filesystem") {
			dataLine = line
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

func FileInfoToJSON(fileInfo os.FileInfo) FileInfoJSON {
	return FileInfoJSON{
		Name:   fileInfo.Name(),
		Size:   fileInfo.Size(),
		Mode:   fileInfo.Mode().String(),
		Modify: fileInfo.ModTime().String(),
		IsDir:  fileInfo.IsDir(),
	}
}

// GetHomeDir by username
func GetHomeDir(username string) string {
	return ""
}

// Close connects
func (s *JumpService) Close() {
	for _, sc := range s.Clients {
		if sc != nil {
			_ = sc.SSHClient.Close()
			_ = sc.SftpClient.Close()
		}
	}
}

// List objects
// @Summary 列出目录下的文件和子目录
// @Description 获取指定路径下的文件和目录列表，返回文件名、大小、权限、修改时间等信息
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param cluster query string true "集群名称" example("hpc1")
// @Param path query string true "目录路径" example("/ai")
// @Param sessionKey header string true "SSH会话密钥" example("tsh_a2e932b625c0d598db3800aa91b92016")
// @Success 200 {object} object{listContent=[]FileInfoJSON,listLength=int,success=string} "成功返回文件列表"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 500 {object} object{error=string} "服务器内部错误或用户未登录"
// @Router /api/v2/files/ [get]
func (j *JumpService) List(c *gin.Context) {
	// list Object
	key, req, err := j.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	}
	if _, ok := j.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	sftpClient := j.Clients[key].SftpClient
	path := j.Clients[key].RepackPath(req.Path)
	fmt.Println("list path:", path, " home path:", j.Clients[key].SSHInfo.HomePath)
	objs, err := sftpClient.ReadDir(path)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	var files []FileInfoJSON
	dirCount := 0
	for _, obj := range objs {
		files = append(files, FileInfoToJSON(obj))
		if obj.IsDir() {
			dirCount++
		}
	}
	//c.JSON(200, map[string]ifapi{}{"objects": files, "filesCount": len(objs), "dirCount": dirCount}) // like ifapi
	c.JSON(200, map[string]interface{}{"listContent": files, "listLength": len(objs), "success": "yes"}) //like api server
	// curl test: curl -X POST -H "Content-Type: application/json" -d "{\"username\":\"root\",\"path\":\"/root/\"}" http://localhost:8080/api/v2/document/files/
}

// Transmission uploads a file
// @Summary 上传文件
// @Description 上传文件到指定路径，支持断点续传和文件覆盖更新功能
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Param sessionKey header string true "SSH会话密钥" example("tsh_a2e932b625c0d598db3800aa91b92016")
// @Param cluster formData string true "集群名称" example("hpc1")
// @Param path formData string true "上传路径" example("/ai/upload/test.txt")
// @Param offset formData string false "文件偏移量（用于断点续传）" example("0")
// @Param update formData string false "是否覆盖已存在文件" Enums(true,false) example("false")
// @Param file formData file true "要上传的文件"
// @Success 200 {object} object{success=string} "上传成功" example({"success":"yes"})
// @Failure 400 {object} object{error=string} "请求参数错误或文件为空"
// @Failure 409 {object} object{error=string} "文件已存在且未设置覆盖标志"
// @Failure 500 {object} object{error=string} "服务器内部错误或用户未登录"
// @Router /api/v2/files/transmission/ [post]
func (j *JumpService) Transmission(c *gin.Context) {
	key, _, err := j.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if _, ok := j.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	sftpClient := j.Clients[key].SftpClient

	path, _ := c.GetPostForm("path")
	path = j.Clients[key].RepackPath(path)
	offsetStr, _ := c.GetPostForm("offset")
	update, ok := c.GetPostForm("update")
	if !ok {
		update = "false"
	}
	if update != "true" && update != "false" {
		c.JSON(http.StatusInternalServerError, errors.New("update must be true,false or empty"))
		return
	}
	offset, _ := strconv.ParseInt(offsetStr, 10, 64)
	input, err := c.FormFile("file")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	_, err = sftpClient.Lstat(path)
	log.Println("upload path:", path, " offset:", offset, " update:", update, " size:", input.Size)
	if input.Size == 0 {
		// upload empty file
		if err == nil {
			// file exist
			if update == "false" {
				log.Println(err)
				c.JSON(http.StatusInternalServerError, errors.New("file exist"))
				return
			} else {
				// force update
				err = sftpClient.Remove(path)
				f, err := sftpClient.Create(path)
				if err != nil {
					log.Println(err)
					c.JSON(http.StatusInternalServerError, err)
					return
				}
				_ = f.Close()
				c.JSON(200, map[string]interface{}{"success": "yes"})
				return
			}
		} else {
			// file not exist
			if strings.Contains(err.Error(), "file does not exist") {
				f, err := sftpClient.Create(path)
				if err != nil {
					log.Println(err)
					c.JSON(http.StatusInternalServerError, err)
					return
				}
				_ = f.Close()
				c.JSON(200, map[string]interface{}{"success": "yes"})
				return
			} else {
				log.Println(err)
				c.JSON(http.StatusInternalServerError, err)
				return
			}
		}
	} else {
		// upload file not empty
		var dstFile *sftp.File
		if err == nil {
			// file exist
			if update == "false" {
				log.Println(err)
				c.JSON(http.StatusInternalServerError, errors.New("file exist"))
				return
			}
		}
		if offset == 0 {
			dstFile, err = sftpClient.OpenFile(path, os.O_CREATE|os.O_RDWR)
		} else {
			dstFile, err = sftpClient.OpenFile(path, os.O_RDWR)
		}
		defer dstFile.Close()
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		_, _ = dstFile.Seek(offset, 0)
		srcFile, err := input.Open()
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		data, err := ioutil.ReadAll(srcFile)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		_, err = dstFile.WriteAt(data, offset)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, map[string]interface{}{"success": "yes"})
	}
}

// Download downloads a file or directory
// @Summary 下载文件或目录
// @Description 下载指定路径的文件或目录。文件直接下载，目录会被打包成ZIP文件下载
// @Tags 文件管理
// @Accept json
// @Produce application/octet-stream
// @Param sessionKey header string true "SSH会话密钥" example("tsh_a2e932b625c0d598db3800aa91b92016")
// @Param cluster query string true "集群名称" example("hpc1")
// @Param path query string true "文件或目录路径" example("/ai/mcp")
// @Success 200 {file} file "下载成功，返回文件内容或ZIP压缩包"
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 404 {object} object{error=string} "文件或目录不存在"
// @Failure 500 {object} object{error=string} "服务器内部错误或用户未登录"
// @Router /api/v2/files/download/ [get]
func (j *JumpService) Download(c *gin.Context) {
	key, req, err := j.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if _, ok := j.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	sftpClient := j.Clients[key].SftpClient
	path := j.Clients[key].RepackPath(req.Path)
	log.Println("download path:", path)
	fileInfo, err := sftpClient.Lstat(path)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	// download dir
	if fileInfo.IsDir() {
		az := zip.NewWriter(c.Writer)
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", fileInfo.Name()))
		walker := sftpClient.Walk(path)
		for walker.Step() {
			if err := walker.Err(); err != nil {
				continue
			}
			fi, _ := sftpClient.Lstat(walker.Path())
			if fi.IsDir() {
				_, _ = az.Create(walker.Path() + "/")
				continue
			}
			ds, _ := az.Create(walker.Path())
			distFile, _ := sftpClient.OpenFile(walker.Path(), os.O_RDONLY)
			_, _ = io.Copy(ds, distFile)
		}
		_ = az.Close()
	} else {
		// download file
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileInfo.Name()))
		c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
		c.Header("Content-Type", "application/octet-stream")
		buffer := make([]byte, 1024*1024)
		dstFile, _ := sftpClient.OpenFile(path, os.O_RDONLY)
		for {
			n, err := dstFile.Read(buffer)
			if err != nil && err != io.EOF {
				log.Println(err)
				c.JSON(http.StatusInternalServerError, err)
				return
			}
			if n == 0 {
				break
			}
			_, err = c.Writer.Write(buffer[:n])
			if err != nil {
				log.Println(err)
				c.JSON(http.StatusInternalServerError, err)
				return
			}
		}
	}
	// curl test: http://localhost:8080/api/v2/document/download/?username=root&path=/root/scritps
}

// GetAttributes gets file or directory attributes
// @Summary 获取文件或目录属性
// @Description 获取指定路径文件或目录的详细属性信息，包括名称、大小、权限、修改时间等
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param sessionKey header string true "SSH会话密钥" example("tsh_a2e932b625c0d598db3800aa91b92016")
// @Param cluster query string true "集群名称" example("hpc1")
// @Param path query string true "文件或目录路径" example("/ai/new_folder_rename/test_renamed.sh")
// @Success 200 {object} object{name=string,size=int,mode=string,modify=string,isDir=bool} "获取属性成功" example({"name":"test_renamed.sh","size":29,"mode":"-rw-r--r--","modify":"2025-08-07 10:16:36 +0800 CST","isDir":false})
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 404 {object} object{error=string} "文件或目录不存在"
// @Failure 500 {object} object{error=string} "服务器内部错误或用户未登录"
// @Router /api/v2/files/attr/ [get]
func (j *JumpService) Attr(c *gin.Context) {
	key, req, err := j.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if _, ok := j.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	sftpClient := j.Clients[key].SftpClient

	path := j.Clients[key].RepackPath(req.Path)
	log.Println("attr path:", path)
	fileInfo, err := sftpClient.Lstat(path)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(200, FileInfoToJSON(fileInfo))
}

// Rename renames a file or directory
// @Summary 重命名文件或目录
// @Description 将指定路径的文件或目录重命名到新路径，支持文件和目录的重命名操作
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param sessionKey header string true "SSH会话密钥" example("tsh_a2e932b625c0d598db3800aa91b92016")
// @Param request body object{old_path=string,new_path=string} true "重命名请求参数" Example({"old_path":"/ai/new_folder","new_path":"/ai/new_folder_rename"})
// @Success 200 {object} object{success=string} "重命名成功" example({"success":"yes"})
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 404 {object} object{error=string} "源文件或目录不存在"
// @Failure 409 {object} object{error=string} "目标路径已存在"
// @Failure 500 {object} object{error=string} "服务器内部错误或用户未登录"
// @Router /api/v2/files/ [put]
func (j *JumpService) Rename(c *gin.Context) {
	key, req, err := j.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if _, ok := j.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	sftpClient := j.Clients[key].SftpClient
	oldPath := j.Clients[key].RepackPath(req.OldPath)
	newPath := j.Clients[key].RepackPath(req.NewPath)
	log.Println("rename oldPath:", oldPath, " newPath:", newPath)

	err = sftpClient.Rename(oldPath, newPath)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(200, map[string]interface{}{"success": "yes"})
}

// New creates a new file or directory
// @Summary 创建新文件或目录
// @Description 在指定路径创建新的文件或目录，支持创建空文件和目录
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param sessionKey header string true "SSH会话密钥" example("tsh_a2e932b625c0d598db3800aa91b92016")
// @Param request body object{path=string,type=string} true "创建请求参数" Example({"path":"/ai/new_folder","type":"dir"})
// @Success 200 {object} object{success=string} "创建成功" example({"success":"yes"})
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 409 {object} object{error=string} "文件或目录已存在"
// @Failure 500 {object} object{error=string} "服务器内部错误或用户未登录"
// @Router /api/v2/files/ [post]
func (j *JumpService) New(c *gin.Context) {
	key, req, err := j.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if _, ok := j.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
	}
	jumpClient := j.Clients[key]
	sftpClient := jumpClient.SftpClient

	path := j.Clients[key].RepackPath(req.Path)
	fileType := req.Type
	log.Println("new path:", path, " type:", fileType, " home path:", jumpClient.SSHInfo.HomePath)
	if fileType != "file" && fileType != "dir" {
		c.JSON(http.StatusInternalServerError, errors.New("type must be file or dir"))
	}
	_, err = sftpClient.Lstat(path)
	if err == nil {
		c.JSON(http.StatusInternalServerError, errors.New("file exist"))
	} else {
		if strings.Contains(err.Error(), "file does not exist") {
			if fileType == "file" {
				f, err := sftpClient.Create(path)
				if err != nil {
					log.Println(err)
					c.JSON(http.StatusInternalServerError, err)
				}
				_ = f.Close()
			} else {
				err = sftpClient.Mkdir(path)
				if err != nil {
					log.Println(err)
					c.JSON(http.StatusInternalServerError, err)
				}
			}
			c.JSON(200, map[string]interface{}{"success": "yes"})
		} else {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, err)
		}
	}
}

// Delete removes a file or directory
// @Summary 删除文件或目录
// @Description 删除指定路径的文件或目录，支持删除单个文件和空目录
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param sessionKey header string true "SSH会话密钥" example("tsh_a2e932b625c0d598db3800aa91b92016")
// @Param request body object{path=string,type=string} true "删除请求参数" Example({"path":"/ai/new_folder/test.sh","type":"file"})
// @Success 200 {object} object{success=string} "删除成功" example({"success":"yes"})
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 404 {object} object{error=string} "文件或目录不存在"
// @Failure 500 {object} object{error=string} "服务器内部错误或用户未登录"
// @Router /api/v2/files/ [delete]
func (j *JumpService) Delete(c *gin.Context) {
	key, req, err := j.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if _, ok := j.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	sftpClient := j.Clients[key].SftpClient
	path := j.Clients[key].RepackPath(req.Path)
	log.Println("delete path:", path)
	fileInfo, err := sftpClient.Lstat(path)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	if fileInfo.IsDir() {
		err = sftpClient.RemoveDirectory(path)
	} else {
		err = sftpClient.Remove(path)
	}
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(200, map[string]interface{}{"success": "yes"})
}

// Copy copies a file or directory
// @Summary 复制文件或目录
// @Description 将源路径的文件或目录复制到目标路径，支持文件和目录的复制操作
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param sessionKey header string true "SSH会话密钥" example("tsh_a2e932b625c0d598db3800aa91b92016")
// @Param request body object{src_path=string,dst_path=string} true "复制请求参数" Example({"src_path":"/ai/new_folder_rename/test_renamed.sh","dst_path":"/ai/new_folder_rename/test_copy.sh"})
// @Success 200 {object} object{success=string} "复制成功" example({"success":"yes"})
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 404 {object} object{error=string} "源文件或目录不存在"
// @Failure 409 {object} object{error=string} "目标路径已存在"
// @Failure 500 {object} object{error=string} "服务器内部错误或用户未登录"
// @Router /api/v2/files/copy/ [post]
func (j *JumpService) Copy(c *gin.Context) {
	key, req, err := j.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if _, ok := j.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	sftpClient := j.Clients[key].SftpClient
	sshClient := j.Clients[key].SSHClient

	srcPath := req.SrcPath
	dstPath := req.DstPath
	srcPath = j.Clients[key].RepackPath(srcPath)
	dstPath = j.Clients[key].RepackPath(dstPath)
	fmt.Println("sftpClient:", sftpClient, "sshClient:", sshClient, " srcPath:", srcPath, "dstPath:", dstPath)

	srcFileInfo, err := sftpClient.Lstat(srcPath)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	_, err = sftpClient.Lstat(dstPath)
	if err == nil {
		// dist path exist
		log.Println(err)
		c.JSON(http.StatusInternalServerError, errors.New("file exist"))
		return
	} else {
		sshSession, err := sshClient.NewSession()
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		defer sshSession.Close()
		if srcFileInfo.IsDir() {
			// copy dir
			cmd := fmt.Sprintf("cp -r %s %s", srcPath, dstPath)
			err = sshSession.Run(cmd)
			if err != nil {
				log.Println(err)
				c.JSON(http.StatusInternalServerError, err)
				return
			}
		} else {
			// copy file
			cmd := fmt.Sprintf("cp %s %s", srcPath, dstPath)
			err = sshSession.Run(cmd)
			if err != nil {
				log.Println(err)
				c.JSON(http.StatusInternalServerError, err)
				return
			}
			c.JSON(200, map[string]interface{}{"success": "yes"})
		}
	}
}

// Move moves a file or directory
// @Summary 移动文件或目录
// @Description 将源路径的文件或目录移动到目标路径，支持文件和目录的移动操作
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param sessionKey header string true "SSH会话密钥" example("tsh_a2e932b625c0d598db3800aa91b92016")
// @Param request body object{src_path=string,dst_path=string} true "移动请求参数" Example({"src_path":"/ai/new_folder_rename/test_copy.sh","dst_path":"/ai/new_folder_rename/test_moved.sh"})
// @Success 200 {object} object{success=string} "移动成功" example({"success":"yes"})
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 404 {object} object{error=string} "源文件或目录不存在"
// @Failure 409 {object} object{error=string} "目标路径已存在"
// @Failure 500 {object} object{error=string} "服务器内部错误或用户未登录"
// @Router /api/v2/files/move/ [post]
func (j *JumpService) Move(c *gin.Context) {
	key, req, err := j.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if _, ok := j.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	sftpClient := j.Clients[key].SftpClient
	sshClient := j.Clients[key].SSHClient

	srcPath := req.SrcPath
	dstPath := req.DstPath
	srcPath = j.Clients[key].RepackPath(srcPath)
	dstPath = j.Clients[key].RepackPath(dstPath)
	log.Println("sftpClient:", sftpClient, "sshClient:", sshClient, " srcPath:", srcPath, "dstPath:", dstPath)

	srcFileInfo, err := sftpClient.Lstat(srcPath)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	_, err = sftpClient.Lstat(dstPath)
	if err == nil {
		// dist path exist
		log.Println(err)
		c.JSON(http.StatusInternalServerError, errors.New("file exist"))
		return
	} else {
		sshSession, err := sshClient.NewSession()
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		defer sshSession.Close()
		if srcFileInfo.IsDir() {
			// move dir
			cmd := fmt.Sprintf("mv %s %s", srcPath, dstPath)
			err = sshSession.Run(cmd)
			if err != nil {
				log.Println(err)
				c.JSON(http.StatusInternalServerError, err)
				return
			}
		} else {
			// move file
			cmd := fmt.Sprintf("mv %s %s", srcPath, dstPath)
			err = sshSession.Run(cmd)
			if err != nil {
				log.Println(err)
				c.JSON(http.StatusInternalServerError, err)
				return
			}
			c.JSON(200, map[string]interface{}{"success": "yes"})
		}
	}
}

// ReadContent reads content from a file
// @Summary 读取文件内容
// @Description 读取指定路径文件的内容，返回文件的文本内容
// @Tags 文件管理
// @Accept json
// @Produce plain
// @Param sessionKey header string true "SSH会话密钥" example("tsh_a2e932b625c0d598db3800aa91b92016")
// @Param cluster query string true "集群名称" example("hpc1")
// @Param path query string true "文件路径" example("/ai/new_folder_rename/test_renamed.sh")
// @Success 200 {string} string "文件内容" example("#!/bin/bash\\necho Hello World")
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 404 {object} object{error=string} "文件不存在"
// @Failure 500 {object} object{error=string} "服务器内部错误或用户未登录"
// @Router /api/v2/files/content/ [get]
func (j *JumpService) ReadFile(c *gin.Context) {
	key, req, err := j.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if _, ok := j.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	sftpClient := j.Clients[key].SftpClient

	path := j.Clients[key].RepackPath(req.Path)
	log.Println("sftpClient:", sftpClient, "path:", path)
	fileInfo, err := sftpClient.Lstat(path)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	if fileInfo.IsDir() {
		c.JSON(http.StatusBadRequest, errors.New("path is a directory"))
		return
	}

	file, err := sftpClient.OpenFile(path, os.O_RDONLY)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.Data(http.StatusOK, "text/plain", content)
}

// WriteContent writes content to a file
// @Summary 写入文件内容
// @Description 将指定内容写入到文件中，如果文件不存在会自动创建
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param sessionKey header string true "SSH会话密钥" example("tsh_a2e932b625c0d598db3800aa91b92016")
// @Param request body object{cluster=string,path=string,content=string} true "写入内容请求参数" Example({"cluster":"hpc1","path":"/ai/new_folder_rename/test_renamed.sh","content":"#!/bin/bash\\necho Hello World"})
// @Success 200 {object} object{content=string,success=string} "写入成功" example({"content":"","success":"yes"})
// @Failure 400 {object} object{error=string} "请求参数错误"
// @Failure 404 {object} object{error=string} "文件路径不存在"
// @Failure 500 {object} object{error=string} "服务器内部错误或用户未登录"
// @Router /api/v2/files/content/ [post]
func (j *JumpService) WriteFile(c *gin.Context) {
	key, req, err := j.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if _, ok := j.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	sftpClient := j.Clients[key].SftpClient
	path := j.Clients[key].RepackPath(req.Path)
	content := req.Content
	log.Println("sftpClient:", sftpClient, "path:", path, " content:", content)

	file, err := sftpClient.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	defer file.Close()

	_, err = file.Write([]byte(content))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(200, map[string]interface{}{"success": "yes"})
}

// ExecuteFile executes a script file
// @Summary 执行脚本文件
// @Description 执行指定路径的脚本文件，使用bash解释器运行并返回执行结果
// @Tags 文件管理
// @Accept json
// @Produce plain
// @Param sessionKey header string true "SSH会话密钥" example("tsh_a2e932b625c0d598db3800aa91b92016")
// @Param cluster query string true "集群名称" example("hpc1")
// @Param path query string true "脚本文件路径" example("/ai/new_folder_rename/test_renamed.sh")
// @Success 200 {string} string "执行成功，返回脚本输出" example("Hello World\nScript executed successfully")
// @Failure 400 {object} object{error=string} "请求参数错误或路径是目录"
// @Failure 404 {object} object{error=string} "脚本文件不存在"
// @Failure 500 {object} object{error=string} "服务器内部错误、用户未登录或脚本执行失败"
// @Router /api/v2/files/execute/ [post]
func (j *JumpService) ExecuteFile(c *gin.Context) {
	key, req, err := j.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if _, ok := j.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	sftpClient := j.Clients[key].SftpClient
	sshClient := j.Clients[key].SSHClient
	path := j.Clients[key].RepackPath(req.Path)
	log.Println("sftpClient:", sftpClient, "sshClient:", sshClient, "path:", path)

	fileInfo, err := sftpClient.Lstat(path)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	if fileInfo.IsDir() {
		c.JSON(http.StatusBadRequest, errors.New("path is a directory"))
		return
	}

	cmd := fmt.Sprintf("bash %s", path)
	session, err := sshClient.NewSession()
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.Data(http.StatusOK, "text/plain", output)
}

// ChangeMode changes file or directory permissions
// @Summary 修改文件或目录权限
// @Description 修改指定路径文件或目录的权限模式，支持数字权限格式（如755、644等）
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param sessionKey header string true "SSH会话密钥" example("tsh_a2e932b625c0d598db3800aa91b92016")
// @Param request body object{path=string,mode=string} true "修改权限请求参数" Example({"path":"/ai/new_folder_rename/test_renamed.sh","mode":"755"})
// @Success 200 {object} object{success=string} "修改权限成功" example({"success":"yes"})
// @Failure 400 {object} object{error=string} "请求参数错误或权限格式无效"
// @Failure 404 {object} object{error=string} "文件或目录不存在"
// @Failure 403 {object} object{error=string} "没有权限修改该文件"
// @Failure 500 {object} object{error=string} "服务器内部错误或用户未登录"
// @Router /api/v2/files/chmod/ [post]
func (j *JumpService) Chmod(c *gin.Context) {
	key, req, err := j.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if _, ok := j.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	sftpClient := j.Clients[key].SftpClient

	path := j.Clients[key].RepackPath(req.Path)
	modeStr := req.Mode
	log.Println("sftpClient:", sftpClient, "path:", path, " mode:", modeStr)

	mode, err := strconv.ParseUint(modeStr, 8, 32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	err = sftpClient.Chmod(path, os.FileMode(mode))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(200, map[string]interface{}{"success": "yes"})
}

// Chown changes file or directory ownership
// @Summary 修改文件或目录所有者
// @Description 修改指定路径文件或目录的所有者和组，需要提供用户ID和组ID
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param sessionKey header string true "SSH会话密钥" example("tsh_a2e932b625c0d598db3800aa91b92016")
// @Param request body object{path=string,owner=string,group=string} true "修改所有者请求参数" Example({"path":"/ai/new_folder_rename/test_renamed.sh","owner":"1000","group":"1000"})
// @Success 200 {object} object{success=string} "修改所有者成功" example({"success":"yes"})
// @Failure 400 {object} object{error=string} "请求参数错误或用户ID/组ID格式无效"
// @Failure 404 {object} object{error=string} "文件或目录不存在"
// @Failure 403 {object} object{error=string} "没有权限修改该文件所有者"
// @Failure 500 {object} object{error=string} "服务器内部错误或用户未登录"
// @Router /api/v2/files/chown/ [post]
func (j *JumpService) Chown(c *gin.Context) {
	key, req, err := j.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if _, ok := j.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	sftpClient := j.Clients[key].SftpClient

	path := j.Clients[key].RepackPath(req.Path)
	owner := req.Owner
	group := req.Group
	log.Println("sftpClient:", sftpClient, "path:", path, " owner:", owner, " group:", group)
	// Validate owner and group
	if owner == "" || group == "" {
		c.JSON(http.StatusBadRequest, errors.New("owner and group must be provided"))
		return
	}
	// Convert owner and group to int
	ownerID, err := strconv.Atoi(owner)
	if err != nil {
		c.JSON(http.StatusInternalServerError, fmt.Errorf("invalid owner ID: %v", err))
		return
	}
	groupID, err := strconv.Atoi(group)
	if err != nil {
		c.JSON(http.StatusInternalServerError, fmt.Errorf("invalid group ID: %v", err))
		return
	}
	err = sftpClient.Chown(path, ownerID, groupID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(200, map[string]interface{}{"success": "yes"})
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
// @Router /api/v2/files/quota/ [get]
func (j *JumpService) Quota(c *gin.Context) {
	key, _, err := j.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if _, ok := j.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	sshClient := j.Clients[key].SSHClient

	session, err := sshClient.NewSession()
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	defer session.Close()

	cmd := "lfs quota -u " + j.Clients[key].SSHInfo.UserName + " /"
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
