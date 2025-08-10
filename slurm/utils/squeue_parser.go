package utils

import (
	"bufio"
	"fmt"
	"slurm-jobacct/models"
	"strconv"
	"strings"
	"time"
)

// SqueueParser 封装 squeue 命令的构建和解析逻辑
type SqueueParser struct{}

// NewSqueueParser 创建新的 SqueueParser 实例
func NewSqueueParser() *SqueueParser {
	return &SqueueParser{}
}

// BuildCommand 根据请求参数构建 squeue 命令
func (p *SqueueParser) BuildCommand(req models.SqueueRequest) string {
	cmd := "squeue"

	// 基本过滤参数
	if len(req.Accounts) > 0 {
		cmd += fmt.Sprintf(" -A %s", strings.Join(req.Accounts, ","))
	}
	if len(req.Jobs) > 0 {
		cmd += fmt.Sprintf(" -j %s", strings.Join(req.Jobs, ","))
	}
	if len(req.Partitions) > 0 {
		cmd += fmt.Sprintf(" -p %s", strings.Join(req.Partitions, ","))
	}
	if len(req.QOS) > 0 {
		cmd += fmt.Sprintf(" -q %s", strings.Join(req.QOS, ","))
	}
	if len(req.States) > 0 {
		cmd += fmt.Sprintf(" -t %s", strings.Join(req.States, ","))
	}
	if len(req.Users) > 0 {
		cmd += fmt.Sprintf(" -u %s", strings.Join(req.Users, ","))
	}
	if len(req.Names) > 0 {
		cmd += fmt.Sprintf(" -n %s", strings.Join(req.Names, ","))
	}
	if len(req.Clusters) > 0 {
		cmd += fmt.Sprintf(" -M %s", strings.Join(req.Clusters, ","))
	}
	if len(req.Licenses) > 0 {
		cmd += fmt.Sprintf(" -L %s", strings.Join(req.Licenses, ","))
	}
	if len(req.NodeList) > 0 {
		cmd += fmt.Sprintf(" -w %s", strings.Join(req.NodeList, ","))
	}
	if len(req.Steps) > 0 {
		cmd += fmt.Sprintf(" -s %s", strings.Join(req.Steps, ","))
	}
	if req.Reservation != "" {
		cmd += fmt.Sprintf(" -R %s", req.Reservation)
	}

	// 输出格式控制
	if req.Format != "" {
		cmd += fmt.Sprintf(" -o %s", req.Format)
	} else if req.FormatLong != "" {
		cmd += fmt.Sprintf(" -O %s", req.FormatLong)
	} else {
		// 默认格式，包含常用字段
		cmd += " -o \"%.10i %.9P %.20j %.8u %.2t %.10M %.10l %.6D %R\""
	}

	if req.NoHeader {
		cmd += " -h"
	}
	if req.Long {
		cmd += " -l"
	}
	if req.NoConvert {
		cmd += " --noconvert"
	}
	if req.Array {
		cmd += " -r"
	}
	if req.Start {
		cmd += " --start"
	}
	if req.Verbose {
		cmd += " -v"
	}
	if req.All {
		cmd += " -a"
	}
	if req.Hide {
		cmd += " --hide"
	}
	if req.Federation {
		cmd += " --federation"
	}
	if req.Local {
		cmd += " --local"
	}
	if req.Sibling {
		cmd += " --sibling"
	}
	if req.OnlyJobState {
		cmd += " --only-job-state"
	}

	// 输出格式
	if req.JSON != "" {
		if req.JSON == "default" {
			cmd += " --json"
		} else {
			cmd += fmt.Sprintf(" --json=%s", req.JSON)
		}
	}
	if req.YAML != "" {
		if req.YAML == "default" {
			cmd += " --yaml"
		} else {
			cmd += fmt.Sprintf(" --yaml=%s", req.YAML)
		}
	}

	// 排序
	if len(req.Sort) > 0 {
		cmd += fmt.Sprintf(" -S %s", strings.Join(req.Sort, ","))
	}

	// 迭代
	if req.Iterate > 0 {
		cmd += fmt.Sprintf(" -i %d", req.Iterate)
	}

	return cmd
}

