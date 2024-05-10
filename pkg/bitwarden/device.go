package bitwarden

import (
	"fmt"
	"math/rand"
	"runtime"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type DeviceType uint

// Device represents a device
type Device struct {
	Name       string     `json:"deviceName"`
	Identifier string     `json:"deviceIdentifier,omitempty"`
	Type       DeviceType `json:"deviceType,omitempty"`
	PushToken  string     `json:"devicePushToken,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	tmplIdentifier = `xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx`
)

const (
	Android DeviceType = iota
	iOS
	ChromeExtension
	FirefoxExtension
	OperaExtension
	EdgeExtension
	WindowsDesktop
	MacOsDesktop
	LinuxDesktop
	ChromeBrowser
	FirefoxBrowser
	OperaBrowser
	EdgeBrowser
	IEBrowser
	UnknownBrowser
	AndroidAmazon
	UWP
	SafariBrowser
	VivaldiBrowser
	VivaldiExtension
	SafariExtension
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Return a new device with a random identifier
func NewDevice(name string) *Device {
	return NewDeviceEx(deviceType(), MakeDeviceIdentifier(), name, "")
}

// Return a new device with a known identifier and type
func NewDeviceEx(deviceType DeviceType, deviceIdentifier, name, pushToken string) *Device {
	return &Device{
		Type:       deviceType,
		Identifier: deviceIdentifier,
		Name:       name,
		PushToken:  pushToken,
	}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (d Device) String() string {
	str := "<device"
	if d.Type != UnknownBrowser {
		str += fmt.Sprintf(" type=%v", d.Type)
	}
	if d.Identifier != "" {
		str += fmt.Sprintf(" identifier=%q", d.Identifier)
	}
	if d.Name != "" {
		str += fmt.Sprintf(" name=%q", d.Name)
	}
	if d.PushToken != "" {
		str += fmt.Sprintf(" pushToken=%q", d.PushToken)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// MakeDeviceIdentifier creates a randomly-assigned device identifier
func MakeDeviceIdentifier() string {
	var uuid string
	for _, r := range tmplIdentifier {
		if r == 'x' || r == 'y' {
			uuid += fmt.Sprintf("%01x", rand.Intn(16))
		} else {
			uuid += string(r)
		}
	}
	return uuid
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func deviceType() DeviceType {
	switch runtime.GOOS {
	case "linux":
		return LinuxDesktop
	case "darwin":
		return MacOsDesktop
	case "windows":
		return WindowsDesktop
	default:
		return UnknownBrowser
	}
}
