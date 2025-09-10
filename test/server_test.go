package test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"webshell/controller"
)

// 测试辅助函数
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// 模拟路由设置
	api := router.Group("/api/v2")
	jumpService := &controller.JumpService{} // 假设这是您的服务结构

	api.GET("/files/", jumpService.ShowDir)
	api.POST("/files/create/", jumpService.CreateFile)
	api.POST("/files/delete/", jumpService.DeleteFile)
	api.POST("/files/rename/", jumpService.RenameFile)
	api.GET("/files/download/", jumpService.DownloadFile)
	api.POST("/files/execute/", jumpService.ExecuteFile)
	api.POST("/files/upload/", jumpService.Transmission)

	return router
}

// 创建测试文件的辅助函数
func createTestFile(t *testing.T, filename, content string) string {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	assert.NoError(t, err)
	return filePath
}

// 创建测试目录的辅助函数
func createTestDir(t *testing.T, dirName string) string {
	tmpDir := t.TempDir()
	dirPath := filepath.Join(tmpDir, dirName)
	err := os.MkdirAll(dirPath, 0755)
	assert.NoError(t, err)
	return dirPath
}

// TestShowDir 测试显示目录接口
func TestShowDir(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name           string
		sessionKey     string
		cluster        string
		path           string
		expectedStatus int
	}{
		{
			name:           "正常请求",
			sessionKey:     "valid_session_key",
			cluster:        "hpc1",
			path:           "/tmp",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "缺少sessionKey",
			sessionKey:     "",
			cluster:        "hpc1",
			path:           "/tmp",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "无效路径",
			sessionKey:     "valid_session_key",
			cluster:        "hpc1",
			path:           "/nonexistent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v2/files/?cluster="+tt.cluster+"&path="+tt.path, nil)
			if tt.sessionKey != "" {
				req.Header.Set("sessionKey", tt.sessionKey)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestCreateFile 测试创建文件接口
func TestCreateFile(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name           string
		sessionKey     string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name:       "创建文件成功",
			sessionKey: "valid_session_key",
			requestBody: map[string]interface{}{
				"cluster": "hpc1",
				"path":    "/tmp/test_file.txt",
				"type":    "file",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "创建目录成功",
			sessionKey: "valid_session_key",
			requestBody: map[string]interface{}{
				"cluster": "hpc1",
				"path":    "/tmp/test_dir",
				"type":    "directory",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "缺少必要参数",
			sessionKey: "valid_session_key",
			requestBody: map[string]interface{}{
				"cluster": "hpc1",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/v2/files/create/", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			if tt.sessionKey != "" {
				req.Header.Set("sessionKey", tt.sessionKey)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestDeleteFile 测试删除文件接口
func TestDeleteFile(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name           string
		sessionKey     string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name:       "删除文件成功",
			sessionKey: "valid_session_key",
			requestBody: map[string]interface{}{
				"cluster": "hpc1",
				"path":    "/tmp/test_file.txt",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "删除不存在的文件",
			sessionKey: "valid_session_key",
			requestBody: map[string]interface{}{
				"cluster": "hpc1",
				"path":    "/tmp/nonexistent.txt",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "缺少sessionKey",
			sessionKey: "",
			requestBody: map[string]interface{}{
				"cluster": "hpc1",
				"path":    "/tmp/test_file.txt",
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/v2/files/delete/", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			if tt.sessionKey != "" {
				req.Header.Set("sessionKey", tt.sessionKey)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestRenameFile 测试重命名文件接口
func TestRenameFile(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name           string
		sessionKey     string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name:       "重命名文件成功",
			sessionKey: "valid_session_key",
			requestBody: map[string]interface{}{
				"cluster": "hpc1",
				"path":    "/tmp/old_name.txt",
				"name":    "new_name.txt",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "重命名不存在的文件",
			sessionKey: "valid_session_key",
			requestBody: map[string]interface{}{
				"cluster": "hpc1",
				"path":    "/tmp/nonexistent.txt",
				"name":    "new_name.txt",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "缺少新名称",
			sessionKey: "valid_session_key",
			requestBody: map[string]interface{}{
				"cluster": "hpc1",
				"path":    "/tmp/old_name.txt",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/v2/files/rename/", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			if tt.sessionKey != "" {
				req.Header.Set("sessionKey", tt.sessionKey)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestDownloadFile 测试下载文件接口
func TestDownloadFile(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name           string
		sessionKey     string
		cluster        string
		path           string
		expectedStatus int
	}{
		{
			name:           "下载文件成功",
			sessionKey:     "valid_session_key",
			cluster:        "hpc1",
			path:           "/tmp/test.txt",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "下载不存在的文件",
			sessionKey:     "valid_session_key",
			cluster:        "hpc1",
			path:           "/tmp/nonexistent.txt",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "下载目录（应该失败）",
			sessionKey:     "valid_session_key",
			cluster:        "hpc1",
			path:           "/tmp",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v2/files/download/?cluster="+tt.cluster+"&path="+tt.path, nil)
			if tt.sessionKey != "" {
				req.Header.Set("sessionKey", tt.sessionKey)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestExecuteFile 测试执行脚本文件接口
func TestExecuteFile(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name           string
		sessionKey     string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name:       "执行脚本成功",
			sessionKey: "valid_session_key",
			requestBody: map[string]interface{}{
				"cluster": "hpc1",
				"path":    "/tmp/test_script.sh",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "执行不存在的脚本",
			sessionKey: "valid_session_key",
			requestBody: map[string]interface{}{
				"cluster": "hpc1",
				"path":    "/tmp/nonexistent.sh",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "执行目录（应该失败）",
			sessionKey: "valid_session_key",
			requestBody: map[string]interface{}{
				"cluster": "hpc1",
				"path":    "/tmp",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/v2/files/execute/", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			if tt.sessionKey != "" {
				req.Header.Set("sessionKey", tt.sessionKey)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestTransmission 测试文件上传接口
func TestTransmission(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name           string
		sessionKey     string
		setupRequest   func() (*http.Request, error)
		expectedStatus int
	}{
		{
			name:       "上传文件成功",
			sessionKey: "valid_session_key",
			setupRequest: func() (*http.Request, error) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				// 添加表单字段
				writer.WriteField("cluster", "hpc1")
				writer.WriteField("path", "/tmp/upload_test.txt")
				writer.WriteField("update", "false")

				// 添加文件
				part, err := writer.CreateFormFile("file", "test.txt")
				if err != nil {
					return nil, err
				}
				part.Write([]byte("test file content"))
				writer.Close()

				req, err := http.NewRequest("POST", "/api/v2/files/upload/", body)
				if err != nil {
					return nil, err
				}
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "上传文件覆盖模式",
			sessionKey: "valid_session_key",
			setupRequest: func() (*http.Request, error) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				writer.WriteField("cluster", "hpc1")
				writer.WriteField("path", "/tmp/upload_test.txt")
				writer.WriteField("update", "true")
				writer.WriteField("offset", "0")

				part, err := writer.CreateFormFile("file", "test.txt")
				if err != nil {
					return nil, err
				}
				part.Write([]byte("updated file content"))
				writer.Close()

				req, err := http.NewRequest("POST", "/api/v2/files/upload/", body)
				if err != nil {
					return nil, err
				}
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "上传空文件",
			sessionKey: "valid_session_key",
			setupRequest: func() (*http.Request, error) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				writer.WriteField("cluster", "hpc1")
				writer.WriteField("path", "/tmp/empty.txt")
				writer.WriteField("update", "false")
				writer.Close()

				req, err := http.NewRequest("POST", "/api/v2/files/upload/", body)
				if err != nil {
					return nil, err
				}
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req, nil
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := tt.setupRequest()
			assert.NoError(t, err)

			if tt.sessionKey != "" {
				req.Header.Set("sessionKey", tt.sessionKey)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestTransmissionWithOffset 测试断点续传
func TestTransmissionWithOffset(t *testing.T) {
	router := setupRouter()

	t.Run("断点续传", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		writer.WriteField("cluster", "hpc1")
		writer.WriteField("path", "/tmp/large_file.txt")
		writer.WriteField("update", "true")
		writer.WriteField("offset", "1024") // 从1KB开始续传

		part, err := writer.CreateFormFile("file", "large_file.txt")
		assert.NoError(t, err)
		part.Write([]byte("continuation of file content"))
		writer.Close()

		req, err := http.NewRequest("POST", "/api/v2/files/upload/", body)
		assert.NoError(t, err)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("sessionKey", "valid_session_key")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestMiddleware 测试中间件功能
func TestMiddleware(t *testing.T) {
	router := setupRouter()

	t.Run("CORS中间件", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", "/api/v2/files/", nil)
		req.Header.Set("Origin", "http://localhost:3000")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Contains(t, w.Header().Get("Access-Control-Allow-Origin"), "*")
	})
}

// BenchmarkShowDir 基准测试
func BenchmarkShowDir(b *testing.B) {
	router := setupRouter()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/v2/files/?cluster=hpc1&path=/tmp", nil)
		req.Header.Set("sessionKey", "valid_session_key")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkTransmission 文件上传基准测试
func BenchmarkTransmission(b *testing.B) {
	router := setupRouter()

	for i := 0; i < b.N; i++ {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		writer.WriteField("cluster", "hpc1")
		writer.WriteField("path", "/tmp/benchmark.txt")
		writer.WriteField("update", "false")

		part, _ := writer.CreateFormFile("file", "benchmark.txt")
		part.Write([]byte("benchmark test content"))
		writer.Close()

		req, _ := http.NewRequest("POST", "/api/v2/files/upload/", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("sessionKey", "valid_session_key")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
