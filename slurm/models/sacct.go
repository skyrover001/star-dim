package models

import "time"

// JobInfo 表示 SLURM 作业信息
type JobInfo struct {
	JobID        string    `json:"jobid"`
	JobIDRaw     string    `json:"jobidraw,omitempty"`
	JobName      string    `json:"jobname"`
	Partition    string    `json:"partition"`
	Account      string    `json:"account"`
	AllocCPUS    int       `json:"alloccpus"`
	State        string    `json:"state"`
	ExitCode     string    `json:"exitcode"`
	Submit       time.Time `json:"submit"`
	Start        time.Time `json:"start"`
	End          time.Time `json:"end"`
	Elapsed      string    `json:"elapsed"`
	ReqMem       string    `json:"reqmem"`
	ReqNodes     int       `json:"reqnodes"`
	AllocNodes   int       `json:"allocnodes"`
	NodeList     string    `json:"nodelist"`
	User         string    `json:"user"`
	Group        string    `json:"group"`
	QOS          string    `json:"qos"`
	WCKey        string    `json:"wckey,omitempty"`
	Cluster      string    `json:"cluster"`
	CPUTime      string    `json:"cputime,omitempty"`
	UserCPU      string    `json:"usercpu,omitempty"`
	SystemCPU    string    `json:"systemcpu,omitempty"`
	TotalCPU     string    `json:"totalcpu,omitempty"`
	MaxRSS       string    `json:"maxrss,omitempty"`
	MaxVMSize    string    `json:"maxvmsize,omitempty"`
	MaxPages     string    `json:"maxpages,omitempty"`
	MaxDiskRead  string    `json:"maxdiskread,omitempty"`
	MaxDiskWrite string    `json:"maxdiskwrite,omitempty"`
	ReqTRES      string    `json:"reqtres,omitempty"`
	AllocTRES    string    `json:"alloctres,omitempty"`
}

type SacctRequest struct {
	JobIDs      []string `json:"jobids,omitempty"`
	Users       []string `json:"users,omitempty"`
	Accounts    []string `json:"accounts,omitempty"`
	Partitions  []string `json:"partitions,omitempty"`
	States      []string `json:"states,omitempty"`
	QOS         []string `json:"qos,omitempty"`
	Clusters    []string `json:"clusters,omitempty"`
	NodeList    []string `json:"nodelist,omitempty"`
	JobNames    []string `json:"jobnames,omitempty"`
	StartTime   string   `json:"starttime,omitempty"`
	EndTime     string   `json:"endtime,omitempty"`
	MinNodes    int      `json:"minnodes,omitempty"`
	MaxNodes    int      `json:"maxnodes,omitempty"`
	MinCPUs     int      `json:"mincpus,omitempty"`
	MaxCPUs     int      `json:"maxcpus,omitempty"`
	Format      string   `json:"format,omitempty"`
	Brief       bool     `json:"brief,omitempty"`
	Long        bool     `json:"long,omitempty"`
	Parsable    bool     `json:"parsable,omitempty"`
	NoHeader    bool     `json:"noheader,omitempty"`
	AllUsers    bool     `json:"allusers,omitempty"`
	AllClusters bool     `json:"allclusters,omitempty"`
	Duplicates  bool     `json:"duplicates,omitempty"`
	Truncate    bool     `json:"truncate,omitempty"`
	ArrayJobs   bool     `json:"arrayjobs,omitempty"`
	Completion  bool     `json:"completion,omitempty"`
	Host        string   `json:"host"`
	Username    string   `json:"username"`
	Password    string   `json:"password,omitempty"`
	PrivateKey  string   `json:"privatekey,omitempty"`
	Port        int      `json:"port,omitempty"`
}

type SacctResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message,omitempty"`
	Data      []JobInfo `json:"data,omitempty"`
	Total     int       `json:"total"`
	Command   string    `json:"command,omitempty"`
	RawOutput string    `json:"raw_output,omitempty"`
}

type SSHConfig struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password,omitempty"`
	PrivateKey string `json:"privatekey,omitempty"`
	Timeout    int    `json:"timeout"`
}

