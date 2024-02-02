package helpers

import (
	"fmt"
	"net"

	corev1 "k8s.io/api/core/v1"
)

// GetPortOrDefault returns the port from an address or a default value.
func GetPortOrDefault(defaultValue int, addr *net.TCPAddr) int {
	if addr != nil {
		return addr.Port
	}
	return defaultValue
}

// CheckProbePort checks that the probe port matches the http port.
func CheckProbePort(port int, probe *corev1.Probe) {
	if probe == nil {
		return
	}

	if probe.ProbeHandler.HTTPGet == nil {
		return
	}

	probePort := probe.ProbeHandler.HTTPGet.Port.IntVal
	if int(probePort) != port {
		panic(fmt.Sprintf(`probe port %d does not match http port %d`, probePort, port))
	}
}
