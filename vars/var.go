package vars

import "time"

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
	Password    string        // password for redis default user 'default'
	Timeout     time.Duration // timeout duration for redis
)

//goland:noinspection ALL
var ForbiddenCmds = map[string]struct{}{
	"DEBUG":    {},
	"FLUSHALL": {},
	"FLUSHDB":  {},
	"SHUTDOWN": {},
	"MONITOR":  {},
}

// filter types for cluster nodes when executing commands
//
//goland:noinspection ALL
const (
	FILTER_NONE = iota
	FILTER_NODEID
	FILTER_ADDR
	FILTER_ROLE
)

// roles when filter type is FILTER_ROLE
//
//goland:noinspection ALL
const (
	ROLE_MASTER = "master"
	ROLE_SLAVE  = "slave"
	ROLE_ALL    = "all"
)

// perf profiler flags
var (
	CPUProfiler bool
	MEMProfiler bool
)
