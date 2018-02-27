package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type countryCache interface {
	Put(ip string, country string, expires time.Time)
	Get(ip string) (string, bool)
}

// serviceConfig contains configuration for country lookup service
type serviceConfig struct {
	URL       string   // base url with %v placeholder for IP address
	ReplyPath []string // path to country code within returned JSON
	Rate      int      // maximum requests per period
	Period    int      // period length in seconds
	Burst     int      // how many requests can be sent at maximum speed, defaults to Rate
}

// appConfig contains configuration for the application
type appConfig struct {
	ListenAddress   string          // listening address for HTTP
	CacheLifetime   int             // how long to keep service lookup result, in seconds
	CacheType       string          // type of cache used
	CacheParameters interface{}     // cache parameters (like passwords etc)
	Services        []serviceConfig // list of lookup services
}

type service struct {
	url       string
	replyPath []string
}

// lookup performs a call to detect country through given service
func (s *service) lookup(ip string) (string, error) {
	url := fmt.Sprintf(s.url, ip)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return decodeResponse(body, s.replyPath)
}

func decodeResponse(s []byte, path []string) (string, error) {
	var v interface{}
	err := json.Unmarshal(s, &v)
	if err != nil {
		return "", err
	}
	for _, field := range path {
		m, ok := v.(map[string]interface{})
		if !ok {
			return "", errors.New("path not found")
		}
		v, ok = m[field]
		if !ok {
			return "", errors.New("path not found")
		}
	}
	if s, ok := v.(string); ok {
		return s, nil
	}
	return "", errors.New("path is not a string")
}

func loadConfig(path string) (*appConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	var config appConfig
	if err := dec.Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

type context struct {
	cache         countryCache
	router        router
	cacheLifetime time.Duration
}

func writeResponse(w http.ResponseWriter, ip, country string) {
	w.Header().Set("Content-Type", "application/json")
	if b, err := json.Marshal(struct{ IP, Country string }{ip, country}); err != nil {
		log.Printf("unexpected error while encoding %v to json: %v", country, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.Write(b)
	}
}

func (c *context) httpHandler(w http.ResponseWriter, r *http.Request) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("error %v while parsing RemoteAddr %v", err, r.RemoteAddr)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if country, ok := c.cache.Get(ip); ok {
		writeResponse(w, ip, country)
		return
	}
	now := time.Now()
	svc, err := c.router.get(now)
	if err != nil {
		log.Printf("error from service router: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	country, err := svc.lookup(ip)
	if err != nil {
		log.Printf("error from service: %v", err) // FIXME: it would be nice to include service name
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.cache.Put(ip, country, now.Add(c.cacheLifetime))
	writeResponse(w, ip, country)
}

func createCache(cacheType string, params interface{}) (countryCache, error) {
	switch cacheType {
	case "builtin":
		return newBuiltinCache(params)
	case "postgres":
		return newPgCache(params)
	default:
		return nil, errors.New("unsupported cache type")
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("configuration file is not given")
	}
	config, err := loadConfig(os.Args[1])
	if err != nil {
		log.Fatalf("error while reading configuration file: %v", err)
	}
	cache, err := createCache(config.CacheType, config.CacheParameters)
	if err != nil {
		log.Fatalf("error while creating cache: %v", err)
	}
	context := context{
		cache:         cache,
		router:        *newServiceRouter(config.Services),
		cacheLifetime: time.Duration(config.CacheLifetime) * time.Second,
	}
	http.HandleFunc("/", context.httpHandler)
	log.Fatal(http.ListenAndServe(config.ListenAddress, nil))
}
