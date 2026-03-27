# netcheck

A command-line interface (CLI) tool for network management and diagnostics on Linux systems.

## Features

- **DNS** — list, add and remove nameservers in `/etc/resolv.conf`
- **Connection Test** — ping, traceroute, packet loss and DNS resolution
- **Change IP** — change network interface IP address with automatic rollback on failure
- **Manage Routes** — add and remove routes with automatic connectivity testing

---

## Requirements

- Ubuntu 22.04 or Debian 11/12
- Go 1.22 or higher
- Root access (sudo)

---

## Step by step

### 1. Install Go

**Ubuntu 22.04:**
```bash
sudo apt update
sudo apt install -y golang-go
```

**Debian** (the official repositories may ship an older version, so install directly from the Go website):
```bash
wget https://go.dev/dl/go1.22.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

Verify the installation:

```bash
go version
```

### 2. Install traceroute

```bash
sudo apt install -y traceroute
```

### 3. Clone the repository

```bash
git clone https://github.com/David-Felix/netcheckGO.git
cd netcheckGO
```

### 4. Download dependencies

```bash
go mod tidy
```

### 5. Build

```bash
go build -o netcheck ./cmd/netcheck/
```

### 6. Run

```bash
sudo ./netcheck
```

> Root is required to manage network interfaces, routes and ICMP packets.

---

## Menu navigation

Use **↑ ↓** arrow keys to navigate and **Enter** to confirm. Press **Ctrl+C** to go back or exit.

---

## Project structure

```
netcheckGO/
├── cmd/
│   └── netcheck/
│       └── main.go              # Application entry point
├── internal/
│   ├── menu/                    # User interface layer
│   │   ├── menu.go
│   │   ├── dns_menu.go
│   │   ├── connection_menu.go
│   │   ├── ip_menu.go
│   │   └── routes_menu.go
│   └── network/                 # Network logic layer
│       ├── validate.go
│       ├── dns.go
│       ├── ping.go
│       ├── connection.go
│       ├── routes.go
│       └── ip.go
├── go.mod
└── go.sum
```

---

## Main dependencies

| Package | Purpose |
|---|---|
| `github.com/manifoldco/promptui` | Interactive terminal menu |
| `github.com/go-ping/ping` | ICMP packet sending |
| `github.com/vishvananda/netlink` | IP address and route management via kernel |
