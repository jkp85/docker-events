package main

var (
	CONTAINER_EVENTS = map[string]uint8{"attach": 1, "commit": 1, "copy": 1, "create": 1, "destroy": 1, "detach": 1,
		"die": 1, "exec_create": 1, "exec_detach": 1, "exec_start": 1, "export": 1, "kill": 1, "oom": 1, "pause": 1,
		"rename": 1, "resize": 1, "restart": 1, "start": 1, "stop": 1, "top": 1, "unpause": 1, "update": 1}
	IMAGE_EVENTS = map[string]uint8{"delete": 1, "import": 1, "load": 1, "pull": 1, "push": 1, "save": 1, "tag": 1,
		"untag": 1}
	VOLUME_EVENTS  = map[string]uint8{"create": 1, "mount": 1, "unmount": 1, "destroy": 1}
	NETWORK_EVENTS = map[string]uint8{"create": 1, "connect": 1, "disconnect": 1, "destroy": 1}
	DAEMON_EVENTS  = map[string]uint8{"reload": 1}
	SERVICE_EVENTS = map[string]uint8{"remove": 1}
)

func validateEvent(eventType, name string) bool {
	ok := false
	switch eventType {
	case "container":
		_, ok = CONTAINER_EVENTS[name]
	case "image":
		_, ok = IMAGE_EVENTS[name]
	case "volume":
		_, ok = VOLUME_EVENTS[name]
	case "network":
		_, ok = NETWORK_EVENTS[name]
	case "daemon":
		_, ok = DAEMON_EVENTS[name]
	case "service":
		_, ok = SERVICE_EVENTS[name]
	}
	return ok
}
