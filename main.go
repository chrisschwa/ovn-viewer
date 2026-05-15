package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"ovn-troubleshooter/config"
	"ovn-troubleshooter/handlers"
	"ovn-troubleshooter/ovn"

	"github.com/gin-gonic/gin"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "", "path to configuration file (JSON)")
	flag.Parse()

	// Load configuration
	cfg := config.DefaultConfig()
	if *configPath != "" {
		if err := loadConfig(*configPath, &cfg); err != nil {
			log.Fatalf("Failed to load config from %s: %v", *configPath, err)
		}
	} else {
		// Try default config paths
		defaultPaths := []string{
			"config.json",
			"./config/config.json",
			"/etc/ovn-troubleshooter/config.json",
		}
		for _, path := range defaultPaths {
			if err := loadConfig(path, &cfg); err == nil {
				log.Printf("Loaded config from %s", path)
				break
			}
		}
	}

	// Allow environment variable overrides
	if port := os.Getenv("OVN_TROUBLESHOOTER_PORT"); port != "" {
		cfg.Server.Port = port
	}
	if host := os.Getenv("OVN_TROUBLESHOOTER_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if controller := os.Getenv("OVN_CONTROLLER_HOST"); controller != "" {
		cfg.OVN.ControllerHost = controller
	}
	if user := os.Getenv("OVN_CONTROLLER_USER"); user != "" {
		cfg.OVN.ControllerUser = user
	}
	if keyPath := os.Getenv("OVN_SSH_KEY_PATH"); keyPath != "" {
		cfg.OVN.SSHKeyPath = keyPath
	}

	// Create OVN executor (mock or SSH)
	var executor ovn.ExecutorInterface
	useMock := os.Getenv("MOCK_OVN") == "1"
	if useMock {
		executor = ovn.NewMockExecutorFactory(cfg.OVN)
		log.Printf("Running in MOCK mode")
	} else {
		executor = ovn.NewExecutorFactory(cfg.OVN)
	}

	if err := executor.Connect(); err != nil {
		log.Printf("Warning: could not connect to OVN controller: %v", err)
	}
	defer executor.Close()

	// Create handler
	h := handlers.NewHandler(executor, cfg)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	// Serve static files (web UI)
	r.Static("/assets", "./web")
	
	// SPA fallback: serve index.html for non-API routes
	r.NoRoute(func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.File("./web/index.html")
		}
	})

	// ===== API ROUTES =====

	api := r.Group("/api")
	{
		// Dashboard
		api.GET("/dashboard", h.GetDashboard)
		api.GET("/connection", h.GetConnectionInfo)

		// Routers
		api.GET("/routers", h.ListRouters)
		api.GET("/routers/:name", h.GetRouter)
		api.GET("/routers/:name/flows", h.GetRouterFlows)
		api.GET("/routers/:name/routes", h.GetRouterStaticRoutes)
		api.GET("/routers/:name/nat", h.GetRouterNATRules)

		// Switches
		api.GET("/switches", h.ListSwitches)
		api.GET("/switches/:name", h.GetSwitch)
		api.GET("/switches/:name/flows", h.GetSwitchFlows)
		api.GET("/switches/:name/acls", h.GetSwitchACLs)

		// Ports
		api.GET("/ports/router", h.ListRouterPorts)
		api.GET("/ports/switch", h.ListSwitchPorts)
		api.GET("/ports/router/:name", h.GetRouterPort)
		api.GET("/ports/switch/:name", h.GetSwitchPort)

		// ACLs
		api.GET("/acls", h.ListACLs)

		// Chassis
		api.GET("/chassis", h.ListChassis)

		// Port Bindings
		api.GET("/port-bindings", h.ListPortBindings)
		api.GET("/port-bindings/:name", h.GetPortBinding)

		// Flows
		api.GET("/flows", h.ListAllFlows)

		// OVS
		api.GET("/ovs/bridges", h.ListBridges)
		api.GET("/ovs/bridges/:name", h.GetBridge)
		api.GET("/ovs/interfaces", h.ListInterfaces)
		api.GET("/ovs/bridges/:name/flows", h.GetOVSFlows)
		api.GET("/ovs/show", h.ShowOVS)

		// Topology
		api.GET("/topology/ovn", h.ShowTopology)

		// Packet Trace
		api.POST("/trace", h.TracePacket)
		api.POST("/rod", h.RadiusOfDarkness)

		// Autocomplete
		api.GET("/autocomplete/routers", h.GetRouterNames)
		api.GET("/autocomplete/switches", h.GetSwitchNames)
		api.GET("/autocomplete/ports", h.GetPortNames)

		// WebSocket
		api.GET("/ws", h.WebSocket)
	}

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("OVN Troubleshooter starting on %s", addr)
	log.Printf("Open http://localhost:%s in your browser", cfg.Server.Port)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// loadConfig loads configuration from a JSON file
func loadConfig(path string, cfg *config.Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, cfg)
}

// corsMiddleware adds CORS headers to all responses
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}