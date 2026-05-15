package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"ovn-troubleshooter/config"
	"ovn-troubleshooter/models"
	"ovn-troubleshooter/ovn"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handler holds the HTTP handlers
type Handler struct {
	executor ovn.ExecutorInterface
	config   config.Config
	mu       sync.RWMutex
	clients  map[*websocket.Conn]bool
}

// NewHandler creates a new handler with the given executor
func NewHandler(executor ovn.ExecutorInterface, cfg config.Config) *Handler {
	return &Handler{
		executor: executor,
		config:   cfg,
		clients:  make(map[*websocket.Conn]bool),
	}
}

// ===== ERROR RESPONSE HELPER =====

func handleError(c *gin.Context, status int, err error) {
	c.JSON(status, gin.H{
		"error": err.Error(),
	})
}

func handleRawOutput(c *gin.Context, output string, parseJSON bool) {
	if parseJSON {
		var jsonData interface{}
		if err := json.Unmarshal([]byte(output), &jsonData); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"raw_output": output,
			})
		} else {
			c.JSON(http.StatusOK, jsonData)
		}
	} else {
		c.JSON(http.StatusOK, gin.H{
			"output": output,
		})
	}
}

// ===== DASHBOARD ENDPOINTS =====

// GetDashboard returns a summary of the OVN environment
func (h *Handler) GetDashboard(c *gin.Context) {
	var summary models.DashboardSummary
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	setError := func(err error) {
		mu.Lock()
		defer mu.Unlock()
		if firstErr == nil {
			firstErr = err
		}
	}

	// Count routers
	wg.Add(1)
	go func() {
		defer wg.Done()
		output, err := h.executor.ListLogicalRouters()
		if err != nil {
			setError(err)
			return
		}
		summary.TotalRouters = countEntities(output)
	}()

	// Count switches
	wg.Add(1)
	go func() {
		defer wg.Done()
		output, err := h.executor.ListLogicalSwitches()
		if err != nil {
			setError(err)
			return
		}
		summary.TotalSwitches = countEntities(output)
	}()

	// Count ports
	wg.Add(1)
	go func() {
		defer wg.Done()
		output, err := h.executor.ListLogicalSwitchPorts()
		if err != nil {
			setError(err)
			return
		}
		summary.TotalPorts = countEntities(output)
	}()

	// Count ACLs
	wg.Add(1)
	go func() {
		defer wg.Done()
		output, err := h.executor.ListACLs()
		if err != nil {
			setError(err)
			return
		}
		summary.TotalACLs = countEntities(output)
	}()

	// Count chassis
	wg.Add(1)
	go func() {
		defer wg.Done()
		output, err := h.executor.ListChassis()
		if err != nil {
			setError(err)
			return
		}
		summary.TotalChassis = countEntities(output)
	}()

	// Count NAT rules
	wg.Add(1)
	go func() {
		defer wg.Done()
		output, err := h.executor.ListNATRules()
		if err != nil {
			setError(err)
			return
		}
		summary.TotalNATRules = countEntities(output)
	}()

	// Count static routes
	wg.Add(1)
	go func() {
		defer wg.Done()
		output, err := h.executor.ListStaticRoutes()
		if err != nil {
			setError(err)
			return
		}
		summary.TotalStaticRoutes = countEntities(output)
	}()

	// Count load balancers
	wg.Add(1)
	go func() {
		defer wg.Done()
		output, err := h.executor.ListLoadBalancers()
		if err != nil {
			setError(err)
			return
		}
		summary.TotalLoadBalancers = countEntities(output)
	}()

	wg.Wait()

	if firstErr != nil {
		handleError(c, http.StatusInternalServerError, firstErr)
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetConnectionInfo returns info about the current OVN connection
func (h *Handler) GetConnectionInfo(c *gin.Context) {
	info := models.ConnectionInfo{
		Connected:      h.executor.IsConnected(),
		ControllerHost: h.config.OVN.ControllerHost,
	}

	var ovnVersion, ovsVersion string
	var ovnErr, ovsErr error

	go func() {
		ovnVersion, ovnErr = h.executor.GetOVNVersion()
	}()

	go func() {
		ovsVersion, ovsErr = h.executor.GetOVSVersion()
	}()

	// Wait a bit for versions
	if ovnErr == nil {
		info.OVNVersion = ovnVersion
		info.NBDBConnected = true
	}
	if ovsErr == nil {
		info.OVSVersion = ovsVersion
		info.SBDBConnected = true
	}

	info.Connected = info.NBDBConnected && info.SBDBConnected

	c.JSON(http.StatusOK, info)
}

// ===== ROUTER ENDPOINTS =====

// ListRouters returns all logical routers
func (h *Handler) ListRouters(c *gin.Context) {
	output, err := h.executor.ListLogicalRouters()
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	handleRawOutput(c, output, true)
}

// GetRouter returns details for a specific router
func (h *Handler) GetRouter(c *gin.Context) {
	name := c.Param("name")
	output, err := h.executor.GetLogicalRouter(name)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	handleRawOutput(c, output, true)
}

// GetRouterFlows returns logical flows for a router
func (h *Handler) GetRouterFlows(c *gin.Context) {
	name := c.Param("name")
	output, err := h.executor.GetLogicalFlowsByRouter(name)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	handleRawOutput(c, output, true)
}

// GetRouterStaticRoutes returns static routes for a router
func (h *Handler) GetRouterStaticRoutes(c *gin.Context) {
	output, err := h.executor.ListStaticRoutes()
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	handleRawOutput(c, output, true)
}

// GetRouterNATRules returns NAT rules for a router
func (h *Handler) GetRouterNATRules(c *gin.Context) {
	output, err := h.executor.ListNATRules()
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	handleRawOutput(c, output, true)
}

// ===== SWITCH ENDPOINTS =====

// ListSwitches returns all logical switches
func (h *Handler) ListSwitches(c *gin.Context) {
	output, err := h.executor.ListLogicalSwitches()
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	handleRawOutput(c, output, true)
}

// GetSwitch returns details for a specific switch
func (h *Handler) GetSwitch(c *gin.Context) {
	name := c.Param("name")
	output, err := h.executor.GetLogicalSwitch(name)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	handleRawOutput(c, output, true)
}

// GetSwitchFlows returns logical flows for a switch
func (h *Handler) GetSwitchFlows(c *gin.Context) {
	name := c.Param("name")
	table := c.Query("table")
	output, err := h.executor.GetLogicalFlowsBySwitch(name, table)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	handleRawOutput(c, output, true)
}

// GetSwitchACLs returns ACLs for a switch
func (h *Handler) GetSwitchACLs(c *gin.Context) {
	name := c.Param("name")
	output, err := h.executor.ListACLsBySwitch(name)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	handleRawOutput(c, output, true)
}

// ===== PORT ENDPOINTS =====

// ListRouterPorts returns all logical router ports
func (h *Handler) ListRouterPorts(c *gin.Context) {
	output, err := h.executor.ListLogicalRouterPorts()
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	handleRawOutput(c, output, true)
}

// ListSwitchPorts returns all logical switch ports
func (h *Handler) ListSwitchPorts(c *gin.Context) {
	output, err := h.executor.ListLogicalSwitchPorts()
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	handleRawOutput(c, output, true)
}

// GetRouterPort returns details for a specific router port
func (h *Handler) GetRouterPort(c *gin.Context) {
	name := c.Param("name")
	output, err := h.executor.GetLRPInfo(name)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	handleRawOutput(c, output, true)
}

// GetSwitchPort returns details for a specific switch port
func (h *Handler) GetSwitchPort(c *gin.Context) {
	name := c.Param("name")
	output, err := h.executor.GetLSPInfo(name)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	handleRawOutput(c, output, true)
}

// ===== ACL ENDPOINTS =====

// ListACLs returns all ACLs
func (h *Handler) ListACLs(c *gin.Context) {
	output, err := h.executor.ListACLs()
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	handleRawOutput(c, output, true)
}

// ===== CHASSIS ENDPOINTS =====

// ListChassis returns all chassis
func (h *Handler) ListChassis(c *gin.Context) {
	output, err := h.executor.ListChassis()
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	handleRawOutput(c, output, true)
}

// ===== PORT BINDING ENDPOINTS =====

// ListPortBindings returns all port bindings
func (h *Handler) ListPortBindings(c *gin.Context) {
	output, err := h.executor.ListPortBindings()
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	handleRawOutput(c, output, true)
}

// GetPortBinding returns details for a specific port binding
func (h *Handler) GetPortBinding(c *gin.Context) {
	name := c.Param("name")
	output, err := h.executor.GetPortBinding(name)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	handleRawOutput(c, output, true)
}

// ===== FLOW ENDPOINTS =====

// ListAllFlows returns all logical flows
func (h *Handler) ListAllFlows(c *gin.Context) {
	output, err := h.executor.ListLogicalFlows()
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	handleRawOutput(c, output, true)
}

// ===== OVS ENDPOINTS =====

// ListBridges returns all OVS bridges
func (h *Handler) ListBridges(c *gin.Context) {
	output, err := h.executor.ListBridges()
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	handleRawOutput(c, output, true)
}

// GetBridge returns details for a specific bridge
func (h *Handler) GetBridge(c *gin.Context) {
	name := c.Param("name")
	output, err := h.executor.GetBridge(name)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	handleRawOutput(c, output, true)
}

// ListInterfaces returns all OVS interfaces
func (h *Handler) ListInterfaces(c *gin.Context) {
	output, err := h.executor.ListInterfaces()
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	handleRawOutput(c, output, true)
}

// GetOVSFlows returns OVS flows for a bridge
func (h *Handler) GetOVSFlows(c *gin.Context) {
	bridge := c.Param("bridge")
	output, err := h.executor.ListOVSFlows(bridge)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"output": output,
	})
}

