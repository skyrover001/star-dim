package main

import (
	"log"
	"slurm-jobacct/controller"
	"slurm-jobacct/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	// 创建 Gin 路由器
	router := gin.Default()

	// 配置 CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Requested-With", "sessionKey"}
	router.Use(cors.New(config))

	// 使用认证中间件
	router.Use(middleware.AuthMiddleware())

	// API 路由组
	api := router.Group("/api/v1")
	{
		// SLURM 作业记账相关路由
		jobacct := api.Group("/jobacct")
		{
			jobacct.GET("/jobs", controller.GetJobs)                 // 获取作业列表
			jobacct.GET("/jobs/:jobid", controller.GetJobDetail)     // 获取单个作业详情
			jobacct.GET("/users/:user/jobs", controller.GetUserJobs) // 获取用户作业
			jobacct.POST("/accounting", controller.GetAccounting)    // 获取记账信息（复杂查询）
		}

		// SLURM 作业提交相关路由
		jobsubmit := api.Group("/jobsubmit")
		{
			jobsubmit.POST("/submit", controller.SubmitJob)                  // 提交作业（完整参数）
			jobsubmit.POST("/submit-script", controller.SubmitJobWithScript) // 通过上传脚本提交作业
			jobsubmit.POST("/quick-submit", controller.QuickSubmit)          // 快速提交作业（简化参数）
		}

		// SLURM 作业队列相关路由
		jobqueue := api.Group("/jobqueue")
		{
			jobqueue.GET("/queue", controller.GetQueue)                 // 获取作业队列列表
			jobqueue.GET("/jobs/:jobid", controller.GetJobQueue)        // 获取单个作业的队列信息
			jobqueue.GET("/users/:user/queue", controller.GetUserQueue) // 获取用户的作业队列
			jobqueue.POST("/query", controller.QueryQueue)              // 复杂队列查询
			jobqueue.GET("/stats", controller.GetQueueStats)            // 获取队列统计信息
		}

		// SLURM 集群信息相关路由
		cluster := api.Group("/cluster")
		{
			cluster.GET("/info", controller.GetClusterInfo)             // 获取集群概览信息
			cluster.GET("/nodes", controller.GetNodeInfo)               // 获取节点信息
			cluster.GET("/partitions", controller.GetPartitionInfo)     // 获取分区信息
			cluster.GET("/reservations", controller.GetReservationInfo) // 获取预留信息
			cluster.POST("/query", controller.QueryClusterInfo)         // 复杂集群查询
		}
	}

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "slurm-api-server",
			"apis":    []string{"jobacct", "jobsubmit", "jobqueue", "cluster"},
		})
	})

	// 启动服务器
	log.Println("SLURM JobAcct API Server starting on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