// ParseOutput 解析 squeue 命令输出
func (p *SqueueParser) ParseOutput(output string, hasHeader bool) ([]models.QueueJobInfo, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return []models.QueueJobInfo{}, nil
	}

	var jobs []models.QueueJobInfo
	startLine := 0

	// 如果有表头，跳过第一行
	if hasHeader && len(lines) > 0 {
		startLine = 1
	}

	for i := startLine; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		job, err := p.parseJobLine(line)
		if err != nil {
			// 记录解析错误但继续处理其他行
			fmt.Printf("Warning: failed to parse line: %s, error: %v\n", line, err)
			continue
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

// parseJobLine 解析单行作业信息
func (p *SqueueParser) parseJobLine(line string) (models.QueueJobInfo, error) {
	// 使用空格分割字段，但要处理可能包含空格的字段
	fields := strings.Fields(line)
	if len(fields) < 8 {
		return models.QueueJobInfo{}, fmt.Errorf("insufficient fields in line: %s", line)
	}

	job := models.QueueJobInfo{}

	// 基本字段解析（按默认格式）
	// %.10i %.9P %.20j %.8u %.2t %.10M %.10l %.6D %R
	job.JobID = fields[0]
	job.Partition = fields[1]
	job.Name = fields[2]
	job.User = fields[3]
	job.State = fields[4]
	job.Time = fields[5]
	job.TimeLeft = fields[6]

	// 节点数解析
	if nodes, err := strconv.Atoi(fields[7]); err == nil {
		job.Nodes = nodes
	}

	// 节点列表或等待原因（最后一个字段可能包含空格）
	if len(fields) > 8 {
		reasonOrNodeList := strings.Join(fields[8:], " ")
		if job.State == "RUNNING" || job.State == "COMPLETING" {
			job.NodeList = reasonOrNodeList
		} else {
			job.Reason = reasonOrNodeList
		}
	}

	return job, nil
}

// ParseDetailedOutput 解析详细输出（-l 选项）
func (p *SqueueParser) ParseDetailedOutput(output string) ([]models.QueueJobInfo, error) {
	jobs := []models.QueueJobInfo{}
	scanner := bufio.NewScanner(strings.NewReader(output))

	var currentJob *models.QueueJobInfo

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// 检查是否是新作业的开始（通常以JobId开头）
		if strings.HasPrefix(line, "JobId=") {
			// 如果有当前作业，保存它
			if currentJob != nil {
				jobs = append(jobs, *currentJob)
			}
			// 创建新作业
			currentJob = &models.QueueJobInfo{}
			p.parseDetailedJobLine(line, currentJob)
		} else if currentJob != nil {
			// 继续解析当前作业的其他字段
			p.parseDetailedJobLine(line, currentJob)
		}
	}

	// 保存最后一个作业
	if currentJob != nil {
		jobs = append(jobs, *currentJob)
	}

	return jobs, scanner.Err()
}

