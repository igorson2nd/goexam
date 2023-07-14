package csvsaver

import "goexam/page"
import "sync"
import "encoding/csv"
import "os"
//import "fmt"

func SavePages(pages *sync.Map) {
	file, err := os.Create("pages.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"id", "lang", "title"}
	writer.Write(header)

	pages.Range(func(id any, val any) bool {
		p, _ := val.(page.Page)
		row := []string{p.ID.GetStrID(), p.ID.GetLang(), p.Title}
		writer.Write(row)
		return true
	})
}