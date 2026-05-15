# OVN Troubleshooter

A web-based application for network administrators to troubleshoot traffic flows and inspect entities in OpenStack OVN/OVS environments.

## Features

- **Dashboard**: Overview of your OVN environment with entity counts and connection status
- **Logical Routers**: Browse routers, inspect static routes, NAT rules, and flows (drill-down on each router)
- **Logical Switches**: Browse switches, inspect ACLs and flows (drill-down on each switch)
- **Ports**: View router ports and switch ports with their configurations, drill-down into individual port details
- **ACLs**: Browse and filter Access Control Lists by direction, action, and match expression
- **Chassis**: View connected OVN chassis (hypervisors) with drill-down to hosted resources (routers via `lrp-set-chassis`, switches via port bindings)
- **Logical Flows**: Browse and filter the OVN pipeline flows
- **Packet Tracer**: Interactive packet trace tool wrapping `ovn-trace` with:
  - Visual packet specification builder (IPv4/IPv6/ARP, TCP/UDP/ICMP)
  - Custom packet specification support
  - Radius of Darkness (ROD) checking
  - Flow-by-flow trace visualization
  - Success/failure verdict display
- **Topology**: View the OVN NB topology (`ovn-nbctl show`)
- **OVS**: Inspect OVS bridges, interfaces, and topology

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                   Web Browser                       │
│  (Single Page Application - HTML/CSS/JS)            │
└────────────────────┬────────────────────────────────┘
                     │ HTTP/WebSocket
┌────────────────────▼────────────────────────────────┐
│              Go Backend (Gin Framework)              │
│                                                     │
│  ┌─────────────┐  ┌─────────────┐  ┌────────────┐  │
│  │  Handlers   │  │   Routes    │  │  WebSocket │  │
│  └──────┬──────┘  └─────────────┘  └────────────┘  │
│         │                                           │
│  ┌──────▼────────────────────────────────────────┐  │
│  │           OVN/OVS Executor                     │  │
│  │   (ovn-nbctl, ovn-sbctl, ovn-trace, ovs-vsctl)│  │
│  └──────────────────┬────────────────────────────┘  │
└─────────────────────┼───────────────────────────────┘
                      │ SSH
              ┌───────▼───────┐
              │  OVN Controller│
              │  (OpenStack)   │
              └───────────────┘
```

## Prerequisites

- Go 1.21+
- SSH access to an OpenStack controller node with OVN tools installed
- OVN tools: `ovn-nbctl`, `ovn-sbctl`, `ovn-trace`
- OVS tools: `ovs-vsctl`, `ovs-ofctl`

## Installation

```bash
# Clone or navigate to the project directory
cd ovn-troubleshooter

# Download dependencies
go mod download

# Build
go build -o ovn-troubleshooter .
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
    "controller_host": "controller.openstack.local",
    "controller_user": "ovnadmin",
    "ssh_port": 22,
    "ssh_key_path": "~/.ssh/id_rsa",
    "command_timeout": 30,

    // Docker mode: wrap commands with docker/podman exec
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
./ovn-troubleshooter -config /path/to/config.json
```

## Running

```bash
# Using default config (config.json in current directory)
./ovn-troubleshooter

# Using custom config
./ovn-troubleshooter -config /path/to/config.json

# Using environment variables
OVN_CONTROLLER_HOST=my-controller.ovn-controller OVN_SSH_KEY_PATH=~/.ssh/mykey ./ovn-troubleshooter
```

Then open `http://localhost:8080` in your browser.

## Running in Docker Mode (Containerized OVN)

In environments where OVN/OVS services run inside containers (e.g., TripleO, Kolla, podified deployments), enable docker mode to wrap commands with `docker exec` or `podman exec`:

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

When enabled, commands like `ovn-nbctl list Logical_router` are automatically wrapped:
- `ovn-nbctl` → `docker exec ovn-northd ovn-nbctl list Logical_router`
- `ovn-sbctl` → `docker exec ovn-southbound ovn-sbctl ...`
- `ovs-vsctl` → `docker exec ovn-controller ovs-vsctl ...`

Default container mappings (when `docker_container_map` is not set):

| Command | Default Container |
|---|---|
| `ovn-nbctl`, `ovn-trace` | `ovn-northd` |
| `ovn-sbctl` | `ovn-southbound` |
| `ovs-vsctl`, `ovs-ofctl` | `ovn-controller` |

## Running in Mock Mode (No OVN Required)

