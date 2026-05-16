package ovn

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/chrisschwa/ovn-viewer/config"
)

// MockExecutor executes OVN/OVS commands using built-in mock data (no SSH required)
type MockExecutor struct {
	config   config.OVNConfig
	scenario string
	mu       sync.Mutex
}

// NewMockExecutor creates a new mock executor
func NewMockExecutor(cfg config.OVNConfig) *MockExecutor {
	scenario := os.Getenv("MOCK_OVN_SCENARIO")
	if scenario == "" {
		scenario = "default"
	}

	return &MockExecutor{
		config:   cfg,
		scenario: scenario,
	}
}

// Connect is a no-op for mock executor
func (e *MockExecutor) Connect() error {
	fmt.Printf("Mock mode: loaded scenario '%s'\n", e.scenario)
	return nil
}

// Close is a no-op
func (e *MockExecutor) Close() error {
	return nil
}

// IsConnected always returns true for mock
func (e *MockExecutor) IsConnected() bool {
	return true
}

// Execute runs a command through the mock (returns empty output)
func (e *MockExecutor) Execute(cmd string) (string, error) {
	return "", fmt.Errorf("mock mode requires ovn-mock server (see docs)")
}

// ExecuteWithContext delegates to Execute
func (e *MockExecutor) ExecuteWithContext(ctx context.Context, cmd string) (string, error) {
	return e.Execute(cmd)
}

// ===== OVN-NBCTL COMMANDS =====

func (e *MockExecutor) ListLogicalRouters() (string, error) {
	return e.Execute("ovn-nbctl list Logical_router")
}

func (e *MockExecutor) GetLogicalRouter(name string) (string, error) {
	return e.Execute(fmt.Sprintf("ovn-nbctl get Logical_router %s", name))
}

func (e *MockExecutor) ListLogicalSwitches() (string, error) {
	return e.Execute("ovn-nbctl list Logical_switch")
}

func (e *MockExecutor) GetLogicalSwitch(name string) (string, error) {
	return e.Execute(fmt.Sprintf("ovn-nbctl get Logical_switch %s", name))
}

func (e *MockExecutor) ListLogicalRouterPorts() (string, error) {
	return e.Execute("ovn-nbctl list Logical_router_port")
}

func (e *MockExecutor) ListLogicalSwitchPorts() (string, error) {
	return e.Execute("ovn-nbctl list Logical_switch_port")
}

func (e *MockExecutor) ListACLs() (string, error) {
	return e.Execute("ovn-nbctl list ACL")
}

func (e *MockExecutor) ListACLsBySwitch(switchName string) (string, error) {
	return e.Execute(fmt.Sprintf("ovn-nbctl list ACL where switch=%s", switchName))
}

func (e *MockExecutor) ListStaticRoutes() (string, error) {
	return e.Execute("ovn-nbctl list Logical_Route")
}

func (e *MockExecutor) ListNATRules() (string, error) {
	return e.Execute("ovn-nbctl list NAT")
}

func (e *MockExecutor) ListLoadBalancers() (string, error) {
	return e.Execute("ovn-nbctl list Load_Balancer")
}

func (e *MockExecutor) GetLoadBalancer(name string) (string, error) {
	return e.Execute(fmt.Sprintf("ovn-nbctl get Load_Balancer %s", name))
}

func (e *MockExecutor) GetOVNVersion() (string, error) {
	return e.Execute("ovn-nbctl --version")
}

func (e *MockExecutor) ShowTopology() (string, error) {
	return e.Execute("ovn-nbctl show")
}

func (e *MockExecutor) GetLRPInfo(name string) (string, error) {
	return e.Execute(fmt.Sprintf("ovn-nbctl get Logical_router_port %s", name))
}

func (e *MockExecutor) GetLSPInfo(name string) (string, error) {
	return e.Execute(fmt.Sprintf("ovn-nbctl get Logical_switch_port %s", name))
}

