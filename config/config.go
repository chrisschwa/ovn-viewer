package config

// Config holds the application configuration
type Config struct {
	Server ServerConfig `json:"server"`
	OVN    OVNConfig    `json:"ovn"`
}

// ServerConfig holds the HTTP server configuration
type ServerConfig struct {
	Port string `json:"port"`
	Host string `json:"host"`
}

// OVNConfig holds the OVN connection configuration
type OVNConfig struct {
	// SSH connection to the controller hosting OVN services
	ControllerHost string `json:"controller_host"`
	ControllerUser string `json:"controller_user"`
	SSHPort        int    `json:"ssh_port"`
	SSHKeyPath     string `json:"ssh_key_path"`

	// OVN database connections (optional, for direct access without SSH)
	NBDBAddress string `json:"nbdb_address"`
	SBDBAddress string `json:"sbdb_address"`

	// Timeout for OVN commands in seconds
	CommandTimeout int `json:"command_timeout"`

	// Docker execution mode: when true, wraps OVN/OVS commands with docker/podman exec
	// e.g., "docker exec ovn-northd ovn-nbctl list Logical_router"
	// When enabled, commands are mapped to containers:
	//   ovn-nbctl, ovn-trace → ovn-northd container
	//   ovn-sbctl             → ovn-southbound container
	//   ovs-vsctl, ovs-ofctl  → ovn-controller container
	// Container names are taken from DockerContainerMap (defaults shown above).
	DockerMode           bool              `json:"docker_mode"`
	DockerRuntime        string            `json:"docker_runtime"` // "docker" or "podman" (default: "docker")
	DockerContainerMap   map[string]string `json:"docker_container_map"` // command → container name
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Port: "8080",
			Host: "0.0.0.0",
		},
		OVN: OVNConfig{
			ControllerHost: "localhost",
			ControllerUser: "ovs",
			SSHPort:        22,
			SSHKeyPath:     "~/.ssh/id_rsa",
			CommandTimeout: 30,
		},
	}
}