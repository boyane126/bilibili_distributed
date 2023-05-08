// 服务注册服务逻辑

package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

const ServerPort = ":3000"
const ServicesUrl = "http://localhost" + ServerPort + "/services"

var reg = registry{
	registrations: make([]Registration, 0),
	mu:            new(sync.RWMutex),
}

type registry struct {
	registrations []Registration
	mu            *sync.RWMutex
}

func (r *registry) add(reg Registration) error {
	r.mu.Lock()
	r.registrations = append(r.registrations, reg)
	r.mu.Unlock()

	// 请求依赖 - 我依赖的服务
	err := r.sendRequireServices(reg)
	// 通知注册表 - 依赖我的服务
	r.notify(patch{Added: []patchEntry{
		{Name: reg.ServerName, URL: reg.ServerURL},
	}})

	return err
}

func (r *registry) sendRequireServices(reg Registration) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var p patch
	for _, regService := range r.registrations {
		for _, requireService := range reg.ServerRequireServices {
			if regService.ServerName == requireService {
				p.Added = append(p.Added, patchEntry{
					Name: regService.ServerName,
					URL:  regService.ServerURL,
				})
			}
		}
	}

	err := r.sendPatch(p, reg.ServerUpdateURL)

	return err
}

func (r *registry) notify(pat patch) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, reg := range r.registrations {
		go func(reg Registration) {
			for _, reqService := range reg.ServerRequireServices {
				p := patch{Added: []patchEntry{}, Removed: []patchEntry{}}
				sendUpdate := false

				for _, entry := range pat.Removed {
					if entry.Name == reqService {
						sendUpdate = true
						p.Removed = append(p.Removed, entry)
					}
				}

				for _, entry := range pat.Added {
					if entry.Name == reqService {
						sendUpdate = true
						p.Added = append(p.Added, entry)
					}
				}

				if sendUpdate {
					err := r.sendPatch(p, reg.ServerUpdateURL)
					if err != nil {
						log.Println(err)
						return
					}
				}
			}
		}(reg)
	}
}

func (r *registry) sendPatch(p patch, updateUrl string) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}

	_, err = http.Post(updateUrl, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	return nil
}

func (r *registry) remove(url string) error {
	for i, registration := range r.registrations {
		if registration.ServerURL == url {
			r.notify(patch{Removed: []patchEntry{{
				Name: registration.ServerName,
				URL:  registration.ServerURL,
			}}})
			r.mu.Lock()
			r.registrations = append(r.registrations[:i], r.registrations[i+1:]...)
			r.mu.Unlock()
			return nil
		}
	}

	return fmt.Errorf("sevice url with %s no found", url)
}

func (r *registry) heartbeat(freq time.Duration) {
	for {
		var wg sync.WaitGroup
		for _, registration := range r.registrations {
			wg.Add(1)
			go func(reg Registration) {
				defer wg.Done()
				success := true
				for attr := 0; attr < 3; attr++ {
					rsp, err := http.Get(reg.HeartbeatUrl)
					if err != nil {
						log.Println(err)
					} else if rsp.StatusCode == http.StatusOK {
						if !success {
							r.add(reg)
						}
						log.Println(fmt.Sprintf("heartbeat check passed for %s", reg.ServerName))
						break
					}
					log.Println(fmt.Sprintf("heartbeat check failed for %s", reg.ServerName))
					if success {
						success = false
						r.remove(reg.ServerURL)
					}
					time.Sleep(time.Second)
				}
			}(registration)
			time.Sleep(freq)
		}
		wg.Wait()
	}
}

var one sync.Once

func StartHeartbeatService() {
	one.Do(func() {
		go reg.heartbeat(3 * time.Second)
	})
}

type RegistryServer struct{}

func (s RegistryServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	log.Println("request received")
	switch request.Method {
	case http.MethodPost:
		decode := json.NewDecoder(request.Body)
		var registration Registration
		err := decode.Decode(&registration)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		err = reg.add(registration)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Println(fmt.Sprintf("Adding Service: %s with URL: %s", registration.ServerName, registration.ServerURL))
	case http.MethodDelete:
		reqData, err := ioutil.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		url := string(reqData)

		log.Println(fmt.Sprintf("removing service with URL: %s", url))

		err = reg.remove(url)
		if err != nil {
			log.Println(err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
	default:
		writer.WriteHeader(http.StatusNotFound)
	}
}
