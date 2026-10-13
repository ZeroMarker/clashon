// Package sysproxy points the OS proxy settings at the local mixed port.
package sysproxy

import "errors"

// Bypass lists hosts that should not go through the proxy.
var Bypass = []string{"localhost", "127.0.0.0/8", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "::1", "*.local"}

// ErrUnsupported means no supported proxy settings backend was found.
var ErrUnsupported = errors.New("system proxy is not supported on this desktop")

// Enable sets HTTP, HTTPS and SOCKS proxies to host:port.
func Enable(host string, port int) error { return enable(host, port) }

// Disable turns the system proxy off.
func Disable() error { return disable() }

// Enabled reports whether the system proxy currently points at host:port.
func Enabled(host string, port int) (bool, error) { return enabled(host, port) }
