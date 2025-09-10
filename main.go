package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"star-dim/api"
	"star-dim/configs"
	_ "star-dim/docs" // 导入 docs 包以注册 Swagger 信息
)

func main() {
	// 定义命令行参数
	var (
		host = flag.String("host", getEnvOrDefault("STARDIM_HOST", "0.0.0.0"), "服务器监听地址")
		port = flag.String("port", getEnvOrDefault("STARDIM_PORT", "8080"), "服务器监听端口")
		help = flag.Bool("help", false, "显示帮助信息")
	)

	flag.Parse()
	// 显示帮助信息
	if *help {
		fmt.Println("star-dim星维融合算力接入服务器")
		fmt.Println("用法:")
		fmt.Printf("  %s [选项]\n\n", os.Args[0])
		fmt.Println("选项:")
		flag.PrintDefaults()
		fmt.Println("\n环境变量:")
		fmt.Println("  WEBSHELL_HOST    服务器监听地址 (默认: 0.0.0.0)")
		fmt.Println("  WEBSHELL_PORT    服务器监听端口 (默认: 8080)")
		fmt.Println("\n示例:")
		fmt.Printf("  %s -host 127.0.0.1 -port 9090\n", os.Args[0])
		fmt.Printf("  WEBSHELL_HOST=192.168.1.100 WEBSHELL_PORT=8888 %s\n", os.Args[0])
		return
	}
	conf := configs.Config{
		Host: *host,
		Port: *port,
	}

	// 构建监听地址
	addr := fmt.Sprintf("%s:%s", *host, *port)
	log.Printf("服务器将在 %s 上启动", addr)
	api.StarHTTP(conf)
}

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
