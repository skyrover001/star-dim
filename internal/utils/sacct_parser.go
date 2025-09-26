package utils

import (
	"encoding/csv"
	"fmt"
	"star-dim/internal/models"
	"strconv"
	"strings"
)

// BuildCommand 根据请求参数构建 sacct 命令
func (p *SlurmParser) BuildSacctCommand(req *models.SacctRequest) string {
	var args []string
	args = append(args, "sacct")

	// 添加作业 ID
	if len(req.JobIDs) > 0 {
		args = append(args, "-j", strings.Join(req.JobIDs, ","))
	}

	// 添加用户
	if len(req.Users) > 0 {
		args = append(args, "-u", strings.Join(req.Users, ","))
	}

	// 添加账户
	if len(req.Accounts) > 0 {
		args = append(args, "-A", strings.Join(req.Accounts, ","))
	}

	// 添加分区
	if len(req.Partitions) > 0 {
		args = append(args, "-r", strings.Join(req.Partitions, ","))
	}

	// 添加状态
	if len(req.States) > 0 {
		args = append(args, "-s", strings.Join(req.States, ","))
	}

	// 添加 QOS
	if len(req.QOS) > 0 {
		args = append(args, "-q", strings.Join(req.QOS, ","))
	}

	// 添加集群
	if len(req.Clusters) > 0 {
		args = append(args, "-M", strings.Join(req.Clusters, ","))
	}

	// 添加节点列表
	if len(req.NodeList) > 0 {
		args = append(args, "-N", strings.Join(req.NodeList, ","))
	}

	// 添加作业名称
	if len(req.JobNames) > 0 {
		args = append(args, "--name", strings.Join(req.JobNames, ","))
	}

	// 添加时间范围
	if req.StartTime != "" {
		args = append(args, "-S", req.StartTime)
	}

	if req.EndTime != "" {
		args = append(args, "-E", req.EndTime)
	}

	// 添加节点数量过滤
	if req.MinNodes > 0 || req.MaxNodes > 0 {
		if req.MaxNodes > 0 {
			args = append(args, "-i", fmt.Sprintf("%d-%d", req.MinNodes, req.MaxNodes))
		} else {
			args = append(args, "-i", strconv.Itoa(req.MinNodes))
		}
	}

	// 添加 CPU 数量过滤
	if req.MinCPUs > 0 || req.MaxCPUs > 0 {
		if req.MaxCPUs > 0 {
			args = append(args, "-I", fmt.Sprintf("%d-%d", req.MinCPUs, req.MaxCPUs))
		} else {
			args = append(args, "-I", strconv.Itoa(req.MinCPUs))
		}
	}

	// 添加输出格式
	if req.Format != "" {
		args = append(args, "-o", req.Format)
	} else if req.Brief {
		args = append(args, "-b")
	} else if req.Long {
		args = append(args, "-l")
	} else {
		// 默认格式，包含最常用的字段
		format := "jobid,jobname,partition,account,alloccpus,state,exitcode,start,end,elapsed,reqmem,nodelist,user"
		args = append(args, "-o", format)
	}

	// 添加其他选项
	if req.AllUsers {
		args = append(args, "-a")
	}

	if req.AllClusters {
		args = append(args, "-L")
	}

	if req.Duplicates {
		args = append(args, "-D")
	}

	if req.Truncate {
		args = append(args, "-T")
	}

	if req.ArrayJobs {
		args = append(args, "--array")
	}

	if req.Completion {
		args = append(args, "-c")
	}

	if req.Parsable {
		args = append(args, "-p")
	}

	if req.NoHeader {
		args = append(args, "-n")
	}

	if req.Allocations {
		args = append(args, "-X")
	}

	return strings.Join(args, " ")
}

