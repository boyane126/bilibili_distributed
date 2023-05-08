package main

import (
	"context"
	"fmt"
	stLog "log"

	"bilibili_distributed/internal/log"
	"bilibili_distributed/internal/registry"
	"bilibili_distributed/internal/service"
)

func main() {
	host, port := "localhost", 4000
	serviceAddr := fmt.Sprintf("http://%s:%d", host, port)

	r := registry.Registration{
		ServerName:            registry.LogService,
		ServerURL:             serviceAddr,
		ServerRequireServices: make([]registry.ServerName, 0),
		ServerUpdateURL:       serviceAddr + "/services",
		HeartbeatUrl:          serviceAddr + "/heartbeat",
	}

	log.Run("./distributed.log")

	ctx, err := service.Start(
		context.Background(),
		port,
		host,
		r,
		log.RegisterLogServers,
	)
	if err != nil {
		stLog.Fatalln(err)
	}

	<-ctx.Done()

	fmt.Println("Shutting down log service.")
}
