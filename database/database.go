package database

import "fmt"
import "time"
import "encoding/json"

import "goexam/page"
import "goexam/cache"

import "gorm.io/gorm"
import "gorm.io/gorm/logger"
import "gorm.io/driver/mysql"

type Page struct {
  ID  string `gorm:"primaryKey"`
  Title string
	TitleTraditional string
}

type Database struct {
	db *gorm.DB
	cache cache.Cache
}

func Connect(dsn string) Database {
	// db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	silent_logger := logger.Default.LogMode(logger.Silent)
  db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: silent_logger})
	if err != nil {
		fmt.Println("Cannot connect to DB: ", err)
	}
	db.AutoMigrate(&Page{})

	return Database{db: db, cache: cache.New()}
}

func (db *Database) SavePage(craw_pag page.Page) {
	id := craw_pag.ID.GetStrID()
	var title, title_trad string
	if craw_pag.ID.GetLang() == "gb" {
		title = craw_pag.Title
	} else {
		title_trad = craw_pag.Title
	}
	
	db_pag := Page{ID: id}
	result := db.db.First(&db_pag)
	if result.RowsAffected == 0 {
		db.db.Create(&Page{ID: id, Title: title, TitleTraditional: title_trad})
	} else {
		if title != "" {
			db_pag.Title = title
		}
		if title_trad != "" {
			db_pag.TitleTraditional = title_trad
		}
		db.db.Save(&db_pag)
	}

	db.cache.SetPage(id, title, title_trad)
}

func (db *Database) GetPage(id string) (Page, int64) {
	// cache first
	found, title, title_trad := db.cache.GetPage(id)
	if found {
		fmt.Println("Found in cache!")
		return Page{ID: id, Title: title, TitleTraditional: title_trad}, 1
	}

	p := Page{ID: id}
	result := db.db.First(&p)
	// fmt.Println(p)
	return p, result.RowsAffected
}

func (db *Database) PutPage(p Page) {
	db.db.Save(&p)
	db.cache.SetPage(p.ID, p.Title, p.TitleTraditional)
}

func (db *Database) DeletePage(p Page) {
	db.db.Delete(&p)
	db.cache.Set(p.ID, "0")
}

func (db *Database) GetAllPages() []Page {
	var pages []Page
	db.db.Find(&pages)
	return pages
}

func (db *Database) GetAllPagesFromCache() []Page{
	var pages []Page
	json_data := db.cache.Get("index_page_json_data")
	err := json.Unmarshal([]byte(json_data), &pages)
	if err == nil {
		// fmt.Println("***found all in cache")
		return pages
	}

	// fmt.Println("Not found :(")
	return db.GetAllPages()
}

func (db *Database) CacheAllPages() {
	pages := db.GetAllPages()
	json_data, _ := json.Marshal(pages)
	db.cache.Set("index_page_json_data", string(json_data))
}

func (db *Database) CacheTickerStart() {
	cache_ticker := time.NewTicker(5 * time.Second)
	go func() {
		for {
			<-cache_ticker.C
			// fmt.Println(">>>caching all now")
			db.CacheAllPages()
		}
	}()
}