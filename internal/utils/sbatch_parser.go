package utils

import (
	"fmt"
	"log"
	"regexp"
	"star-dim/internal/models"
	"strconv"
	"strings"
)

// BuildCommand 根据请求参数构建 sbatch 命令
func (p *SlurmParser) BuildSbatchCommand(req *models.SbatchRequest) (string, error) {
	var args []string
	log.Println("req:", req)
	args = append(args, "sbatch")

	// 并行运行选项
	if req.Array != "" {
		args = append(args, "-a", req.Array)
	}

	if req.Account != "" {
		args = append(args, "-A", req.Account)
	}

	if req.Begin != "" {
		args = append(args, "-b", req.Begin)
	}

	if req.Comment != "" {
		args = append(args, "--comment", req.Comment)
	}

	if req.CPUFreq != "" {
		args = append(args, "--cpu-freq", req.CPUFreq)
	}

	if req.CPUsPerTask > 0 {
		args = append(args, "-c", strconv.Itoa(req.CPUsPerTask))
	}

	if req.Dependency != "" {
		args = append(args, "-d", req.Dependency)
	}

	if req.Deadline != "" {
		args = append(args, "--deadline", req.Deadline)
	}

	if req.DelayBoot > 0 {
		args = append(args, "--delay-boot", strconv.Itoa(req.DelayBoot))
	}

	if req.Chdir != "" {
		args = append(args, "-D", req.Chdir)
	}

	if req.Error != "" {
		args = append(args, "-e", req.Error)
	}

	if req.Export != "" {
		args = append(args, "--export", req.Export)
	}

	if req.ExportFile != "" {
		args = append(args, "--export-file", req.ExportFile)
	}

	if req.GetUserEnv {
		args = append(args, "--get-user-env")
	}

	if req.GID != "" {
		args = append(args, "--gid", req.GID)
	}

	if req.GRES != "" {
		args = append(args, "--gres", req.GRES)
	}

	if req.GRESFlags != "" {
		args = append(args, "--gres-flags", req.GRESFlags)
	}

	if req.Hold {
		args = append(args, "-H")
	}

	if req.IgnorePBS {
		args = append(args, "--ignore-pbs")
	}

	if req.Input != "" {
		args = append(args, "-i", req.Input)
	}

	if req.JobName != "" {
		args = append(args, "-J", req.JobName)
	}

	if req.NoKill {
		args = append(args, "-k")
	}

	if req.Licenses != "" {
		args = append(args, "-L", req.Licenses)
	}

	if len(req.Clusters) > 0 {
		args = append(args, "-M", strings.Join(req.Clusters, ","))
	}

	if req.Container != "" {
		args = append(args, "--container", req.Container)
	}

	if req.ContainerID != "" {
		args = append(args, "--container-id", req.ContainerID)
	}

	if req.Distribution != "" {
		args = append(args, "-m", req.Distribution)
	}

	if req.MailType != "" {
		args = append(args, "--mail-type", req.MailType)
	}

	if req.MailUser != "" {
		args = append(args, "--mail-user", req.MailUser)
	}

	if req.MCSLabel != "" {
		args = append(args, "--mcs-label", req.MCSLabel)
	}

	if req.NTasks > 0 {
		args = append(args, "-n", strconv.Itoa(req.NTasks))
	}

	if req.Nice > 0 {
		args = append(args, "--nice", strconv.Itoa(req.Nice))
	}

	if req.NoRequeue {
		args = append(args, "--no-requeue")
	}

	if req.NTasksPerNode > 0 {
		args = append(args, "--ntasks-per-node", strconv.Itoa(req.NTasksPerNode))
	}

	if req.Nodes != "" {
		args = append(args, "-N", req.Nodes)
	}

	if req.Output != "" {
		args = append(args, "-o", req.Output)
	}

	if req.Overcommit {
		args = append(args, "-O")
	}

	if req.Partition != "" {
		args = append(args, "-p", req.Partition)
	}

	if req.Parsable {
		args = append(args, "--parsable")
	}

	if req.Power != "" {
		args = append(args, "--power", req.Power)
	}

	if req.Priority > 0 {
		args = append(args, "--priority", strconv.Itoa(req.Priority))
	}

	if req.Profile != "" {
		args = append(args, "--profile", req.Profile)
	}

	if req.Propagate != "" {
		args = append(args, "--propagate", req.Propagate)
	}

	if req.QOS != "" {
		args = append(args, "-q", req.QOS)
	}

	if req.Quiet {
		args = append(args, "-Q")
	}

	if req.Reboot {
		args = append(args, "--reboot")
	}

	if req.Requeue {
		args = append(args, "--requeue")
	}

	if req.Oversubscribe {
		args = append(args, "-s")
	}

	if req.CoreSpec > 0 {
		args = append(args, "-S", strconv.Itoa(req.CoreSpec))
	}

	if req.Signal != "" {
		args = append(args, "--signal", req.Signal)
	}

	if req.SpreadJob {
		args = append(args, "--spread-job")
	}

	if req.Switches != "" {
		args = append(args, "--switches", req.Switches)
	}

	if req.ThreadSpec > 0 {
		args = append(args, "--thread-spec", strconv.Itoa(req.ThreadSpec))
	}

	if req.Time != "" {
		args = append(args, "-t", req.Time)
	}

	if req.TimeMin != "" {
		args = append(args, "--time-min", req.TimeMin)
	}

	if req.TRESBind != "" {
		args = append(args, "--tres-bind", req.TRESBind)
	}

	if req.TRESPerTask != "" {
		args = append(args, "--tres-per-task", req.TRESPerTask)
	}

	if req.UID != "" {
		args = append(args, "--uid", req.UID)
	}

	if req.UseMinNodes {
		args = append(args, "--use-min-nodes")
	}

	if req.Verbose {
		args = append(args, "-v")
	}

	if req.Wait {
		args = append(args, "-W")
	}

	if req.WCKey != "" {
		args = append(args, "--wckey", req.WCKey)
	}

	// 约束选项
	if req.ClusterConstraint != "" {
		args = append(args, "--cluster-constraint", req.ClusterConstraint)
	}

	if req.Contiguous {
		args = append(args, "--contiguous")
	}

	if req.Constraint != "" {
		args = append(args, "-C", req.Constraint)
	}

	if req.NodeFile != "" {
		args = append(args, "-F", req.NodeFile)
	}

	if req.Memory != "" {
		args = append(args, "--mem", req.Memory)
	}

	if req.MinCPUs > 0 {
		args = append(args, "--mincpus", strconv.Itoa(req.MinCPUs))
	}

	if req.Reservation != "" {
		args = append(args, "--reservation", req.Reservation)
	}

	if req.TmpDisk != "" {
		args = append(args, "--tmp", req.TmpDisk)
	}

	if len(req.NodeList) > 0 {
		args = append(args, "-w", strings.Join(req.NodeList, ","))
	}

	if len(req.ExcludeNodes) > 0 {
		args = append(args, "-x", strings.Join(req.ExcludeNodes, ","))
	}

	// 可消费资源相关选项
	if req.Exclusive != "" {
		args = append(args, "--exclusive", req.Exclusive)
	}

	if req.MemPerCPU != "" {
		args = append(args, "--mem-per-cpu", req.MemPerCPU)
	}

	if req.ResvPorts {
		args = append(args, "--resv-ports")
	}

	// 亲和性/多核心选项
	if req.SocketsPerNode > 0 {
		args = append(args, "--sockets-per-node", strconv.Itoa(req.SocketsPerNode))
	}

	if req.CoresPerSocket > 0 {
		args = append(args, "--cores-per-socket", strconv.Itoa(req.CoresPerSocket))
	}

	if req.ThreadsPerCore > 0 {
		args = append(args, "--threads-per-core", strconv.Itoa(req.ThreadsPerCore))
	}

	if req.ExtraNodeInfo != "" {
		args = append(args, "-B", req.ExtraNodeInfo)
	}

	if req.NTasksPerCore > 0 {
		args = append(args, "--ntasks-per-core", strconv.Itoa(req.NTasksPerCore))
	}

	if req.NTasksPerSocket > 0 {
		args = append(args, "--ntasks-per-socket", strconv.Itoa(req.NTasksPerSocket))
	}

	if req.Hint != "" {
		args = append(args, "--hint", req.Hint)
	}

	if req.MemBind != "" {
		args = append(args, "--mem-bind", req.MemBind)
	}

	// GPU 调度选项
	if req.CPUsPerGPU > 0 {
		args = append(args, "--cpus-per-gpu", strconv.Itoa(req.CPUsPerGPU))
	}

	if req.GPUs > 0 {
		args = append(args, "-G", strconv.Itoa(req.GPUs))
	}

	if req.GPUBind != "" {
		args = append(args, "--gpu-bind", req.GPUBind)
	}

	if req.GPUFreq != "" {
		args = append(args, "--gpu-freq", req.GPUFreq)
	}

	if req.GPUsPerNode > 0 {
		args = append(args, "--gpus-per-node", strconv.Itoa(req.GPUsPerNode))
	}

	if req.GPUsPerSocket > 0 {
		args = append(args, "--gpus-per-socket", strconv.Itoa(req.GPUsPerSocket))
	}

	if req.GPUsPerTask > 0 {
		args = append(args, "--gpus-per-task", strconv.Itoa(req.GPUsPerTask))
	}

	if req.MemPerGPU != "" {
		args = append(args, "--mem-per-gpu", req.MemPerGPU)
	}

	// 处理脚本和参数
	if req.Wrap != "" {
		// 使用 --wrap 选项
		args = append(args, "--wrap", req.Wrap)
	} else if req.ScriptFile != "" {
		// 使用现有的脚本文件
		log.Println("Using existing script file:", req.ScriptFile)
		log.Println("parser home path:", p.HomePath)
		if p.HomePath != "" {
			scriptPath := p.processScriptPath(req.ScriptFile, p.HomePath)
			args = append(args, scriptPath)
		} else {
			args = append(args, req.ScriptFile)
		}
		// 添加脚本参数
		if len(req.ScriptArgs) > 0 {
			args = append(args, req.ScriptArgs...)
		}
	} else {
		return "", fmt.Errorf("no script provided (script_file, or wrap required)")
	}

	return strings.Join(args, " "), nil
}

