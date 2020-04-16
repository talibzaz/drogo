package http

import (
	db "code.drogo.gw.com/database"
	"github.com/kataras/iris"
	"code.drogo.gw.com/database"
	"encoding/json"
	"code.drogo.gw.com/auth"
	log "github.com/sirupsen/logrus"
	"code.drogo.gw.com/analytics"
	"context"
)

func Login(ctx iris.Context){
	var user User
	err := ctx.ReadJSON(&user)
	if err != nil {
		log.Error("error in reading json: ", err)
		ctx.StatusCode(400)
		return
	}

	res, err := database.ValidateUser(user.Email, user.Password)
	if err != nil {
		log.Error(res.Message, ":",  err)
		ctx.StatusCode(400)
		return
	}

	if res.Status {
		token, err := auth.GenerateToken()
		if err != nil {
			log.Error("error in token generation: ", err)
			return
		}
		res.Token = token

		j, err := json.Marshal(res)
		if err != nil {
			log.Error(err)
			ctx.StatusCode(500)
			return
		}

		ctx.Write(j)
		return
	}

	j, err := json.Marshal(res)
	if err != nil {
		log.Error(err)
		ctx.StatusCode(500)
		return
	}
	ctx.StatusCode(401)
	ctx.Write(j)

	return
}

func PageHits(ctx iris.Context) {
	duration := ctx.URLParam("duration")
	eventName := ctx.URLParam("eventName")

	events, err := analytics.EventViews(duration, eventName)
	if err != nil {
		log.Error(err)
		ctx.StatusCode(500)
		return
	}

	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		ctx.StatusCode(500)
		return
	}
	res, err := arango.GetEventNameByEventID(events)
	if err != nil {
		log.Error(err)
		ctx.StatusCode(500)
		return
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		log.Error(err)
		ctx.StatusCode(500)
		return
	}

	ctx.Write(bytes)

	return
}