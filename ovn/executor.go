package ovn

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/chrisschwa/ovn-viewer/config"

	"golang.org/x/crypto/ssh"
)

// Executor handles execution of OVN/OVS commands via SSH
type Executor struct {
	config config.OVNConfig
	client *ssh.Client
}

// NewExecutor creates a new command executor
func NewExecutor(cfg config.OVNConfig) *Executor {
	return &Executor{
		config: cfg,
	}
}

// Connect establishes SSH connection to the OVN controller
func (e *Executor) Connect() error {
	if e.client != nil {
		return nil
	}

	// Expand SSH key path
	keyPath := e.config.SSHKeyPath
	if strings.HasPrefix(keyPath, "~") {
		u, err := user.Current()
		if err != nil {
			return fmt.Errorf("failed to get current user: %w", err)
		}
		keyPath = strings.Replace(keyPath, "~", u.HomeDir, 1)
	}

	// Read SSH key
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read SSH key: %w", err)
	}
	key, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return fmt.Errorf("failed to parse SSH key: %w", err)
	}

	sshConfig := &ssh.ClientConfig{
		User:            e.config.ControllerUser,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(key)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", e.config.ControllerHost, e.config.SSHPort)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to OVN controller: %w", err)
	}

	e.client = client
	log.Printf("Connected to OVN controller at %s", addr)
	return nil
}

// Close closes the SSH connection
func (e *Executor) Close() error {
	if e.client != nil {
		return e.client.Close()
	}
	return nil
}

// IsConnected checks if the SSH connection is active
func (e *Executor) IsConnected() bool {
	return e.client != nil
}

// Execute runs an OVN/OVS command and returns the output
func (e *Executor) Execute(cmd string) (string, error) {
	return e.ExecuteWithContext(context.Background(), cmd)
}

// ExecuteWithContext runs a command with a context for timeout control
func (e *Executor) ExecuteWithContext(ctx context.Context, cmd string) (string, error) {
	if err := e.Connect(); err != nil {
		return "", err
	}

	session, err := e.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Duration(e.config.CommandTimeout)*time.Second)
	defer cancel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	// Wrap command with docker/podman exec if docker mode is enabled
	wrappedCmd := e.wrapWithDockerExec(cmd)

	err = session.Run(wrappedCmd)
	if err != nil {
		errMsg := stderr.String()
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("command '%s' failed: %s", cmd, errMsg)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// wrapWithDockerExec wraps a command with docker/podman exec if docker mode is enabled
func (e *Executor) wrapWithDockerExec(cmd string) string {
	if !e.config.DockerMode {
		return cmd
	}
	
	runtime := e.config.DockerRuntime
	if runtime == "" {
		runtime = "docker"
	}
	
	// Determine which container to use based on the command
	container := e.resolveContainer(cmd)
	if container == "" {
		// No container mapping found, run without wrapping
		return cmd
	}
	
	return fmt.Sprintf("%s exec %s %s", runtime, container, cmd)
}

// resolveContainer determines which container a command should run in
func (e *Executor) resolveContainer(cmd string) string {
	// Extract the base command (first word, handling paths)
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return ""
	}
	base := parts[0]
	// Strip path if present
	if idx := strings.LastIndex(base, "/"); idx != -1 {
		base = base[idx+1:]
	}
	
	// Check user-defined mapping first
	if e.config.DockerContainerMap != nil {
		if c, ok := e.config.DockerContainerMap[base]; ok {
			return c
		}
	}
	
	// Default container mapping for common OVN/OVS commands
	switch base {
	case "ovn-nbctl", "ovn-trace":
		return "ovn-northd"
	case "ovn-sbctl":
		return "ovn-southbound"
	case "ovs-vsctl", "ovs-ofctl":
		return "ovn-controller"
	default:
		return ""
	}
}

// ===== OVN-NBCTL COMMANDS =====

// ListLogicalRouters returns all logical routers
func (e *Executor) ListLogicalRouters() (string, error) {
	return e.Execute("ovn-nbctl --format=json --data=bare --no-heading --columns=name,uuid,_uuid list Logical_router")
}

// GetLogicalRouter returns details for a specific logical router
func (e *Executor) GetLogicalRouter(name string) (string, error) {
	return e.Execute(fmt.Sprintf("ovn-nbctl --format=json --data=bare --no-heading --columns=name,uuid,ports,static_routes,nat,policies,external_ids,options get Logical_router %s", name))
}

// ListLogicalSwitches returns all logical switches
func (e *Executor) ListLogicalSwitches() (string, error) {
	return e.Execute("ovn-nbctl --format=json --data=bare --no-heading --columns=name,uuid,_uuid list Logical_switch")
}

// GetLogicalSwitch returns details for a specific logical switch
func (e *Executor) GetLogicalSwitch(name string) (string, error) {
	return e.Execute(fmt.Sprintf("ovn-nbctl --format=json --data=bare --no-heading --columns=name,uuid,ports,acls,external_ids,options get Logical_switch %s", name))
}

