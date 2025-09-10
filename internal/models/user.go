package models

type User struct {
	Cluster    *Cluster `json:"cluster"`
	Name       string   `json:"username"`
	Password   string   `json:"password"`
	PrivateKey string   `json:"private_key"`
	HomePath   string   `json:"home_path"`
}

type LoginInfo struct {
	User      *User      `json:"user"`
	LoginNode *LoginNode `json:"login_node"`
}
