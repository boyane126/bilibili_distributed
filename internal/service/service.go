// 统一注册服务逻辑

package service

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"bilibili_distributed/internal/registry"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func Start(ctx context.Context, port int, host string, registration registry.Registration, registerFun func()) (context.Context, error) {
	registerFun()

	ctx = startService(ctx, port, host, string(registration.ServerName))

	if err := registry.RegistryService(registration); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func startService(ctx context.Context, port int, host, serverName string) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	var srv http.Server

	srv.Addr = fmt.Sprintf("%s:%d", host, port)

	go func() {
		log.Println(srv.ListenAndServe())
		err := registry.ShutDownService(fmt.Sprintf("http://%s:%d", host, port))
		if err != nil {
			log.Println(err)
		}
		cancel()
	}()

	go func() {
		fmt.Println(fmt.Sprintf("%s started. Press any key to stop.", serverName))
		var s string
		fmt.Scanln(&s)
		err := registry.ShutDownService(fmt.Sprintf("http://%s:%d", host, port))
		if err != nil {
			log.Println(err)
		}
		srv.Shutdown(ctx)
		cancel()
	}()

	return ctx
}
