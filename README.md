# cisco-exporter

Exporter for metrics from Cisco devices running NX-OS, IOS XE or IOS via SSH.

## Usage
```
Usage of ./cisco-exporter:
  -config.file string
    	Configuration file (default "cisco-exporter.yml")
  -scrape.timeout duration
    	Duration after which to abort a scrape (default 50s)
  -ssh.keep-alive-interval duration
    	Duration to wait between keep alive messages (default 10s)
  -ssh.keep-alive-timeout duration
    	Duration to wait for keep alive message response (default 15s)
  -ssh.reconnect-interval duration
    	Duration to wait before reconnecting to a device after connection got lost (default 30s)
  -version
    	Print version and exit
  -web.listen-address string
    	Address to listen on (default "[::]:9457")
  -web.telemetry-path string
    	Path under which to expose metrics (default "/metrics")
```

## Installation
Binary releases can be downloaded from the [releases page](https://gitlab.com/wobcom/cisco-exporter/-/releases).

## Configuration
Monitored devices are provided in a configuration file:
```yaml
devices:
  hostname.example.com:
    port: 1337  # optional: SSH port of the remote device
    enabled_collectors:  # required: See below for a list of collectors
      - cpu
      - memory
      - interfaces
    enabled_vlans: # optional: Some devices (BNGs) might have thousands of vlans 
      - 100
    interfaces:  # optional: Some devices (BNGs) might have thousands of interfaces
      - HundredGigE0/0/0  # you can specify the interfaces to be scraped
      - GigabitEthernet0
    username: monitoring  # required: Username to use for SSH auth
    kefile: /path/to/a/private.key  # optional: Private key to use for SSH auth
    password: correcthorsebatterystaple  # optional: Password for SSH auth
    ConnectTimeout: 5  # optional: Timeout for establishing the SSH conenction
    CommandTimeout: 10  # optional: Timeout for running a single command on the remote
```

## Available collectors
Multiple collectors are available, you **must** specify which one to use.

* **`aaa`**: Collects metrics about radius servers by running `show aaa servers`.
* **`bgp`**: Collects metrics about IPv4 / IPv6 unicast BGP peers by both running `show bgp ipv4 unicast neighbors` and `show bgp ipv6 unicast neighbors`.
* **`cpu`**: Collects metrics about CPU usage by running `show processes cpu`.
* **`environment`**: Collects metrics about the device's environment by running `show environment` or `show env all`.
* **`interfaces`**: Collects interface counters. Note that you can optionally limit which interfaces to scrape.
* **`mpls`**: Collects mpls specific metrics by both executing `show mpls forwarding-table` and `show mpls memory`.
* **`memory`**: Collects metrics about memory usage by running `show system resources` (NX-OS) or `show memory statistics`.
* **`nat`**: Collectrs metrics about network address translation by scraping the outputs of `show ip nat statistics` and multiple `show ip nat pool name ...`.
* **`optics`**: Collects transceiver status by issueing a `show interfaces transceiver detail` (IOS and NX-OS) or a `show inventory raw` followed by multiple `show hw-module subslot ...` commands on IOS XE.
* **`pppoe`**: Collects PPPoE statistics by issueing a `show pppoe statistics`.
* **`vlans`**: Collects VLAN counters returned by a `show vlans`.
* **`nat`**: Collects general NAT counters `show ip nat statistics` and NAT Pool counters `show ip nat pool name $name`.
* **`local_pools`**: Collects general information about local pools by using `show ip local pool`.

## Implementation details
Upon start cisco-exporter will try to connect with all the scrape targets.
Established SSH connections are kept alive as long as possible, to reduce scrape latency, load on the tacacs server and logged events.
