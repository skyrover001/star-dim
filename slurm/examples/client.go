package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"slurm-jobacct/models"
)

const (
	baseURL       = "http://localhost:8080/api/v1/jobacct"
	sbatchBaseURL = "http://localhost:8080/api/v1/jobsubmit"
)

func main() {
	fmt.Println("=== SLURM API 测试客户端 ===")
	fmt.Println("1. JobAcct API (作业记账查询)")
	fmt.Println("2. JobSubmit API (作业提交)")
	fmt.Println()

	// 测试作业记账 API
	fmt.Println("=== 测试作业记账 API ===")
	testSacctAPIs()

	fmt.Println()

	// 测试作业提交 API
	fmt.Println("=== 测试作业提交 API ===")
	testSbatchAPIs()
}

func testSacctAPIs() {
	// 示例：获取作业列表
	fmt.Println("1. 测试获取作业列表")
	testGetJobs()

	fmt.Println("\n2. 测试获取单个作业详情")
	testGetJobDetail()

	fmt.Println("\n3. 测试获取用户作业")
	testGetUserJobs()

	fmt.Println("\n4. 测试复杂查询")
	testComplexQuery()
}

func testSbatchAPIs() {
	// 示例：快速提交作业
	fmt.Println("1. 测试快速提交作业（使用 --wrap）")
	testQuickSubmit()

	fmt.Println("\n2. 测试完整参数提交作业")
	testFullSubmit()

	fmt.Println("\n3. 测试通过脚本文件提交作业")
	testSubmitWithScript()

	fmt.Println("\n4. 测试批量作业提交")
	testArrayJobSubmit()
}

func testGetJobs() {
	url := fmt.Sprintf("%s/jobs?host=your-slurm-host&username=your-username&password=your-password&allusers=true&states=COMPLETED,FAILED", baseURL)

	resp, err := makeRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("响应: %s\n", resp)
}

func testGetJobDetail() {
	jobID := "12345" // 替换为实际的作业 ID
	url := fmt.Sprintf("%s/jobs/%s?host=your-slurm-host&username=your-username&password=your-password", baseURL, jobID)

	resp, err := makeRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("响应: %s\n", resp)
}

func testGetUserJobs() {
	user := "your-username" // 替换为实际的用户名
	url := fmt.Sprintf("%s/users/%s/jobs?host=your-slurm-host&username=your-username&password=your-password&states=RUNNING,COMPLETED", baseURL, user)

	resp, err := makeRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("响应: %s\n", resp)
}

func testComplexQuery() {
	// 构建复杂查询请求
	request := models.SacctRequest{
		Host:       "your-slurm-host", // 替换为实际的主机
		Username:   "your-username",   // 替换为实际的用户名
		Password:   "your-password",   // 替换为实际的密码
		Users:      []string{"user1", "user2"},
		Accounts:   []string{"default"},
		Partitions: []string{"compute", "gpu"},
		States:     []string{"COMPLETED", "FAILED"},
		StartTime:  "2024-01-01",
		EndTime:    "2024-01-31",
		Format:     "jobid,jobname,state,elapsed,exitcode,user,account",
		AllUsers:   false,
		Parsable:   true,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("JSON 编码错误: %v\n", err)
		return
	}

	url := fmt.Sprintf("%s/accounting", baseURL)
	resp, err := makeRequest("POST", url, jsonData)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("响应: %s\n", resp)
}

func makeRequest(method, url string, body []byte) (string, error) {
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return "", err
	}

	// 添加认证头
	req.Header.Set("sessionKey", "test-session-key")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 格式化 JSON 输出
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, responseBody, "", "  ")
	if err != nil {
		return string(responseBody), nil
	}

	return prettyJSON.String(), nil
}

// === SBATCH API 测试函数 ===

func testQuickSubmit() {
	// 构建快速提交请求
	request := map[string]interface{}{
		"host":          "your-slurm-host",     // 替换为实际的主机
		"username":      "your-username",       // 替换为实际的用户名
		"password":      "your-password",       // 替换为实际的密码
		"wrap":          "echo 'Hello SLURM!'", // 简单的命令
		"job_name":      "test_quick_job",
		"partition":     "compute",
		"time":          "00:10:00", // 10分钟
		"ntasks":        1,
		"cpus_per_task": 1,
		"memory":        "1G",
		"output":        "quick_job_%j.out",
		"error":         "quick_job_%j.err",
	}

	jsonData, _ := json.MarshalIndent(request, "", "  ")
	fmt.Printf("请求数据:\n%s\n\n", jsonData)

	url := fmt.Sprintf("%s/quick-submit", sbatchBaseURL)
	resp, err := makeRequest("POST", url, nil)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("响应: %s\n", resp)
}

