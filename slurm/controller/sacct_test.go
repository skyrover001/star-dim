package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slurm-jobacct/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	api := router.Group("/api/v1")
	{
		jobacct := api.Group("/jobacct")
		{
			jobacct.GET("/jobs", GetJobs)
			jobacct.GET("/jobs/:jobid", GetJobDetail)
			jobacct.GET("/users/:user/jobs", GetUserJobs)
			jobacct.POST("/accounting", GetAccounting)
		}
	}

	return router
}

func TestGetJobs_MissingParameters(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "缺少 host 参数",
			queryParams:    "username=testuser",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Host and username are required",
		},
		{
			name:           "缺少 username 参数",
			queryParams:    "host=localhost",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Host and username are required",
		},
		{
			name:           "缺少认证信息",
			queryParams:    "host=localhost&username=testuser",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Password or private key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/jobacct/jobs?"+tt.queryParams, nil)
			req.Header.Set("sessionKey", "test-session")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response models.SacctResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.False(t, response.Success)
			assert.Contains(t, response.Message, tt.expectedError)
		})
	}
}

func TestGetJobDetail_MissingJobID(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest("GET", "/api/v1/jobacct/jobs/", nil)
	req.Header.Set("sessionKey", "test-session")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 这应该返回 404，因为路径不匹配
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetUserJobs_ValidUser(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest("GET", "/api/v1/jobacct/users/testuser/jobs?host=localhost&username=testuser", nil)
	req.Header.Set("sessionKey", "test-session")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 由于缺少认证信息，应该返回 400
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.SacctResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Password or private key is required")
}

func TestGetAccounting_InvalidJSON(t *testing.T) {
	router := setupRouter()

	invalidJSON := `{"host": "localhost", "username": "testuser", invalid}`
	req, _ := http.NewRequest("POST", "/api/v1/jobacct/accounting", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("sessionKey", "test-session")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.SacctResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Invalid request format")
}

func TestGetAccounting_ValidRequest(t *testing.T) {
	router := setupRouter()

	requestBody := models.SacctRequest{
		Host:     "localhost",
		Username: "testuser",
		Password: "testpass",
		Users:    []string{"testuser"},
		States:   []string{"COMPLETED"},
	}

	jsonData, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/jobacct/accounting", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("sessionKey", "test-session")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 由于无法连接到实际的 SSH 服务器，应该返回 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response models.SacctResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Failed to connect to SSH server")
}
