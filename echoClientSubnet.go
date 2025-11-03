// echoClientSubnet is a CoreDNS plugin that returns a query's EDNS Client Subnet information in a TXT record.
package echoClientSubnet

import (
	"context"
	"fmt"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
)

const pluginName = "echoClientSubnet"

var log = clog.NewWithPlugin(pluginName)

type echoClientSubnet struct {
	Next plugin.Handler
}

// ServeDNS implements the plugin.Handler interface. This method gets called when echoClientSubnet is used in a Server.
func (e echoClientSubnet) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	// Only respond to TXT queries
	if r.Question[0].Qtype != dns.TypeTXT {
		return plugin.NextOrFailure(e.Name(), e.Next, ctx, w, r)
	}

	state := request.Request{W: w, Req: r}
	qname := state.Name()

	answers := make([]dns.RR, 0, 10)
	resp := new(dns.TXT)
	resp.Hdr = dns.RR_Header{Name: dns.Fqdn(qname), Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 50}

	msg := new(dns.Msg)
	msg.SetReply(r)

	opt := state.Req.IsEdns0()
	if opt == nil {
		log.Debug("No EDNS options in request")
		resp.Txt = []string{fmt.Sprintf("No EDNS options found. Remote address: %s", state.RemoteAddr())}
		answers = append(answers, resp)
		msg.Answer = answers
		w.WriteMsg(msg)
		return 0, nil
	}

	for i := range opt.Option {
		option := opt.Option[i]
		if option.Option() == dns.EDNS0SUBNET {
			resp.Txt = []string{fmt.Sprintf("%s. Remote address: %s", option.String(), state.RemoteAddr())}
			answers = append(answers, resp)
			msg.Answer = answers

			w.WriteMsg(msg)
			return 0, nil
		}
	}

	log.Debug("No EDNS client subnet option in request")

	resp.Txt = []string{fmt.Sprintf("No EDNS Client subnet option found. Remote address: %s", state.RemoteAddr())}
	answers = append(answers, resp)
	msg.Answer = answers

	w.WriteMsg(msg)
	return 0, nil
}

// Name implements the Handler interface.
func (e echoClientSubnet) Name() string { return pluginName }
