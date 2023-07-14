package href

// import "fmt"
import "regexp"
import "strconv"
import "strings"

const base = "/news/gb"
const default_date = "1970/01/01"

type Href struct {
	str string
	raw_str bool
	traditional bool
	id uint32	
}

func New(url string) Href {
	// remove prot. & domain
	url = strings.ReplaceAll(url, "https://www.secretchina.com", "")

	// trad. vs simpl.
	standard_prefix, _ := regexp.MatchString("(?i)^/news/[bg][5b]/", url)
	if ! standard_prefix {
		return Href{str: url, raw_str: true}
	}
	traditional, _ := regexp.MatchString("(?i)^/news/b5/", url)

	// remove date
	date_match := regexp.MustCompile("(?i)^/news/[bg][5b]/(\\d{4}/\\d{2}/\\d{2})/").FindStringSubmatch(url)
	if date_match != nil {
		url = strings.ReplaceAll(url, date_match[1], "§")
	}

	// remove prefix & suffix
	url = strings.ReplaceAll(url, "/news/b5/", "")
	url = strings.ReplaceAll(url, "/news/gb/", "")
	url = strings.ReplaceAll(url, ".html", "")

	// remove integer ID
	int_match := regexp.MustCompile("\\d{3,}").FindStringSubmatch(url)
	if int_match != nil {
		int_id, _ := strconv.Atoi(int_match[0])
		url = strings.ReplaceAll(url, int_match[0], "*")
		return Href{str: url, traditional: traditional, id: uint32(int_id)}
	}

	return Href{str: url, traditional: traditional}
}

func (h *Href) ToStr() string {
	if h.raw_str {
		return h.str
	}

	prefix := "/news/gb/"
	if h.traditional {
		prefix = "/news/b5/"
	}

	url := h.str

	str_id := strconv.Itoa(int(h.id))
	url = strings.ReplaceAll(url, "§", default_date)
	url = strings.ReplaceAll(url, "*", str_id)

	return prefix + url + ".html"
}

func (h *Href) GetStrID() string {
	url := h.ToStr()
	id_match := regexp.MustCompile("(?i)/([^/]+).html$").FindStringSubmatch(url)
	if id_match == nil {
		return ""
	}

	return id_match[1]
}

func (h *Href) GetLang() string {
	if h.raw_str {
		return "raw"
	}
	if h.traditional {
		return "b5"
	}
	return "gb"
}