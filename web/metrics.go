package web

// metrics module provides various metrics about our server
//
// Copyright (c) 2020 - Valentin Kuznetsov <vkuznet AT gmail dot com>

import (
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/dmwm/dbs2go/dbs"
	"github.com/dmwm/dbs2go/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

// TotalGetRequests counts total number of GET requests received by the server
var TotalGetRequests uint64

// TotalPostRequests counts total number of POST requests received by the server
var TotalPostRequests uint64

// TotalPutRequests counts total number of PUT requests received by the server
var TotalPutRequests uint64

// MetricsLastUpdateTime keeps track of last update time of the metrics
var MetricsLastUpdateTime time.Time

// RPS represents requests per second for a given server
var RPS float64

// RPSPhysical represents requests per second for a given server times number of physical CPU cores
var RPSPhysical float64

// RPSLogical represents requests per second for a given server times number of logical CPU cores
var RPSLogical float64

// NumPhysicalCores represents number of cores in our node
var NumPhysicalCores int

// NumLogicalCores represents number of cores in our node
var NumLogicalCores int

// AvgGetRequestTime represents average GET request time
var AvgGetRequestTime float64

// AvgPostRequestTime represents average POST request time
var AvgPostRequestTime float64

// AvgPutRequestTime represents average PUT request time
var AvgPutRequestTime float64

// RequestStats holds metrics related to number of requests on a server
type RequestStats struct {
	TotalGetRequests  uint64
	TotalPostRequests uint64
	TotalPutRequests  uint64
	Time              time.Time
	NumPhysicalCores  int
	NumLogicalCores   int
}

// Update RequestStatus metrics
func (r *RequestStats) Update() {
	r.TotalGetRequests = TotalGetRequests
	r.TotalPostRequests = TotalPostRequests
	r.TotalPutRequests = TotalPutRequests
	r.NumPhysicalCores = NumPhysicalCores
	r.NumLogicalCores = NumLogicalCores
	r.Time = time.Now()
}

var rstat *RequestStats

// Memory structure keeps track of server memory
type Memory struct {
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
}

// Mem structure keeps track of virtual/swap memory of the server
type Mem struct {
	Virtual Memory `json:"virtual"` // virtual memory metrics from gopsutils
	Swap    Memory `json:"swap"`    // swap memory metrics from gopsutils
}

// Metrics provide various metrics about our server
type Metrics struct {
	CPU                []float64               `json:"cpu"`                // cpu metrics from gopsutils
	CpuPercent         float64                 `json:"cpu_pct"`            // cpu percent
	Connections        []net.ConnectionStat    `json:"connections"`        // connections metrics from gopsutils
	Load               load.AvgStat            `json:"load"`               // load metrics from gopsutils
	Memory             Mem                     `json:"memory"`             // memory metrics from gopsutils
	OpenFiles          []process.OpenFilesStat `json:"openFiles"`          // open files metrics from gopsutils
	GoRoutines         uint64                  `json:"goroutines"`         // total number of go routines at run-time
	Uptime             float64                 `json:"uptime"`             // uptime of the server
	GetRequests        uint64                  `json:"getRequests"`        // total number of get requests across all services
	PostRequests       uint64                  `json:"postRequests"`       // total number of post requests across all services
	PutRequests        uint64                  `json:"putRequests"`        // total number of post requests across all services
	AvgGetTime         float64                 `json:"avgGetTime"`         // avg GET request time
	AvgPostTime        float64                 `json:"avgPostTime"`        // avg POST request time
	AvgPutTime         float64                 `json:"avgPutTime"`         // avg PUT request time
	RPS                float64                 `json:"rps"`                // throughput req/sec
	RPSPhysical        float64                 `json:"rpsPhysical"`        // throughput req/sec using physical cpu
	RPSLogical         float64                 `json:"rpsLogical"`         // throughput req/sec using logical cpu
	ProcFS             utils.ProcFS            `json:"procfs"`             // metrics from prometheus procfs
	DBStats            sql.DBStats             `json:"dbstats"`            // metrics about database
	MaxDBConnections   uint64                  `json:"maxDBConnections"`   // max number of DB connections
	MaxIdleConnections uint64                  `json:"maxIdleConnections"` // max number of idle DB connections

	// Migration server metrics
	MigrationRequests   uint64 `json:"migrationRequests"`   // total number of migration requests across all services
	MigrationPending    uint64 `json:"migrationPending"`    // total number of pending migration requests across all services
	MigrationInProgress uint64 `json:"migrationInProgress"` // total number of in progress migration requests across all services
	MigrationFailed     uint64 `json:"migrationFailed"`     // total number of failed migration requests across all services
	MigrationTermFailed uint64 `json:"migrationTermFailed"` // total number of term failed migration requests across all services
	MigrationCompleted  uint64 `json:"migrationCompleted"`  // total number of completed migration requests across all services
	MigrationQueued     uint64 `json:"migrationQueued"`     // total number of queued migration requests across all services
	MigrationExistInDB  uint64 `json:"migrationExistInDB"`  // total number of exist in db migration requests across all services
}

// DBSMetrics stores custom metrics related to DBS.
// commented out metrics are implemented as CounterFunc in InitMetrics
type DBSMetrics struct {
	cpuInfo []prometheus.Gauge

	totalConnections       prometheus.Gauge
	establishedConnections prometheus.Gauge
	listenConnections      prometheus.Gauge

	// procfs metrics
	cpuTotal prometheus.Gauge
	vsize    prometheus.Gauge
	rss      prometheus.Gauge
	openfds  prometheus.Gauge
	maxfds   prometheus.Gauge
	maxvsize prometheus.Gauge

	// procfs /proc/stat metrics
	sumusercpus   prometheus.Gauge
	sumsystemcpus prometheus.Gauge

	// cpu percent
	cpupct prometheus.Gauge

	// load
	load1  prometheus.Gauge
	load5  prometheus.Gauge
	load15 prometheus.Gauge

	// memory virtual
	memvirttotal prometheus.Gauge
	memvirtfree  prometheus.Gauge
	memvirtused  prometheus.Gauge
	memvirtpct   prometheus.Gauge

	// memory swap
	memswaptotal prometheus.Gauge
	memswapfree  prometheus.Gauge
	memswapused  prometheus.Gauge
	memswappct   prometheus.Gauge

	// open files
	openfiles prometheus.Gauge

	// go routines
	// goroutines prometheus.Counter

	// uptime
	// uptime prometheus.Counter

	// total requests
	// getrequests  prometheus.Counter
	// postrequests prometheus.Counter
	// putrequests  prometheus.Counter

	// throughput, rps, rps physical cpu, rps logical cpu
	rps            prometheus.Gauge
	avggettime     prometheus.Gauge
	avgposttime    prometheus.Gauge
	avgputtime     prometheus.Gauge
	rpsphysicalcpu prometheus.Gauge
	rpslogicalcpu  prometheus.Gauge

	// database metrics
	maxdbconn   prometheus.Gauge
	maxidleconn prometheus.Gauge

	// see https://pkg.go.dev/database/sql#DBStats
	maxopenconn prometheus.Gauge
	openconn    prometheus.Gauge
	inuseconn   prometheus.Gauge
	idleconn    prometheus.Gauge
	// waitcount         prometheus.Counter
	// waitduration      prometheus.Counter
	// maxidleclosed     prometheus.Counter
	// maxidletimeclosed prometheus.Counter
	// maxlifetimeclosed prometheus.Counter

	// migration server metrics
	// requests   prometheus.Counter
	// pending    prometheus.Counter
	// inprogress prometheus.Counter
	// failed     prometheus.Counter
	// termfailed prometheus.Counter
	// completed  prometheus.Counter
	// queued     prometheus.Counter
	// existindb  prometheus.Counter
}

// contains the prometheus gauges and counters for DBSMetrics using the default registry
var dbsMetrics DBSMetrics

// InitMetrics initializes DBS metrics into the default Prometheus metrics registry
func InitMetrics(prefix string) {
	// prometheus.MustRegister(prometheus.NewGoCollector())
	c, _ := cpu.Percent(time.Millisecond, true)
	dbsMetrics.cpuInfo = make([]prometheus.Gauge, len(c))
	for i := range dbsMetrics.cpuInfo {
		dbsMetrics.cpuInfo[i] = promauto.NewGauge(prometheus.GaugeOpts{
			Namespace:   prefix,
			Name:        "cpu",
			Help:        "percentage of cpu used per cpu",
			ConstLabels: prometheus.Labels{"core": strconv.FormatInt(int64(i), 10)},
		})
	}

	dbsMetrics.totalConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "total_connections",
		Help:      "total number of connections",
	})
	dbsMetrics.establishedConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "established_connections",
		Help:      "established connections",
	})
	dbsMetrics.listenConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "listen_connections",
		Help:      "listen connections",
	})

	// procfs metrics
	dbsMetrics.cpuTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "procf_cputotal",
		Help:      "ProcFS CPU Total",
	})
	dbsMetrics.vsize = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "procf_vsize",
		Help:      "ProcFS vsize",
	})
	dbsMetrics.rss = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "procfs_rss",
		Help:      "ProcFS RSS",
	})
	dbsMetrics.openfds = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "procfs_openfds",
		Help:      "ProcFS OpenFDs",
	})
	dbsMetrics.maxfds = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      fmt.Sprintf("%s_procfs_maxfds", prefix),
		Help:      "ProcFS MaxFDs",
	})
	dbsMetrics.maxvsize = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "procfs_maxvsize",
		Help:      "ProcFS MaxVSize",
	})

	// procfs /proc/stat metrics
	dbsMetrics.sumusercpus = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "procfs_sumusercpus",
		Help:      "ProcFS SumUserCPUs",
	})
	dbsMetrics.sumsystemcpus = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "procfs_sumsystemcpus",
		Help:      "ProcFS SumSystemCPUs",
	})

	// cpu percent
	dbsMetrics.cpupct = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "cpu_pct",
	})

	// load
	dbsMetrics.load1 = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "load1",
	})
	dbsMetrics.load5 = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "load5",
	})
	dbsMetrics.load15 = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "load15",
	})

	// memory virtual
	dbsMetrics.memvirttotal = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "mem_virt_total",
		Help:      "reports total virtual memory in bytes",
	})
	dbsMetrics.memvirtfree = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "mem_virt_free",
		Help:      "reports free virtual memory in bytes",
	})
	dbsMetrics.memvirtused = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "mem_virt_used",
		Help:      "reports used virtual memory in bytes",
	})
	dbsMetrics.memvirtpct = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "mem_virt_pct",
		Help:      "reports percentage of virtual memory",
	})

	// memory swap
	dbsMetrics.memswaptotal = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "mem_swap_total",
		Help:      "reports total swap memory in bytes",
	})
	dbsMetrics.memswapfree = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "mem_swap_free",
		Help:      "reports free swap memory in bytes",
	})
	dbsMetrics.memswapused = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "mem_swap_used",
		Help:      "reports used swap memory in bytes",
	})
	dbsMetrics.memswappct = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "mem_swap_pct",
		Help:      "reports percentage swap memory",
	})

	// open files
	dbsMetrics.openfiles = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "open_files",
		Help:      "reports total number of open file descriptors",
	})

	// go routines
	if err := prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: prefix,
			Name:      "goroutines",
			Help:      "reports total number of go routines",
		},
		func() float64 { return float64(runtime.NumGoroutine()) },
	)); err == nil {
		fmt.Println("CounterFunc '_goroutines' registered")
	}

	// go uptime
	if err := prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: prefix,
			Name:      "uptime",
			Help:      "reports server uptime in seconds",
		},
		func() float64 { return float64(time.Since(StartTime).Seconds()) },
	)); err == nil {
		fmt.Println("CounterFunc '_uptime' registered")
	}

	// go uptime
	if err := prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: prefix,
			Name:      "get_requests",
			Help:      "reports total number of HTTP GET requests",
		},
		func() float64 { return float64(TotalGetRequests) },
	)); err == nil {
		fmt.Println("CounterFunc '_get_requests' registered")
	}
	if err := prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: prefix,
			Name:      "post_requests",
			Help:      "reports total number of HTTP POST requests",
		},
		func() float64 { return float64(TotalPostRequests) },
	)); err == nil {
		fmt.Println("CounterFunc '_post_requests' registered")
	}
	if err := prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: prefix,
			Name:      "put_requests",
			Help:      "reports total number of HTTP PUT requests",
		},
		func() float64 { return float64(TotalPutRequests) },
	)); err == nil {
		fmt.Println("CounterFunc '_put_requests' registered")
	}

	// throughput, rps, rps physical cpu, rps logical cpu
	dbsMetrics.rps = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "rps",
		Help:      "reports request per second average",
	})
	dbsMetrics.avggettime = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "avg_get_time",
		Help:      "reports average get request time",
	})
	dbsMetrics.avgposttime = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "avg_post_time",
		Help:      "reports average post request time",
	})
	dbsMetrics.avgputtime = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "avg_put_time",
		Help:      "reports average put request time",
	})
	dbsMetrics.rpsphysicalcpu = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "rps_physical_cpu",
		Help:      "reports request per second average weighted by physical CPU cores",
	})
	dbsMetrics.rpslogicalcpu = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "rps_logical_cpu",
		Help:      "reports request per second average weighted by logical CPU cores",
	})

	// database metrics
	dbsMetrics.maxdbconn = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "max_db_connections",
		Help:      "reports max number of DB conenctions",
	})
	dbsMetrics.maxidleconn = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "max_idle_connections",
		Help:      "reports max number of idle DB connections",
	})

	dbsMetrics.maxopenconn = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "max_open_connections",
		Help:      "reports max number of open DB connections",
	})
	dbsMetrics.openconn = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "open_connections",
		Help:      "reports number of established to database (both in use and idle)",
	})
	dbsMetrics.inuseconn = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "in_use_connections",
		Help:      "reports number of in use database connections",
	})
	dbsMetrics.idleconn = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prefix,
		Name:      "idle_connections",
		Help:      "reports number of idle database connections",
	})
	// register wait_count
	if err := prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: prefix,
			Name:      "wait_count",
			Help:      "reports total number of connections waited for",
		},
		func() float64 { return float64(dbs.DB.Stats().WaitCount) },
	)); err == nil {
		fmt.Println("CounterFunc '_wait_count' registered")
	}
	// register wait_duration
	if err := prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: prefix,
			Name:      "wait_duration",
			Help:      "reports total time (in sec) blocked waiting for a new connection",
		},
		func() float64 { return float64(dbs.DB.Stats().WaitDuration.Seconds()) },
	)); err == nil {
		fmt.Println("CounterFunc '_wait_duration' registered")
	}
	// register max_idle_closed
	if err := prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: prefix,
			Name:      "max_idle_closed",
			Help:      "reports total number of connections closed due to SetMaxIdleConns",
		},
		func() float64 { return float64(dbs.DB.Stats().MaxIdleClosed) },
	)); err == nil {
		fmt.Println("CounterFunc '_max_idle_closed' registered")
	}
	// register max_idle_time_closed
	if err := prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: prefix,
			Name:      "max_idle_time_closed",
			Help:      "reports total number of connections closed due to SetConnMaxIdleTime",
		},
		func() float64 { return float64(dbs.DB.Stats().MaxIdleTimeClosed) },
	)); err == nil {
		fmt.Println("CounterFunc '_max_idle_time_closed' registered")
	}
	// register max_lifetime_closed
	if err := prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: prefix,
			Name:      "max_lifetime_closed",
			Help:      "reports total number of connections closed due to SetConnMaxLifetime",
		},
		func() float64 { return float64(dbs.DB.Stats().MaxLifetimeClosed) },
	)); err == nil {
		fmt.Println("CounterFunc '_max_lifetime_closed' registered")
	}

	// migration server
	// register _requests
	if err := prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: prefix,
			Name:      "requests",
			Help:      "reports total number of migration requests",
		},
		func() float64 { return float64(dbs.TotalMigrationRequests) },
	)); err == nil {
		fmt.Println("CounterFunc '_requests' registered")
	}
	// register _pending
	if err := prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: prefix,
			Name:      "pending",
			Help:      "reports total number of pending migration requests",
		},
		func() float64 { return float64(dbs.TotalPending) },
	)); err == nil {
		fmt.Println("CounterFunc '_pending' registered")
	}
	// register _in_progress
	if err := prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: prefix,
			Name:      "in_progress",
			Help:      "reports total number of in progress migration requests",
		},
		func() float64 { return float64(dbs.TotalInProgress) },
	)); err == nil {
		fmt.Println("CounterFunc '_in_progress' registered")
	}
	// register _failed
	if err := prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: prefix,
			Name:      "failed",
			Help:      "reports total number of failed migration requests",
		},
		func() float64 { return float64(dbs.TotalFailed) },
	)); err == nil {
		fmt.Println("CounterFunc '_failed' registered")
	}
	// register _term_failed
	if err := prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: prefix,
			Name:      "term_failed",
			Help:      "reports total number of term failed migration requests",
		},
		func() float64 { return float64(dbs.TotalTermFailed) },
	)); err == nil {
		fmt.Println("CounterFunc '_term_failed' registered")
	}
	// register _completed
	if err := prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: prefix,
			Name:      "completed",
			Help:      "reports total number of completed migration requests",
		},
		func() float64 { return float64(dbs.TotalCompleted) },
	)); err == nil {
		fmt.Println("CounterFunc '_completed' registered")
	}
	// register _queued
	if err := prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: prefix,
			Name:      "queued",
			Help:      "reports total number of queued migration requests",
		},
		func() float64 { return float64(dbs.TotalQueued) },
	)); err == nil {
		fmt.Println("CounterFunc '_queued' registered")
	}
	// register _exist_in_db
	if err := prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Namespace: prefix,
			Name:      "exist_in_db",
			Help:      "reports total number of exist in db migration requests",
		},
		func() float64 { return float64(dbs.TotalExistInDB) },
	)); err == nil {
		fmt.Println("CounterFunc '_exist_in_db' registered")
	}
}

