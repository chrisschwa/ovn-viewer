package ovn

import (
	"context"
	"github.com/chrisschwa/ovn-viewer/config"
)

// ExecutorInterface defines the interface for executing OVN/OVS commands
type ExecutorInterface interface {
	Connect() error
	Close() error
	IsConnected() bool
	Execute(cmd string) (string, error)
	ExecuteWithContext(ctx context.Context, cmd string) (string, error)

	// OVN-NBCTL
	ListLogicalRouters() (string, error)
	GetLogicalRouter(name string) (string, error)
	ListLogicalSwitches() (string, error)
	GetLogicalSwitch(name string) (string, error)
	ListLogicalRouterPorts() (string, error)
	ListLogicalSwitchPorts() (string, error)
	ListACLs() (string, error)
	ListACLsBySwitch(switchName string) (string, error)
	ListStaticRoutes() (string, error)
	ListNATRules() (string, error)
	ListLoadBalancers() (string, error)
	GetLoadBalancer(name string) (string, error)
	GetOVNVersion() (string, error)
	ShowTopology() (string, error)
	GetLRPInfo(name string) (string, error)
	GetLSPInfo(name string) (string, error)

	// OVN-SBCTL
	ListChassis() (string, error)
	ListPortBindings() (string, error)
	GetPortBinding(name string) (string, error)
	ListLogicalFlows() (string, error)
	GetLogicalFlowsBySwitch(switchName string, table string) (string, error)
	GetLogicalFlowsByRouter(routerName string) (string, error)
	GetChassisRedirect() (string, error)
	GetGatewayChassis() (string, error)

	// OVN-TRACE
	TracePacket(lr string, lsp string, pkt string, checkRod bool) (string, error)
	TracePacketWithRouter(router string, lrp string, pkt string, checkRod bool) (string, error)
	TracePacketFromLs(sw string, lsp string, pkt string, checkRod bool) (string, error)
	RadiusOfDarkness(lr string, lsp string, pkt string) (string, error)

	// OVS
	ListBridges() (string, error)
	ListInterfaces() (string, error)
	GetBridge(name string) (string, error)
	ListOVSFlows(bridge string) (string, error)
	GetOVSFlowsByTable(bridge string, table int) (string, error)
	GetOVSVersion() (string, error)
	ShowOVS() (string, error)
	GetMeterStatistics() (string, error)
	GetQoSRules(switchName string) (string, error)
}

// NewExecutorFactory creates the appropriate executor based on configuration
func NewExecutorFactory(cfg config.OVNConfig) ExecutorInterface {
	return NewExecutor(cfg)
}

// NewMockExecutorFactory creates a mock executor
func NewMockExecutorFactory(cfg config.OVNConfig) ExecutorInterface {
	return NewMockExecutor(cfg)
}