package utils

import (
	"bufio"
	"fmt"
	"log"
	"star-dim/internal/models"
	"strconv"
	"strings"
)

// BuildCommand 根据请求参数构建 sinfo 命令
func (p *SlurmParser) BuildSinfoCommand(req models.SinfoRequest) string {
	cmd := "sinfo"
	log.Println("req:", req)
	// 基本过滤参数
	if len(req.Nodes) > 0 {
		cmd += fmt.Sprintf(" -n %s", strings.Join(req.Nodes, ","))
	}
	if len(req.Partitions) > 0 {
		cmd += fmt.Sprintf(" -p %s", strings.Join(req.Partitions, ","))
	}
	if len(req.States) > 0 {
		cmd += fmt.Sprintf(" -t %s", strings.Join(req.States, ","))
	}
	if len(req.Clusters) > 0 {
		cmd += fmt.Sprintf(" -M %s", strings.Join(req.Clusters, ","))
	}

	// 显示选项
	if req.All {
		cmd += " -a"
	}
	if req.Dead {
		cmd += " -d"
	}
	if req.Exact {
		cmd += " -e"
	}
	if req.Future {
		cmd += " -F"
	}
	if req.Hide {
		cmd += " --hide"
	}
	if req.Long {
		cmd += " -l"
	}
	if req.NodeCentric {
		cmd += " -N"
	}
	if req.Responding {
		cmd += " -r"
	}
	if req.ListReasons {
		cmd += " -R"
	}
	if req.Summarize {
		cmd += " -s"
	}
	if req.Reservation {
		cmd += " -T"
	}
	if req.Verbose {
		cmd += " -v"
	}
	if req.Federation {
		cmd += " --federation"
	}
	if req.Local {
		cmd += " --local"
	}
	if req.NoConvert {
		cmd += " --noconvert"
	}
	if req.NoHeader {
		cmd += " -h"
	}

	// 输出格式控制
	if req.Format != "" {
		cmd += fmt.Sprintf(" -o %s", req.Format)
	} else if req.FormatLong != "" {
		cmd += fmt.Sprintf(" -O %s", req.FormatLong)
	} else if req.NodeCentric {
		// 节点中心格式的默认字段
		cmd += " -o \"%.15N %.9P %.6t %.4c %.8z %.6m %.8d %.6w %.8f %.20E\""
	} else if req.Summarize {
		// 摘要模式的默认字段
		cmd += " -o \"%.9P %.5a %.6t %.4c %.8z %.6m %.8d %.6w %.8f %.20E\""
	} else {
		// 默认格式
		cmd += " -o \"%.15N %.9P %.6t %.4c %.8z %.6m %.8d %.6w %.8f %.20E\""
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

	log.Println("sinfo command:", cmd)
	return cmd
}

// ParseOutput 解析 sinfo 命令输出
func (p *SlurmParser) ParseSinfoOutput(output string, req models.SinfoRequest) (interface{}, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return []models.NodeInfo{}, nil
	}

	if req.Summarize {
		return p.parseSummaryOutput(output, !req.NoHeader)
	} else if req.Reservation {
		return p.parseReservationOutput(output, !req.NoHeader)
	} else {
		return p.parseNodeOutput(output, !req.NoHeader, req.NodeCentric)
	}
}