func recordMetrics(m *Metrics) {
	for i := range m.CPU {
		dbsMetrics.cpuInfo[i].Set(m.CPU[i])
	}

	totCon, estCon, lisCon := getConnections(m.Connections)
	dbsMetrics.totalConnections.Set(totCon)
	dbsMetrics.establishedConnections.Set(estCon)
	dbsMetrics.listenConnections.Set(lisCon)

	// procfs metrics
	dbsMetrics.cpuTotal.Set(m.ProcFS.CpuTotal)
	dbsMetrics.vsize.Set(m.ProcFS.Vsize)
	dbsMetrics.rss.Set(m.ProcFS.Rss)
	dbsMetrics.openfds.Set(m.ProcFS.OpenFDs)
	dbsMetrics.maxfds.Set(m.ProcFS.MaxFDs)
	dbsMetrics.maxvsize.Set(m.ProcFS.MaxVsize)

	// procfs /proc/stat metrics
	dbsMetrics.sumusercpus.Set(m.ProcFS.SumUserCPUs)
	dbsMetrics.sumsystemcpus.Set(m.ProcFS.SumSystemCPUs)

	// cpu percent
	dbsMetrics.cpupct.Set(m.CpuPercent)

	// load
	dbsMetrics.load1.Set(m.Load.Load1)
	dbsMetrics.load5.Set(m.Load.Load5)
	dbsMetrics.load15.Set(m.Load.Load15)

	// memory virtual
	dbsMetrics.memvirttotal.Set(float64(m.Memory.Virtual.Total))
	dbsMetrics.memvirtfree.Set(float64(m.Memory.Virtual.Free))
	dbsMetrics.memvirtused.Set(float64(m.Memory.Virtual.Used))
	dbsMetrics.memvirtpct.Set(m.Memory.Virtual.UsedPercent)

	// memory swap
	dbsMetrics.memswaptotal.Set(float64(m.Memory.Swap.Total))
	dbsMetrics.memswapfree.Set(float64(m.Memory.Swap.Free))
	dbsMetrics.memswapused.Set(float64(m.Memory.Swap.Used))
	dbsMetrics.memswappct.Set(m.Memory.Swap.UsedPercent)

	// open files
	dbsMetrics.openfiles.Set(float64(len(m.OpenFiles)))

	// go routines
	// registered in InitMetrics

	// uptime
	// registered in InitMetrics

	// total requests
	// registered in InitMetrics

	// throughput, rps, rps physical cpu, rps logical cpu
	dbsMetrics.rps.Set(m.RPS)
	dbsMetrics.avggettime.Set(m.AvgGetTime)
	dbsMetrics.avgposttime.Set(m.AvgPostTime)
	dbsMetrics.avgputtime.Set(m.AvgPutTime)
	dbsMetrics.rpsphysicalcpu.Set(m.RPSPhysical)
	dbsMetrics.rpslogicalcpu.Set(m.RPSLogical)

	// database metrics
	dbsMetrics.maxdbconn.Set(float64(m.MaxDBConnections))
	dbsMetrics.maxidleconn.Set(float64(m.MaxIdleConnections))

	// see https://pkg.go.dev/database/sql#DBStats
	dbsMetrics.maxopenconn.Set(float64(m.DBStats.MaxOpenConnections))
	dbsMetrics.openconn.Set(float64(m.DBStats.OpenConnections))
	dbsMetrics.inuseconn.Set(float64(m.DBStats.InUse))
	dbsMetrics.idleconn.Set(float64(m.DBStats.Idle))
}

