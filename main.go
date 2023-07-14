package main

import "os"
import "fmt"
import "time"
import "syscall"
import "os/signal"

import "goexam/server"
import "goexam/crawler"
import "goexam/database"
import "goexam/csvsaver"
import "goexam/confreader"

func main() {
	fmt.Println("Loading config ...")
	config := confreader.Read()

	db := database.Connect(config.DbDsn)
	db.CacheTickerStart()
	s := server.Server{
		DB: db, 
		LoginPass: config.LoginPass,
		JwtSecret: config.JwtSecret, 
	}
	go s.Start()
	time.Sleep(200 * time.Millisecond)

	sigterm_channel := make(chan os.Signal)
  signal.Notify(sigterm_channel, os.Interrupt, syscall.SIGTERM)

	fmt.Println("Crawling ...")
	c := crawler.Crawler{
		BaseUrl: config.BaseUrl, 
		Crawl_limit: 4294967295, 
		Concurrency: config.Concurrency, 
		DB: db,
		SignalChannel: sigterm_channel,
	}
	ticker := time.NewTicker(time.Duration(config.TickInterval) * time.Minute)
	for {
		c.Crawl()
		if c.Interrupted() {
			break
		}
		<-ticker.C
	}

	fmt.Println("Saving CSV ...")
	csvsaver.SavePages(&c.Pages)
	fmt.Println("CSV Saved.")
	
	fmt.Println("Press Ctrl-C to stop HTTP server.")

	<-sigterm_channel
	fmt.Println(">>> Programm end.")
}
