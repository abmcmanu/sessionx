module github.com/abmcmanu/sessionx/optional/store/redis

go 1.23.0

toolchain go1.24.3

require (
	github.com/abmcmanu/sessionx v0.1.0
	github.com/redis/go-redis/v9 v9.5.5
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
)

replace github.com/abmcmanu/sessionx => ../../../
