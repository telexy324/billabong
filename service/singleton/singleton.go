package singleton

import (
	_ "embed"
	"gorm.io/driver/mysql"
	"iter"
	"log"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
	"sigs.k8s.io/yaml"

	"github.com/telexy324/billabong/model"
	"github.com/telexy324/billabong/pkg/utils"
)

var Version = "debug"

var (
	Cache             *cache.Cache
	DB                *gorm.DB
	Loc               *time.Location
	FrontendTemplates []model.FrontendTemplate
	DashboardBootTime = uint64(time.Now().Unix())

	ServerShared          *ServerClass
	ServiceSentinelShared *ServiceSentinel
	DDNSShared            *DDNSClass
	NotificationShared    *NotificationClass
	NATShared             *NATClass
	CronShared            *CronClass
)

//go:embed frontend-templates.yaml
var frontendTemplatesYAML []byte

func InitTimezoneAndCache() {
	var err error
	Loc, err = time.LoadLocation(Conf.Location)
	if err != nil {
		panic(err)
	}

	Cache = cache.New(5*time.Minute, 10*time.Minute)
}

// LoadSingleton 加载子服务并执行
func LoadSingleton() {
	initUser()                                  // 加载用户ID绑定表
	initI18n()                                  // 加载本地化服务
	NotificationShared = NewNotificationClass() // 加载通知服务
	ServerShared = NewServerClass()             // 加载服务器列表
	CronShared = NewCronClass()                 // 加载定时任务
	NATShared = NewNATClass()
	DDNSShared = NewDDNSClass()
}

// InitFrontendTemplates 从内置文件中加载FrontendTemplates
func InitFrontendTemplates() {
	err := yaml.Unmarshal(frontendTemplatesYAML, &FrontendTemplates)
	if err != nil {
		panic(err)
	}
}

// InitDBFromPath 从给出的文件路径中加载数据库
func InitDBFromPath(path string) {
	//var err error
	//DB, err = gorm.Open(sqlite.Open(path), &gorm.Config{
	//	CreateBatchSize: 200,
	//})
	//if err != nil {
	//	panic(err)
	//}
	//if Conf.Debug {
	//	DB = DB.Debug()
	//}
	//err = DB.AutoMigrate(model.Server{}, model.User{}, model.ServerGroup{}, model.NotificationGroup{},
	//	model.Notification{}, model.AlertRule{}, model.Service{}, model.NotificationGroupNotification{},
	//	model.ServiceHistory{}, model.Cron{}, model.Transfer{}, model.ServerGroupServer{},
	//	model.NAT{}, model.DDNSProfile{}, model.NotificationGroupNotification{},
	//	model.WAF{}, model.Oauth2Bind{})
	//if err != nil {
	//	panic(err)
	//}
	mysqlConfig := mysql.Config{
		DSN:                       path,  // DSN data source name
		DefaultStringSize:         191,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据版本自动配置
	}
	if db, err := gorm.Open(mysql.New(mysqlConfig), gormConfig()); err != nil {
		panic(err)
	} else {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(0)
		sqlDB.SetMaxOpenConns(0)
		DB = db.Debug()
	}
	err := DB.AutoMigrate(model.Server{}, model.User{}, model.ServerGroup{}, model.NotificationGroup{},
		model.Notification{}, model.AlertRule{}, model.Service{}, model.NotificationGroupNotification{},
		model.ServiceHistory{}, model.Cron{}, model.Transfer{}, model.ServerGroupServer{},
		model.NAT{}, model.DDNSProfile{}, model.NotificationGroupNotification{},
		model.WAF{}, model.Oauth2Bind{}, model.Tool{}, model.ToolGroup{}, model.ToolGroupTool{}, model.Upload{},
		model.Topic{}, model.TopicGroup{}, model.TopicGroupTopic{}, model.Favorite{}, model.UserLike{}, model.Comment{},
		model.UserAdditionalInfo{})
	if err != nil {
		panic(err)
	}
}

func gormConfig() *gorm.Config {
	config := &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true}
	return config
}

// RecordTransferHourlyUsage 对流量记录进行打点
func RecordTransferHourlyUsage() {
	ServerShared.listMu.RLock()
	defer ServerShared.listMu.RUnlock()

	now := time.Now()
	nowTrimSeconds := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
	var txs []model.Transfer
	for id, server := range ServerShared.list {
		tx := model.Transfer{
			ServerID: id,
			In:       utils.Uint64SubInt64(server.State.NetInTransfer, server.PrevTransferInSnapshot),
			Out:      utils.Uint64SubInt64(server.State.NetOutTransfer, server.PrevTransferOutSnapshot),
		}
		if tx.In == 0 && tx.Out == 0 {
			continue
		}
		server.PrevTransferInSnapshot = int64(server.State.NetInTransfer)
		server.PrevTransferOutSnapshot = int64(server.State.NetOutTransfer)
		tx.CreatedAt = nowTrimSeconds
		txs = append(txs, tx)
	}
	if len(txs) == 0 {
		return
	}
	log.Printf("NEZHA>> Saved traffic metrics to database. Affected %d row(s), Error: %v", len(txs), DB.Create(txs).Error)
}