// ParseOutput 解析 sbatch 命令输出
func (p *SlurmParser) ParseSbatchOutput(output string) (*models.SbatchResponse, error) {
	output = strings.TrimSpace(output)

	response := &models.SbatchResponse{
		Success:   "no",
		RawOutput: output,
	}

	if output == "" {
		response.Message = "Empty output from sbatch command"
		return response, nil
	}

	// 解析不同格式的输出

	// 1. 标准成功输出格式: "Submitted batch job 12345"
	standardPattern := regexp.MustCompile(`^Submitted batch job (\d+)$`)
	if matches := standardPattern.FindStringSubmatch(output); len(matches) > 1 {
		response.Success = "yes"
		response.JobID = matches[1]
		response.Message = "Job submitted successfully"
		return response, nil
	}

	// 2. Parsable 格式输出: "12345;cluster" 或 "12345"
	parsablePattern := regexp.MustCompile(`^(\d+)(?:;(\w+))?$`)
	if matches := parsablePattern.FindStringSubmatch(output); len(matches) > 1 {
		response.Success = "no"
		response.JobID = matches[1]
		if len(matches) > 2 && matches[2] != "" {
			response.Cluster = matches[2]
		}
		response.Message = "Job submitted successfully"
		return response, nil
	}

	// 3. 多行输出，提取作业ID
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 尝试标准格式
		if matches := standardPattern.FindStringSubmatch(line); len(matches) > 1 {
			response.Success = "yes"
			response.JobID = matches[1]
			response.Message = "Job submitted successfully"
			return response, nil
		}

		// 尝试 parsable 格式
		if matches := parsablePattern.FindStringSubmatch(line); len(matches) > 1 {
			response.Success = "yes"
			response.JobID = matches[1]
			if len(matches) > 2 && matches[2] != "" {
				response.Cluster = matches[2]
			}
			response.Message = "Job submitted successfully"
			return response, nil
		}
	}

	// 4. 检查错误信息
	errorKeywords := []string{
		"error", "Error", "ERROR",
		"invalid", "Invalid", "INVALID",
		"denied", "Denied", "DENIED",
		"failed", "Failed", "FAILED",
		"reject", "Reject", "REJECT",
	}

	for _, keyword := range errorKeywords {
		if strings.Contains(output, keyword) {
			response.Message = "Job submission failed: " + output
			return response, nil
		}
	}

	// 5. 无法解析的输出
	response.Message = "Unable to parse sbatch output: " + output
	return response, nil
}

// ValidateRequest 验证 sbatch 请求参数
func (p *SlurmParser) ValidateSbatchRequest(req *models.SbatchRequest) error {
	// 检查脚本信息
	if req.ScriptFile == "" && req.Wrap == "" {
		return fmt.Errorf("script, script_file, or wrap is required")
	}

	// 检查时间格式
	if req.Time != "" {
		if !p.isValidTimeFormat(req.Time) {
			return fmt.Errorf("invalid time format: %s", req.Time)
		}
	}

	if req.TimeMin != "" {
		if !p.isValidTimeFormat(req.TimeMin) {
			return fmt.Errorf("invalid time_min format: %s", req.TimeMin)
		}
	}

	// 检查节点数量格式
	if req.Nodes != "" {
		if !p.isValidNodeFormat(req.Nodes) {
			return fmt.Errorf("invalid nodes format: %s", req.Nodes)
		}
	}

	return nil
}

func (p *SlurmParser) processScriptPath(scriptFile string, homePath string) string {
	// 检查是否为绝对路径（以 / 开头）
	if strings.HasPrefix(scriptFile, "/") {
		// 如果是绝对路径，则添加 homePath 前缀
		return homePath + scriptFile
	}
	// 如果是相对路径，则直接返回
	return scriptFile
}
