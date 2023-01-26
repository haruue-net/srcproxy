srcproxy
=================

srcproxy is a TCP only proxy protocol designed to preserve the source address
of the client side. The proxy client of srcproxy sends source address
information to the proxy server, and the proxy server will establish TCP
connection to the proxy target with the original source address
(with `sysctl net.ipv{4,6}.ip_nonlocal_bind=1`). 

![](./docs/srcproxy.drawio.png)

Comparing to other Layer 4 proxy, srcproxy allows you preserve the address of
client side. For instance, providing proxy service to multiple clients and
allow them has their own source prefixes.

Comparing to Layer 3 VPN, srcproxy let you utilize customized tools to optimize
the TCP stack.

srcproxy only supports Linux now, as it requires iptables REDIRECT target for
inbound and `net.ipv*.ip_nonlocal_bind` to bind any address in a specified
prefix.


## Configuration

### Server

```bash
./srcproxy server config.json
```

```json5
{
  "listen": "192.0.2.1:21829", // The listen address for connection from clients, it is recommanded to use a specified address instead of 0.0.0.0
  "acl": [  // (Optional) The ACL policy to restrict the source addresses can be used by different clients.
    // An empty "acl" array would completely disable the ACL feature.
    {
      "auth": "613200f2-0af9-40e3-9dc2-a5d6f365db1b",  // The "auth" field in the client config, used to distinguish each client.
      "allowed_src_ips": [  // (Optional) Prefixes that this client can use as source address.
        "192.0.2.128/25",
        "2001:db8:aaaa::/64"
      ],
    },
    {
      "auth": "", // An empty "auth" is acceptable, and it only matches empty "auth" field in client config.
      "allowed_src_ips": [
        // An empty "allowed_src_ips" array is acceptable, and it allows this client use any address as source address.
      ],
    },
  ],
  "timeout": "5m", // (Optional) Connect & idle timeout of each TCP connection, accept strings like "5m", "300s", and integer number (for seconds) as other proxy application.
  "log_level": "info", // (Optional) Log level, supports "info" and "error"
}
```

You need to enable `ip_nonlocal_bind` and add prefixes that uses as source
addresses into route table manually.

```bash
sysctl net.ipv4.ip_nonlocal_bind=1
sysctl net.ipv6.ip_nonlocal_bind=1
ip -6 route add local 2001:db8:aaaa::/64 dev lo table local
```

Please be aware that `ip_nonlocal_bind` is a global options for all application
in this netns, and add a `dev lo table local` route would make the prefix only
to be routed to local process. It is recommanded to run the srcproxy sevver in
a standalone netns for a easier configration.


### Client

```bash
./srcproxy client config.json
```

```json5
{
  "inbound": {
    "mode": "redirect", // (Optional) Only supports iptables -j REDIRECT now.
    "listen": ":9001"  // The listen address for REDIRECT --to-port, only specify the port part so it would listen for both IPv4 and IPv6.
  },
  "outbound": {
    "server": "192.0.2.1:21829",  // The address of the srcproxy server.
    "auth": "613200f2-0af9-40e3-9dc2-a5d6f365db1b", // (Optional) The auth credential for server side ACL policy.
    "local_addr": "192.0.2.2%eth0", // (Optional) Bind to specified source address / interface when connect to the server.
    "fwmark": 0x1234,  // (Optional) Set specified fwmark for connection to the server.
  },
  "timeout": "5m", // (Optional) Connect & idle timeout of each TCP connection, accept strings like "5m", "300s", and integer number (for seconds) as other proxy application.
  "log_level": "info", // (Optional) Log level, supports "info" and "error"
}
```

You need manually add iptables rules to redirect TCP traffic to the srcproxy client.

```bash
ip6tables -t nat -A PREROUTING -s 2001:db8:aaaa::/64 -j REDIRECT --to-port 9001
```

As srcproxy only supports TCP, you might need an Layer 3 VPN for other
protocols such like UDP and ICMP.


## Security

The security of srcproxy should be same as plain SOCKS5, as there is no
data/key encrypt in this protocol.

I personally recommand to forward the port of srcproxy server with the "relay"
feature of other secure proxy tools (such like SSH and Hysteria).

