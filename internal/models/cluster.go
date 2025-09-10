package models

type Cluster struct {
	Name       string       `json:"name"`
	LoginNodes []*LoginNode `json:"login_nodes"`
}
