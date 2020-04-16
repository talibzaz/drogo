package mailClient

import (
	"google.golang.org/grpc"
	"code.drogo.gw.com/mail"
	log "github.com/sirupsen/logrus"
	db "code.drogo.gw.com/database"
	"context"
	"strings"
)

func mailGRPCClient() (mail.MailServiceClient, error){
	conn, err := grpc.Dial("139.59.40.163:34567", grpc.WithInsecure())
	if err != nil {
		log.Errorf("error while connecting to grpc server: %v", err)
		return nil,err
	}
	client := mail.NewMailServiceClient(conn)

	return client, nil
}

func ApproveOrganizerMail(organizer, email string) (db.Response, error){
	client, err := mailGRPCClient()
	if err != nil {
		return db.Response{}, err
	}

	res, err := client.ApprovalEmail(context.Background(), &mail.CreationRequestDetail{OrganizerName: organizer, EmailId: email})
	if err != nil {
		return db.Response{}, err
	}

	var status bool
	if strings.ToLower(res.Status) == "ok" {
		status = true
	} else {
		status = false
	}

	return db.Response{Status: status, Message: res.Message}, nil
}