// ===== OVN-SBCTL COMMANDS =====

func (e *MockExecutor) ListChassis() (string, error) {
	return e.Execute("ovn-sbctl list Chassis")
}

func (e *MockExecutor) ListPortBindings() (string, error) {
	return e.Execute("ovn-sbctl list Port_Binding")
}

func (e *MockExecutor) GetPortBinding(name string) (string, error) {
	return e.Execute(fmt.Sprintf("ovn-sbctl get Port_Binding %s", name))
}

func (e *MockExecutor) ListLogicalFlows() (string, error) {
	return e.Execute("ovn-sbctl list Logical_Flow")
}

func (e *MockExecutor) GetLogicalFlowsBySwitch(switchName string, table string) (string, error) {
	args := fmt.Sprintf("ovn-sbctl list Logical_Flow where switch=%s", switchName)
	if table != "" {
		args += fmt.Sprintf(" and table=%s", table)
	}
	return e.Execute(args)
}

func (e *MockExecutor) GetLogicalFlowsByRouter(routerName string) (string, error) {
	return e.Execute(fmt.Sprintf("ovn-sbctl list Logical_Flow where router=%s", routerName))
}

func (e *MockExecutor) GetChassisRedirect() (string, error) {
	return e.Execute("ovn-sbctl list Chassis_Redirect")
}

func (e *MockExecutor) GetGatewayChassis() (string, error) {
	return e.Execute("ovn-sbctl list Gateway_Chassis")
}

// ===== OVN-TRACE COMMANDS =====

func (e *MockExecutor) TracePacket(lr string, lsp string, pkt string, checkRod bool) (string, error) {
	return e.Execute(fmt.Sprintf("ovn-trace %s %s %s", lr, lsp, pkt))
}

func (e *MockExecutor) TracePacketWithRouter(router string, lrp string, pkt string, checkRod bool) (string, error) {
	return e.Execute(fmt.Sprintf("ovn-trace %s %s %s", router, lrp, pkt))
}

func (e *MockExecutor) TracePacketFromLs(sw string, lsp string, pkt string, checkRod bool) (string, error) {
	return e.Execute(fmt.Sprintf("ovn-trace %s %s %s", sw, lsp, pkt))
}

func (e *MockExecutor) RadiusOfDarkness(lr string, lsp string, pkt string) (string, error) {
	return e.Execute(fmt.Sprintf("ovn-sbctl rod %s %s %s", lr, lsp, pkt))
}

// ===== OVS COMMANDS =====

func (e *MockExecutor) ListBridges() (string, error) {
	return e.Execute("ovs-vsctl list Bridge")
}

func (e *MockExecutor) ListInterfaces() (string, error) {
	return e.Execute("ovs-vsctl list Interface")
}

func (e *MockExecutor) GetBridge(name string) (string, error) {
	return e.Execute(fmt.Sprintf("ovs-vsctl get Bridge %s", name))
}

func (e *MockExecutor) ListOVSFlows(bridge string) (string, error) {
	return e.Execute(fmt.Sprintf("ovs-ofctl dump-flows %s", bridge))
}

func (e *MockExecutor) GetOVSFlowsByTable(bridge string, table int) (string, error) {
	return e.Execute(fmt.Sprintf("ovs-ofctl dump-flows %s table=%d", bridge, table))
}

func (e *MockExecutor) GetOVSVersion() (string, error) {
	return e.Execute("ovs-vsctl --version")
}

func (e *MockExecutor) ShowOVS() (string, error) {
	return e.Execute("ovs-vsctl show")
}

func (e *MockExecutor) GetMeterStatistics() (string, error) {
	return e.Execute("ovn-nbctl list Meter")
}

func (e *MockExecutor) GetQoSRules(switchName string) (string, error) {
	return e.Execute(fmt.Sprintf("ovn-nbctl list QoS where port=%s", switchName))
}