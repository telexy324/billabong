package singleton

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/libdns/cloudflare"
	tencentcloud "github.com/nezhahq/libdns-tencentcloud"

	"github.com/telexy324/billabong/model"
	ddns2 "github.com/telexy324/billabong/pkg/ddns"
	"github.com/telexy324/billabong/pkg/ddns/dummy"
	"github.com/telexy324/billabong/pkg/ddns/webhook"
	"github.com/telexy324/billabong/pkg/utils"
)

type DDNSClass struct {
	class[uint64, *model.DDNSProfile]
}

func NewDDNSClass() *DDNSClass {
	var sortedList []*model.DDNSProfile

	DB.Find(&sortedList)
	list := make(map[uint64]*model.DDNSProfile, len(sortedList))
	for _, profile := range sortedList {
		list[profile.ID] = profile
	}

	dc := &DDNSClass{
		class: class[uint64, *model.DDNSProfile]{
			list:       list,
			sortedList: sortedList,
		},
	}

	OnNameserverUpdate()
	return dc
}

func (c *DDNSClass) Update(p *model.DDNSProfile) {
	c.listMu.Lock()
	c.list[p.ID] = p
	c.listMu.Unlock()

	c.sortList()
}

func (c *DDNSClass) Delete(idList []uint64) {
	c.listMu.Lock()
	for _, id := range idList {
		delete(c.list, id)
	}
	c.listMu.Unlock()

	c.sortList()
}

func (c *DDNSClass) GetDDNSProvidersFromProfiles(profileId []uint64, ip *model.IP) ([]*ddns2.Provider, error) {
	profiles := make([]*model.DDNSProfile, 0, len(profileId))

	c.listMu.RLock()
	for _, id := range profileId {
		if profile, ok := c.list[id]; ok {
			profiles = append(profiles, profile)
		} else {
			c.listMu.RUnlock()
			return nil, fmt.Errorf("无法找到DDNS配置 ID %d", id)
		}
	}
	c.listMu.RUnlock()

	providers := make([]*ddns2.Provider, 0, len(profiles))
	for _, profile := range profiles {
		provider := &ddns2.Provider{DDNSProfile: profile, IPAddrs: ip}
		switch profile.Provider {
		case model.ProviderDummy:
			provider.Setter = &dummy.Provider{}
			providers = append(providers, provider)
		case model.ProviderWebHook:
			provider.Setter = &webhook.Provider{DDNSProfile: profile}
			providers = append(providers, provider)
		case model.ProviderCloudflare:
			provider.Setter = &cloudflare.Provider{APIToken: profile.AccessSecret}
			providers = append(providers, provider)
		case model.ProviderTencentCloud:
			provider.Setter = &tencentcloud.Provider{SecretId: profile.AccessID, SecretKey: profile.AccessSecret}
			providers = append(providers, provider)
		default:
			return nil, fmt.Errorf("无法找到配置的DDNS提供者 %s", profile.Provider)
		}
	}
	return providers, nil
}

func (c *DDNSClass) sortList() {
	c.listMu.RLock()
	defer c.listMu.RUnlock()

	sortedList := utils.MapValuesToSlice(c.list)
	slices.SortFunc(sortedList, func(a, b *model.DDNSProfile) int {
		return cmp.Compare(a.ID, b.ID)
	})

	c.sortedListMu.Lock()
	defer c.sortedListMu.Unlock()
	c.sortedList = sortedList
}

func OnNameserverUpdate() {
	ddns2.InitDNSServers(Conf.DNSServers)
}