// ListLogicalRouterPorts returns all logical router ports
func (e *Executor) ListLogicalRouterPorts() (string, error) {
	return e.Execute("ovn-nbctl --format=json --data=bare --no-heading --columns=name,uuid,networks,mac,external_ids list Logical_router_port")
}

// ListLogicalSwitchPorts returns all logical switch ports
func (e *Executor) ListLogicalSwitchPorts() (string, error) {
	return e.Execute("ovn-nbctl --format=json --data=bare --no-heading --columns=name,uuid,type,addresses,options,external_ids list Logical_switch_port")
}

// ListACLs returns all ACLs
func (e *Executor) ListACLs() (string, error) {
	return e.Execute("ovn-nbctl --format=json --data=bare --no-heading --columns=name,uuid,direction,priority,action,match,external_ids list ACL")
}

// ListACLsBySwitch returns ACLs for a specific switch
func (e *Executor) ListACLsBySwitch(switchName string) (string, error) {
	return e.Execute(fmt.Sprintf("ovn-nbctl --format=json --data=bare --no-heading --columns=uuid,direction,priority,action,match,external_ids list ACL where switch=%s", switchName))
}

// ListStaticRoutes returns all static routes
func (e *Executor) ListStaticRoutes() (string, error) {
	return e.Execute("ovn-nbctl --format=json --data=bare --no-heading --columns=uuid,policy,prefix,nexthop,output_port,external_ids list Logical_Route")
}

// ListNATRules returns all NAT rules
func (e *Executor) ListNATRules() (string, error) {
	return e.Execute("ovn-nbctl --format=json --data=bare --no-heading --columns=uuid,name,external_ip,internal_ip,type,protocol,logical_ip,options,external_ids list NAT")
}

// ListLoadBalancers returns all load balancers
func (e *Executor) ListLoadBalancers() (string, error) {
	return e.Execute("ovn-nbctl --format=json --data=bare --no-heading --columns=name,uuid,vips,protocol,options,external_ids list Load_Balancer")
}

// GetLoadBalancer returns details for a specific load balancer
func (e *Executor) GetLoadBalancer(name string) (string, error) {
	return e.Execute(fmt.Sprintf("ovn-nbctl --format=json get Load_Balancer %s", name))
}

// GetOVNVersion returns the OVN version
func (e *Executor) GetOVNVersion() (string, error) {
	return e.Execute("ovn-nbctl --version")
}

// ShowTopology returns the full OVN NB topology
func (e *Executor) ShowTopology() (string, error) {
	return e.Execute("ovn-nbctl show")
}

// GetLRPInfo returns logical router port info
func (e *Executor) GetLRPInfo(name string) (string, error) {
	return e.Execute(fmt.Sprintf("ovn-nbctl --format=json get Logical_router_port %s", name))
}

// GetLSPInfo returns logical switch port info
func (e *Executor) GetLSPInfo(name string) (string, error) {
	return e.Execute(fmt.Sprintf("ovn-nbctl --format=json get Logical_switch_port %s", name))
}

// ===== OVN-SBCTL COMMANDS =====

// ListChassis returns all chassis
func (e *Executor) ListChassis() (string, error) {
	return e.Execute("ovn-sbctl --format=json --data=bare --no-heading --columns=name,uuid,hostname,arch,transport_zones,encap_types,external_ids list Chassis")
}

// ListPortBindings returns all port bindings
func (e *Executor) ListPortBindings() (string, error) {
	return e.Execute("ovn-sbctl --format=json --data=bare --no-heading --columns=name,uuid,chassis,mac,addresses,type,external_ids list Port_Binding")
}

// GetPortBinding returns details for a specific port binding
func (e *Executor) GetPortBinding(name string) (string, error) {
	return e.Execute(fmt.Sprintf("ovn-sbctl --format=json get Port_Binding %s", name))
}

// ListLogicalFlows returns all logical flows
func (e *Executor) ListLogicalFlows() (string, error) {
	return e.Execute("ovn-sbctl --format=json list Logical_Flow")
}

// GetLogicalFlowsBySwitch returns logical flows for a specific switch
func (e *Executor) GetLogicalFlowsBySwitch(switchName string, table string) (string, error) {
	args := fmt.Sprintf("ovn-sbctl --format=json list Logical_Flow where switch=%s", switchName)
	if table != "" {
		args += fmt.Sprintf(" and table=%s", table)
	}
	return e.Execute(args)
}

// GetLogicalFlowsByRouter returns logical flows for a specific router
func (e *Executor) GetLogicalFlowsByRouter(routerName string) (string, error) {
	return e.Execute(fmt.Sprintf("ovn-sbctl --format=json list Logical_Flow where router=%s", routerName))
}

// GetChassisRedirect returns chassis redirect information
func (e *Executor) GetChassisRedirect() (string, error) {
	return e.Execute("ovn-sbctl --format=json list Chassis_Redirect")
}

// GetGatewayChassis returns gateway chassis information
func (e *Executor) GetGatewayChassis() (string, error) {
	return e.Execute("ovn-sbctl --format=json list Gateway_Chassis")
}

