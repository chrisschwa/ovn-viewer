# OVN Viewer

A web-based tool for network administrators to inspect and troubleshoot traffic flows in OVN (Open Virtual Network) environments.

Browse logical routers, switches, ports, ACLs, and chassis — trace packets through the OVN pipeline — all from your browser.

![Architecture](https://img.shields.io/badge/Built%20with-Go%20%2B%20Gin-blue)
![License](https://img.shields.io/badge/License-MIT-yellow)

## Features

- **Dashboard** — Overview of your OVN environment with entity counts and connection status
- **Logical Routers** — Browse routers, inspect static routes, NAT rules, and flows
- **Logical Switches** — Browse switches, inspect ACLs and flows
- **Ports** — View router ports and switch ports with their configurations
- **ACLs** — Browse and filter Access Control Lists
- **IPs** — Overview of all IP addresses: NAT floating IPs, router port IPs, switch port IPs
- **Chassis** — View connected OVN chassis (hypervisors)
- **Logical Flows** — Browse and filter the OVN pipeline flows
- **Packet Tracer** — Interactive packet trace tool wrapping `ovn-trace`:
  - Visual packet specification builder (IPv4/IPv6/ARP, TCP/UDP/ICMP)
  - Custom packet specification support
  - Radius of Darkness (ROD) checking
  - Flow-by-flow trace visualization
- **Topology** — View the OVN NB topology (`ovn-nbctl show`)
- **OVS** — Inspect OVS bridges, interfaces, and flows

## Architecture

```
┌─────────────────────────────────────────────────┐
│                 Web Browser                      │
│  (Single Page Application - HTML/CSS/JS)         │
└────────────────────┬────────────────────────────┘
                     │ HTTP/WebSocket
┌────────────────────▼────────────────────────────┐
│           Go Backend (Gin Framework)             │
│                                                 │
│  ┌──────────┐  ┌──────────┐  ┌──────────────┐  │
│  │ Handlers │  │  Routes  │  │   WebSocket  │  │
│  └────┬─────┘  └──────────┘  └──────────────┘  │
│       │                                          │
│  ┌────▼───────────────────────────────────────┐  │
│  │         OVN/OVS Executor                    │  │
│  │  (ovn-nbctl, ovn-sbctl, ovn-trace, ovs-...)│  │
│  └──────────────────┬─────────────────────────┘  │
└─────────────────────┼────────────────────────────┘
                      │ SSH
              ┌───────▼───────┐
              │  OVN Controller│
              │  (any host)    │
              └───────────────┘
```

## Prerequisites

- Go 1.21+
- SSH access to a host with OVN tools installed
- OVN tools: `ovn-nbctl`, `ovn-sbctl`, `ovn-trace`
- OVS tools: `ovs-vsctl`, `ovs-ofctl`

## Installation

```bash
# Clone the repository
git clone https://github.com/chrisschwa/ovn-viewer.git
cd ovn-viewer

# Download dependencies
go mod download

# Build
go build -o ovn-viewer .
```

## Configuration

### Option 1: Configuration File

Copy and edit the example configuration:

```bash
cp config.json.example config.json
```

Edit `config.json` with your environment settings:

```json
{
  "server": {
    "port": "8080",
    "host": "0.0.0.0"
  },
  "ovn": {
    "controller_host": "your-ovn-controller.example.com",
    "controller_user": "ovs",
    "ssh_port": 22,
    "ssh_key_path": "~/.ssh/id_rsa",
    "command_timeout": 30,
    "docker_mode": false
  }
}
```

### Option 2: Environment Variables

Override any setting with environment variables:

| Variable | Description |
|---|---|
| `OVN_TROUBLESHOOTER_PORT` | HTTP server port |
| `OVN_TROUBLESHOOTER_HOST` | HTTP server host |
| `OVN_CONTROLLER_HOST` | OVN controller hostname |
| `OVN_CONTROLLER_USER` | SSH user for controller |
| `OVN_SSH_KEY_PATH` | Path to SSH private key |

### Option 3: Command Line

```bash
./ovn-viewer -config /path/to/config.json
```

## Running

```bash
# Using default config (config.json in current directory)
./ovn-viewer

# Using custom config
./ovn-viewer -config /path/to/config.json

# Using environment variables
OVN_CONTROLLER_HOST=my-controller OVN_SSH_KEY_PATH=~/.ssh/mykey ./ovn-viewer
```

Then open `http://localhost:8080` in your browser.

## Mock Mode (No OVN Required)

Set `MOCK_OVN=1` to run entirely offline using built-in mock data. This is useful for testing, demos, or development without an OVN environment:

```bash
# Linux / macOS
MOCK_OVN=1 ./ovn-viewer

# Windows PowerShell
$env:MOCK_OVN="1"; .\ovn-viewer.exe
```

Optionally select a scenario with `MOCK_OVN_SCENARIO`:

| Scenario | Description |
|---|---|
| `default` | Basic setup: 2 routers, 3 switches, ACLs, routes |
| `demo` | Rich environment: 4 routers, 6 switches, LBs, complex ACLs |
| `broken-routing` | Missing static routes causing packet drops |
| `acl-drop` | ACLs blocking HTTP, SSH, and ICMP traffic |
| `nat-issue` | Misconfigured NAT rules (wrong internal IP) |

```bash
MOCK_OVN=1 MOCK_OVN_SCENARIO=broken-routing ./ovn-viewer
```

## Docker Mode

In environments where OVN/OVS services run inside containers (e.g., TripleO, Kolla, podified deployments), enable docker mode:

```json
{
  "ovn": {
    "docker_mode": true,
    "docker_runtime": "docker",
    "docker_container_map": {
      "ovn-nbctl": "ovn-northd",
      "ovn-sbctl": "ovn-southbound",
      "ovn-trace": "ovn-northd",
      "ovs-vsctl": "ovn-controller",
      "ovs-ofctl": "ovn-controller"
    }
  }
}
```

When enabled, commands are automatically wrapped with `docker exec` or `podman exec`.

## API Endpoints

### Dashboard

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/dashboard` | Entity count summary |
| GET | `/api/connection` | Connection status |

### Routers

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/routers` | List all logical routers |
| GET | `/api/routers/:name` | Router details |
| GET | `/api/routers/:name/flows` | Router logical flows |
| GET | `/api/routers/:name/routes` | Static routes |
| GET | `/api/routers/:name/nat` | NAT rules |

### Switches

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/switches` | List all logical switches |
| GET | `/api/switches/:name` | Switch details |
| GET | `/api/switches/:name/flows` | Switch logical flows |
| GET | `/api/switches/:name/acls` | Switch ACLs |

### Ports

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/ports/router` | List router ports |
| GET | `/api/ports/switch` | List switch ports |
| GET | `/api/ports/router/:name` | Router port details |
| GET | `/api/ports/switch/:name` | Switch port details |

### Other

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/acls` | List all ACLs |
| GET | `/api/chassis` | List all chassis |
| GET | `/api/port-bindings` | List port bindings |
| GET | `/api/flows` | List all logical flows |
| GET | `/api/topology/ovn` | OVN NB topology |
| GET | `/api/ovs/bridges` | List OVS bridges |
| GET | `/api/ovs/interfaces` | List OVS interfaces |
| GET | `/api/ovs/show` | OVS topology |

### IP Overview

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/ips` | NAT rules, router port IPs, switch port IPs |

### Packet Tracing

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/trace` | Trace a packet through OVN |
| POST | `/api/rod` | Check Radius of Darkness |

### WebSocket

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/ws` | WebSocket for live updates |

## OVN Troubleshooting Concepts

### ovn-trace
The `ovn-trace` command simulates a packet through the OVN logical pipeline and shows which flows match and the final verdict.

### Radius of Darkness (ROD)
Validates that compiled datapath flows in the SB DB correctly implement the logical flows from the NB DB.

### OVN Pipeline Tables

**Logical Switch Pipeline**: port_sec (0) → in_port (1) → lb_skip (2) → acl_in_host (3) → lb_source (4) → acl_in_port (5) → lb (6) → acl_in_datapath (7) → lrstat (8) → ldap (9) → reroute (10) → dup (11)

**Logical Router Pipeline**: lr_nac (0) → lr_policy (1) → lr_lb_skip (2) → lr_stat (3) → lr_ad (4) → lr_reroute (5) → lr_lb (6) → lr_dp (7)

## License

MIT