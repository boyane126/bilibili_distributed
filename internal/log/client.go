package log

import (
	"bilibili_distributed/internal/registry"
	"bytes"
	"fmt"
	stLog "log"
	"net/http"
)

func SetClientLogger(name registry.ServerName, url string) {
	stLog.SetPrefix(fmt.Sprintf("[%v] - ", name))
	stLog.SetFlags(0)
	stLog.SetOutput(&clientLogger{
		url: url,
	})
}

type clientLogger struct {
	url string
}

func (l clientLogger) Write(data []byte) (n int, err error) {
	p, err := http.Post(l.url+"/log", "text/plain", bytes.NewBuffer(data))
	if err != nil {
		return 0, err
	}
	if p.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to send log message. Service responded with %v", p.StatusCode)
	}

	return len(data), nil
}