// parseNodeOutput 解析节点信息输出
func (p *SlurmParser) parseNodeOutput(output string, hasHeader bool, nodeCentric bool) ([]models.NodeInfo, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return []models.NodeInfo{}, nil
	}

	var nodes []models.NodeInfo
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

		node, err := p.parseNodeLine(line, nodeCentric)
		if err != nil {
			// 记录解析错误但继续处理其他行
			fmt.Printf("Warning: failed to parse line: %s, error: %v\n", line, err)
			continue
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// parseNodeLine 解析单行节点信息
func (p *SlurmParser) parseNodeLine(line string, nodeCentric bool) (models.NodeInfo, error) {
	// 使用空格分割字段
	fields := strings.Fields(line)
	if len(fields) < 8 {
		return models.NodeInfo{}, fmt.Errorf("insufficient fields in line: %s", line)
	}

	node := models.NodeInfo{}

	// 基本字段解析（按默认格式）
	// %.15N %.9P %.6t %.4c %.8z %.6m %.8d %.6w %.8f %.20E
	node.NodeName = fields[0]
	node.Partition = fields[1]
	node.State = fields[2]

	// CPU数解析
	if cpus, err := strconv.Atoi(fields[3]); err == nil {
		node.CPUs = cpus
	}

	// 内存解析 (字段4: z - 内存大小)
	if mem := p.parseMemory(fields[5]); mem > 0 {
		node.Memory = mem
	}

	// 临时磁盘解析 (字段5: d - 临时磁盘)
	if tmpDisk := p.parseMemory(fields[6]); tmpDisk > 0 {
		node.TmpDisk = tmpDisk
	}

	// 权重解析 (字段6: w - 权重)
	if weight, err := strconv.Atoi(fields[7]); err == nil {
		node.Weight = weight
	}

	// 特性 (字段7: f - 特性)
	if len(fields) > 8 {
		node.Features = fields[8]
	}

	// 原因 (字段8: E - 原因)
	if len(fields) > 9 {
		node.Reason = strings.Join(fields[9:], " ")
	}

	return node, nil
}

// parseSummaryOutput 解析摘要输出
func (p *SlurmParser) parseSummaryOutput(output string, hasHeader bool) ([]models.NodeSummary, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return []models.NodeSummary{}, nil
	}

	var summaries []models.NodeSummary
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

		summary, err := p.parseSummaryLine(line)
		if err != nil {
			fmt.Printf("Warning: failed to parse summary line: %s, error: %v\n", line, err)
			continue
		}

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// parseSummaryLine 解析单行摘要信息
func (p *SlurmParser) parseSummaryLine(line string) (models.NodeSummary, error) {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return models.NodeSummary{}, fmt.Errorf("insufficient fields in summary line: %s", line)
	}

	summary := models.NodeSummary{}
	summary.State = fields[0]

	if count, err := strconv.Atoi(fields[1]); err == nil {
		summary.Count = count
	}

	if cpus, err := strconv.Atoi(fields[2]); err == nil {
		summary.CPUs = cpus
	}

	return summary, nil
}

// parseReservationOutput 解析预留信息输出
func (p *SlurmParser) parseReservationOutput(output string, hasHeader bool) ([]models.ReservationInfo, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return []models.ReservationInfo{}, nil
	}

	var reservations []models.ReservationInfo
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

		reservation, err := p.parseReservationLine(line)
		if err != nil {
			fmt.Printf("Warning: failed to parse reservation line: %s, error: %v\n", line, err)
			continue
		}

		reservations = append(reservations, reservation)
	}

	return reservations, nil
}

// parseReservationLine 解析单行预留信息
func (p *SlurmParser) parseReservationLine(line string) (models.ReservationInfo, error) {
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return models.ReservationInfo{}, fmt.Errorf("insufficient fields in reservation line: %s", line)
	}

	reservation := models.ReservationInfo{}
	reservation.ReservationName = fields[0]
	reservation.State = fields[1]
	reservation.StartTime = fields[2]
	reservation.EndTime = fields[3]
	reservation.NodeList = fields[4]

	return reservation, nil
}

// ParseDetailedOutput 解析详细输出（-l 选项）
func (p *SlurmParser) ParseDetailedOutput(output string) ([]models.NodeInfo, error) {
	nodes := []models.NodeInfo{}
	scanner := bufio.NewScanner(strings.NewReader(output))

	var currentNode *models.NodeInfo

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// 检查是否是新节点的开始（通常以NodeName开头）
		if strings.HasPrefix(line, "NodeName=") {
			// 如果有当前节点，保存它
			if currentNode != nil {
				nodes = append(nodes, *currentNode)
			}
			// 创建新节点
			currentNode = &models.NodeInfo{}
			p.parseDetailedNodeLine(line, currentNode)
		} else if currentNode != nil {
			// 继续解析当前节点的其他字段
			p.parseDetailedNodeLine(line, currentNode)
		}
	}

	// 保存最后一个节点
	if currentNode != nil {
		nodes = append(nodes, *currentNode)
	}

	return nodes, scanner.Err()
}

