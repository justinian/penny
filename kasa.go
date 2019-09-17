package main

import (
	"fmt"
	"strconv"

	"github.com/justinian/gokasa"
)

type powerStrip struct {
	powerStrip *gokasa.PowerStrip
}

func NewKasaPowerstrip(di *DeviceInfo) (Device, error) {
	strip, err := gokasa.NewPowerStrip(di.Connection)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to power strip: %v", err)
	}

	return &powerStrip{strip}, nil
}

func (kps *powerStrip) Subdevice(sub string) (PowerToggler, error) {
	idx, err := strconv.ParseInt(sub, 10, 0)
	if err != nil {
		return nil, fmt.Errorf("Invalid plug index: %s: %v", sub, err)
	}

	if int(idx) > len(kps.powerStrip.Plugs) {
		return nil, fmt.Errorf("Invalid plug index: %s out of bounds", sub)
	}

	return kps.powerStrip.Plugs[idx], nil
}
