package main

import "time"

type builtinCacheEntry struct {
	country string
	expires time.Time
}

type builtinCountryCache map[string]builtinCacheEntry

func newBuiltinCache(_ interface{}) (countryCache, error) {
	return make(builtinCountryCache), nil
}

func (c builtinCountryCache) Put(ip, country string, expires time.Time) {
	c[ip] = builtinCacheEntry{country, expires}
}

func (c builtinCountryCache) Get(ip string) (string, bool) {
	if entry, ok := c[ip]; ok && entry.expires.After(time.Now()) {
		return entry.country, true
	}
	return "", false
}
