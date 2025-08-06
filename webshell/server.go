package main

import (
	"github.com/gin-gonic/gin"
	//"log"
	controller2 "webshell/controller"
	utils "webshell/utils"
)

func main() {
	Clients := make(map[string]*controller2.JumpClient)
	controller := controller2.JumpController{Clients: Clients, Record: true, RecordPath: "rec"}
	controller.Cache = &utils.Cache{
		Address:  "192.168.2.35:6379",
		Password: "",
		DB:       0,
	}
	//err := controller.Cache.InitRedis()
	//if err != nil {
	//	log.Fatal(err)
	//}
	router := gin.Default()

	router.POST("/api/v2/document/login/", controller.Login)
	router.GET("/api/v2/document/logout/", controller.Logout)

	router.GET("/api/v2/web/shell/", controller.SSH)

	router.GET("/api/v2/files/", controller.List)                       //request param: path!,cluster? systemUsername? ok!
	router.POST("/api/v2/files/", controller.New)                       //request param: path!,cluster? systemUsername? ok!
	router.DELETE("/api/v2/files/", controller.Delete)                  //request param: path!,cluster? systemUsername? ok!
	router.PUT("/api/v2/files/", controller.Rename)                     //request param: path!,newPath!,cluster? systemUsername? ok!
	router.GET("/api/v2/files/content/", controller.ReadFile)           //request param: path!,content!,cluster? systemUsername? ok!
	router.POST("/api/v2/files/content/", controller.WriteFile)         //request param: path!,cluster? systemUsername? ok!
	router.POST("/api/v2/files/copy/", controller.Copy)                 //request param: path!,cluster? systemUsername? ok!
	router.POST("/api/v2/files/move/", controller.Move)                 //request param: path!,cluster? systemUsername? ok!
	router.GET("/api/v2/files/attr/", controller.Attr)                  //request param: path!,cluster? systemUsername? ok!
	router.GET("/api/v2/files/quota/", controller.Quota)                //request param: path!,cluster? systemUsername? no quota cmd!
	router.POST("/api/v2/files/chmod/", controller.Chmod)               //request param: path!,cluster? systemUsername?
	router.POST("/api/v2/files/chown/", controller.Chown)               //request param: path!,cluster? systemUsername?
	router.POST("/api/v2/files/transmission/", controller.Transmission) //request param: path!,cluster? systemUsername?
	router.GET("/api/v2/files/download/", controller.Download)          //request param: path!,cluster? systemUsername? ok!

	err := router.Run("0.0.0.0:8080")
	if err != nil {
		// close all sftp connections
		controller.Close()
	}
}