// fetches dbs2go metrics
func metrics() Metrics {
	if rstat == nil {
		rstat = &RequestStats{}
		rstat.Time = time.Now()
	}

	// get cpu and mem profiles
	m, _ := mem.VirtualMemory()
	s, _ := mem.SwapMemory()
	l, _ := load.Avg()
	c, _ := cpu.Percent(time.Millisecond, true)
	process, perr := process.NewProcess(int32(os.Getpid()))

	// get unfinished queries
	metrics := Metrics{}
	metrics.GoRoutines = uint64(runtime.NumGoroutine())
	virt := Memory{Total: m.Total, Free: m.Free, Used: m.Used, UsedPercent: m.UsedPercent}
	swap := Memory{Total: s.Total, Free: s.Free, Used: s.Used, UsedPercent: s.UsedPercent}
	metrics.Memory = Mem{Virtual: virt, Swap: swap}
	metrics.Load = *l
	metrics.CPU = c
	if perr == nil { // if we got process info
		conn, err := process.Connections()
		if err == nil {
			metrics.Connections = conn
		}
		openFiles, err := process.OpenFiles()
		if err == nil {
			metrics.OpenFiles = openFiles
		}
	}

	// get cpu percent
	cpuPct, err := process.Percent(time.Duration(1 * time.Second))
	if err == nil {
		metrics.CpuPercent = cpuPct
	}

	metrics.MaxIdleConnections = uint64(Config.MaxIdleConnections)
	metrics.MaxDBConnections = uint64(Config.MaxDBConnections)
	metrics.DBStats = dbs.DB.Stats()
	metrics.ProcFS = utils.ProcFSMetrics()
	metrics.Uptime = time.Since(StartTime).Seconds()

	metrics.AvgGetTime = AvgGetRequestTime
	metrics.AvgPostTime = AvgPostRequestTime
	metrics.AvgPutTime = AvgPutRequestTime

	metrics.GetRequests = TotalGetRequests
	metrics.PostRequests = TotalPostRequests
	metrics.PutRequests = TotalPutRequests

	lapse := time.Since(rstat.Time).Seconds()
	total := float64(TotalGetRequests + TotalPostRequests + TotalPutRequests)
	metrics.RPS = (total - float64(rstat.TotalGetRequests+rstat.TotalPostRequests+rstat.TotalPutRequests)) / lapse
	metrics.RPSLogical = float64(rstat.NumLogicalCores-NumLogicalCores) / lapse
	metrics.RPSPhysical = float64(rstat.NumPhysicalCores-NumPhysicalCores) / lapse

	// migration server metrics
	metrics.MigrationRequests = dbs.TotalMigrationRequests
	metrics.MigrationPending = dbs.TotalPending
	metrics.MigrationInProgress = dbs.TotalInProgress
	metrics.MigrationFailed = dbs.TotalFailed
	metrics.MigrationTermFailed = dbs.TotalTermFailed
	metrics.MigrationCompleted = dbs.TotalCompleted
	metrics.MigrationQueued = dbs.TotalQueued
	metrics.MigrationExistInDB = dbs.TotalExistInDB

	rstat.Update()

	return metrics
}

