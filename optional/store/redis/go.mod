module github.com/abmcmanu/sessionx/optional/store/redis

go 1.23.0

toolchain go1.24.3

require (
	github.com/abmcmanu/sessionx v0.0.0
	github.com/redis/go-redis/v9 v9.7.0
)

replace github.com/abmcmanu/sessionx => ../../../