package singleton

import (
	"cmp"
	"slices"

	"github.com/telexy324/billabong/model"
	"github.com/telexy324/billabong/pkg/utils"
)

type ServerClass struct {
	class[uint64, *model.Server]

	uuidToID map[string]uint64

	sortedListForGuest []*model.Server
}

func NewServerClass() *ServerClass {
	sc := &ServerClass{
		class: class[uint64, *model.Server]{
			list: make(map[uint64]*model.Server),
		},
		uuidToID: make(map[string]uint64),
	}

	var servers []model.Server
	DB.Find(&servers)
	for _, s := range servers {
		innerS := s
		model.InitServer(&innerS)
		sc.list[innerS.ID] = &innerS
		sc.uuidToID[innerS.UUID] = innerS.ID
	}
	sc.sortList()

	return sc
}

func (c *ServerClass) Update(s *model.Server, uuid string) {
	c.listMu.Lock()

	c.list[s.ID] = s
	if uuid != "" {
		c.uuidToID[uuid] = s.ID
	}

	c.listMu.Unlock()

	c.sortList()
}

func (c *ServerClass) Delete(idList []uint64) {
	c.listMu.Lock()

	for _, id := range idList {
		serverUUID := c.list[id].UUID
		delete(c.uuidToID, serverUUID)
		delete(c.list, id)
	}

	c.listMu.Unlock()

	c.sortList()
}

func (c *ServerClass) GetSortedListForGuest() []*model.Server {
	c.sortedListMu.RLock()
	defer c.sortedListMu.RUnlock()

	return slices.Clone(c.sortedListForGuest)
}

func (c *ServerClass) UUIDToID(uuid string) (id uint64, ok bool) {
	c.listMu.RLock()
	defer c.listMu.RUnlock()

	id, ok = c.uuidToID[uuid]
	return
}

func (c *ServerClass) sortList() {
	c.listMu.RLock()
	defer c.listMu.RUnlock()
	c.sortedListMu.Lock()
	defer c.sortedListMu.Unlock()

	c.sortedList = utils.MapValuesToSlice(c.list)
	// 按照服务器 ID 排序的具体实现（ID越大越靠前）
	slices.SortStableFunc(c.sortedList, func(a, b *model.Server) int {
		if a.DisplayIndex == b.DisplayIndex {
			return cmp.Compare(a.ID, b.ID)
		}
		return cmp.Compare(b.DisplayIndex, a.DisplayIndex)
	})

	c.sortedListForGuest = make([]*model.Server, 0, len(c.sortedList))
	for _, s := range c.sortedList {
		if !s.HideForGuest {
			c.sortedListForGuest = append(c.sortedListForGuest, s)
		}
	}
}
