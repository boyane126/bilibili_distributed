package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"bilibili_distributed/internal/registry"
)

func main() {
	registry.StartHeartbeatService()
	http.Handle("/services", registry.RegistryServer{})

	var srv http.Server
	srv.Addr = registry.ServerPort

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		log.Println(srv.ListenAndServe())
		cancel()
	}()

	go func() {
		fmt.Println(fmt.Sprintf("%s started. Press any key to stop.", "registryService"))
		var s string
		fmt.Scanln(&s)
		srv.Shutdown(ctx)
		cancel()
	}()

	<-ctx.Done()

	log.Println("registryService shutdown")
}
