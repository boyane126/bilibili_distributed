package main

import (
	"context"
	"fmt"
	stLog "log"
	"os"

	"bilibili_distributed/internal/log"
	"bilibili_distributed/internal/portal"
	"bilibili_distributed/internal/registry"
	"bilibili_distributed/internal/service"
)

func main() {
	str, _ := os.Getwd()
	err := portal.ImportTemplates(str + "/../../internal/portal")
	if err != nil {
		stLog.Fatal(err)
	}

	host, port := "localhost", 7000
	serviceAddr := fmt.Sprintf("http://%s:%d", host, port)

	r := registry.Registration{
		ServerName: registry.PortalService,
		ServerURL:  serviceAddr,
		ServerRequireServices: []registry.ServerName{
			registry.LogService,
			registry.GradeService,
		},
		ServerUpdateURL: serviceAddr + "/services",
		HeartbeatUrl:    serviceAddr + "/heartbeat",
	}

	ctx, err := service.Start(
		context.Background(),
		port,
		host,
		r,
		portal.RegisterHandlers,
	)
	if err != nil {
		stLog.Fatal(err)
	}

	if provider, err := registry.GetProvider(registry.LogService); err == nil {
		log.SetClientLogger(registry.LogService, provider)
	}

	<-ctx.Done()

	fmt.Println("Shutting down service portal")
}