When `MOCK_OVN=1` is set, the troubleshooter runs entirely offline using a built-in mock executor that simulates OVN/OVS commands. This is useful for testing, demos, or development without an OVN environment.

```bash
# Run with built-in mock data
$env:MOCK_OVN="1"; .\ovn-troubleshooter.exe
```

The mock mode includes the same predefined scenarios as [ovn-mock](../ovn-mock) (default, demo, broken-routing, acl-drop, nat-issue). The `MOCK_OVN_SCENARIO` environment variable selects which scenario to use.

For more advanced mock testing (e.g., remote SSH-based testing), pair with the [ovn-mock](../ovn-mock) SSH server.

## API Endpoints

### Dashboard
| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/dashboard` | Entity count summary |
| GET | `/api/connection` | Connection status info |

### Routers
| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/routers` | List all logical routers |
| GET | `/api/routers/:name` | Get router details |
| GET | `/api/routers/:name/flows` | Get router logical flows |
| GET | `/api/routers/:name/routes` | Get static routes |
| GET | `/api/routers/:name/nat` | Get NAT rules |

### Switches
| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/switches` | List all logical switches |
| GET | `/api/switches/:name` | Get switch details |
| GET | `/api/switches/:name/flows` | Get switch logical flows |
| GET | `/api/switches/:name/acls` | Get switch ACLs |

### Ports
| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/ports/router` | List router ports |
| GET | `/api/ports/switch` | List switch ports |
| GET | `/api/ports/router/:name` | Get router port details |
| GET | `/api/ports/switch/:name` | Get switch port details |

### Other
| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/acls` | List all ACLs |
| GET | `/api/chassis` | List all chassis |
| GET | `/api/port-bindings` | List port bindings |
| GET | `/api/flows` | List all logical flows |
| GET | `/api/topology/ovn` | OVN NB topology |

### OVS
| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/ovs/show` | OVS topology |
| GET | `/api/ovs/bridges` | List OVS bridges |
| GET | `/api/ovs/interfaces` | List OVS interfaces |
| GET | `/api/ovs/bridges/:name/flows` | Get bridge flows |

### Packet Tracing
| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/trace` | Trace a packet through OVN pipeline |
| POST | `/api/rod` | Check Radius of Darkness |

### Autocomplete
| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/autocomplete/routers` | Router names |
| GET | `/api/autocomplete/switches` | Switch names |
| GET | `/api/autocomplete/ports` | Port names |

### WebSocket
| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/ws` | WebSocket connection for live updates |

## Packet Trace Request Format

```json
{
  "router": "lr0",
  "lrp": "lr0-int",
  "lsp": "int-br-e16e53b6-eb",
  "pkt": "ip4.src==10.0.0.10 && ip4.dst==10.0.1.20 && tcp.dst==80",
  "check_rod": true
}
```

Or for switch-to-switch:

```json
{
  "switch": "int-br-e16e53b6-eb",
  "lsp": "int-br-e16e53b6-eb",
  "pkt": "ip4.src==10.0.0.10 && ip4.dst==10.0.1.20",
  "check_rod": true
}
```

## OVN Troubleshooting Concepts

### ovn-trace
The `ovn-trace` command simulates a packet through the OVN logical pipeline. It shows which logical flows match and the final verdict (delivered or dropped).

### Radius of Darkness (ROD)
The ROD check validates that the compiled datapath flows in the SB DB correctly implement the logical flows from the NB DB. A failing ROD check indicates a potential datapath bug.

### OVN Pipeline Tables
- **Logical Switch Pipeline**: port_sec (0) → in_port (1) → lb_skip (2) → acl_in_host (3) → lb_source (4) → acl_in_port (5) → lb (6) → acl_in_datapath (7) → lrstat (8) → ldap (9) → reroute (10) → dup (11)
- **Logical Router Pipeline**: lr_nac (0) → lr_policy (1) → lr_lb_skip (2) → lr_stat (3) → lr_ad (4) → lr_reroute (5) → lr_lb (6) → lr_dp (7)
- **Output Pipeline**: acl_out (80) → meter (81) → ratelimit (90) → port_sec_out (91) → ipsec_out (92)

## References

- [OVN Traffic Flow Troubleshooting](https://cloudification.io/de/cloud-blog/ovn-traffic-flow-troubleshooting-in-openstack/)
- [Tracing Packets with OVN](https://lewisdenny.io/tracing_packets_out_an_external_network_with_ovn/)
- [OVN Documentation](https://ovn.org/docs/)

## License

MIT