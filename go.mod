module github.com/dmwm/dbs2go

go 1.18

require (
	github.com/dmwm/cmsauth v0.0.0-20220120183156-5495692d4ca7
	github.com/go-playground/validator/v10 v10.10.1
	github.com/google/uuid v1.3.0
	github.com/gorilla/csrf v1.7.1
	github.com/gorilla/mux v1.8.0
	github.com/graph-gophers/graphql-go v1.3.0
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible
	github.com/mattn/go-oci8 v0.1.1
	github.com/mattn/go-sqlite3 v1.14.12
	github.com/prometheus/procfs v0.8.0
	github.com/shirou/gopsutil v3.21.11+incompatible
	github.com/ulule/limiter/v3 v3.10.0
	github.com/vkuznet/auth-proxy-server/logging v0.0.0-20220406163751-c36feb20c750
	github.com/vkuznet/limiter v2.2.2+incompatible
	github.com/vkuznet/x509proxy v0.0.0-20210801171832-e47b94db99b6
	golang.org/x/exp v0.0.0-20220428152302-39d4317da171
	gopkg.in/rana/ora.v4 v4.1.15
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lestrrat-go/strftime v1.0.5 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.14.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/r3labs/diff/v3 v3.0.0 // indirect
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/tklauser/numcpus v0.4.0 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	golang.org/x/crypto v0.0.0-20220331220935-ae2d96664a29 // indirect
	golang.org/x/exp/errors v0.0.0-20220916125017-b168a2c6b86b // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)

replace github.com/ulule/limiter/v3 => github.com/vkuznet/limiter/v3 v3.10.2
