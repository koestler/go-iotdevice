// Package mqtttopic implements the matching logic of MQTT topics and routes accroding to the mqtt 5.0 specification.
// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901241
package mqtttopic

import "strings"

func Match(route, topic string) bool {
	if route == topic {
		return true
	}

	routeElems := routeSplit(route)

	// MQTT-4.7.2-1: wildcards at the first level must not match topics starting with $
	if strings.HasPrefix(topic, "$") && (routeElems[0] == "#" || routeElems[0] == "+") {
		return false
	}

	topicElems := strings.Split(topic, "/")

	return matchElements(routeElems, topicElems)
}

func matchElements(route []string, topic []string) bool {
	if len(route) == 0 {
		return len(topic) == 0
	}

	if len(topic) == 0 {
		return route[0] == "#"
	}

	if route[0] == "#" {
		return true
	}

	if (route[0] == "+") || (route[0] == topic[0]) {
		return matchElements(route[1:], topic[1:])
	}
	return false
}

func routeSplit(route string) []string {
	parts := strings.Split(route, "/")
	if len(parts) >= 3 && parts[0] == "$share" {
		// skip $share/{group}
		return parts[2:]
	}
	return parts
}
