package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"star-dim/api/handler/filesystem"
	"star-dim/api/handler/slurm"
	"star-dim/api/handler/user"
	"star-dim/api/public"
)

func SetupRouters(r *gin.Engine, server *public.Server) {
	userHandler := user.NewUserHandler(server)
	filesHandler := filesystem.NewFilesHandler(server)
	slurmHandler := slurm.NewSlurmHandler(server)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")

	userRouter := v1.Group("user")
	userRouter.POST("/login", userHandler.Login)
	userRouter.POST("/logout", userHandler.Logout)

	fileRouter := v1.Group("filesystem")
	fileRouter.GET("/files/", filesHandler.List)                       //request param: path!,cluster? systemUsername? ok!
	fileRouter.POST("files/", filesHandler.New)                        //request param: path!,cluster? systemUsername? ok!
	fileRouter.DELETE("files/", filesHandler.Delete)                   //request param: path!,cluster? systemUsername? ok!
	fileRouter.PUT("/files/", filesHandler.Rename)                     //request param: path!,newPath!,cluster? systemUsername? ok!
	fileRouter.GET("/files/content/", filesHandler.ReadFile)           //request param: path!,content!,cluster? systemUsername? ok!
	fileRouter.POST("/files/content/", filesHandler.WriteFile)         //request param: path!,cluster? systemUsername? ok!
	fileRouter.POST("/files/copy/", filesHandler.Copy)                 //request param: path!,cluster? systemUsername? ok!
	fileRouter.POST("/files/move/", filesHandler.Move)                 //request param: path!,cluster? systemUsername? ok!
	fileRouter.GET("/files/attr/", filesHandler.Attr)                  //request param: path!,cluster? systemUsername? ok!
	fileRouter.POST("/files/chmod/", filesHandler.Chmod)               //request param: path!,cluster? systemUsername? ok!
	fileRouter.POST("/files/chown/", filesHandler.Chown)               //request param: path!,cluster? systemUsername?
	fileRouter.POST("/files/transmission/", filesHandler.Transmission) //request param: path!,cluster? systemUsername?
	fileRouter.GET("/files/download/", filesHandler.Download)          //request param: path!,cluster? systemUsername? ok!
	fileRouter.GET("/quota/", filesHandler.Quota)                      //request param: path!,cluster? systemUsername? no quota cmd!
	fileRouter.POST("/files/execute/", filesHandler.ExecuteFile)

	//slurmRouter := v1.Group("/slurm")
	//sacctRouter := slurmRouter.Group("/sacct")
	//sacctRouter.POST("/accounting/", slurmHandler.GetAccounting)
	//sacctRouter.POST("/jobs/", slurmHandler.GetJobs)
	//sacctRouter.POST("/jobs/:jobid/", slurmHandler.GetJobDetail)
	//sacctRouter.POST("/jobs/user/:user/", slurmHandler.GetUserJobs)
	//sbatchRouter := slurmRouter.Group("/sbatch")
	//sbatchRouter.POST("/job/", slurmHandler.SubmitJob)
	//scancelRouter := slurmRouter.Group("/scancel")
	//scancelRouter.POST("/job/", slurmHandler.CancelJob)
	//squeueRouter := slurmRouter.Group("/squeue")
	//squeueRouter.POST("/jobs/", slurmHandler.GetQueue)
	//sinfoRouter := slurmRouter.Group("/sinfo")
	//sinfoRouter.POST("/cluster/", slurmHandler.GetClusterInfo)

	slurmRouter := v1.Group("/slurm")
	slurmRouter.POST("/job/", slurmHandler.SubmitJob)
	slurmRouter.DELETE("/job/", slurmHandler.CancelJob)
	slurmRouter.POST("/jobs/", slurmHandler.GetQueue)
	slurmRouter.POST("/account/", slurmHandler.GetAccounting)
	slurmRouter.POST("/cluster/", slurmHandler.GetClusterInfo)
}