// SbatchRequest 表示 sbatch 命令请求参数
type SbatchRequest struct {
	Host              string   `json:"host"`
	Port              int      `json:"port,omitempty"`
	Username          string   `json:"username"`
	Password          string   `json:"password,omitempty"`
	PrivateKey        string   `json:"privatekey,omitempty"`
	Script            string   `json:"script"`
	ScriptFile        string   `json:"script_file,omitempty"`
	ScriptArgs        []string `json:"script_args,omitempty"`
	Array             string   `json:"array,omitempty"`
	Account           string   `json:"account,omitempty"`
	Begin             string   `json:"begin,omitempty"`
	Comment           string   `json:"comment,omitempty"`
	CPUFreq           string   `json:"cpu_freq,omitempty"`
	CPUsPerTask       int      `json:"cpus_per_task,omitempty"`
	Dependency        string   `json:"dependency,omitempty"`
	Deadline          string   `json:"deadline,omitempty"`
	DelayBoot         int      `json:"delay_boot,omitempty"`
	Chdir             string   `json:"chdir,omitempty"`
	Error             string   `json:"error,omitempty"`
	Export            string   `json:"export,omitempty"`
	ExportFile        string   `json:"export_file,omitempty"`
	GetUserEnv        bool     `json:"get_user_env,omitempty"`
	GID               string   `json:"gid,omitempty"`
	GRES              string   `json:"gres,omitempty"`
	GRESFlags         string   `json:"gres_flags,omitempty"`
	Hold              bool     `json:"hold,omitempty"`
	IgnorePBS         bool     `json:"ignore_pbs,omitempty"`
	Input             string   `json:"input,omitempty"`
	JobName           string   `json:"job_name,omitempty"`
	NoKill            bool     `json:"no_kill,omitempty"`
	Licenses          string   `json:"licenses,omitempty"`
	Clusters          []string `json:"clusters,omitempty"`
	Container         string   `json:"container,omitempty"`
	ContainerID       string   `json:"container_id,omitempty"`
	Distribution      string   `json:"distribution,omitempty"`
	MailType          string   `json:"mail_type,omitempty"`
	MailUser          string   `json:"mail_user,omitempty"`
	MCSLabel          string   `json:"mcs_label,omitempty"`
	NTasks            int      `json:"ntasks,omitempty"`
	Nice              int      `json:"nice,omitempty"`
	NoRequeue         bool     `json:"no_requeue,omitempty"`
	NTasksPerNode     int      `json:"ntasks_per_node,omitempty"`
	Nodes             string   `json:"nodes,omitempty"`
	Output            string   `json:"output,omitempty"`
	Overcommit        bool     `json:"overcommit,omitempty"`
	Partition         string   `json:"partition,omitempty"`
	Parsable          bool     `json:"parsable,omitempty"`
	Power             string   `json:"power,omitempty"`
	Priority          int      `json:"priority,omitempty"`
	Profile           string   `json:"profile,omitempty"`
	Propagate         string   `json:"propagate,omitempty"`
	QOS               string   `json:"qos,omitempty"`
	Quiet             bool     `json:"quiet,omitempty"`
	Reboot            bool     `json:"reboot,omitempty"`
	Requeue           bool     `json:"requeue,omitempty"`
	Oversubscribe     bool     `json:"oversubscribe,omitempty"`
	CoreSpec          int      `json:"core_spec,omitempty"`
	Signal            string   `json:"signal,omitempty"`
	SpreadJob         bool     `json:"spread_job,omitempty"`
	Switches          string   `json:"switches,omitempty"`
	ThreadSpec        int      `json:"thread_spec,omitempty"`
	Time              string   `json:"time,omitempty"`
	TimeMin           string   `json:"time_min,omitempty"`
	TRESBind          string   `json:"tres_bind,omitempty"`
	TRESPerTask       string   `json:"tres_per_task,omitempty"`
	UID               string   `json:"uid,omitempty"`
	UseMinNodes       bool     `json:"use_min_nodes,omitempty"`
	Verbose           bool     `json:"verbose,omitempty"`
	Wait              bool     `json:"wait,omitempty"`
	WCKey             string   `json:"wckey,omitempty"`
	Wrap              string   `json:"wrap,omitempty"`
	Memory            string   `json:"memory,omitempty"`
	MinCPUs           int      `json:"min_cpus,omitempty"`
	Reservation       string   `json:"reservation,omitempty"`
	TmpDisk           string   `json:"tmp_disk,omitempty"`
	NodeList          []string `json:"node_list,omitempty"`
	ExcludeNodes      []string `json:"exclude_nodes,omitempty"`
	Exclusive         string   `json:"exclusive,omitempty"`
	MemPerCPU         string   `json:"mem_per_cpu,omitempty"`
	ResvPorts         bool     `json:"resv_ports,omitempty"`
	ClusterConstraint string   `json:"cluster_constraint,omitempty"`
	Contiguous        bool     `json:"contiguous,omitempty"`
	Constraint        string   `json:"constraint,omitempty"`
	NodeFile          string   `json:"node_file,omitempty"`
	SocketsPerNode    int      `json:"sockets_per_node,omitempty"`
	CoresPerSocket    int      `json:"cores_per_socket,omitempty"`
	ThreadsPerCore    int      `json:"threads_per_core,omitempty"`
	ExtraNodeInfo     string   `json:"extra_node_info,omitempty"`
	NTasksPerCore     int      `json:"ntasks_per_core,omitempty"`
	NTasksPerSocket   int      `json:"ntasks_per_socket,omitempty"`
	Hint              string   `json:"hint,omitempty"`
	MemBind           string   `json:"mem_bind,omitempty"`
	CPUsPerGPU        int      `json:"cpus_per_gpu,omitempty"`
	GPUs              int      `json:"gpus,omitempty"`
	GPUBind           string   `json:"gpu_bind,omitempty"`
	GPUFreq           string   `json:"gpu_freq,omitempty"`
	GPUsPerNode       int      `json:"gpus_per_node,omitempty"`
	GPUsPerSocket     int      `json:"gpus_per_socket,omitempty"`
	GPUsPerTask       int      `json:"gpus_per_task,omitempty"`
	MemPerGPU         string   `json:"mem_per_gpu,omitempty"`
}

type SbatchResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	JobID     string `json:"job_id,omitempty"`
	Cluster   string `json:"cluster,omitempty"`
	Command   string `json:"command,omitempty"`
	RawOutput string `json:"raw_output,omitempty"`
}

// QueueJobInfo 表示 squeue 返回的作业队列信息
type QueueJobInfo struct {
	JobID       string    `json:"jobid"`
	Partition   string    `json:"partition"`
	Name        string    `json:"name"`
	User        string    `json:"user"`
	State       string    `json:"state"`
	Time        string    `json:"time"`
	TimeLeft    string    `json:"time_left"`
	Nodes       int       `json:"nodes"`
	NodeList    string    `json:"nodelist"`
	Reason      string    `json:"reason"`
	Priority    int64     `json:"priority"`
	QOS         string    `json:"qos"`
	Account     string    `json:"account"`
	CPUS        int       `json:"cpus"`
	MinMemory   string    `json:"min_memory"`
	SubmitTime  time.Time `json:"submit_time"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	PreemptTime time.Time `json:"preempt_time"`
	Dependency  string    `json:"dependency"`
	ArrayJobID  string    `json:"array_job_id"`
	ArrayTaskID string    `json:"array_task_id"`
	GroupID     int       `json:"group_id"`
	BatchHost   string    `json:"batch_host"`
	Command     string    `json:"command"`
	WorkDir     string    `json:"work_dir"`
	StdOut      string    `json:"std_out"`
	StdErr      string    `json:"std_err"`
	Licenses    string    `json:"licenses"`
	ReqNodes    string    `json:"req_nodes"`
	ExcNodes    string    `json:"exc_nodes"`
	Features    string    `json:"features"`
	Gres        string    `json:"gres"`
	Reservation string    `json:"reservation"`
	Network     string    `json:"network"`
	Sockets     int       `json:"sockets"`
	Cores       int       `json:"cores"`
	Threads     int       `json:"threads"`
	NiceValue   int       `json:"nice_value"`
}

type SqueueRequest struct {
	Host         string   `json:"host" binding:"required"`
	Username     string   `json:"username" binding:"required"`
	Password     string   `json:"password,omitempty"`
	PrivateKey   string   `json:"privatekey,omitempty"`
	Port         int      `json:"port,omitempty"`
	Accounts     []string `json:"accounts,omitempty"`
	Jobs         []string `json:"jobs,omitempty"`
	Partitions   []string `json:"partitions,omitempty"`
	QOS          []string `json:"qos,omitempty"`
	States       []string `json:"states,omitempty"`
	Users        []string `json:"users,omitempty"`
	Names        []string `json:"names,omitempty"`
	Clusters     []string `json:"clusters,omitempty"`
	Licenses     []string `json:"licenses,omitempty"`
	NodeList     []string `json:"nodelist,omitempty"`
	Steps        []string `json:"steps,omitempty"`
	Reservation  string   `json:"reservation,omitempty"`
	Format       string   `json:"format,omitempty"`
	FormatLong   string   `json:"format_long,omitempty"`
	NoHeader     bool     `json:"noheader,omitempty"`
	Long         bool     `json:"long,omitempty"`
	NoConvert    bool     `json:"noconvert,omitempty"`
	Array        bool     `json:"array,omitempty"`
	Start        bool     `json:"start,omitempty"`
	Verbose      bool     `json:"verbose,omitempty"`
	All          bool     `json:"all,omitempty"`
	Hide         bool     `json:"hide,omitempty"`
	Federation   bool     `json:"federation,omitempty"`
	Local        bool     `json:"local,omitempty"`
	Sibling      bool     `json:"sibling,omitempty"`
	OnlyJobState bool     `json:"only_job_state,omitempty"`
	JSON         string   `json:"json,omitempty"`
	YAML         string   `json:"yaml,omitempty"`
	Sort         []string `json:"sort,omitempty"`
	Iterate      int      `json:"iterate,omitempty"`
}

type SqueueResponse struct {
	Success   bool           `json:"success"`
	Message   string         `json:"message,omitempty"`
	Data      []QueueJobInfo `json:"data,omitempty"`
	Total     int            `json:"total"`
	Command   string         `json:"command,omitempty"`
	RawOutput string         `json:"raw_output,omitempty"`
}

// SinfoRequest 表示 sinfo 命令请求参数
type SinfoRequest struct {
	Host        string   `json:"host" binding:"required"`
	Username    string   `json:"username" binding:"required"`
	Password    string   `json:"password,omitempty"`
	PrivateKey  string   `json:"privatekey,omitempty"`
	Port        int      `json:"port,omitempty"`
	Nodes       []string `json:"nodes,omitempty"`
	Partitions  []string `json:"partitions,omitempty"`
	States      []string `json:"states,omitempty"`
	Clusters    []string `json:"clusters,omitempty"`
	All         bool     `json:"all,omitempty"`
	Dead        bool     `json:"dead,omitempty"`
	Exact       bool     `json:"exact,omitempty"`
	Future      bool     `json:"future,omitempty"`
	Hide        bool     `json:"hide,omitempty"`
	Long        bool     `json:"long,omitempty"`
	NodeCentric bool     `json:"node_centric,omitempty"`
	Responding  bool     `json:"responding,omitempty"`
	ListReasons bool     `json:"list_reasons,omitempty"`
	Summarize   bool     `json:"summarize,omitempty"`
	Reservation bool     `json:"reservation,omitempty"`
	Verbose     bool     `json:"verbose,omitempty"`
	Federation  bool     `json:"federation,omitempty"`
	Local       bool     `json:"local,omitempty"`
	NoConvert   bool     `json:"noconvert,omitempty"`
	NoHeader    bool     `json:"noheader,omitempty"`
	Format      string   `json:"format,omitempty"`
	FormatLong  string   `json:"format_long,omitempty"`
	JSON        string   `json:"json,omitempty"`
	YAML        string   `json:"yaml,omitempty"`
	Sort        []string `json:"sort,omitempty"`
	Iterate     int      `json:"iterate,omitempty"`
}

type SinfoResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Total     int         `json:"total"`
	Command   string      `json:"command,omitempty"`
	RawOutput string      `json:"raw_output,omitempty"`
}

type NodeInfo struct {
	NodeName        string `json:"nodename"`
	Partition       string `json:"partition"`
	State           string `json:"state"`
	CPUs            int    `json:"cpus"`
	Memory          int    `json:"memory"`
	TmpDisk         int    `json:"tmp_disk"`
	Weight          int    `json:"weight"`
	Features        string `json:"features"`
	Gres            string `json:"gres"`
	Reason          string `json:"reason"`
	User            string `json:"user"`
	Timestamp       string `json:"timestamp"`
	AllocCPUs       int    `json:"alloc_cpus"`
	IdleCPUs        int    `json:"idle_cpus"`
	OtherCPUs       int    `json:"other_cpus"`
	AllocMemory     int    `json:"alloc_memory"`
	IdleMemory      int    `json:"idle_memory"`
	OtherMemory     int    `json:"other_memory"`
	Sockets         int    `json:"sockets"`
	Cores           int    `json:"cores"`
	Threads         int    `json:"threads"`
	SlurmdStartTime string `json:"slurmd_start_time"`
	BootTime        string `json:"boot_time"`
	OS              string `json:"os"`
	Architecture    string `json:"architecture"`
	Load            string `json:"load"`
	FreeMem         int    `json:"free_mem"`
}

type PartitionInfo struct {
	PartitionName string   `json:"partition_name"`
	Availability  string   `json:"availability"`
	TimeLimit     string   `json:"time_limit"`
	Nodes         int      `json:"nodes"`
	NodeList      string   `json:"node_list"`
	State         string   `json:"state"`
	Root          bool     `json:"root"`
	OverSubscribe string   `json:"over_subscribe"`
	Groups        []string `json:"groups"`
	Priority      int      `json:"priority"`
}

type ReservationInfo struct {
	ReservationName string `json:"reservation_name"`
	State           string `json:"state"`
	StartTime       string `json:"start_time"`
	EndTime         string `json:"end_time"`
	Duration        string `json:"duration"`
	NodeList        string `json:"node_list"`
	NodeCount       int    `json:"node_count"`
	CoreCount       int    `json:"core_count"`
	Features        string `json:"features"`
	PartitionName   string `json:"partition_name"`
	Flags           string `json:"flags"`
	TRES            string `json:"tres"`
	Users           string `json:"users"`
	Accounts        string `json:"accounts"`
}

type NodeSummary struct {
	State string `json:"state"`
	Count int    `json:"count"`
	CPUs  int    `json:"cpus"`
}
