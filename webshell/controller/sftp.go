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
func (s *JumpController) Close() {
	for _, sc := range s.Clients {
		if sc != nil {
			_ = sc.SSHClient.Close()
			_ = sc.SftpClient.Close()
		}
	}
}

// List objects
func (j *JumpController) List(c *gin.Context) {
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

// Transmission data (upload file)
func (j *JumpController) Transmission(c *gin.Context) {
	key, _, err := j.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	}
	if _, ok := j.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	sftpClient := j.Clients[key].SftpClient

	path, _ := c.GetPostForm("path")
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

// Download file and directory by zip
func (j *JumpController) Download(c *gin.Context) {
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

// show attribute of file or directory
func (j *JumpController) Attr(c *gin.Context) {
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
	log.Println("attr path:", path)
	fileInfo, err := sftpClient.Lstat(path)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(200, FileInfoToJSON(fileInfo))
}

// rename file or directory
func (j *JumpController) Rename(c *gin.Context) {
	key, req, err := j.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	}
	if _, ok := j.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	sftpClient := j.Clients[key].SftpClient
	oldPath := req.OldPath
	newPath := req.NewPath

	err = sftpClient.Rename(oldPath, newPath)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(200, map[string]interface{}{"success": "yes"})
}

// new file or directory
func (j *JumpController) New(c *gin.Context) {
	key, req, err := j.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
	}
	if _, ok := j.Clients[key]; !ok {
		c.JSON(http.StatusInternalServerError, errors.New("user not login"))
		return
	}
	jumpClient := j.Clients[key]
	sftpClient := jumpClient.SftpClient

	path := j.Clients[key].RepackPath(req.Path)
	fileType := req.Type
	log.Println("new path:", path, " type:", fileType, " home path:", jumpClient.SSHInfo.HomePath)
	if fileType != "file" && fileType != "dir" {
		c.JSON(http.StatusInternalServerError, errors.New("type must be file or dir"))
		return
	}
	_, err = sftpClient.Lstat(path)
	if err == nil {
		// file exist
		c.JSON(http.StatusInternalServerError, errors.New("file exist"))
		return
	} else {
		if strings.Contains(err.Error(), "file does not exist") {
			if fileType == "file" {
				f, err := sftpClient.Create(path)
				if err != nil {
					log.Println(err)
					c.JSON(http.StatusInternalServerError, err)
					return
				}
				_ = f.Close()
			} else {
				err = sftpClient.Mkdir(path)
				if err != nil {
					log.Println(err)
					c.JSON(http.StatusInternalServerError, err)
					return
				}
			}
			c.JSON(200, map[string]interface{}{"success": "yes"})
			return
		} else {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, err)
			return
		}
	}
}

// Delete file or directory
func (j *JumpController) Delete(c *gin.Context) {
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

// copy file or directory
func (j *JumpController) Copy(c *gin.Context) {
	key, req, err := j.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
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

// Move file or directory
func (j *JumpController) Move(c *gin.Context) {
	key, req, err := j.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
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

// read file content
func (j *JumpController) ReadFile(c *gin.Context) {
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

// WriteFile content to file
func (j *JumpController) WriteFile(c *gin.Context) {
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

// execute file content
func (j *JumpController) ExecuteFile(c *gin.Context) {
	key, req, err := j.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
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

// chmod file or directory
func (j *JumpController) Chmod(c *gin.Context) {
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

// chown file or directory content
func (j *JumpController) Chown(c *gin.Context) {
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

// get quato information use lustre command
func (j *JumpController) Quota(c *gin.Context) {
	key, _, err := j.GetKeyFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
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
