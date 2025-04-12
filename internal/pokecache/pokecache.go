package pokecache
import (
	"time"
	"sync"
)

type Cache struct {
	cache map[string]cacheEntry
	mu	   sync.Mutex
	interval time.Duration
}
type cacheEntry struct {
	createdAt time.Time
	val []byte
}

func NewCache(interval time.Duration) *Cache {
	c:= &Cache{
		cache: make(map[string]cacheEntry),
		interval: interval,
	}
	c.reapLoop()
	return c


}
func (c *Cache) Add(key string, val []byte){
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = cacheEntry{
		createdAt: time.Now(),
		val: val,
	}

}
func (c *Cache) Get(key string) ([]byte, bool){
	c.mu.Lock()
	defer c.mu.Unlock()
	entry,ok:=c.cache[key]
	if !ok{
		return []byte{}, false
	}
	
	return entry.val, true
}
func (c *Cache) reapLoop(){
	ticker := time.NewTicker(c.interval)
	go func(){
		for {
			<-ticker.C
			c.mu.Lock()
			now:=time.Now()
			for key, entry := range c.cache {
				if now.Sub(entry.createdAt) > c.interval {
					delete(c.cache, key)
				}
			}
			c.mu.Unlock()
		}
	}()
}