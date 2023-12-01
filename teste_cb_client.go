package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/sony/gobreaker"
)

// On client side, we defind a simple function to call the upstream service.
func DoReq() error {
	var respBody []byte
	resp, err := http.Get("http://localhost:8080/metrics")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	convert, err := prom2json(string(respBody))
	if err != nil {
		return err
	}

	fmt.Println(string(convert))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("bad response")
	}

	return nil
}

func prom2json(str string) ([]byte, error) {
	parser := &expfmt.TextParser{}
	families, err := parser.TextToMetricFamilies(strings.NewReader(str))
	if err != nil {
		return nil, err
	}

	out := make(map[string][]map[string]map[string]any)

	for key, val := range families {
		family := out[key]

		for _, m := range val.GetMetric() {
			metric := make(map[string]any)
			for _, label := range m.GetLabel() {
				metric[label.GetName()] = label.GetValue()
			}
			switch val.GetType() {
			case dto.MetricType_COUNTER:
				metric["value"] = m.GetCounter().GetValue()
			case dto.MetricType_GAUGE:
				metric["value"] = m.GetGauge().GetValue()
			case dto.MetricType_SUMMARY:
				metric["value"] = m.GetSummary().GetQuantile()
			case dto.MetricType_HISTOGRAM:
				metric["value"] = m.GetHistogram().GetBucket()
			default:
				return nil, fmt.Errorf("unsupported type: %v", val.GetType())
			}
			family = append(family, map[string]map[string]any{
				val.GetName(): metric,
			})
		}

		out[key] = family
	}

	output, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("failed to encode json: %w", err)
	}

	return output, nil
}

func DoDb() error {
	resp, err := http.Get("http://localhost:8080/ping")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("bad response")
	}

	return nil
}

func main() {
	// call with circuit breaker
	cb := gobreaker.NewCircuitBreaker(
		gobreaker.Settings{
			Name:        "circuit-breaker-teste",
			MaxRequests: 3,
			Timeout:     3 * time.Second,
			Interval:    1 * time.Second,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > 3
			},
			OnStateChange: func(name string, from, to gobreaker.State) {
				fmt.Printf("CircuitBreaker '%s' changed from '%s' to '%s'\n", name, from, to)
			},
		},
	)
	// fmt.Println("Call with circuit breaker")
	for {
		_, err := cb.Execute(func() (interface{}, error) {
			err := DoReq()
			return nil, err
		})
		if err != nil {
			fmt.Println(err.Error() + " for client teste")
		}
		fmt.Println("Call with circuit breaker")
		time.Sleep(300 * time.Millisecond)
	}
}
