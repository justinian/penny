# Penny Tank Manager

`penny` is a manager for terrarium environmental control and monitoring.
Sensors are exported as Prometheus metrics, and can also be configured to
control devices based on their readings.

This project is named for its primary user: my Ball Python, Penelope.

## Sensor support

Sensors are currently read as files on the filesystem. This makes it easy to
read from devices via existing Linux drivers like `w1-therm`. The currently
supported file formats are:

* `w1-therm` For thermometers on a OneWire bus. (Here's an [article about
  hooking up a DS18B20 to a Raspberry Pi][ds18b20] via the `w1-therm` driver.)

[ds18b20]: https://thepihut.com/blogs/raspberry-pi-tutorials/ds18b20-one-wire-digital-temperature-sensor-and-the-raspberry-pi

## Device control support

Supported device control:

* TP-Link Kasa smart power strip via [gokasa][] 

[gokasa]: https://github.com/justinian/gokasa
