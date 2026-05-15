package models

// ===== OVN NORTHBOUND ENTITIES =====

// LogicalRouter represents an OVN Logical Router
type LogicalRouter struct {
	UUID             string             `json:"uuid"`
	Name             string             `json:"name"`
	ExternalIDs      map[string]string  `json:"external_ids"`
	Options          map[string]string  `json:"options"`
	StaticRoutes     []StaticRoute      `json:"static_routes"`
	NATRules         []NATRule          `json:"nat_rules"`
	Policies         []RouterPolicy     `json:"policies"`
	PortNames        []string           `json:"ports"`
	PortUUIDs        []string           `json:"port_uuids"`
	CopyARPForRoutes bool               `json:"copy_arp_for_routes"`
}

// LogicalSwitch represents an OVN Logical Switch
type LogicalSwitch struct {
	UUID         string            `json:"uuid"`
	Name         string            `json:"name"`
	ExternalIDs  map[string]string `json:"external_ids"`
	Options      map[string]string `json:"options"`
	Ports        []string          `json:"ports"`
	PortUUIDs    []string          `json:"port_uuids"`
	ACLs         []string          `json:"acls"`
	FakePorts    []string          `json:"fake_ports"`
	OtherConfig  map[string]string `json:"other_config"`
}

// LogicalRouterPort represents a port on a Logical Router
type LogicalRouterPort struct {
	UUID        string            `json:"uuid"`
	Name        string            `json:"name"`
	Networks    []string          `json:"networks"`
	MAC         string            `json:"mac"`
	Options     map[string]string `json:"options"`
	ExternalIDs map[string]string `json:"external_ids"`
}

