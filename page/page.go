package page

import "goexam/href"
import "fmt"
import "net/http"
import "io/ioutil"
import "regexp"
import "strings"
import "time"
import "os"
import "errors"
import "compress/gzip"
import "bytes"

type Page struct {
  ID href.Href
  Title string
}

func get_body(url string) string {
  resp, err := http.Get(url)
  if err != nil {
    fmt.Println("Error getting the page: ", err)
    fmt.Println(url)
    return ""
  }
  defer resp.Body.Close()
  body, _ := ioutil.ReadAll(resp.Body)
  return string(body)
}

func get_title(body string) string {
  title := regexp.MustCompile("(?i)<title>(.+)</title>").FindStringSubmatch(body)
  if title == nil {return ""}
  title_ret := title[1]
  title_ret = strings.ReplaceAll(title_ret, "看中国栏目： ", "") // prefix
  title_ret = strings.ReplaceAll(title_ret, " - 看中国网", "")  // suffix 
  title_ret = regexp.MustCompile("\\|.+$").ReplaceAllString(title_ret, "") // suffix 
  // title_ret = strings.ReplaceAll(title_ret, " || 看中国网", "")  // suffix 
  // title_ret = strings.ReplaceAll(title_ret, "| 看中国网", "")  // suffix 
  return title_ret
}

// relative only
func get_hrefs(body string) []href.Href {
  pattern := regexp.MustCompile("(?i)<a.+?a>")
  all_a := pattern.FindAllString(body, -1)
  var ret []href.Href
  for _, a := range all_a {
    href_pat :=  regexp.MustCompile("(?i)href=['\"](/.+?.html)['\"]")
    href_match := href_pat.FindStringSubmatch(a)
    if href_match != nil {
      new_href := href.New(href_match[1])
      ret = append(ret, new_href)
    }    
  }
  return ret
}

func save_page_to_file(url string, body string) {
  ref := href.New(url)
  str_id := ref.GetStrID()
  if str_id == "" {
    return
  }

  dir_path := "data/" + time.Now().Format("2006-01-02")
  dir_err := os.Mkdir(dir_path, os.ModePerm)
  if dir_err != nil {
    if ! errors.Is(dir_err, os.ErrExist) {
      fmt.Println("Error creating directory: ", dir_err)
      fmt.Println(dir_path)
    }
  }

  filename := str_id + "." + ref.GetLang() + ".html.gz"
  filepath := dir_path + "/" + filename

  file_stat, _ := os.Stat(filepath)
  if file_stat != nil {
    return
  }

  var buff bytes.Buffer
  gz := gzip.NewWriter(&buff)
  gz.Write([]byte(body))
  gz.Flush()
  gz.Close()

  file_err := os.WriteFile(filepath, buff.Bytes(), 0644)
  if file_err != nil {
    fmt.Println("Error writing to file: ", file_err)
    fmt.Println(filepath)
  }
}

func CheckSavedPage(id string, trad bool) bool {
  lang := "gb"
  if trad { lang = "b5" }
  dir_path := "data/" + time.Now().Format("2006-01-02")
  filename := id + "." + lang + ".html.gz"
  filepath := dir_path + "/" + filename
  _, err := os.Stat(filepath)
  return err == nil
}

func LoadPageFromFile(lang string, id string) string {
  dir_path := "data/" + time.Now().Format("2006-01-02")
  filename := id + "." + lang + ".html.gz"
  filepath := dir_path + "/" + filename

  content, err := os.ReadFile(filepath)
  if err != nil {
    fmt.Println(err)
    return "File not found"
  }

  reader := bytes.NewReader([]byte(content))
  gz, _ := gzip.NewReader(reader);
  output, err := ioutil.ReadAll(gz);
  if err != nil {
    fmt.Println(err);
    return "Problem serving file"
  }
  return string(output);
}

func SearchPage(url string) (Page, []href.Href) {
  body := get_body(url)

  save_page_to_file(url, body)

  ret_page := Page{
    ID: href.New(url), 
    Title: get_title(body),
  }
  return ret_page, get_hrefs(body)
}