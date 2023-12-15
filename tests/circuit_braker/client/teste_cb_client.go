package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
)

// On client side, we defind a simple function to call the upstream service.
func DoReq() error {
	resp, err := http.Get("http://localhost:8080/ping")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("bad response")
	}

	return nil
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
