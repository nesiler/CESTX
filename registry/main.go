package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nesiler/cestx/common"
	"github.com/robfig/cron/v3"
)

type Service struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Address     string      `json:"address"`
	Port        int         `json:"port"`
	HealthCheck HealthCheck `json:"healthCheck"`
}

type HealthCheck struct {
	Endpoint string `json:"endpoint"`
	Interval string `json:"interval"`
	Timeout  string `json:"timeout"`
}

type Config struct {
	ExternalServices map[string]ServiceInfo `json:"externalServices"`
}

type ServiceInfo struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
	DBName   string `json:"dbname,omitempty"`
}

var (
	rdb        *redis.Client
	ctx        = context.Background()
	configData Config
	c          = cron.New()
)

func main() {
	// Load configuration file
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		common.Fatal("Error reading config file: %v\n", err)
	}

	err = json.Unmarshal(configFile, &configData)
	if err != nil {
		common.Fatal("Error parsing config file: %v\n", err)
	}

	// Initialize Redis client using config data
	redisConfig, ok := configData.ExternalServices["redis"]
	if !ok {
		common.Fatal("Redis configuration not found in config file\n")
	}

	rdb = redis.NewClient(&redis.Options{
		Addr: redisConfig.Host + ":" + strconv.Itoa(redisConfig.Port),
	})

	http.HandleFunc("/register", registerServiceHandler)
	http.HandleFunc("/service/", getServiceHandler)
	http.HandleFunc("/config/", getConfigHandler)

	// Start the cron scheduler
	c.Start()

	common.Info("Server started on 192.168.4.63:3434")
	log.Fatal(http.ListenAndServe("192.168.4.63:3434", nil))
}

func registerServiceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var service Service
	err := json.NewDecoder(r.Body).Decode(&service)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	registerService(service)
	scheduleHealthCheck(service)
	w.WriteHeader(http.StatusOK)
}

func registerService(service Service) {
	common.Info("Registering service: %s", service.Name)

	serviceData, err := json.Marshal(service)
	if err != nil {
		common.Fatal("Error serializing service data: %v\n", err)
	}

	err = rdb.Set(ctx, "service:"+service.ID, serviceData, 0).Err()
	if err != nil {
		common.Fatal("Error storing service data in Redis: %v\n", err)
	}
}

func getServiceHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Path[len("/service/"):]

	serviceData, err := rdb.Get(ctx, "service:"+serviceID).Result()
	if err == redis.Nil {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(serviceData))
}

func getConfigHandler(w http.ResponseWriter, r *http.Request) {
	configName := r.URL.Path[len("/config/"):]

	serviceInfo, ok := configData.ExternalServices[configName]
	if !ok {
		http.Error(w, "Configuration not found", http.StatusNotFound)
		return
	}

	serviceInfoData, err := json.Marshal(serviceInfo)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(serviceInfoData)
}

func scheduleHealthCheck(service Service) {
	interval, err := time.ParseDuration(service.HealthCheck.Interval)
	if err != nil {
		common.Fatal("Error parsing interval: %v\n", err)
	}

	cronSpec := "@every " + interval.String()
	c.AddFunc(cronSpec, func() {
		monitorService(service)
	})
}

func monitorService(service Service) {
	resp, err := http.Get("http://" + service.Address + ":" + strconv.Itoa(service.Port) + service.HealthCheck.Endpoint)
	status := "unhealthy"
	if err == nil && resp.StatusCode == http.StatusOK {
		status = "healthy"
	}

	err = rdb.HSet(ctx, "service:"+service.ID, "status", status).Err()
	if err != nil {
		common.Warn("Error updating status for service %s: %v", service.Name, err)
	}
}