package utils

import (
	"encoding/csv"
	"fmt"
	"slurm-jobacct/models"
	"strconv"
	"strings"
	"time"
)

// SacctParser SLURM sacct 命令解析器
type SacctParser struct{}

// NewSacctParser 创建新的解析器
func NewSacctParser() *SacctParser {
	return &SacctParser{}
}

// BuildCommand 根据请求参数构建 sacct 命令
func (p *SacctParser) BuildCommand(req *models.SacctRequest) string {
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

	return strings.Join(args, " ")
}

// ParseOutput 解析 sacct 命令输出
func (p *SacctParser) ParseOutput(output string, format string) ([]models.JobInfo, error) {
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

		job, err := p.parseJobLine(line, headers)
		if err != nil {
			// 记录错误但继续处理其他行
			continue
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

// inferHeadersFromFormat 从格式字符串推断表头
func (p *SacctParser) inferHeadersFromFormat(format string) []string {
	if format == "" {
		// 默认格式
		return []string{"jobid", "jobname", "partition", "account", "alloccpus", "state", "exitcode", "start", "end", "elapsed", "reqmem", "nodelist", "user"}
	}

	// 解析格式字符串
	return strings.Split(strings.ReplaceAll(format, " ", ""), ",")
}

// parseJobLine 解析单个作业行
func (p *SacctParser) parseJobLine(line string, headers []string) (models.JobInfo, error) {
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

// parseTime 解析时间字符串
func (p *SacctParser) parseTime(timeStr string) (time.Time, error) {
	if timeStr == "" || timeStr == "Unknown" || timeStr == "None" {
		return time.Time{}, fmt.Errorf("invalid time string")
	}

	// 常见的时间格式
	formats := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}

// GetDefaultFormat 获取默认输出格式
func (p *SacctParser) GetDefaultFormat() string {
	return "jobid,jobname,partition,account,alloccpus,state,exitcode,start,end,elapsed,reqmem,nodelist,user"
}

// GetDetailedFormat 获取详细输出格式
func (p *SacctParser) GetDetailedFormat() string {
	return "jobid,jobname,partition,account,alloccpus,state,exitcode,start,end,elapsed,reqmem,reqnodes,allocnodes,nodelist,user,group,qos,submit,cputime,totalcpu,maxrss,maxvmsize"
}