// ParseOutput 解析 sacct 命令输出
func (p *SlurmParser) ParseSacctOutput(output string, format string) ([]models.JobInfo, error) {
	if strings.TrimSpace(output) == "" {
		return []models.JobInfo{}, nil
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return []models.JobInfo{}, nil
	}

	var jobs []models.JobInfo
	var headers []string

	// 解析表头（如果有的话）
	startIdx := 0
	if !strings.Contains(lines[0], "|") && len(lines) > 1 {
		// 第一行是表头
		headers = strings.Fields(lines[0])
		startIdx = 1
	} else if strings.Contains(lines[0], "|") {
		// 管道分隔的输出，需要根据格式推断表头
		headers = p.inferHeadersFromFormat(format)
	}

	// 解析数据行
	for i := startIdx; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		job, err := p.parseSacctJobLine(line, headers)
		if err != nil {
			// 记录错误但继续处理其他行
			continue
		}

		// 过滤掉空作业记录
		if p.isInvalidJobRecord(job) {
			continue
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

// isInvalidJobRecord 检查是否为空作业记录
func (p *SlurmParser) isInvalidJobRecord(job models.JobInfo) bool {
	// 检查作业ID是否为无效值（如分隔符）
	if job.JobID == "------------" || job.JobID == "" {
		return true
	}

	// 检查作业名是否为无效值
	if job.JobName == "----------" || job.JobName == "" {
		return true
	}

	// 检查状态是否为无效值
	if job.State == "----------" || job.State == "" {
		return true
	}

	// 检查节点列表是否为无效值
	if job.NodeList == "---------------" {
		return true
	}

	// 检查用户是否为无效值
	if job.User == "---------" || job.User == "" {
		return true
	}

	// 检查分区是否为无效值
	if job.Partition == "----------" {
		return true
	}

	// 检查账户是否为无效值
	if job.Account == "----------" {
		return true
	}

	return false
}

// parseJobLine 解析单个作业行
func (p *SlurmParser) parseSacctJobLine(line string, headers []string) (models.JobInfo, error) {
	var job models.JobInfo
	var fields []string

	// 解析字段
	if strings.Contains(line, "|") {
		// 管道分隔
		reader := csv.NewReader(strings.NewReader(line))
		reader.Comma = '|'
		record, err := reader.Read()
		if err != nil {
			return job, err
		}
		fields = record
	} else {
		// 空格分隔
		fields = strings.Fields(line)
	}

	// 映射字段到结构体
	for i, field := range fields {
		if i >= len(headers) {
			break
		}

		header := strings.ToLower(headers[i])
		field = strings.TrimSpace(field)

		switch header {
		case "jobid":
			job.JobID = field
		case "jobidraw":
			job.JobIDRaw = field
		case "jobname":
			job.JobName = field
		case "partition":
			job.Partition = field
		case "account":
			job.Account = field
		case "alloccpus":
			if val, err := strconv.Atoi(field); err == nil {
				job.AllocCPUS = val
			}
		case "state":
			job.State = field
		case "exitcode":
			job.ExitCode = field
		case "start":
			if t, err := p.parseTime(field); err == nil {
				job.Start = t
			}
		case "end":
			if t, err := p.parseTime(field); err == nil {
				job.End = t
			}
		case "submit":
			if t, err := p.parseTime(field); err == nil {
				job.Submit = t
			}
		case "elapsed":
			job.Elapsed = field
		case "reqmem":
			job.ReqMem = field
		case "reqnodes":
			if val, err := strconv.Atoi(field); err == nil {
				job.ReqNodes = val
			}
		case "allocnodes":
			if val, err := strconv.Atoi(field); err == nil {
				job.AllocNodes = val
			}
		case "nodelist":
			job.NodeList = field
		case "user":
			job.User = field
		case "group":
			job.Group = field
		case "qos":
			job.QOS = field
		case "wckey":
			job.WCKey = field
		case "cluster":
			job.Cluster = field
		case "cputime":
			job.CPUTime = field
		case "usercpu":
			job.UserCPU = field
		case "systemcpu":
			job.SystemCPU = field
		case "totalcpu":
			job.TotalCPU = field
		case "maxrss":
			job.MaxRSS = field
		case "maxvmsize":
			job.MaxVMSize = field
		case "maxpages":
			job.MaxPages = field
		case "maxdiskread":
			job.MaxDiskRead = field
		case "maxdiskwrite":
			job.MaxDiskWrite = field
		case "reqtres":
			job.ReqTRES = field
		case "alloctres":
			job.AllocTRES = field
		}
	}

	return job, nil
}
