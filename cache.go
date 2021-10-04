package spacetraders

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	LOANS       CacheKey = "loans"
	SHIPS       CacheKey = "ships"
	MYLOCATIONS CacheKey = "my locations"
	LOCATIONS   CacheKey = "all locations"
	SYSTEMS     CacheKey = "systems"
	FLIGHTS     CacheKey = "flights"
	FLIGHTDESTS CacheKey = "flight destinations"
	CARGO       CacheKey = "cargo"
)

var (
	shortToID  = make(map[string]string)
	idToShort  = make(map[string]string)
	shortIndex = make(map[CacheKey]int)
)

type Cache struct {
	data map[CacheKey]*CacheItem
}
type CacheKey string
type CacheItem struct {
	expiresOn time.Time
	data      []string
	shorts    []string
}

func GetCache() *Cache {
	var c *Cache
	var once sync.Once
	once.Do(func() {
		cargos := &CacheItem{
			expiresOn: time.Now().Add(24 * time.Hour),
			data:      []string{},
		}
		for _, c := range []string{"FUEL", "METALS", "NONE"} {
			cargos.data = append(cargos.data, c)
		}
		c = &Cache{
			data: map[CacheKey]*CacheItem{CARGO: cargos},
		}
	})
	return c

}

// Add a new value to a key, create a short if needed
func (c *Cache) Add(key CacheKey, data string) {
	short := makeShort(key, data)
	if _, ok := c.data[key]; !ok {
		c.data[key] = &CacheItem{}
	}
	newKey := c.data[key]
	newKey.data = sort.StringSlice(append(c.data[key].data, data))
	newKey.shorts = sort.StringSlice(append(c.data[key].shorts, short))
	newKey.expiresOn = time.Now().Add(time.Hour)
	c.data[key] = newKey
}

// Add multiple new values (both long and short) to a key
// Note: not creating shorts if they aren't provided
func (c *Cache) Extend(key CacheKey, data []string, shorts []string) {
	sort.Strings(data)
	sort.Strings(shorts)
	item, ok := c.data[key]
	if !ok {
		c.Store(key, data, shorts)
		return
	}

	var set = make(map[string]bool)
	for _, v := range append(data, item.data...) {
		set[strings.ToUpper(v)] = true
	}
	item.data = []string{}
	for v := range set {
		item.data = append(item.data, v)
	}
	sort.Strings(item.data)

	set = make(map[string]bool)
	for _, v := range append(shorts, item.shorts...) {
		set[strings.ToUpper(v)] = true
	}
	item.shorts = []string{}
	for v := range set {
		item.shorts = append(item.shorts, v)
	}
	sort.Strings(item.shorts)
	c.data[key] = item
}

// Replace a key with new longs and shorts
func (c *Cache) Store(key CacheKey, data []string, shorts []string) {
	sort.Strings(data)
	c.data[key] = &CacheItem{expiresOn: time.Now().Add(time.Hour), data: data, shorts: shorts}
}

// Get the current cached value for a key
func (c *Cache) Restore(key CacheKey) []string {
	cached, ok := c.data[key]
	if !ok || cached.expiresOn.Before(time.Now()) {
		log.Printf("Cache miss: %q", key)
		if err := c.data[key]; err != nil {
			log.Printf("Error caching %q: %v", key, err)
			return []string{}
		}
		cached = c.data[key]
	} else {
		log.Printf("Cache hit: %q", key)
	}
	if cached.shorts != nil {
		return append(cached.shorts, cached.data...)
	}
	return cached.data
}

// Create a short name for a given identifier, per type
func makeShort(key CacheKey, data string) string {
	short, ok := idToShort[data]
	if ok {
		return short
	}
	var prefix string
	switch key {
	case LOANS:
		prefix = "ln"
	case SHIPS:
		prefix = "s"
	case FLIGHTS:
		prefix = "f"
	case CARGO:
		return strings.ToUpper(data)
	case FLIGHTDESTS:
		return data
	default:
		log.Printf("Unknown prefix for %s", key)
		prefix = "X"
	}

	shortIndex[key]++
	short = fmt.Sprintf("%s-%d", prefix, shortIndex[key])
	idToShort[data] = short
	shortToID[short] = data
	log.Printf("Created short %q in %q for %q", short, key, data)
	return short
}

// Get the identifier a short is associated with
func makeLong(id string) string {
	if long, ok := shortToID[id]; ok {
		return long
	}
	return id
}

// Create short identifiers in bulk
func getShorts(key CacheKey, data []string) []string {
	res := []string{}
	for _, d := range data {
		res = append(res, makeShort(key, d))
	}

	return res
}