func testFullSubmit() {
	// 构建完整参数提交请求
	request := models.SbatchRequest{
		Host:     "your-slurm-host", // 替换为实际的主机
		Username: "your-username",   // 替换为实际的用户名
		Password: "your-password",   // 替换为实际的密码

		// 脚本内容
		Script: `#!/bin/bash
#SBATCH --job-name=full_test_job
#SBATCH --output=full_test_%j.out
#SBATCH --error=full_test_%j.err

echo "作业开始时间: $(date)"
echo "运行节点: $(hostname)"
echo "作业ID: $SLURM_JOB_ID"
echo "分区: $SLURM_JOB_PARTITION"

# 模拟一些工作
for i in {1..10}; do
    echo "处理步骤 $i/10"
    sleep 2
done

echo "作业完成时间: $(date)"
`,

		// 作业参数
		JobName:     "full_test_job",
		Partition:   "compute",
		Time:        "00:15:00", // 15分钟
		NTasks:      2,
		CPUsPerTask: 2,
		Memory:      "4G",
		Output:      "full_test_%j.out",
		Error:       "full_test_%j.err",

		// 邮件通知
		MailType: "ALL",
		MailUser: "your-email@example.com", // 替换为实际邮箱

		// 其他选项
		Account:  "default",
		QOS:      "normal",
		Parsable: true,
		Verbose:  true,
	}

	jsonData, _ := json.Marshal(request)
	url := fmt.Sprintf("%s/submit", sbatchBaseURL)
	resp, err := makeRequest("POST", url, jsonData)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("响应: %s\n", resp)
}

func testSubmitWithScript() {
	fmt.Println("脚本上传提交功能需要使用 multipart/form-data")
	fmt.Println("这个功能可以通过 curl 或其他工具测试")
	fmt.Println("示例 curl 命令:")
	fmt.Println(`curl -X POST http://localhost:8080/api/v1/jobsubmit/submit-script \
  -H "sessionKey: test-session-key" \
  -F "script=@your_script.sh" \
  -F "host=your-slurm-host" \
  -F "username=your-username" \
  -F "password=your-password" \
  -F "parsable=true"`)
}

func testArrayJobSubmit() {
	// 构建数组作业提交请求
	request := models.SbatchRequest{
		Host:     "your-slurm-host", // 替换为实际的主机
		Username: "your-username",   // 替换为实际的用户名
		Password: "your-password",   // 替换为实际的密码

		// 数组作业脚本
		Script: `#!/bin/bash
#SBATCH --job-name=array_test_job
#SBATCH --output=array_test_%A_%a.out
#SBATCH --error=array_test_%A_%a.err

echo "数组作业ID: $SLURM_ARRAY_JOB_ID"
echo "数组任务ID: $SLURM_ARRAY_TASK_ID"
echo "作业ID: $SLURM_JOB_ID"
echo "开始时间: $(date)"

# 根据数组索引执行不同的任务
case $SLURM_ARRAY_TASK_ID in
    1)
        echo "执行任务1: 数据预处理"
        sleep 30
        ;;
    2)
        echo "执行任务2: 模型训练"
        sleep 60
        ;;
    3)
        echo "执行任务3: 结果分析"
        sleep 45
        ;;
    *)
        echo "执行通用任务: $SLURM_ARRAY_TASK_ID"
        sleep 20
        ;;
esac

echo "任务完成时间: $(date)"
`,

		// 数组作业参数
		Array:       "1-3", // 提交3个任务
		JobName:     "array_test_job",
		Partition:   "compute",
		Time:        "00:10:00", // 每个任务最多10分钟
		NTasks:      1,
		CPUsPerTask: 1,
		Memory:      "2G",
		Output:      "array_test_%A_%a.out", // %A=数组作业ID, %a=数组任务ID
		Error:       "array_test_%A_%a.err",

		// 其他选项
		Account:  "default",
		Parsable: true,
	}

	jsonData, _ := json.Marshal(request)
	url := fmt.Sprintf("%s/submit", sbatchBaseURL)
	resp, err := makeRequest("POST", url, jsonData)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("响应: %s\n", resp)
}
