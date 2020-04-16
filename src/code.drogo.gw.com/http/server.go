package http

import (
	"github.com/kataras/iris"
	"github.com/iris-contrib/middleware/cors"
	"github.com/gqlgen/handler"
	"code.drogo.gw.com/gql"
	"code.drogo.gw.com/vue"
	"code.drogo.gw.com/service"
	"github.com/dgrijalva/jwt-go"
	jwtmiddleware "github.com/iris-contrib/middleware/jwt"
)

type Server struct {

}

type User struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

func NewDrogoServer(isProd bool) *iris.Application  {

	service.NewServer()

	authKey := "DrogoAuthorizationTokenKey"

	app := iris.Default()

	gqlApp := &gql.App{}

	crs := cors.New(cors.Options{
		AllowedOrigins:   []string{"*", "http://admin.eventackle.com","https://admin.eventackle.com", "http://eventackle.surge.sh", "http://localhost:8080"},
		AllowCredentials: true,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "CONNECT", "HEAD", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "Accept", "Origin", "X-Requested-With"},
	})

	if isProd {
		app.Get("/index", iris.FromStd(vue.Index("Eventackle Admin Panel", "")))
	}

	jwtHandler := jwtmiddleware.New(jwtmiddleware.Config{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte(authKey), nil
		},

		SigningMethod: jwt.SigningMethodHS256,
	})

	v1 := app.Party("/api/v1", crs).AllowMethods(iris.MethodPost, iris.MethodOptions, iris.MethodGet)
	v1.Get("/gql-play", iris.FromStd(handler.Playground("Eventackle Admin Panel", "/api/v1/query")))
	v1.Post("/query", jwtHandler.Serve, iris.FromStd(handler.GraphQL(gql.NewExecutableSchema(gqlApp))))

	v1.Post("/login", Login)
	v1.Get("/page-views", PageHits)

	return app
}