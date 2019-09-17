package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/justinian/gokasa"
)

type Powerstrip struct {
	Hostname string `toml:"hostname"`
}

type Config struct {
	Interval   string     `toml:"interval"`
	Powerstrip Powerstrip `toml:"power_strip"`
	Sensors    []*Sensor  `toml:"sensors"`

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

	strip, err := gokasa.NewPowerStrip(config.Powerstrip.Hostname)
	if err != nil {
		return fmt.Errorf("Error connecting to power strip: %v", err)
	}

	for _, s := range config.Sensors {
		if s.PlugIndex < 0 || s.PlugIndex >= len(strip.Plugs) {
			return fmt.Errorf("Device %s has invalid plug index %d. (Max %d)",
				s.Name, s.PlugIndex, len(strip.Plugs)-1)
		}
		s.plug = strip.Plugs[s.PlugIndex]

		log.Printf("Sensor %s configured. Range: %g - %g.", s.Name,
			s.Target-s.Range, s.Target+s.Range)
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
