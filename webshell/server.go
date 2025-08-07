package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
	"webshell/controller"
	"webshell/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "webshell/docs"
)

func main() {
	// 定义命令行参数
	var (
		host = flag.String("host", getEnvOrDefault("WEBSHELL_HOST", "0.0.0.0"), "服务器监听地址")
		port = flag.String("port", getEnvOrDefault("WEBSHELL_PORT", "8080"), "服务器监听端口")
		help = flag.Bool("help", false, "显示帮助信息")
	)
	
	flag.Parse()
	
	// 显示帮助信息
	if *help {
		fmt.Println("WebShell 服务器")
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
	
	// 构建监听地址
	addr := fmt.Sprintf("%s:%s", *host, *port)
	log.Printf("服务器将在 %s 上启动", addr)

	Clients := make(map[string]*controller.JumpClient)
	service := controller.JumpService{Clients: Clients, Record: true, RecordPath: "rec"}
	service.Cache = &utils.Cache{
		Address:  "192.168.2.35:6379",
		Password: "",
		DB:       0,
	}
	//err := controller.Cache.InitRedis()
	//if err != nil {
	//	log.Fatal(err)
	//}
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{fmt.Sprintf("http://%s", addr), "http://localhost:8080", "http://127.0.0.1:8080"},  // 动态允许当前服务器地址
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // 允许的请求方法
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "sessionKey"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.POST("/api/v2/login/", service.Login)
	router.GET("/api/v2/logout/", service.Logout)

	router.GET("/api/v2/webshell/", service.SSH)

	router.GET("/api/v2/files/", service.List)                       //request param: path!,cluster? systemUsername? ok!
	router.POST("/api/v2/files/", service.New)                       //request param: path!,cluster? systemUsername? ok!
	router.DELETE("/api/v2/files/", service.Delete)                  //request param: path!,cluster? systemUsername? ok!
	router.PUT("/api/v2/files/", service.Rename)                     //request param: path!,newPath!,cluster? systemUsername? ok!
	router.GET("/api/v2/files/content/", service.ReadFile)           //request param: path!,content!,cluster? systemUsername? ok!
	router.POST("/api/v2/files/content/", service.WriteFile)         //request param: path!,cluster? systemUsername? ok!
	router.POST("/api/v2/files/copy/", service.Copy)                 //request param: path!,cluster? systemUsername? ok!
	router.POST("/api/v2/files/move/", service.Move)                 //request param: path!,cluster? systemUsername? ok!
	router.GET("/api/v2/files/attr/", service.Attr)                  //request param: path!,cluster? systemUsername? ok!
	router.GET("/api/v2/files/quota/", service.Quota)                //request param: path!,cluster? systemUsername? no quota cmd!
	router.POST("/api/v2/files/chmod/", service.Chmod)               //request param: path!,cluster? systemUsername? ok!
	router.POST("/api/v2/files/chown/", service.Chown)               //request param: path!,cluster? systemUsername?
	router.POST("/api/v2/files/transmission/", service.Transmission) //request param: path!,cluster? systemUsername?
	router.GET("/api/v2/files/download/", service.Download)          //request param: path!,cluster? systemUsername? ok!

	log.Printf("启动服务器在 %s", addr)
	err := router.Run(addr)
	if err != nil {
		log.Printf("服务器启动失败: %v", err)
		// close all sftp connections
		service.Close()
	}
}

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