// ===== OVN-TRACE COMMANDS =====

// TracePacket executes ovn-trace to trace a packet through the OVN pipeline
// This is the core troubleshooting command
func (e *Executor) TracePacket(lr string, lsp string, pkt string, checkRod bool) (string, error) {
	cmd := fmt.Sprintf("ovn-trace --disable-ct --check=%s %s %s", lr, lsp, pkt)
	if checkRod {
		cmd += " --rod"
	}
	return e.Execute(cmd)
}

// TracePacketWithRouter traces packet with router context using the router syntax
func (e *Executor) TracePacketWithRouter(router string, lrp string, pkt string, checkRod bool) (string, error) {
	cmd := fmt.Sprintf("ovn-trace --disable-ct %s %s %s", router, lrp, pkt)
	if checkRod {
		cmd += " --rod"
	}
	return e.Execute(cmd)
}

// TracePacketFromLs traces packet starting from a logical switch
func (e *Executor) TracePacketFromLs(sw string, lsp string, pkt string, checkRod bool) (string, error) {
	cmd := fmt.Sprintf("ovn-trace --disable-ct %s %s %s", sw, lsp, pkt)
	if checkRod {
		cmd += " --rod"
	}
	return e.Execute(cmd)
}

// RadiusOfDarkness checks the Radius of Darkness for a given flow
func (e *Executor) RadiusOfDarkness(lr string, lsp string, pkt string) (string, error) {
	cmd := fmt.Sprintf("ovn-sbctl rod %s %s %s", lr, lsp, pkt)
	return e.Execute(cmd)
}

// ===== OVS COMMANDS =====

// ListBridges returns all OVS bridges
func (e *Executor) ListBridges() (string, error) {
	return e.Execute("ovs-vsctl --format=json list Bridge")
}

// ListInterfaces returns all OVS interfaces
func (e *Executor) ListInterfaces() (string, error) {
	return e.Execute("ovs-vsctl --format=json list Interface")
}

// GetBridge returns details for a specific bridge
func (e *Executor) GetBridge(name string) (string, error) {
	return e.Execute(fmt.Sprintf("ovs-vsctl --format=json get Bridge %s", name))
}

// ListOVSFlows returns all flows for a bridge
func (e *Executor) ListOVSFlows(bridge string) (string, error) {
	return e.Execute(fmt.Sprintf("ovs-ofctl dump-flows %s", bridge))
}

// GetOVSFlowsByTable returns flows for a specific table
func (e *Executor) GetOVSFlowsByTable(bridge string, table int) (string, error) {
	return e.Execute(fmt.Sprintf("ovs-ofctl dump-flows %s table=%d", bridge, table))
}

// GetOVSVersion returns the OVS version
func (e *Executor) GetOVSVersion() (string, error) {
	return e.Execute("ovs-vsctl --version")
}

// ShowOVS shows the OVS topology
func (e *Executor) ShowOVS() (string, error) {
	return e.Execute("ovs-vsctl show")
}

// GetMeterStatistics returns meter statistics
func (e *Executor) GetMeterStatistics() (string, error) {
	return e.Execute("ovn-nbctl --format=json list Meter")
}

// GetQoSRules returns QoS rules for a switch
func (e *Executor) GetQoSRules(switchName string) (string, error) {
	return e.Execute(fmt.Sprintf("ovn-nbctl --format=json list QoS where port=%s", switchName))
}

// ===== HELPER FUNCTIONS =====

// ParseJSONOutput helps parse JSON array output from OVN commands
func ParseJSONOutput(output string) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	err := json.Unmarshal([]byte(output), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return result, nil
}

// FormatCommandOutput formats raw command output for display
func FormatCommandOutput(output string) string {
	lines := strings.Split(output, "\n")
	var formatted []string
	for i, line := range lines {
		if i == 0 {
			formatted = append(formatted, line)
		} else {
			formatted = append(formatted, fmt.Sprintf("  %s", line))
		}
	}
	return strings.Join(formatted, "\n")
}

// TableInfo represents information about an OVN pipeline table
type TableInfo struct {
	Name     string `json:"name"`
	ID       int    `json:"id"`
	Priority int    `json:"priority"`
	Match    string `json:"match"`
	Action   string `json:"action"`
}

// ParseTableID converts a table name or ID to its numeric ID
func ParseTableID(table string) int {
	if id, err := strconv.Atoi(table); err == nil {
		return id
	}
	// Common OVN pipeline table names (logical switch pipeline)
	tableMap := map[string]int{
		"port_sec":       0,
		"in_port":        1,
		"lb_skip":        2,
		"acl_in_host":    3,
		"lb_source":      4,
		"acl_in_port":    5,
		"lb":             6,
		"acl_in_datapath": 7,
		"lrstat":         8,
		"ldap":           9,
		"reroute":        10,
		"dup":            11,
		"acl_out":        80,
		"meter":          81,
		"ratelimit":      90,
		"port_sec_out":   91,
		"ipsec_out":      92,
	}
	if id, ok := tableMap[table]; ok {
		return id
	}
	return -1
}