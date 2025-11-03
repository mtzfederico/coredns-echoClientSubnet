package echoClientSubnet

import (
	"context"
	"net"
	"testing"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"

	"github.com/miekg/dns"
)

func TestWithIPv4(t *testing.T) {
	req := new(dns.Msg)
	req.SetQuestion("clientSubnet.example.invalid.", dns.TypeTXT)

	subnet := new(dns.EDNS0_SUBNET)
	subnet.Code = dns.EDNS0SUBNET
	// 1 for IPv4, 2 for IPv6
	subnet.Family = 1
	subnet.SourceNetmask = 24
	// 0 in a request
	subnet.SourceScope = 0
	subnet.Address = net.IPv4(192, 168, 0, 0)

	edns := new(dns.OPT)
	edns.Hdr.Name = "."
	edns.Hdr.Rrtype = dns.TypeOPT
	edns.Option = []dns.EDNS0{subnet}
	// https://www.dnsflagday.net/2020/
	edns.SetUDPSize(1232)
	edns.SetDo()
	req.Extra = append(req.Extra, edns)

	a := &echoClientSubnet{}

	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	_, err := a.ServeDNS(context.Background(), rec, req)

	if err != nil {
		t.Errorf("Expected no error, but got %q", err)
	}

	value := rec.Msg.Answer[0].(*dns.TXT)

	if value.Txt[0] != "192.168.0.0/24/0" && value.Txt[1] != "Remote address: 10.240.0.1:40212" {
		t.Errorf("IPv4 test Failed. got %q", rec.Msg.Answer[0].(*dns.TXT).String())
	}
}

func TestWithIPv6(t *testing.T) {
	req := new(dns.Msg)
	req.SetQuestion("clientSubnet.example.invalid.", dns.TypeTXT)

	subnet := new(dns.EDNS0_SUBNET)
	subnet.Code = dns.EDNS0SUBNET
	// 1 for IPv4, 2 for IPv6
	subnet.Family = 2
	subnet.SourceNetmask = 64
	// 0 in a request
	subnet.SourceScope = 0
	subnet.Address = net.ParseIP("2a11:f2c0:fff7:1234::")

	edns := new(dns.OPT)
	edns.Hdr.Name = "."
	edns.Hdr.Rrtype = dns.TypeOPT
	edns.Option = []dns.EDNS0{subnet}
	// https://www.dnsflagday.net/2020/
	edns.SetUDPSize(1232)
	edns.SetDo()
	req.Extra = append(req.Extra, edns)

	a := &echoClientSubnet{}

	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	_, err := a.ServeDNS(context.Background(), rec, req)

	if err != nil {
		t.Errorf("Expected no error, but got %q", err)
	}

	value := rec.Msg.Answer[0].(*dns.TXT)

	if value.Txt[0] != "2a11:f2c0:fff7:1234::/64/0" && value.Txt[1] != "Remote address: 10.240.0.1:40212" {
		t.Errorf("IPv6 test Failed. got %q", rec.Msg.Answer[0].(*dns.TXT).String())
	}
}

func TestAnyNonTXTQuery(t *testing.T) {
	tests := []struct {
		name  string
		qtype uint16
	}{
		{"A query", dns.TypeA},
		{"AAAA query", dns.TypeAAAA},
		{"MX query", dns.TypeMX},
		{"CNAME query", dns.TypeCNAME},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := new(dns.Msg)
			req.SetQuestion("example.org.", tt.qtype)

			nextCalled := false
			a := &echoClientSubnet{
				Next: test.HandlerFunc(func(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
					nextCalled = true
					return 0, nil
				}),
			}

			rec := dnstest.NewRecorder(&test.ResponseWriter{})
			_, err := a.ServeDNS(context.TODO(), rec, req)

			if err != nil {
				t.Errorf("Expected no error, but got %q", err)
			}

			if !nextCalled {
				t.Error("Expected Next handler to be called for non-ANY query")
			}
		})
	}
}