// parseDetailedJobLine 解析详细输出的单行
func (p *SqueueParser) parseDetailedJobLine(line string, job *models.QueueJobInfo) {
	// 解析 key=value 格式的字段
	parts := strings.Split(line, " ")
	for _, part := range parts {
		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) != 2 {
				continue
			}
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])

			switch key {
			case "JobId":
				job.JobID = value
			case "JobName", "Name":
				job.Name = value
			case "UserId":
				// 格式通常是 "username(uid)"
				if idx := strings.Index(value, "("); idx > 0 {
					job.User = value[:idx]
				} else {
					job.User = value
				}
			case "GroupId":
				// 解析组ID
				if idx := strings.Index(value, "("); idx > 0 {
					if gid, err := strconv.Atoi(value[idx+1 : strings.Index(value, ")")]); err == nil {
						job.GroupID = gid
					}
				}
			case "Priority":
				if priority, err := strconv.ParseInt(value, 10, 64); err == nil {
					job.Priority = priority
				}
			case "Nice":
				if nice, err := strconv.Atoi(value); err == nil {
					job.NiceValue = nice
				}
			case "Account":
				job.Account = value
			case "QOS":
				job.QOS = value
			case "JobState":
				job.State = value
			case "Reason":
				job.Reason = value
			case "Dependency":
				job.Dependency = value
			case "Requeue":
				// 处理重新排队信息
			case "Restarts":
				// 处理重启次数
			case "BatchFlag":
				// 处理批处理标志
			case "Reboot":
				// 处理重启信息
			case "ExitCode":
				// 处理退出码
			case "RunTime":
				job.Time = value
			case "TimeLimit":
				job.TimeLeft = value
			case "TimeMin":
				// 处理最小时间
			case "SubmitTime":
				if t, err := time.Parse("2006-01-02T15:04:05", value); err == nil {
					job.SubmitTime = t
				}
			case "EligibleTime":
				// 处理可调度时间
			case "AccrueTime":
				// 处理累计时间
			case "StartTime":
				if t, err := time.Parse("2006-01-02T15:04:05", value); err == nil {
					job.StartTime = t
				}
			case "EndTime":
				if t, err := time.Parse("2006-01-02T15:04:05", value); err == nil {
					job.EndTime = t
				}
			case "Deadline":
				// 处理截止时间
			case "SuspendTime":
				// 处理暂停时间
			case "SecsPreSuspend":
				// 处理暂停前时间
			case "LastSchedEval":
				// 处理最后调度评估时间
			case "Scheduler":
				// 处理调度器信息
			case "Partition":
				job.Partition = value
			case "AllocNode:Sid":
				// 处理分配节点信息
			case "ReqNodeList":
				job.ReqNodes = value
			case "ExcNodeList":
				job.ExcNodes = value
			case "NodeList":
				job.NodeList = value
			case "BatchHost":
				job.BatchHost = value
			case "NumNodes":
				if nodes, err := strconv.Atoi(value); err == nil {
					job.Nodes = nodes
				}
			case "NumCPUs":
				if cpus, err := strconv.Atoi(value); err == nil {
					job.CPUS = cpus
				}
			case "NumTasks":
				// 处理任务数
			case "CPUs/Task":
				// 处理每任务CPU数
			case "ReqB:S:C:T":
				// 处理板:插槽:核心:线程要求
			case "TRES":
				// 处理TRES信息
			case "Socks/Node":
				// 处理每节点Socket数
			case "NtasksPerN:B:S:C":
				// 处理每节点:板:插槽:核心任务数
			case "CoreSpec":
				// 处理核心规格
			case "MinCPUsNode":
				// 处理最小CPU节点
			case "MinMemoryNode", "MinMemoryCPU":
				job.MinMemory = value
			case "MinTmpDiskNode":
				// 处理最小临时磁盘
			case "Features":
				job.Features = value
			case "DelayBoot":
				// 处理延迟启动
			case "OverSubscribe":
				// 处理超额订阅
			case "Contiguous":
				// 处理连续性
			case "Licenses":
				job.Licenses = value
			case "Network":
				job.Network = value
			case "Command":
				job.Command = value
			case "WorkDir":
				job.WorkDir = value
			case "StdErr":
				job.StdErr = value
			case "StdIn":
				// 处理标准输入
			case "StdOut":
				job.StdOut = value
			case "Power":
				// 处理电源管理
			case "MailUser":
				// 处理邮件用户
			case "MailType":
				// 处理邮件类型
			}
		}
	}
}

// ValidateRequest 验证请求参数
func (p *SqueueParser) ValidateRequest(req models.SqueueRequest) error {
	// SSH 连接验证
	if req.Host == "" {
		return fmt.Errorf("host is required")
	}
	if req.Username == "" {
		return fmt.Errorf("username is required")
	}
	if req.Password == "" && req.PrivateKey == "" {
		return fmt.Errorf("either password or private key is required")
	}

	// 端口验证
	if req.Port != 0 && (req.Port < 1 || req.Port > 65535) {
		return fmt.Errorf("invalid port number: %d", req.Port)
	}

	// 迭代间隔验证
	if req.Iterate < 0 {
		return fmt.Errorf("iterate interval cannot be negative")
	}

	// 状态验证
	validStates := map[string]bool{
		"PENDING": true, "RUNNING": true, "SUSPENDED": true, "COMPLETING": true,
		"COMPLETED": true, "CANCELLED": true, "FAILED": true, "TIMEOUT": true,
		"PREEMPTED": true, "NODE_FAIL": true, "REVOKED": true, "SPECIAL_EXIT": true,
		"PD": true, "R": true, "S": true, "CG": true, "CD": true, "CA": true,
		"F": true, "TO": true, "PR": true, "NF": true, "RV": true, "SE": true,
		"all": true,
	}

	for _, state := range req.States {
		if !validStates[strings.ToUpper(state)] {
			return fmt.Errorf("invalid state: %s", state)
		}
	}

	return nil
}
