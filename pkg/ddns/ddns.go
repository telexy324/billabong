package ddns

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/libdns/libdns"
	"github.com/miekg/dns"

	"github.com/telexy324/billabong/model"
	"github.com/telexy324/billabong/pkg/utils"
)

var (
	dnsTimeOut       = 10 * time.Second
	customDNSServers []string
)

type Provider struct {
	ctx        context.Context
	ipAddr     string
	recordType string
	prefix     string
	zone       string

	DDNSProfile *model.DDNSProfile
	IPAddrs     *model.IP
	Setter      libdns.RecordSetter
}

func InitDNSServers(s string) {
	if s != "" {
		customDNSServers = strings.Split(s, ",")
	}
}

func (provider *Provider) GetProfileID() uint64 {
	return provider.DDNSProfile.ID
}

func (provider *Provider) UpdateDomain(ctx context.Context, overrideDomains ...string) {
	provider.ctx = ctx
	for _, domain := range utils.IfOr(len(overrideDomains) > 0, overrideDomains, provider.DDNSProfile.Domains) {
		for retries := 0; retries < int(provider.DDNSProfile.MaxRetries); retries++ {
			log.Printf("NEZHA>> Updating DNS Record of domain %s: %d/%d", domain, retries+1, provider.DDNSProfile.MaxRetries)
			if err := provider.updateDomain(domain); err != nil {
				log.Printf("NEZHA>> Failed to update DNS record of domain %s: %v", domain, err)
			} else {
				log.Printf("NEZHA>> Update DNS record of domain %s succeeded", domain)
				break
			}
		}
	}
}

func (provider *Provider) updateDomain(domain string) error {
	var err error
	provider.prefix, provider.zone, err = splitDomainSOA(domain)
	if err != nil {
		return err
	}

	// 当IPv4和IPv6同时成功才算作成功
	if *provider.DDNSProfile.EnableIPv4 {
		provider.recordType = getRecordString(true)
		provider.ipAddr = provider.IPAddrs.IPv4Addr
		if err = provider.addDomainRecord(); err != nil {
			return err
		}
	}

	if *provider.DDNSProfile.EnableIPv6 {
		provider.recordType = getRecordString(false)
		provider.ipAddr = provider.IPAddrs.IPv6Addr
		if err = provider.addDomainRecord(); err != nil {
			return err
		}
	}

	return nil
}

func (provider *Provider) addDomainRecord() error {
	_, err := provider.Setter.SetRecords(provider.ctx, provider.zone,
		[]libdns.Record{
			{
				Type:  provider.recordType,
				Name:  provider.prefix,
				Value: provider.ipAddr,
				TTL:   time.Minute,
			},
		})
	return err
}

func splitDomainSOA(domain string) (prefix string, zone string, err error) {
	c := &dns.Client{Timeout: dnsTimeOut}

	domain += "."
	indexes := dns.Split(domain)

	servers := utils.DNSServers
	if len(customDNSServers) > 0 {
		servers = customDNSServers
	}

	var r *dns.Msg
	for _, idx := range indexes {
		var m dns.Msg
		m.SetQuestion(domain[idx:], dns.TypeSOA)

		for _, server := range servers {
			r, _, err = c.Exchange(&m, server)
			if err != nil {
				return
			}
			if len(r.Answer) > 0 {
				if soa, ok := r.Answer[0].(*dns.SOA); ok {
					zone = soa.Hdr.Name
					prefix = libdns.RelativeName(domain, zone)
					return
				}
			}
		}
	}

	return "", "", fmt.Errorf("SOA record not found for domain: %s", domain)
}

func getRecordString(isIpv4 bool) string {
	if isIpv4 {
		return "A"
	}
	return "AAAA"
}
