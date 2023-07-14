# goexam

```
git clone https://github.com/igorson2nd/goexam.git
cd goexam
docker compose up --build
go run main.go
```

Web server will start on port `:3000` and crawler will start crawling BaseUrl from `config.json`.

Press `Ctrl-C` to stop crawling.

Press `Ctrl-C` again to stop web server.

JWT auth is active if JwtSecret is set to non empty string. 

Setting JwtSecret to empty string `"JwtSecret": ""` in config file will disable JWT.

## Login
```
POST /login

# curl login example:
curl -d "username=ted" -d "password=pass" http://localhost:3000/login
```
Login password is set in config file as `LoginPass`.

## API & Index
```
# index page: article list in HTML table
GET /

# get article by id in JSON
GET /articles/:id

# edit title (PUT JSON body)
PUT /articles/:id

# delete article
DELETE /articles/:id
```