// LogicalSwitchPort represents a port on a Logical Switch
type LogicalSwitchPort struct {
	UUID        string            `json:"uuid"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Options     map[string]string `json:"options"`
	ExternalIDs map[string]string `json:"external_ids"`
	Addresses   []string          `json:"addresses"`
	Peer        string            `json:"peer"`
}

// ACL represents an OVN Logical ACL
type ACL struct {
	UUID        string            `json:"uuid"`
	ExternalIDs map[string]string `json:"external_ids"`
	Direction   string            `json:"direction"`     // from-lport, to-lport, to-lport-delimited
	Priority    int               `json:"priority"`
	Action      string            `json:"action"`        // allow, allow-related, deny, deny-related, related, related-and-established, established, accept, drop, reject
	Meter       string            `json:"meter"`
	IPVersion   string            `json:"ip_version"`
	Log         bool              `json:"log"`
	LogOptions  map[string]string `json:"log_options"`
	Severity    string            `json:"severity"`
	Match       string            `json:"match"`
	Switch      string            `json:"switch_uuid"`
	SwitchName  string            `json:"switch_name"`
}

// StaticRoute represents a static route in a Logical Router
type StaticRoute struct {
	UUID          string            `json:"uuid"`
	Policy        string            `json:"policy"`          // simple, routable, source
	Prefix        string            `json:"prefix"`
	Nexthop       string            `json:"nexthop"`
	OutputIP      string            `json:"output_ip"`
	OutputPort    string            `json:"output_port"`
	ExternalIDs   map[string]string `json:"external_ids"`
}

// NATRule represents a NAT rule in a Logical Router
type NATRule struct {
	UUID        string            `json:"uuid"`
	Name        string            `json:"name"`
	ExternalIDs map[string]string `json:"external_ids"`
	ExternalIP  string            `json:"external_ip"`
	InternalIP  string            `json:"internal_ip"`
	Type        string            `json:"type"`        // snat, dnat, dnat_and_snat
	Protocol    string            `json:"protocol"`    // tcp, udp
	LogicalIP   string            `json:"logical_ip"`
	Options     map[string]string `json:"options"`
}

// RouterPolicy represents a routing policy in a Logical Router
type RouterPolicy struct {
	UUID       string            `json:"uuid"`
	Priority   int               `json:"priority"`
	Policy     string            `json:"policy"`      // next_hop, routetable
	Match      string            `json:"match"`
	Action     string            `json:"action"`      // allow, drop
	Nexthop    string            `json:"nexthop"`
	ExternalIDs map[string]string `json:"external_ids"`
}

// LoadBalancer represents an OVN Load Balancer
type LoadBalancer struct {
	UUID        string            `json:"uuid"`
	Name        string            `json:"name"`
	ExternalIDs map[string]string `json:"external_ids"`
	VPNs        []string          `json:"vips"`
	Protocol    string            `json:"protocol"`
	Options     map[string]string `json:"options"`
	VIPs        map[string]string `json:"vips_map"`
}

// ===== OVN SOUTHBOUND ENTITIES =====

// Chassis represents an OVN Chassis (hypervisor with OVS+OVN agent)
type Chassis struct {
	UUID        string            `json:"uuid"`
	Name        string            `json:"name"`
	Hostname    string            `json:"hostname"`
	Arch        string            `json:"arch"`
	TransportZones []string       `json:"transport_zones"`
	EncapTypes  []string          `json:"encap_types"`
	ExternalIDs map[string]string `json:"external_ids"`
	Operators   []string          `json:"operators"`
}

// LogicalFlow represents a compiled logical flow in the OVN pipeline
type LogicalFlow struct {
	UUID        string            `json:"uuid"`
	Table       string            `json:"table"`
	Priority    int               `json:"priority"`
	Match       string            `json:"match"`
	Action      string            `json:"action"`
	Type        string            `json:"type"`        // chassis, logical, lsp, lr
	Done        bool              `json:"done"`
	ExternalIDs map[string]string `json:"external_ids"`
	Severity    string            `json:"severity"`
	Skipped     bool              `json:"skipped"`
}

// PortBinding represents a port binding in the SB DB
type PortBinding struct {
	UUID        string            `json:"uuid"`
	PortName    string            `json:"port_name"`
	Chassis     []string          `json:"chassis"`
	MAC         []string          `json:"mac"`
	Addresses   []string          `json:"addresses"`
	Type        string            `json:"type"`
	Options     map[string]string `json:"options"`
	ExternalIDs map[string]string `json:"external_ids"`
	ChassisRequest string         `json:"chassis_request"`
}

// ===== OVS ENTITIES =====

// Bridge represents an OVS Bridge
type Bridge struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	ExternalIDs map[string]string `json:"external_ids"`
	Ports       []string          `json:"ports"`
	OtherConfig map[string]string `json:"other_config"`
}

// Interface represents an OVS Interface
type Interface struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	ExternalIDs map[string]string `json:"external_ids"`
	Statistics  map[string]int    `json:"statistics"`
	OtherConfig map[string]string `json:"other_config"`
	Status      map[string]string `json:"status"`
	MtuRequest  int               `json:"mtu_request"`
}

// OVSFlow represents a flow in the OVS pipeline
type OVSFlow struct {
	Cookie    string `json:"cookie"`
	Duration  float64 `json:"duration"`
	TableID   int    `json:"table_id"`
	Priority  int    `json:"priority"`
	Reg       string `json:"reg"`
	Match     string `json:"match"`
	Action    string `json:"action"`
	DurationDetails string `json:"duration_details"`
}

// ===== PACKET TRACE ENTITIES =====

// TraceRequest represents a packet trace request
type TraceRequest struct {
	Lr      string `json:"lr"`             // logical router name
	Lsp     string `json:"lsp"`            // logical switch port name
	InPort  string `json:"inport"`         // input port
	OutPort string `json:"outport"`        // output port (optional)
	Pkt     string `json:"pkt"`            // packet specification
	Switch  string `json:"switch"`         // logical switch name
	Router  string `json:"router"`         // logical router name
	Lrp     string `json:"lrp"`            // logical router port name
	CheckRod bool   `json:"check_rod"`     // include Radius of Darkness check
}

// TraceResult represents the result of a packet trace
type TraceResult struct {
	Success       bool              `json:"success"`
	Output        string            `json:"output"`
	SrcIP         string            `json:"src_ip"`
	DstIP         string            `json:"dst_ip"`
	SrcMAC        string            `json:"src_mac"`
	DstMAC        string            `json:"dst_mac"`
	Protocol      string            `json:"protocol"`
	SrcPort       int               `json:"src_port"`
	DstPort       int               `json:"dst_port"`
	LosToLrouter  bool              `json:"ls_to_lr"`
	LosToLswitch  bool              `json:"ls_to_ls"`
	Verdict       string            `json:"verdict"`
	Flows         []LogicalFlow     `json:"flows"`
	RadiusOfDarkness RadiusOfDarknessResult `json:"radius_of_darkness"`
	Error         string            `json:"error,omitempty"`
}

// RadiusOfDarknessResult represents the ROD check result
type RadiusOfDarknessResult struct {
	Pass        bool    `json:"pass"`
	MaxPassLoss float64 `json:"max_pass_loss"`
	MaxDropLoss float64 `json:"max_drop_loss"`
	Loss        float64 `json:"loss"`
	Margin      float64 `json:"margin"`
}

// ===== DASHBOARD & SUMMARY =====

// DashboardSummary provides an overview of the OVN environment
type DashboardSummary struct {
	TotalRouters       int `json:"total_routers"`
	TotalSwitches      int `json:"total_switches"`
	TotalPorts         int `json:"total_ports"`
	TotalACLs          int `json:"total_acls"`
	TotalChassis       int `json:"total_chassis"`
	TotalNATRules      int `json:"total_nat_rules"`
	TotalStaticRoutes  int `json:"total_static_routes"`
	TotalLoadBalancers int `json:"total_load_balancers"`
}

// ConnectionInfo holds info about the current OVN connection
type ConnectionInfo struct {
	Connected       bool   `json:"connected"`
	ControllerHost  string `json:"controller_host"`
	NBDBConnected   bool   `json:"nbdb_connected"`
	SBDBConnected   bool   `json:"sbdb_connected"`
	OVNVersion      string `json:"ovn_version"`
	OVSVersion      string `json:"ovs_version"`
}