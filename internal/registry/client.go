package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
)

func RegistryService(r Registration) error {
	serviceURL, err := url.Parse(r.ServerUpdateURL)
	if err != nil {
		return err
	}
	http.Handle(serviceURL.Path, &serviceUpdateHandler{})

	heartbeatUrl, err := url.Parse(r.HeartbeatUrl)
	if err != nil {
		return err
	}
	http.HandleFunc(heartbeatUrl.Path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	buff := new(bytes.Buffer)
	encode := json.NewEncoder(buff)
	err = encode.Encode(&r)
	if err != nil {
		return err
	}
	rsp, err := http.Post(ServicesUrl, "application/json", buff)
	if err != nil {
		return err
	}

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to registry sevice. registry service responsed with code = %v", rsp.StatusCode)
	}

	return nil
}

type serviceUpdateHandler struct{}

func (s serviceUpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	dec := json.NewDecoder(r.Body)
	var p patch
	err := dec.Decode(&p)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("Updated received %v\n", p)

	prov.Update(p)
}

func ShutDownService(url string) error {
	req, err := http.NewRequest(http.MethodDelete, ServicesUrl, bytes.NewBuffer([]byte(url)))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "text/plain")

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to deregister service. Registry service responded with code %d ", rsp.StatusCode)
	}

	return nil
}

type providers struct {
	services map[ServerName][]string
	mu       *sync.RWMutex
}

// 更新依赖
func (p *providers) Update(pat patch) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, entry := range pat.Added {
		if _, ok := p.services[entry.Name]; !ok {
			p.services[entry.Name] = make([]string, 0)
		}
		p.services[entry.Name] = append(p.services[entry.Name], entry.URL)
	}

	for _, entry := range pat.Removed {
		if providerURLs, ok := p.services[entry.Name]; ok {
			for i, provider := range providerURLs {
				if provider == entry.URL {
					p.services[entry.Name] = append(providerURLs[:i], providerURLs[i+1:]...)
				}
			}
		}
	}
}

func (p providers) get(name ServerName) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	provider, ok := p.services[name]
	if !ok {
		return "", fmt.Errorf("no providers available for service %v", name)
	}

	idx := int(rand.Float32() * float32(len(provider)))

	return provider[idx], nil
}

func GetProvider(name ServerName) (string, error) {
	return prov.get(name)
}

var prov = providers{
	services: make(map[ServerName][]string, 0),
	mu:       new(sync.RWMutex),
}
