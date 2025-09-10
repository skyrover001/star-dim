package api

import (
	"github.com/gin-gonic/gin"
	"log"
	"star-dim/api/public"
	"star-dim/api/router"
	"star-dim/configs"
)

func StarHTTP(conf configs.Config) {
	r := gin.Default()
	server := public.Server{
		Clients:     nil,
		Record:      false,
		RecordPath:  "",
		Log:         false,
		LogFilePath: "",
	}
	server.Clients = make(map[string]*public.UserClient)
	server.Record = true
	server.RecordPath = "./"
	server.Log = true
	router.SetupRouters(r, &server)

	err := r.Run(conf.Host + ":" + conf.Port)
	if err != nil {
		log.Fatal("Error while starting server:", err)
	}
}
