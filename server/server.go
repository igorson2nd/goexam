package server


import "fmt"
import "bytes"
import "strconv"
import "net/http"
import "math/rand"
import "goexam/page"
import "html/template"
import "goexam/database"

import "github.com/labstack/echo/v4"
import "github.com/labstack/echo-jwt/v4"
import "github.com/labstack/echo/v4/middleware"
import "github.com/golang-jwt/jwt/v5"

type Server struct {
	DB database.Database
	JwtSecret string
	LoginPass string
}

type jwt_claims struct {
	Username string
	jwt.RegisteredClaims
}

func (s *Server) Start() {
	e := echo.New()

	e.GET("/", s.index)
	e.POST("/login", s.jwt_login)
	e.GET("/articles/:id", s.get_page)
	e.PUT("/articles/:id", s.put_page)
	e.DELETE("/articles/:id", s.delete_page)

	e.GET("/article_content/:lang/:id", s.article_content)
	e.Use(middleware.Static("data"))

	jwt_config := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(jwt_claims)
		},
		SigningKey: []byte(s.JwtSecret),
		Skipper: func(c echo.Context) bool {
			return c.Request().URL.Path == "/login"
	 },
	}
	if s.JwtSecret != "" {
		e.Use(echojwt.WithConfig(jwt_config))
	}

	e.Logger.Fatal(e.Start(":3000"))
}

func (s *Server) index(c echo.Context) error {
	process_cookie(c)
	pages := s.DB.GetAllPagesFromCache()

	var out bytes.Buffer
	funcs := template.FuncMap{"check": page.CheckSavedPage}
	tmpl, err := template.New("index.html.tmpl").Funcs(funcs).ParseFiles("index.html.tmpl")
	if err != nil {
		fmt.Println(err)
	}
	tmpl.Execute(&out, pages)
	return c.HTML(http.StatusOK, out.String())
}

// json api
func (s *Server) get_page(c echo.Context) error {
	process_cookie(c)
	id := c.Param("id")
	p, rows_found := s.DB.GetPage(id)
	if rows_found == 0 {
		return c.String(http.StatusNotFound, "404: page not found")
	}
	return c.JSON(http.StatusOK, p)
}

func (s *Server) put_page(c echo.Context) error {
	process_cookie(c)
	new_page := new(database.Page)
  if err := c.Bind(new_page); err != nil {
    return c.String(http.StatusBadRequest, "Bad request")
  }
	if new_page.ID != c.Param("id") {
		return c.String(http.StatusBadRequest, "Error: param ID and json body ID do not match")
	}
	if new_page.Title == "" {
		return c.String(http.StatusBadRequest, "Bad request, missing field")
	}

	_, rows_found := s.DB.GetPage(new_page.ID)
	if rows_found == 0 {
		return c.String(http.StatusNotFound, "404: page not found")
	}

	s.DB.PutPage(*new_page)

	return c.JSON(http.StatusOK, new_page)
}

func (s *Server) delete_page(c echo.Context) error {
	process_cookie(c)
	id := c.Param("id")
	p, rows_found := s.DB.GetPage(id)
	if rows_found == 0 {
		return c.String(http.StatusNotFound, "404: page not found")
	}
	s.DB.DeletePage(p)
	return c.JSON(http.StatusOK, p)
}

func process_cookie(c echo.Context) {
	name := "session_id"
	cookie, err := c.Cookie(name)
	if err != nil {
		val := strconv.Itoa(rand.Int())
		c.SetCookie(&http.Cookie{Name: name, Value: val})
		fmt.Println("New cookie created:", name, val)
		return
	}
	fmt.Println("Cookie found:", cookie.Name, cookie.Value)
}

func (s *Server) jwt_login(c echo.Context) error {
	process_cookie(c)
	username := c.FormValue("username")
	pass := c.FormValue("password")
	if username == "" {
		return c.String(http.StatusBadRequest, "Field missing")
	}
	if pass != s.LoginPass {
		return echo.ErrUnauthorized
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt_claims{Username: username})
	t, err := token.SignedString([]byte(s.JwtSecret))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{
		"token": t,
	})
}

func (s *Server) article_content(c echo.Context) error {
	process_cookie(c)
	lang := c.Param("lang")
	id := c.Param("id")

	content := page.LoadPageFromFile(lang, id)

	return c.HTML(http.StatusOK, content)
}