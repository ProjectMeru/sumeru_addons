module sumeru_addons

go 1.26.2

replace sumeru => ../sumeru

require sumeru v0.0.0-00010101000000-000000000000

require (
	github.com/lib/pq v1.12.3 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.51.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)
