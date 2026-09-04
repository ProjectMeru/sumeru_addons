module sumeru_addons

go 1.26.6

replace sumeru => ../sumeru

require sumeru v0.0.0-00010101000000-000000000000

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/gpdf-dev/gpdf v1.0.12 // indirect
	github.com/lib/pq v1.12.3 // indirect
	golang.org/x/crypto v0.56.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)
