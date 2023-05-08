// 日志服务业务逻辑

package log

import (
	"io/ioutil"
	stLog "log"
	"net/http"
	"os"
)

var log *stLog.Logger

func Run(distributed string) {
	log = stLog.New(fileLog(distributed), "[go] - ", stLog.LstdFlags)
}

type fileLog string

func (fl fileLog) Write(p []byte) (n int, err error) {
	f, err := os.OpenFile(string(fl), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return f.Write(p)
}

// 注册web服务
func RegisterLogServers() {
	http.HandleFunc("/log", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodPost:
			data, err := ioutil.ReadAll(request.Body)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				return
			}

			write(string(data))
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	})
}

func write(data string) {
	log.Printf("%s\n", data)
}
