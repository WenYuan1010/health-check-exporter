package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var healthURLs string
var listenAddr string
var timeoutSeconds int
var labels string // 新的标签参数

func init() {
	flag.StringVar(&healthURLs, "health-urls", "", "Comma-separated list of URLs for health check endpoints")
	flag.StringVar(&listenAddr, "listen-addr", ":8080", "listen address")
	flag.IntVar(&timeoutSeconds, "timeout-seconds", 1, "Timeout in seconds for HTTP requests")
	flag.StringVar(&labels, "labels", "unknown", "Comma-separated list of labels for the health URLs")
	flag.Parse()
}

// HealthResponse 表示来自 /actuator/health 端点的响应
type HealthResponse struct {
	Status string `json:"status"`
}

// CustomCollector 是一个自定义的 Prometheus 收集器
type CustomCollector struct {
	statusGauge       *prometheus.GaugeVec
	systemHealthGauge *prometheus.GaugeVec
}

func (c *CustomCollector) Describe(ch chan<- *prometheus.Desc) {
	c.statusGauge.Describe(ch)
	c.systemHealthGauge.Describe(ch)
}

func (c *CustomCollector) Collect(ch chan<- prometheus.Metric) {
	urls := strings.Split(healthURLs, ",")
	labelValues := make(map[string]float64) // 用于存储每个标签的健康状态
	labelList := strings.Split(labels, ",") // 将标签参数解析为列表
	for _, label := range labelList {
		labelValues[label] = 1.0
	}
	// Create a new HTTP client with a timeout
	httpClient := &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}

	for _, url := range urls {
		// 提取状态值并将其设置为 Gauge 值
		statusValue := 0.0
		if strings.Contains(url, "health") {
			//健康检查探针

			// 获取健康端点并解析响应
			resp, err := httpClient.Get(url)
			if err != nil {
				fmt.Println("获取健康端点时出错:", err)
				c.statusGauge.WithLabelValues(url).Set(statusValue)
				for _, label := range labelList {
					if strings.Contains(url, strings.Replace(label, ":", "", -1)) {
						labelValues[label] = math.Min(labelValues[label], statusValue) // 取最小值
					}
				}
				continue
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("读取响应体时出错:", err)
				c.statusGauge.WithLabelValues(url).Set(statusValue)
				for _, label := range labelList {
					if strings.Contains(url, label) {
						labelValues[label] = math.Min(labelValues[label], statusValue) // 取最小值
					}
				}
				continue
			}

			var healthResp HealthResponse
			err = json.Unmarshal(body, &healthResp)
			if err != nil {
				fmt.Println("解析 JSON 时出错:", err)
				c.statusGauge.WithLabelValues(url).Set(statusValue)
				for _, label := range labelList {
					if strings.Contains(url, label) {
						labelValues[label] = math.Min(labelValues[label], statusValue) // 取最小值
					}
				}
				continue
			}

			// Output the status and corresponding URL
			fmt.Printf("URL: %s, 状态: %s\n", url, healthResp.Status)

			if healthResp.Status == "UP" {
				statusValue = 1.0
			}
			c.statusGauge.WithLabelValues(url).Set(statusValue)

			for _, label := range labelList {
				if strings.Contains(url, label) {
					labelValues[label] = math.Min(labelValues[label], statusValue) // 取最小值
				}
			}
		} else {
			//http探针

			// 发送HTTP GET请求
			resp, err := httpClient.Get(url)
			if err != nil {
				fmt.Printf("%v\n", err)
				c.statusGauge.WithLabelValues(url).Set(statusValue)
				continue
			}
			defer resp.Body.Close()

			// 检查响应状态码
			if resp.StatusCode == http.StatusOK {
				fmt.Printf("%s可达,状态码:%d\n", url, resp.StatusCode)
				statusValue = 1.0
				c.statusGauge.WithLabelValues(url).Set(statusValue)
			} else {
				fmt.Printf("%s不可达,状态码:%d\n", url, resp.StatusCode)
				c.statusGauge.WithLabelValues(url).Set(statusValue)
			}
		}
	}

	// 设置系统健康指标的值
	for label, value := range labelValues {
		c.systemHealthGauge.WithLabelValues(label).Set(value)
	}

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
		systemHealthGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "system_health",
			Help: "System health status (1 if all application_health metrics are 1, 0 otherwise).",
		}, []string{"label"}),
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
