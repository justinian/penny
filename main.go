package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Sensors []*SensorInfo `toml:"sensors"`
	Devices []*DeviceInfo `toml:"devices"`

	Interval       string `toml:"interval"`
	MetricsAddress string `toml:"metrics_address"`
}

func run(configPath string) error {
	log.Printf("Penny starting, config: %s", configPath)

	var config Config
	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		return fmt.Errorf("Error reading config: %v", err)
	}

	interval, err := time.ParseDuration(config.Interval)
	if err != nil {
		return fmt.Errorf("Invalid interval %s: %v", config.Interval, err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	registry, err := NewRegistry(&config)
	if err != nil {
		return fmt.Errorf("Failed to initialize devices: %v", err)
	}

	for _, s := range config.Sensors {
		err = registry.InitSensor(s)
		if err != nil {
			return fmt.Errorf("Sensor %s failed device binding: %v", s.Name, err)
		}

		if s.device == nil {
			log.Printf("Sensor %s configured, logging only.", s.Name)
		} else {
			log.Printf("Sensor %s configured. Range: %g - %g.", s.Name,
				s.Target-s.Range, s.Target+s.Range)
		}
	}

	log.Printf("Penny running, updating every %v", interval)

	go serveMetrics(config)

	for {
		select {
		case <-ticker.C:
			for _, s := range config.Sensors {
				if err := s.Update(); err != nil {
					log.Printf("Sensor %s error: %v", s.Name, err)
				}
			}

		case <-sigs:
			log.Printf("Received interrupt, shutting down")
			return nil
		}
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <config file>\n", os.Args[0])
		os.Exit(1)
	}

	err := run(os.Args[1])
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
