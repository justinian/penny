package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	metricValue = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "penny",
			Subsystem: "sensors",
			Name:      "value",
			Help:      "Value read from one of penny's sensors",
		},
		[]string{"sensor"},
	)
)

func init() {
	prometheus.MustRegister(metricValue)
}

type SensorInfo struct {
	Name string `toml:"name"`
	Kind string `toml:"kind"`

	ReadFrom string  `toml:"read_from"`
	Target   float64 `toml:"target"`
	Range    float64 `toml:"range"`

	Device    string `toml:"device"`
	Subdevice string `toml:"subdevice"`

	device PowerToggler
}

type valueReader func(filename string) (float64, error)

var kinds = map[string]valueReader{
	"w1-therm": readW1Therm,
}

func (s *SensorInfo) Read() (float64, error) {
	reader, ok := kinds[s.Kind]
	if !ok {
		return 0, fmt.Errorf("Unknown sensor kind: %s", s.Kind)
	}

	return reader(s.ReadFrom)
}

func (s *SensorInfo) Update() error {
	val, err := s.Read()
	if err != nil {
		return err
	}

	metricValue.WithLabelValues(s.Name).Set(val)

	if s.device != nil {
		if !s.device.IsOn() && val < (s.Target-s.Range) {
			log.Printf("%s read %g - turning on.", s.Name, val)
			return s.device.On()
		} else if s.device.IsOn() && val > (s.Target+s.Range) {
			log.Printf("%s read %g - turning off.", s.Name, val)
			return s.device.Off()
		}
	}

	return nil
}

func readW1Therm(filename string) (float64, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("%s wrong format: %d lines", filename, len(lines))
	}

	if !strings.HasSuffix(lines[0], "YES") {
		return 0, fmt.Errorf("%s not ready", filename)
	}

	parts := strings.Split(lines[1], " ")
	tempString := parts[len(parts)-1]
	if !strings.HasPrefix(tempString, "t=") {
		return 0, fmt.Errorf("%s wrong format: no temperature found", filename)
	}

	temp, err := strconv.ParseFloat(tempString[2:], 64)
	if err != nil {
		return 0, fmt.Errorf("%s wrong format: %v", filename, err)
	}

	return temp / 1000.0, nil
}