// ShowOVS shows the OVS topology
func (h *Handler) ShowOVS(c *gin.Context) {
	output, err := h.executor.ShowOVS()
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"output": output,
	})
}

// ShowTopology shows the OVN NB topology
func (h *Handler) ShowTopology(c *gin.Context) {
	output, err := h.executor.ShowTopology()
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"output": output,
	})
}

// ===== PACKET TRACE ENDPOINTS =====

// TracePacket traces a packet through the OVN pipeline
func (h *Handler) TracePacket(c *gin.Context) {
	var req models.TraceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	var output string
	var err error

	// Determine which trace method to use based on request parameters
	if req.Router != "" && req.Lrp != "" {
		output, err = h.executor.TracePacketWithRouter(req.Router, req.Lrp, req.Pkt, req.CheckRod)
	} else if req.Switch != "" && req.Lsp != "" {
		output, err = h.executor.TracePacketFromLs(req.Switch, req.Lsp, req.Pkt, req.CheckRod)
	} else if req.Lr != "" && req.Lsp != "" {
		output, err = h.executor.TracePacket(req.Lr, req.Lsp, req.Pkt, req.CheckRod)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "must specify either (router + lrp) or (switch + lsp) or (lr + lsp)",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Parse the trace output to extract useful information
	traceResult := parseTraceOutput(output)
	traceResult.Output = output
	traceResult.Success = true

	c.JSON(http.StatusOK, traceResult)
}

// RadiusOfDarkness checks the Radius of Darkness
func (h *Handler) RadiusOfDarkness(c *gin.Context) {
	var req models.TraceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	if req.Lr == "" || req.Lsp == "" || req.Pkt == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "lr, lsp, and pkt are required",
		})
		return
	}

	output, err := h.executor.RadiusOfDarkness(req.Lr, req.Lsp, req.Pkt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"output": output,
	})
}