// CleanServiceHistory 清理无效或过时的 监控记录 和 流量记录
func CleanServiceHistory() {
	// 清理已被删除的服务器的监控记录与流量记录
	DB.Unscoped().Delete(&model.ServiceHistory{}, "created_at < ? OR service_id NOT IN (SELECT `id` FROM services)", time.Now().AddDate(0, 0, -30))
	// 由于网络监控记录的数据较多，并且前端仅使用了 1 天的数据
	// 考虑到 sqlite 数据量问题，仅保留一天数据，
	// server_id = 0 的数据会用于/service页面的可用性展示
	DB.Unscoped().Delete(&model.ServiceHistory{}, "(created_at < ? AND server_id != 0) OR service_id NOT IN (SELECT `id` FROM services)", time.Now().AddDate(0, 0, -1))
	DB.Unscoped().Delete(&model.Transfer{}, "server_id NOT IN (SELECT `id` FROM servers)")
	// 计算可清理流量记录的时长
	var allServerKeep time.Time
	specialServerKeep := make(map[uint64]time.Time)
	var specialServerIDs []uint64
	var alerts []model.AlertRule
	DB.Find(&alerts)
	for _, alert := range alerts {
		for _, rule := range alert.Rules {
			// 是不是流量记录规则
			if !rule.IsTransferDurationRule() {
				continue
			}
			dataCouldRemoveBefore := rule.GetTransferDurationStart().UTC()
			// 判断规则影响的机器范围
			if rule.Cover == model.RuleCoverAll {
				// 更新全局可以清理的数据点
				if allServerKeep.IsZero() || allServerKeep.After(dataCouldRemoveBefore) {
					allServerKeep = dataCouldRemoveBefore
				}
			} else {
				// 更新特定机器可以清理数据点
				for id := range rule.Ignore {
					if specialServerKeep[id].IsZero() || specialServerKeep[id].After(dataCouldRemoveBefore) {
						specialServerKeep[id] = dataCouldRemoveBefore
						specialServerIDs = append(specialServerIDs, id)
					}
				}
			}
		}
	}
	for id, couldRemove := range specialServerKeep {
		DB.Unscoped().Delete(&model.Transfer{}, "server_id = ? AND datetime(`created_at`) < datetime(?)", id, couldRemove)
	}
	if allServerKeep.IsZero() {
		DB.Unscoped().Delete(&model.Transfer{}, "server_id NOT IN (?)", specialServerIDs)
	} else {
		DB.Unscoped().Delete(&model.Transfer{}, "server_id NOT IN (?) AND datetime(`created_at`) < datetime(?)", specialServerIDs, allServerKeep)
	}
}

// IPDesensitize 根据设置选择是否对IP进行打码处理 返回处理后的IP(关闭打码则返回原IP)
func IPDesensitize(ip string) string {
	if Conf.EnablePlainIPInNotification {
		return ip
	}
	return utils.IPDesensitize(ip)
}

type class[K comparable, V model.CommonInterface] struct {
	list   map[K]V
	listMu sync.RWMutex

	sortedList   []V
	sortedListMu sync.RWMutex
}

func (c *class[K, V]) Get(id K) (s V, ok bool) {
	c.listMu.RLock()
	defer c.listMu.RUnlock()

	s, ok = c.list[id]
	return
}

func (c *class[K, V]) GetList() map[K]V {
	c.listMu.RLock()
	defer c.listMu.RUnlock()

	return maps.Clone(c.list)
}

func (c *class[K, V]) GetSortedList() []V {
	c.sortedListMu.RLock()
	defer c.sortedListMu.RUnlock()

	return slices.Clone(c.sortedList)
}

func (c *class[K, V]) Range(fn func(k K, v V) bool) {
	c.listMu.RLock()
	defer c.listMu.RUnlock()

	for k, v := range c.list {
		if !fn(k, v) {
			break
		}
	}
}

func (c *class[K, V]) CheckPermission(ctx *gin.Context, idList iter.Seq[K]) bool {
	c.listMu.RLock()
	defer c.listMu.RUnlock()

	for id := range idList {
		if s, ok := c.list[id]; ok {
			if !s.HasPermission(ctx) {
				return false
			}
		}
	}
	return true
}
