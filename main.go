package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var healthURLs string
var listenAddr string
var timeoutSeconds int

func init() {
	flag.StringVar(&healthURLs, "health-urls", "", "Comma-separated list of URLs for health check endpoints")
	flag.StringVar(&listenAddr, "listen-addr", ":8080", "listen address")
	flag.IntVar(&timeoutSeconds, "timeout-seconds", 1, "Timeout in seconds for HTTP requests")
	flag.Parse()
}

// HealthResponse 表示来自 /actuator/health 端点的响应
type HealthResponse struct {
	Status string `json:"status"`
}

// CustomCollector 是一个自定义的 Prometheus 收集器
type CustomCollector struct {
	statusGauge       *prometheus.GaugeVec
	systemHealthGauge prometheus.Gauge
}

func (c *CustomCollector) Describe(ch chan<- *prometheus.Desc) {
	c.statusGauge.Describe(ch)
	ch <- c.systemHealthGauge.Desc()
}

func (c *CustomCollector) Collect(ch chan<- prometheus.Metric) {
	urls := strings.Split(healthURLs, ",")
	systemHealth := 1.0 // 假设系统健康
	// Create a new HTTP client with a timeout
	httpClient := &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}

	for _, url := range urls {
		// 提取状态值并将其设置为 Gauge 值
		statusValue := 0.0

		// 获取健康端点并解析响应
		resp, err := httpClient.Get(url)
		if err != nil {
			fmt.Println("获取健康端点时出错:", err)
			c.statusGauge.WithLabelValues(url).Set(statusValue)
			systemHealth = 0.0 // 如果有一个URL不健康，系统健康就为0
			continue
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("读取响应体时出错:", err)
			systemHealth = 0.0 // 如果有一个URL不健康，系统健康就为0
			continue
		}

		var healthResp HealthResponse
		err = json.Unmarshal(body, &healthResp)
		if err != nil {
			fmt.Println("解析 JSON 时出错:", err)
			systemHealth = 0.0 // 如果有一个URL不健康，系统健康就为0
			continue
		}

		// Output the status and corresponding URL
		fmt.Printf("URL: %s, 状态: %s\n", url, healthResp.Status)

		if healthResp.Status == "UP" {
			statusValue = 1.0
		} else {
			systemHealth = 0.0 // 如果有一个应用程序状态不为 1，设置系统健康为 0
		}
		c.statusGauge.WithLabelValues(url).Set(statusValue)
	}

	// 设置系统健康指标的值
	c.systemHealthGauge.Set(systemHealth)

	// 收集所有指标后一次性注册
	c.statusGauge.Collect(ch)
	c.systemHealthGauge.Collect(ch)
}

func newCustomCollector() *CustomCollector {
	return &CustomCollector{
		statusGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "application_health",
			Help: "Status of the application health.",
		}, []string{"url"}),
		systemHealthGauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "system_health",
			Help: "System health status (1 if all application_health metrics are 1, 0 otherwise).",
		}),
	}
}

func main() {
	// 检查是否提供了 healthURLs
	if healthURLs == "" {
		panic("必须提供 healthURLs 标志")
	}

	// 注册自定义收集器
	customCollector := newCustomCollector()
	prometheus.MustRegister(customCollector)

	// 通过 HTTP 公开指标
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(listenAddr, nil)
}