// parseDetailedNodeLine 解析详细输出的单行
func (p *SlurmParser) parseDetailedNodeLine(line string, node *models.NodeInfo) {
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
			case "NodeName":
				node.NodeName = value
			case "Arch", "Architecture":
				node.Architecture = value
			case "CoresPerSocket":
				if cores, err := strconv.Atoi(value); err == nil {
					node.Cores = cores
				}
			case "CPUAlloc":
				if cpus, err := strconv.Atoi(value); err == nil {
					node.AllocCPUs = cpus
				}
			case "CPUIdle":
				if cpus, err := strconv.Atoi(value); err == nil {
					node.IdleCPUs = cpus
				}
			case "CPUOther":
				if cpus, err := strconv.Atoi(value); err == nil {
					node.OtherCPUs = cpus
				}
			case "CPUTotal":
				if cpus, err := strconv.Atoi(value); err == nil {
					node.CPUs = cpus
				}
			case "CPULoad":
				node.Load = value
			case "AvailableFeatures", "ActiveFeatures":
				if node.Features == "" {
					node.Features = value
				} else {
					node.Features += "," + value
				}
			case "Gres":
				node.Gres = value
			case "NodeAddr", "NodeHostName":
				// 处理节点地址信息
			case "OS":
				node.OS = value
			case "RealMemory":
				if mem, err := strconv.Atoi(value); err == nil {
					node.Memory = mem
				}
			case "AllocMem":
				if mem, err := strconv.Atoi(value); err == nil {
					node.AllocMemory = mem
				}
			case "FreeMem":
				if mem, err := strconv.Atoi(value); err == nil {
					node.FreeMem = mem
				}
			case "Sockets":
				if sockets, err := strconv.Atoi(value); err == nil {
					node.Sockets = sockets
				}
			case "Boards":
				// 处理板信息
			case "State":
				node.State = value
			case "ThreadsPerCore":
				if threads, err := strconv.Atoi(value); err == nil {
					node.Threads = threads
				}
			case "TmpDisk":
				if tmpDisk, err := strconv.Atoi(value); err == nil {
					node.TmpDisk = tmpDisk
				}
			case "Weight":
				if weight, err := strconv.Atoi(value); err == nil {
					node.Weight = weight
				}
			case "Owner":
				node.User = value
			case "MCS_label":
				// 处理MCS标签
			case "Partitions":
				node.Partition = value
			case "BootTime":
				node.BootTime = value
			case "SlurmdStartTime":
				node.SlurmdStartTime = value
			case "LastBusyTime":
				// 处理最后忙碌时间
			case "CfgTRES":
				// 处理配置的TRES
			case "AllocTRES":
				// 处理分配的TRES
			case "CapWatts":
				// 处理功耗上限
			case "CurrentWatts":
				// 处理当前功耗
			case "AveWatts":
				// 处理平均功耗
			case "ExtSensorsJoules":
				// 处理外部传感器焦耳
			case "ExtSensorsWatts":
				// 处理外部传感器瓦特
			case "ExtSensorsTemp":
				// 处理外部传感器温度
			case "Reason":
				node.Reason = value
			case "ReasonTime":
				node.Timestamp = value
			case "ReasonUid":
				// 处理原因用户ID
			}
		}
	}
}

// parseMemory 解析内存字符串，返回MB值
func (p *SlurmParser) parseMemory(memStr string) int {
	if memStr == "" || memStr == "N/A" {
		return 0
	}

	// 移除可能的单位后缀并转换
	memStr = strings.ToUpper(memStr)
	multiplier := 1

	if strings.HasSuffix(memStr, "K") {
		multiplier = 1
		memStr = strings.TrimSuffix(memStr, "K")
	} else if strings.HasSuffix(memStr, "M") {
		multiplier = 1
		memStr = strings.TrimSuffix(memStr, "M")
	} else if strings.HasSuffix(memStr, "G") {
		multiplier = 1024
		memStr = strings.TrimSuffix(memStr, "G")
	} else if strings.HasSuffix(memStr, "T") {
		multiplier = 1024 * 1024
		memStr = strings.TrimSuffix(memStr, "T")
	}

	if value, err := strconv.Atoi(memStr); err == nil {
		return value * multiplier
	}

	return 0
}

// ValidateRequest 验证请求参数
func (p *SlurmParser) ValidateSinfoRequest(req models.SinfoRequest) error {
	// 迭代间隔验证
	if req.Iterate < 0 {
		return fmt.Errorf("iterate interval cannot be negative")
	}

	// 状态验证
	validStates := map[string]bool{
		"ALLOC": true, "ALLOCATED": true, "COMP": true, "COMPLETING": true,
		"DOWN": true, "DRAIN": true, "DRAINING": true, "FAIL": true,
		"FAILING": true, "FUTURE": true, "IDLE": true, "MAINT": true,
		"MIX": true, "MIXED": true, "NO_RESPOND": true, "NPC": true,
		"PERFCTRS": true, "POWER_DOWN": true, "POWER_UP": true, "RESV": true,
		"RESERVED": true, "UNK": true, "UNKNOWN": true,
	}

	for _, state := range req.States {
		if !validStates[strings.ToUpper(state)] {
			return fmt.Errorf("invalid state: %s", state)
		}
	}

	return nil
}
