package models

type RequestInfo struct {
	Path           string `json:"path"`
	SrcPath        string `json:"src_path"`
	DstPath        string `json:"dst_path"`
	Type           string `json:"type"`
	Cluster        string `json:"cluster"`
	SystemUsername string `json:"system_username"`
	OldPath        string `json:"old_path"`
	NewPath        string `json:"new_path"`
	Content        string `json:"content"`
	Mode           string `json:"mode"`
	Owner          string `json:"owner"`
	Group          string `json:"group"`
	ForceDir       bool   `json:"force_dir"`
	CommandParams  string `json:"command_params"`
}
