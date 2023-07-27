package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var healthURLs string

func init() {
	// Define and parse command-line flags
	flag.StringVar(&healthURLs, "health-urls", "", "Comma-separated list of URLs for health check endpoints")
	flag.Parse()
}

// HealthResponse represents the response from the /actuator/health endpoint
type HealthResponse struct {
	Status string `json:"status"`
}

// CustomCollector is a custom Prometheus collector
type CustomCollector struct {
	statusGauge *prometheus.GaugeVec
}

func (c *CustomCollector) Describe(ch chan<- *prometheus.Desc) {
	c.statusGauge.Describe(ch)
}

func (c *CustomCollector) Collect(ch chan<- prometheus.Metric) {
	urls := strings.Split(healthURLs, ",")
	for _, url := range urls {
		// Extract the status value and set it as the Gauge value
		statusValue := 0.0

		// Fetch the health endpoint and parse the response
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("Error fetching health endpoint:", err)
			c.statusGauge.WithLabelValues(url).Set(statusValue)
			continue
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			continue
		}

		var healthResp HealthResponse
		err = json.Unmarshal(body, &healthResp)
		if err != nil {
			fmt.Println("Error unmarshalling JSON:", err)
			continue
		}

		// Print the health status and URL
		fmt.Printf("URL: %s, Health Status: %s\n", url, healthResp.Status)

		if healthResp.Status == "UP" {
			statusValue = 1.0
		}
		c.statusGauge.WithLabelValues(url).Set(statusValue)
	}

	// Collect all metrics and register them at once
	c.statusGauge.Collect(ch)
}

func newCustomCollector() *CustomCollector {
	return &CustomCollector{
		statusGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "application_health",
			Help: "Status of the application health.",
		}, []string{"url"}),
	}
}

func main() {
	// Check if healthURLs are provided
	if healthURLs == "" {
		panic("healthURLs flag is required")
	}

	// Register the custom collector
	customCollector := newCustomCollector()
	prometheus.MustRegister(customCollector)

	// Expose metrics via HTTP
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}
