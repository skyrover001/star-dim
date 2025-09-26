package utils

import (
	"fmt"
	"regexp"
	"star-dim/internal/models"
	"strings"
	"time"
)

// SacctParser SLURM sacct 命令解析器
type SlurmParser struct {
	HomePath string
}

// NewSacctParser 创建新的解析器
func NewSlurmParser(slurmUser *models.User) *SlurmParser {
	if slurmUser == nil {
		return &SlurmParser{
			HomePath: "",
		}
	}
	return &SlurmParser{
		HomePath: slurmUser.HomePath,
	}
}

// inferHeadersFromFormat 从格式字符串推断表头
func (p *SlurmParser) inferHeadersFromFormat(format string) []string {
	if format == "" {
		// 默认格式
		return []string{"jobid", "jobname", "partition", "account", "alloccpus", "state", "exitcode", "start", "end", "elapsed", "reqmem", "nodelist", "user"}
	}

	// 解析格式字符串
	return strings.Split(strings.ReplaceAll(format, " ", ""), ",")
}

// parseTime 解析时间字符串
func (p *SlurmParser) parseTime(timeStr string) (time.Time, error) {
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
func (p *SlurmParser) GetDefaultFormat() string {
	return "jobid,jobname,partition,account,alloccpus,state,exitcode,start,end,elapsed,reqmem,nodelist,user"
}

// GetDetailedFormat 获取详细输出格式
func (p *SlurmParser) GetDetailedFormat() string {
	return "jobid,jobname,partition,account,alloccpus,state,exitcode,start,end,elapsed,reqmem,reqnodes,allocnodes,nodelist,user,group,qos,submit,cputime,totalcpu,maxrss,maxvmsize"
}

// isValidTimeFormat 检查时间格式是否有效
func (p *SlurmParser) isValidTimeFormat(timeStr string) bool {
	// 支持的格式：
	// - minutes
	// - minutes:seconds
	// - hours:minutes:seconds
	// - days-hours
	// - days-hours:minutes
	// - days-hours:minutes:seconds
	patterns := []string{
		`^\d+$`,             // minutes
		`^\d+:\d+$`,         // minutes:seconds
		`^\d+:\d+:\d+$`,     // hours:minutes:seconds
		`^\d+-\d+$`,         // days-hours
		`^\d+-\d+:\d+$`,     // days-hours:minutes
		`^\d+-\d+:\d+:\d+$`, // days-hours:minutes:seconds
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, timeStr); matched {
			return true
		}
	}

	return false
}

// isValidNodeFormat 检查节点格式是否有效
func (p *SlurmParser) isValidNodeFormat(nodeStr string) bool {
	// 支持的格式：
	// - N (单个数字)
	// - N-M (范围)
	patterns := []string{
		`^\d+$`,     // N
		`^\d+-\d+$`, // N-M
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, nodeStr); matched {
			return true
		}
	}

	return false
}
