module github.com/nesiler/cestx/registry

go 1.22.3

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/nesiler/cestx/common v0.0.0-00010101000000-000000000000
	github.com/robfig/cron/v3 v3.0.1
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fatih/color v1.17.0 // indirect
	github.com/joho/godotenv v1.5.1
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sys v0.18.0 // indirect
)

// for local development
replace github.com/nesiler/cestx/common => ../common
