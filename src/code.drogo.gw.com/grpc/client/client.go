package client

import (
	"google.golang.org/grpc"
	log "github.com/sirupsen/logrus"
	"code.drogo.gw.com/event"
	"context"
	"code.drogo.gw.com/database"
	"strconv"
	"strings"
)


func GRPCClient() (event.EventServiceClient, error) {
	conn, err := grpc.Dial("139.59.40.163:34567", grpc.WithInsecure())
	if err != nil {
		log.Errorf("error while connecting to grpc server: %v", err)
		return nil,err
	}
	client := event.NewEventServiceClient(conn)
	return client, nil
}

func FetchEventById(eventId string) (database.Event, error) {

	client, err := GRPCClient()
	if err != nil {
		return database.Event{}, err
	}

	event, err := client.GetEventById(context.Background(), &event.EventRequest{EventId: eventId, UserId: ""})
	if err != nil {
		log.Info("error while getting event by id", "")
		log.Error("error: ", err)
	}

	if event.EventDetail != nil {

		uploadId := event.EventDetail.CoverImageUploadId
		imageData := "https://minio.eventackle.com/uploads/" + uploadId

		// Calculate soldAmount and add tax if separate
		var soldAmount float64

		if strings.ToLower(event.Tax.ShouldAddTax) == "separate" {
			taxRate, _ := strconv.Atoi(event.Tax.TaxRate)
			tax := (float64(taxRate)/100) * event.Ticket.Price
			soldAmount = float64(event.Ticket.Sold) * (event.Ticket.Price + tax)
		} else {
			soldAmount = float64(event.Ticket.Sold) * event.Ticket.Price
		}

		// Convert isFeatured to bool
		isFeatured := false
		if event.IsFeatured == "yes" {
			isFeatured = true
		}

		returnEvent := database.Event{
			Name:        event.EventDetail.Name,
			PublishDate: event.EventDetail.CreatedOn,
			SoldAmount:  soldAmount,
			Currency: 	 event.Ticket.Currency,
			TicketsSold: int(event.Ticket.Sold),
			Location:    event.EventDetail.VenueCity + ", " + event.EventDetail.VenueState + ", " + event.EventDetail.VenueCountry,
			StartDate:   event.EventDetail.StartDate,
			StartTime:   event.EventDetail.StartTime,
			EndDate:     event.EventDetail.EndDate,
			EndTime:     event.EventDetail.EndTime,
			Categories:  event.Categories,
			Types:       event.EventTypes,
			ImageData:   imageData,
			Description: event.EventDetail.BriefDescription,
			IsFeatured:  isFeatured,
			Deactivated: event.Deactivated,
		}
		return returnEvent, nil
	}
	return database.Event{}, err
}


func UpdateFeaturedEvent(eventId string, feature bool) (database.Response, error) {
	client, err := GRPCClient()
	if err != nil {
		log.Info("")
		log.Error("error: ", err)
		return database.Response{
			Status: false,
			Message: "",
		}, err
	}
	isFeatured := ""
	if feature {
		isFeatured = "yes"
	} else {
		isFeatured = "no"
	}

	_, err = client.UpdateFeaturedEventById(context.Background(), &event.UpdateFeaturedRequest{EventId: eventId, Featured: isFeatured})
	if err != nil {
		log.Info("error while updating event")
		log.Error("error: ", err)
		return database.Response{
			Status: false,
			Message: "failed to update event",
		}, err
	}
	returnResponse := database.Response{
		Status:  true,
		Message: "updated successfully",
	}
	return returnResponse, nil
}