// ===== AUTOCOMPLETE ENDPOINTS =====

// GetRouterNames returns all router names for autocomplete
func (h *Handler) GetRouterNames(c *gin.Context) {
	output, err := h.executor.ListLogicalRouters()
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	names := extractNames(output)
	c.JSON(http.StatusOK, names)
}

// GetSwitchNames returns all switch names for autocomplete
func (h *Handler) GetSwitchNames(c *gin.Context) {
	output, err := h.executor.ListLogicalSwitches()
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	names := extractNames(output)
	c.JSON(http.StatusOK, names)
}

// GetPortNames returns all port names for autocomplete
func (h *Handler) GetPortNames(c *gin.Context) {
	var routerPorts, switchPorts []string

	output1, err := h.executor.ListLogicalRouterPorts()
	if err == nil {
		routerPorts = extractNames(output1)
	}

	output2, err := h.executor.ListLogicalSwitchPorts()
	if err == nil {
		switchPorts = extractNames(output2)
	}

	result := gin.H{
		"router_ports": routerPorts,
		"switch_ports": switchPorts,
	}

	c.JSON(http.StatusOK, result)
}

// ===== WEBSOCKET ENDPOINT =====

// WebSocket handles WebSocket connections for live updates
func (h *Handler) WebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket: %v", err)
		return
	}

	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, conn)
		h.mu.Unlock()
		conn.Close()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

// Broadcast sends a message to all connected WebSocket clients
func (h *Handler) Broadcast(msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal broadcast message: %v", err)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		if err := client.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("Failed to send to client: %v", err)
			client.Close()
			delete(h.clients, client)
		}
	}
}

// ===== HELPER FUNCTIONS =====

func countEntities(output string) int {
	parsed, err := ovn.ParseJSONOutput(output)
	if err != nil {
		return 0
	}
	return len(parsed)
}

func extractNames(output string) []string {
	parsed, err := ovn.ParseJSONOutput(output)
	if err != nil {
		return []string{}
	}

	var names []string
	for _, item := range parsed {
		if name, ok := item["name"].(string); ok {
			names = append(names, name)
		}
	}
	return names
}

func parseTraceOutput(output string) models.TraceResult {
	result := models.TraceResult{}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Extract verdict
		if strings.Contains(line, "returned success") {
			result.Verdict = "success"
		} else if strings.Contains(line, "returned failure") {
			result.Verdict = "dropped"
		}

		// Extract flow information
		if strings.HasPrefix(line, "table=") || strings.HasPrefix(line, " Table=") {
			flow := parseFlowLine(line)
			result.Flows = append(result.Flows, flow)
		}
	}

	return result
}

func parseFlowLine(line string) models.LogicalFlow {
	flow := models.LogicalFlow{}

	// Extract table name
	if idx := strings.Index(line, "Table="); idx != -1 {
		end := strings.IndexAny(line[idx:], " []")
		if end != -1 {
			flow.Table = line[idx+6 : idx+end]
		}
	}

	// Extract priority
	if idx := strings.Index(line, "priority="); idx != -1 {
		end := strings.IndexAny(line[idx+9:], " ,]")
		if end != -1 {
			flow.Match = line[idx+9 : idx+9+end]
		}
	}

	flow.Action = strings.TrimSpace(line)
	return flow
}