// get latest metrics and records them to prometheus registry
func promMetrics2() {
	data := metrics()

	recordMetrics(&data)
}

// get connection metrics
func getConnections(c []net.ConnectionStat) (float64, float64, float64) {
	var totCon, estCon, lisCon float64
	for _, c := range c {
		v := c.Status
		switch v {
		case "ESTABLISHED":
			estCon++
		case "LISTEN":
			lisCon++
		}
	}
	totCon = float64(len(c))
	return totCon, estCon, lisCon
}

// helper function to generate metrics in prometheus format
func promMetrics(prefix string) string {
	var out string
	data := metrics()

	// cpu info
	out += fmt.Sprintf("# HELP %s_cpu percentage of cpu used per CPU\n", prefix)
	out += fmt.Sprintf("# TYPE %s_cpu gauge\n", prefix)
	for i, v := range data.CPU {
		out += fmt.Sprintf("%s_cpu{core=\"%d\"} %v\n", prefix, i, v)
	}

	// connections
	totCon, estCon, lisCon := getConnections(data.Connections)
	out += fmt.Sprintf("# HELP %s_total_connections\n", prefix)
	out += fmt.Sprintf("# TYPE %s_total_connections gauge\n", prefix)
	out += fmt.Sprintf("%s_total_connections %v\n", prefix, totCon)
	out += fmt.Sprintf("# HELP %s_established_connections\n", prefix)
	out += fmt.Sprintf("# TYPE %s_established_connections gauge\n", prefix)
	out += fmt.Sprintf("%s_established_connections %v\n", prefix, estCon)
	out += fmt.Sprintf("# HELP %s_listen_connections\n", prefix)
	out += fmt.Sprintf("# TYPE %s_listen_connections gauge\n", prefix)
	out += fmt.Sprintf("%s_listen_connections %v\n", prefix, lisCon)

	// procfs metrics
	// cpuTotal, vsize, rss, openFDs, maxFDs, maxVsize
	out += fmt.Sprintf("# HELP %s_procfs_cputotal\n", prefix)
	out += fmt.Sprintf("# TYPE %s_procfs_cputotal gauge\n", prefix)
	out += fmt.Sprintf("%s_procfs_cputotal %v\n", prefix, data.ProcFS.CpuTotal)
	out += fmt.Sprintf("# HELP %s_procfs_vsize\n", prefix)
	out += fmt.Sprintf("# TYPE %s_procfs_vsize gauge\n", prefix)
	out += fmt.Sprintf("%s_procfs_vsize %v\n", prefix, data.ProcFS.Vsize)
	out += fmt.Sprintf("# HELP %s_procfs_rss\n", prefix)
	out += fmt.Sprintf("# TYPE %s_procfs_rss gauge\n", prefix)
	out += fmt.Sprintf("%s_procfs_rss %v\n", prefix, data.ProcFS.Rss)
	out += fmt.Sprintf("# HELP %s_procfs_openfds\n", prefix)
	out += fmt.Sprintf("# TYPE %s_procfs_openfds gauge\n", prefix)
	out += fmt.Sprintf("%s_procfs_openfds %v\n", prefix, data.ProcFS.OpenFDs)
	out += fmt.Sprintf("# HELP %s_procfs_maxfds\n", prefix)
	out += fmt.Sprintf("# TYPE %s_procfs_maxfds gauge\n", prefix)
	out += fmt.Sprintf("%s_procfs_maxfds %v\n", prefix, data.ProcFS.MaxFDs)
	out += fmt.Sprintf("# HELP %s_procfs_maxvsize\n", prefix)
	out += fmt.Sprintf("# TYPE %s_procfs_maxvsize gauge\n", prefix)
	out += fmt.Sprintf("%s_procfs_maxvsize %v\n", prefix, data.ProcFS.MaxVsize)

	// procfs /proc/stat metrics
	out += fmt.Sprintf("# HELP %s_procfs_sumusercpus\n", prefix)
	out += fmt.Sprintf("# TYPE %s_procfs_sumusercpus gauge\n", prefix)
	out += fmt.Sprintf("%s_procfs_sumusercpus %v\n", prefix, data.ProcFS.SumUserCPUs)
	out += fmt.Sprintf("# HELP %s_procfs_sumsystemcpus\n", prefix)
	out += fmt.Sprintf("# TYPE %s_procfs_sumsystemcpus gauge\n", prefix)
	out += fmt.Sprintf("%s_procfs_sumsystemcpus %v\n", prefix, data.ProcFS.SumSystemCPUs)

	// cpu percent
	out += fmt.Sprintf("# HELP %s_cpu_pct\n", prefix)
	out += fmt.Sprintf("# TYPE %s_cpu_pct gauge\n", prefix)
	out += fmt.Sprintf("%s_cpu_pct %v\n", prefix, data.CpuPercent)

	// load
	out += fmt.Sprintf("# HELP %s_load1\n", prefix)
	out += fmt.Sprintf("# TYPE %s_load1 gauge\n", prefix)
	out += fmt.Sprintf("%s_load1 %v\n", prefix, data.Load.Load1)
	out += fmt.Sprintf("# HELP %s_load5\n", prefix)
	out += fmt.Sprintf("# TYPE %s_load5 gauge\n", prefix)
	out += fmt.Sprintf("%s_load5 %v\n", prefix, data.Load.Load5)
	out += fmt.Sprintf("# HELP %s_load15\n", prefix)
	out += fmt.Sprintf("# TYPE %s_load15 gauge\n", prefix)
	out += fmt.Sprintf("%s_load15 %v\n", prefix, data.Load.Load15)

	// memory virtual
	out += fmt.Sprintf("# HELP %s_mem_virt_total reports total virtual memory in bytes\n", prefix)
	out += fmt.Sprintf("# TYPE %s_mem_virt_total gauge\n", prefix)
	out += fmt.Sprintf("%s_mem_virt_total %v\n", prefix, data.Memory.Virtual.Total)
	out += fmt.Sprintf("# HELP %s_mem_virt_free reports free virtual memory in bytes\n", prefix)
	out += fmt.Sprintf("# TYPE %s_mem_virt_free gauge\n", prefix)
	out += fmt.Sprintf("%s_mem_virt_free %v\n", prefix, data.Memory.Virtual.Free)
	out += fmt.Sprintf("# HELP %s_mem_virt_used reports used virtual memory in bytes\n", prefix)
	out += fmt.Sprintf("# TYPE %s_mem_virt_used gauge\n", prefix)
	out += fmt.Sprintf("%s_mem_virt_used %v\n", prefix, data.Memory.Virtual.Used)
	out += fmt.Sprintf("# HELP %s_mem_virt_pct reports percentage of virtual memory\n", prefix)
	out += fmt.Sprintf("# TYPE %s_mem_virt_pct gauge\n", prefix)
	out += fmt.Sprintf("%s_mem_virt_pct %v\n", prefix, data.Memory.Virtual.UsedPercent)

	// memory swap
	out += fmt.Sprintf("# HELP %s_mem_swap_total reports total swap memory in bytes\n", prefix)
	out += fmt.Sprintf("# TYPE %s_mem_swap_total gauge\n", prefix)
	out += fmt.Sprintf("%s_mem_swap_total %v\n", prefix, data.Memory.Swap.Total)
	out += fmt.Sprintf("# HELP %s_mem_swap_free reports free swap memory in bytes\n", prefix)
	out += fmt.Sprintf("# TYPE %s_mem_swap_free gauge\n", prefix)
	out += fmt.Sprintf("%s_mem_swap_free %v\n", prefix, data.Memory.Swap.Free)
	out += fmt.Sprintf("# HELP %s_mem_swap_used reports used swap memory in bytes\n", prefix)
	out += fmt.Sprintf("# TYPE %s_mem_swap_used gauge\n", prefix)
	out += fmt.Sprintf("%s_mem_swap_used %v\n", prefix, data.Memory.Swap.Used)
	out += fmt.Sprintf("# HELP %s_mem_swap_pct reports percentage swap memory\n", prefix)
	out += fmt.Sprintf("# TYPE %s_mem_swap_pct gauge\n", prefix)
	out += fmt.Sprintf("%s_mem_swap_pct %v\n", prefix, data.Memory.Swap.UsedPercent)

	// open files
	out += fmt.Sprintf("# HELP %s_open_files reports total number of open file descriptors\n", prefix)
	out += fmt.Sprintf("# TYPE %s_open_files gauge\n", prefix)
	out += fmt.Sprintf("%s_open_files %v\n", prefix, len(data.OpenFiles))

	// go routines
	out += fmt.Sprintf("# HELP %s_goroutines reports total number of go routines\n", prefix)
	out += fmt.Sprintf("# TYPE %s_goroutines counter\n", prefix)
	out += fmt.Sprintf("%s_goroutines %v\n", prefix, data.GoRoutines)

	// uptime
	out += fmt.Sprintf("# HELP %s_uptime reports server uptime in seconds\n", prefix)
	out += fmt.Sprintf("# TYPE %s_uptime counter\n", prefix)
	out += fmt.Sprintf("%s_uptime %v\n", prefix, data.Uptime)

	// total requests
	out += fmt.Sprintf("# HELP %s_get_requests reports total number of HTTP GET requests\n", prefix)
	out += fmt.Sprintf("# TYPE %s_get_requests counter\n", prefix)
	out += fmt.Sprintf("%s_get_requests %v\n", prefix, data.GetRequests)
	out += fmt.Sprintf("# HELP %s_post_requests reports total number of HTTP POST requests\n", prefix)
	out += fmt.Sprintf("# TYPE %s_post_requests counter\n", prefix)
	out += fmt.Sprintf("%s_post_requests %v\n", prefix, data.PostRequests)
	out += fmt.Sprintf("# HELP %s_put_requests reports total number of HTTP PUT requests\n", prefix)
	out += fmt.Sprintf("# TYPE %s_put_requests counter\n", prefix)
	out += fmt.Sprintf("%s_put_requests %v\n", prefix, data.PutRequests)

	// throughput, rps, rps physical cpu, rps logical cpu
	out += fmt.Sprintf("# HELP %s_rps reports request per second average\n", prefix)
	out += fmt.Sprintf("# TYPE %s_rps gauge\n", prefix)
	out += fmt.Sprintf("%s_rps %v\n", prefix, data.RPS)

	out += fmt.Sprintf("# HELP %s_avg_get_time reports average get request time\n", prefix)
	out += fmt.Sprintf("# TYPE %s_avg_get_time gauge\n", prefix)
	out += fmt.Sprintf("%s_avg_get_time %v\n", prefix, data.AvgGetTime)

	out += fmt.Sprintf("# HELP %s_avg_post_time reports average post request time\n", prefix)
	out += fmt.Sprintf("# TYPE %s_avg_post_time gauge\n", prefix)
	out += fmt.Sprintf("%s_avg_post_time %v\n", prefix, data.AvgPostTime)

	out += fmt.Sprintf("# HELP %s_avg_put_time reports average put request time\n", prefix)
	out += fmt.Sprintf("# TYPE %s_avg_put_time gauge\n", prefix)
	out += fmt.Sprintf("%s_avg_put_time %v\n", prefix, data.AvgPutTime)

	out += fmt.Sprintf("# HELP %s_rps_physical_cpu reports request per second average weighted by physical CPU cores\n", prefix)
	out += fmt.Sprintf("# TYPE %s_rps_physical_cpu gauge\n", prefix)
	out += fmt.Sprintf("%s_rps_physical_cpu %v\n", prefix, data.RPSPhysical)

	out += fmt.Sprintf("# HELP %s_rps_logical_cpu reports request per second average weighted by logical CPU cores\n", prefix)
	out += fmt.Sprintf("# TYPE %s_rps_logical_cpu gauge\n", prefix)
	out += fmt.Sprintf("%s_rps_logical_cpu %v\n", prefix, data.RPSLogical)

	// database metrics
	out += fmt.Sprintf("# HELP %s_max_db_connections reports max number of DB conenctions\n", prefix)
	out += fmt.Sprintf("# TYPE %s_max_db_connections gauge\n", prefix)
	out += fmt.Sprintf("%s_max_db_connections %v\n", prefix, data.MaxDBConnections)

	out += fmt.Sprintf("# HELP %s_max_idle_connections reports max number of idle DB connections\n", prefix)
	out += fmt.Sprintf("# TYPE %s_max_idle_connections gauge\n", prefix)
	out += fmt.Sprintf("%s_max_idle_connections %v\n", prefix, data.MaxIdleConnections)

	// see https://pkg.go.dev/database/sql#DBStats
	out += fmt.Sprintf("# HELP %s_max_open_connections reports max number of open DB connections\n", prefix)
	out += fmt.Sprintf("# TYPE %s_max_open_connections gauge\n", prefix)
	out += fmt.Sprintf("%s_max_open_connections %v\n", prefix, data.DBStats.MaxOpenConnections)

	out += fmt.Sprintf("# HELP %s_open_connections reports number of established to database (both in use and idle)\n", prefix)
	out += fmt.Sprintf("# TYPE %s_open_connections gauge\n", prefix)
	out += fmt.Sprintf("%s_open_connections %v\n", prefix, data.DBStats.OpenConnections)

	out += fmt.Sprintf("# HELP %s_in_use_connections reports number of in use database connections\n", prefix)
	out += fmt.Sprintf("# TYPE %s_in_use_connections gauge\n", prefix)
	out += fmt.Sprintf("%s_in_use_connections %v\n", prefix, data.DBStats.InUse)

	out += fmt.Sprintf("# HELP %s_idle_connections reports number of idle database connections\n", prefix)
	out += fmt.Sprintf("# TYPE %s_idle_connections gauge\n", prefix)
	out += fmt.Sprintf("%s_idle_connections %v\n", prefix, data.DBStats.Idle)

	out += fmt.Sprintf("# HELP %s_wait_count reports total number of connections waited for\n", prefix)
	out += fmt.Sprintf("# TYPE %s_wait_count counter\n", prefix)
	out += fmt.Sprintf("%s_wait_count %v\n", prefix, data.DBStats.WaitCount)

	out += fmt.Sprintf("# HELP %s_wait_duration reports total time (in sec) blocked waiting for a new connection\n", prefix)
	out += fmt.Sprintf("# TYPE %s_wait_duration counter\n", prefix)
	out += fmt.Sprintf("%s_wait_duration %v\n", prefix, data.DBStats.WaitDuration.Seconds())

	out += fmt.Sprintf("# HELP %s_max_idle_closed reports total number of connections closed due to SetMaxIdleConns\n", prefix)
	out += fmt.Sprintf("# TYPE %s_max_idle_closed counter\n", prefix)
	out += fmt.Sprintf("%s_max_idle_closed %v\n", prefix, data.DBStats.MaxIdleClosed)

	out += fmt.Sprintf("# HELP %s_max_idle_time_closed reports total number of connections closed due to SetConnMaxIdleTime\n", prefix)
	out += fmt.Sprintf("# TYPE %s_max_idle_time_closed counter\n", prefix)
	out += fmt.Sprintf("%s_max_idle_time_closed %v\n", prefix, data.DBStats.MaxIdleTimeClosed)

	out += fmt.Sprintf("# HELP %s_max_lifetime_closed reports total number of connections closed due to SetConnMaxLifetime\n", prefix)
	out += fmt.Sprintf("# TYPE %s_max_lifetime_closed counter\n", prefix)
	out += fmt.Sprintf("%s_max_lifetime_closed %v\n", prefix, data.DBStats.MaxLifetimeClosed)

	// migration server metrics
	out += fmt.Sprintf("# HELP %s_requests reports total number of migration requests\n", prefix)
	out += fmt.Sprintf("# TYPE %s_requests counter\n", prefix)
	out += fmt.Sprintf("%s_requests %v\n", prefix, data.MigrationRequests)

	out += fmt.Sprintf("# HELP %s_pending reports total number of pending migration requests\n", prefix)
	out += fmt.Sprintf("# TYPE %s_pending counter\n", prefix)
	out += fmt.Sprintf("%s_pending %v\n", prefix, data.MigrationPending)

	out += fmt.Sprintf("# HELP %s_in_progress reports total number of in progress migration requests\n", prefix)
	out += fmt.Sprintf("# TYPE %s_in_progress counter\n", prefix)
	out += fmt.Sprintf("%s_in_progress %v\n", prefix, data.MigrationInProgress)

	out += fmt.Sprintf("# HELP %s_failed reports total number of failed migration requests\n", prefix)
	out += fmt.Sprintf("# TYPE %s_failed counter\n", prefix)
	out += fmt.Sprintf("%s_failed %v\n", prefix, data.MigrationFailed)

	out += fmt.Sprintf("# HELP %s_term_failed reports total number of term failed migration requests\n", prefix)
	out += fmt.Sprintf("# TYPE %s_term_failed counter\n", prefix)
	out += fmt.Sprintf("%s_term_failed %v\n", prefix, data.MigrationTermFailed)

	out += fmt.Sprintf("# HELP %s_completed reports total number of completed migration requests\n", prefix)
	out += fmt.Sprintf("# TYPE %s_completed counter\n", prefix)
	out += fmt.Sprintf("%s_completed %v\n", prefix, data.MigrationCompleted)

	out += fmt.Sprintf("# HELP %s_queued reports total number of queued migration requests\n", prefix)
	out += fmt.Sprintf("# TYPE %s_queued counter\n", prefix)
	out += fmt.Sprintf("%s_queued %v\n", prefix, data.MigrationQueued)

	out += fmt.Sprintf("# HELP %s_exist_in_db reports total number of exist in db migration requests\n", prefix)
	out += fmt.Sprintf("# TYPE %s_exist_in_db counter\n", prefix)
	out += fmt.Sprintf("%s_exist_in_db %v\n", prefix, data.MigrationExistInDB)
	return out
}

