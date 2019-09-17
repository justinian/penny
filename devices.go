package main

import (
	"fmt"
)

type PowerToggler interface {
	IsOn() bool
	On() error
	Off() error
}

type Device interface {
	Subdevice(string) (PowerToggler, error)
}

type DeviceInfo struct {
	Name       string `toml:"name"`
	Kind       string `toml:"kind"`
	Connection string `toml:"connection"`
}

type Registry struct {
	devices map[string]Device
}

func NewRegistry(config *Config) (*Registry, error) {
	r := &Registry{
		devices: make(map[string]Device),
	}

	var err error
	for _, di := range config.Devices {
		var device Device
		switch di.Kind {
		case "kasa":
			device, err = NewKasaPowerstrip(di)
		default:
			device = nil
		}

		if err != nil {
			return nil, fmt.Errorf("Error initializing device %s: %v", di.Name, err)
		}

		r.devices[di.Name] = device
	}

	return r, nil
}

func (r *Registry) InitSensor(s *SensorInfo) error {
	if s.Device == "" {
		return nil
	}

	d, ok := r.devices[s.Device]
	if !ok {
		return fmt.Errorf("Device %s not found", s.Device)
	}

	sd, err := d.Subdevice(s.Subdevice)
	if err != nil {
		return fmt.Errorf("Subdevice %s: %v ", s.Subdevice, err)
	}

	s.device = sd
	return nil
}
