package cache

import "fmt"
import "context"

import "github.com/redis/go-redis/v9"

type Cache struct {
	rdb *redis.Client
	ctx context.Context
}

type Page struct {
  ID  string
  Title string
	TitleTraditional string
}

func New() Cache {
	return Cache{
		rdb: redis.NewClient(&redis.Options{Addr: "localhost:6379"}),
		ctx: context.Background(),
	}
}

func (c *Cache) Set(key string, value string) {
	err := c.rdb.Set(c.ctx, key, value, 0).Err()
	if err != nil {
		fmt.Println(err)
	}
}

func (c *Cache) Get(key string) string {
	val, _ := c.rdb.Get(c.ctx, key).Result()
	// if err != nil {
	// 	fmt.Println("Failed to get Redis value key", key, err)
	// }
	return val
}

func (c *Cache) SetPage(id string, title string, title_trad string) {
	c.Set(id, "1")
	c.Set("title:" + id, title)
	c.Set("title_trad:" + id, title_trad)
}

func (c *Cache) GetPage(id string) (bool, string, string) {
	found := c.Get(id)
	if found != "1" {
		return false, "", ""
	}
	return true, c.Get("title:" + id), c.Get("title_trad:" + id)
}