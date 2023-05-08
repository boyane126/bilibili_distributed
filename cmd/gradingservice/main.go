package main

import (
	"context"
	"fmt"
	stLog "log"

	"bilibili_distributed/internal/grades"
	"bilibili_distributed/internal/log"
	"bilibili_distributed/internal/registry"
	"bilibili_distributed/internal/service"
)

func main() {
	host, port := "localhost", 6000
	serviceAddr := fmt.Sprintf("http://%s:%d", host, port)

	r := registry.Registration{
		ServerName: registry.GradeService,
		ServerURL:  serviceAddr,
		ServerRequireServices: []registry.ServerName{
			registry.LogService,
		},
		ServerUpdateURL: serviceAddr + "/services",
		HeartbeatUrl:    serviceAddr + "/heartbeat",
	}
	ctx, err := service.Start(
		context.Background(),
		port,
		host,
		r,
		grades.RegisterHandlers,
	)
	if err != nil {
		stLog.Fatal(err)
	}

	if provider, err := registry.GetProvider(registry.LogService); err == nil {
		fmt.Printf("Logging service found at: %s\n", provider)
		log.SetClientLogger(registry.LogService, provider)
	}
	stLog.Println("hello world")

	<-ctx.Done()

	fmt.Println("Shutting down grading service")

}