// helper function to update RPS values
func updateRPS() {
	total := float64(TotalGetRequests + TotalPostRequests + TotalPutRequests)
	oldLogical := float64(NumLogicalCores)
	oldPhysical := float64(NumPhysicalCores)
	time.Sleep(1 * time.Minute)
	for {
		RPS = (float64(TotalGetRequests+TotalPostRequests+TotalPutRequests) - total) / 3600
		RPSLogical = (float64(NumLogicalCores) - oldLogical) / 3600.
		RPSPhysical = (float64(NumPhysicalCores) - oldPhysical) / 3600.

		total = float64(TotalGetRequests + TotalPostRequests + TotalPutRequests)
		oldLogical = float64(NumLogicalCores)
		oldPhysical = float64(NumPhysicalCores)
		time.Sleep(1 * time.Minute)
	}
}

// helper function to update avg get request time
func updateGetRequestTime(time0 time.Time) {
	AvgGetRequestTime += time.Since(time0).Seconds() / float64(TotalGetRequests)
}

// helper function to update avg post request time
func updatePostRequestTime(time0 time.Time) {
	AvgPostRequestTime += time.Since(time0).Seconds() / float64(TotalPostRequests)
}

// helper function to update avg put request time
func updatePutRequestTime(time0 time.Time) {
	AvgPutRequestTime += time.Since(time0).Seconds() / float64(TotalPutRequests)
}
