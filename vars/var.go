package vars

// app version info
var (
	AppName    = "rcm" // redis-cluster-manager
	AppVersion = "unknown"
	GoVersion  = "unknown"
	BuildTime  = "unknown"
	GitCommit  = "unknown"
	GitRemote  = "unknown"
)

// redis server info, ClusterName&ClusterID not supported now
// HostPort is the seed node address, e.g. "x.x.x.x:6379"
var (
	HostPort    string
	ClusterName string
	ClusterID   string
	Password    string // password for redis default user 'default'
)

var ForbiddenCmds = map[string]struct{}{
	"DEBUG":    {},
	"FLUSHALL": {},
	"FLUSHDB":  {},
	"SHUTDOWN": {},
	"MONITOR":  {},
}

// filter types for cluster nodes when executing commands
const (
	FILTER_NONE = iota
	FILTER_NODEID
	FILTER_ADDR
	FILTER_ROLE
)
