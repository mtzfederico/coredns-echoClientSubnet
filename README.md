# Echo Client Subnet
This is a CoreDNS plugin that responds to TXT queries with the EDNS Client Subnet's data.

## Usage
```
edns.example.com {
	echoClientSubnet
}
```

## Install plugin
Add `echoClientSubnet:github.com/mtzfederico/coredns-echoClientSubnet` to `plugin.cfg` and run `make`

## Querying using Dig
`dig +short txt edns.example.com +subnet=203.0.113.0/24`

Output:
`"1203.0.113.0/24/0. Remote address: [::1]:51505"`