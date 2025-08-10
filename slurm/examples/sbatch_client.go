package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"slurm-jobacct/models"
)

const (
	sbatchBaseURL = "http://localhost:8080/api/v1/jobsubmit"
)

func mainSbatch() {
	fmt.Println("=== SLURM JobSubmit API 测试客户端 ===\n")

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
	resp, err := makeSbatchJSONRequest("POST", url, request)
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

	jsonData, _ := json.MarshalIndent(request, "", "  ")
	fmt.Printf("请求数据:\n%s\n\n", jsonData)

	url := fmt.Sprintf("%s/submit", sbatchBaseURL)
	resp, err := makeSbatchJSONRequest("POST", url, request)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("响应: %s\n", resp)
}

func testSubmitWithScript() {
	// 创建脚本内容
	scriptContent := `#!/bin/bash
#SBATCH --job-name=script_upload_test
#SBATCH --partition=compute
#SBATCH --time=00:05:00
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=1
#SBATCH --mem=1G
#SBATCH --output=script_test_%j.out

echo "这是通过脚本上传提交的作业"
echo "当前时间: $(date)"
echo "工作目录: $(pwd)"
echo "环境变量:"
env | grep SLURM | head -10

# 简单的计算任务
echo "执行一些计算..."
python3 -c "
import time
import math
for i in range(1000000):
    result = math.sqrt(i)
    if i % 100000 == 0:
        print(f'处理进度: {i/10000:.1f}%')
        time.sleep(0.1)
print('计算完成!')
"

echo "作业执行完成"
`

	// 使用 multipart/form-data 上传
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加脚本文件
	part, err := writer.CreateFormFile("script", "test_script.sh")
	if err != nil {
		fmt.Printf("创建文件字段失败: %v\n", err)
		return
	}
	part.Write([]byte(scriptContent))

	// 添加其他表单字段
	writer.WriteField("host", "your-slurm-host")   // 替换为实际的主机
	writer.WriteField("username", "your-username") // 替换为实际的用户名
	writer.WriteField("password", "your-password") // 替换为实际的密码
	writer.WriteField("parsable", "true")
	writer.WriteField("wait", "false")

	writer.Close()

	// 发送请求
	url := fmt.Sprintf("%s/submit-script", sbatchBaseURL)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("sessionKey", "test-session-key")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("发送请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return
	}

	// 格式化输出
	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, responseBody, "", "  ")
	fmt.Printf("响应: %s\n", prettyJSON.String())
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

	jsonData, _ := json.MarshalIndent(request, "", "  ")
	fmt.Printf("请求数据:\n%s\n\n", jsonData)

	url := fmt.Sprintf("%s/submit", sbatchBaseURL)
	resp, err := makeSbatchJSONRequest("POST", url, request)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("响应: %s\n", resp)
}

func makeSbatchJSONRequest(method, url string, data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("sessionKey", "test-session-key")

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
