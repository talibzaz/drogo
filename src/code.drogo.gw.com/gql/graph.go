package gql

import (
	db "code.drogo.gw.com/database"
	"context"
	"code.drogo.gw.com/grpc/client"
	log "github.com/sirupsen/logrus"
	"code.drogo.gw.com/service"
	"code.drogo.gw.com/analytics"
	"code.drogo.gw.com/elastic"
	"code.drogo.gw.com/grpc/mailClient"
)
type App struct {}

func (a *App) Query() QueryResolver {
	return &queryResolver{a}
}

func (a *App) Mutation() MutationResolver {
	return &mutationResolver{ a}
}

type mutationResolver struct { *App }

type queryResolver struct { *App }

func (a *App) GetCategories(ctx context.Context, offset int) ([]db.Category, error) {

	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Errorf("error: %v", err)
		return nil, err
	}

	categories, err := arango.ListEventCategory(offset)
	if err != nil {
		log.Errorf("error: %v", err)
		return nil, err
	}

	return categories, nil
}

func (a *App) UpdateFeaturedEvent(ctx context.Context, id string, feature bool) (db.Response, error) {

	response, err := client.UpdateFeaturedEvent(id, feature)
	if err != nil {
		log.Error("not able to update the event")
		log.Errorf("error: %v", err)
		return response, err
	}
	return response, nil
}

func (a *App) ReportsEventList(ctx context.Context, id string) ([]db.Event, error) {
	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return nil, err
	}
	events, err := arango.ReportsEventList(id)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	return events, nil
}

func (a *App) GetEventAttendees(ctx context.Context, eventId, organizerId string, offset int) ([]db.Attendee, error) {
	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return nil, err
	}
	attendees, err := arango.GetEventAttendees(eventId, organizerId, offset)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return attendees, nil
}

func (a *App) GetEventById(ctx context.Context, id string) (db.Event, error) {

	event, err := client.FetchEventById(id)
	if err != nil {
		log.Error("not able to fetch event from grpc server")
		log.Errorf("error: %v", err)
		return event, nil
	}
	return event, nil
}

func (a *App) GetEventList(ctx context.Context, offset int) ([]db.Event, error) {

	var events []db.Event
	arangoDb, err := service.NewArangoDB(context.Background())
	if err != nil {
		log.Errorf("error: %v", err)
		return nil, err
	}

	events, err = arangoDb.FetchEvents(offset)
	if err != nil {
		log.Errorf("error: %v", err)
		return events, err
	}

	return events, nil
}

func (a *App) GetTopEvents(ctx context.Context, duration string) ([]db.Event, error){
	var events []db.Event
	arangoDb, err := service.NewArangoDB(context.Background())
	if err != nil {
		log.Errorf("error: %v", err)
		return nil, err
	}

	events, err = arangoDb.GetTopEvents(duration)
	if err != nil {
		log.Errorf("error: %v", err)
		return events, err
	}

	return events, nil
}

func (a *App) GetOngoingEvents(ctx context.Context, offset int) ([]db.Event, error) {

	var events []db.Event
	arangoDb, err := service.NewArangoDB(context.Background())
	if err != nil {
		log.Errorf("error: %v", err)
		return nil, err
	}

	events, err = arangoDb.FetchOngoingEvents(offset)
	if err != nil {
		log.Errorf("error: %v", err)
		return events, err
	}

	return events, nil
}

func (a *App) CreateCategory(ctx context.Context, name, description ,imageSize , imageData string)(db.Response, error) {

	category := db.Category{
		Name: name,
		Description: description,
		ImageSize: imageSize,
		ImageData: imageData,
	}

	return db.AddCategory(category), nil
}

func (a *App) EditCategory(ctx context.Context, id string, newName string, newDescription string, imageSize string, imageData string)(db.Response, error) {

	category := db.Category{
		Name: newName,
		Description: newDescription,
		ImageSize: imageSize,
		ImageData: imageData,
	}
	log.Error(newName, newDescription, imageData)
	return db.UpdateCategory(category, id), nil
}

func (a *App) GetTicketList(ctx context.Context, offset int) ([]*db.Ticket, error) {

	var tickets []*db.Ticket
	arangoDb, err := service.NewArangoDB(context.Background())
	if err != nil {
		log.Errorf("error: %v", err)
		return nil, err
	}

	tickets, err = arangoDb.GetTicketList(offset)
	if err != nil {
		log.Error("could not get tikcets")
		log.Errorf("err: %v", err)
		return tickets, err
	}
	return tickets, nil
}

func (a *App) GetCategoryById(ctx context.Context, id string) (db.Category, error) {

	category, err := db.FetchCategoryById(id)
	if err != nil {
		log.Errorf("error: %v", err)
		return category, err
	}

	return category, nil
}

func (a *queryResolver) GetOrganizerByID(ctx context.Context, id string) (db.OrganizerData, error) {
	organizer, err := db.GetOrganizerByID(id)
	if err != nil {
		log.Error(err)
		return organizer, err
	}
	return organizer, nil
}

func (a *queryResolver) GetOrganizersEventList(ctx context.Context, id string, offset int) ([]db.Event, error) {
	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return nil, err
	}
	events, err := arango.OrganizersEventList(id, offset)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	return events, nil
}

func (a *queryResolver) OrganizerProfileList(ctx context.Context, status string, offset int) ([]db.OrganizerProfile, error) {
	organizerProfile, err := db.OrganizerProfileList(status, offset)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return organizerProfile, nil
}

func (a *queryResolver) GetApprovedOrganizersList(ctx context.Context)([]db.OrganizerProfile, error) {
	organizer, err := db.GetApprovedOrganizersList()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return organizer, nil
}

func (a *mutationResolver) ApproveOrganizer(ctx context.Context, id, organizer, email string) (db.Response, error) {
	res, err := mailClient.ApproveOrganizerMail(organizer, email)
	if err != nil {
		return db.Response{}, err
	}

	if res.Status {
		arango, err := db.NewArangoDB(context.Background())
		if err != nil {
			log.Error(err)
			return db.Response{}, err
		}
		response, err := arango.ApproveOrganizer(id)
		if err != nil {
			log.Error(err)
			return db.Response{}, err
		}
		return response, nil
	}

	return db.Response{Status:false, Message: res.Message}, nil
}

func (a *queryResolver) GetUserList(ctx context.Context, offset int) ([]db.User, error) {
	users, err := db.GetUserList(offset)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return users, nil
}

func (a *queryResolver) GetPositionByID(ctx context.Context, id int) (db.Position, error) {
	position, err := db.GetPositionByID(id)
	if err != nil {
		log.Error(err)
		return db.Position{},err
	}

	return position, nil
}

func (a *queryResolver) GetEventTypeById(ctx context.Context, id int) (db.EventType, error) {
	eventType, err := db.GetEventTypeByID(id)
	if err != nil {
		log.Error(err)
		return db.EventType{}, err
	}
	return eventType, nil
}

func (a *mutationResolver) CreateNewEventType(ctx context.Context, eventTypeName string, desc string) (db.Response, error) {
	response, err := db.CreateNewEventType(eventTypeName, desc)
	if err != nil {
		log.Error(err)
		return db.Response{}, err
	}
	return response, nil
}

func (a *mutationResolver) UpdateEventType(ctx context.Context,id int, eventTypeName string, desc string) (db.Response, error) {
	response, err := db.UpdateEventType(id, eventTypeName, desc)
	if err != nil {
		log.Error(err)
		return db.Response{}, err
	}
	return response, nil
}

func (a *mutationResolver) CreateNewJobPosition(ctx context.Context, name string, desc string) (db.Response, error) {
	response, err := db.CreateNewJobPosition(name, desc)
	if err != nil {
		log.Error(err)
		return response, err
	}

	return response, nil
}

func (a *mutationResolver) UpdateJobPositionById(ctx context.Context, id int, name string, desc string) (db.Response, error) {
	response, err := db.UpdateJobPositionByID(id, name, desc)
	if err != nil {
		log.Error(err)
		return db.Response{}, err
	}
	return response, nil
}

func (a *mutationResolver) RejectOrganizerByID(ctx context.Context, id string, reason string, desc string) (db.Response, error) {
	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return db.Response{}, err
	}

	response, err := arango.RejectOrganizerByID(id, reason, desc)
	if err != nil {
		log.Error(err)
		return response, err
	}
	return response, nil
}

func (a *mutationResolver) UpdateEventEtCommission(ctx context.Context, eventId string, etCommission float64) (db.Response, error){
	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return db.Response{}, err
	}

	res, err := arango.UpdateEventEtCommissionRate(eventId, etCommission)
	if err != nil {
		return db.Response{}, err
	}

	return res, nil
}

func (a *mutationResolver) UpdateOrganizer(ctx context.Context, description, website, agreement_uid, o_id, fname, lname, salutation, mobile, blogUrl, user_id string, etCommission float64, etNotes string) (db.Response, error) {
	var organizer db.OrganizerData
	organizer.OrganizerProfile.Description = description
	organizer.OrganizerProfile.Website = website
	organizer.OrganizerProfile.AgreementUploadId = agreement_uid
	organizer.OrganizerProfile.ID = o_id
	organizer.OrganizerProfile.EtCommission = etCommission
	organizer.UserProfile.FirstName = fname
	organizer.UserProfile.LastName = lname
	organizer.UserProfile.Salutation = salutation
	organizer.UserProfile.MobileNumber = mobile
	organizer.UserProfile.Email = blogUrl
	organizer.UserProfile.UserId = user_id
	organizer.OrganizerProfile.EtNotes = etNotes

	res, err := db.UpdateOrganizer(organizer)
	if err != nil {
		log.Error(err)
		return res, err
	}
	return res, nil
}

func (a *mutationResolver) CreateAreaOfInterest(ctx context.Context, name string, desc string) (db.Response, error) {
	response, err := db.CreateAreaOfInterest(name, desc)
	if err != nil {
		log.Error(err)
		return response, err
	}
	return response, nil
}

func (a *queryResolver) GetInterestByID(ctx context.Context, id int) (db.Interest, error) {
	interest, err := db.GetInterestByID(id)
	if err != nil {
		log.Error(err)
		return interest, err
	}
	return interest, nil
}

func (a *mutationResolver) UpdateAreaOfInterest(ctx context.Context, id int, name string, desc string) (db.Response, error) {
	response, err := db.UpdateAreaOfInterest(id, name, desc)
	if err != nil {
		log.Error(err)
		return response, err
	}
	return response, nil
}

func (a *queryResolver) ListOfInterests(ctx context.Context, offset int) ([]db.Interest, error) {
	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return nil, err
	}
	interests, err := arango.ListOfInterests(offset)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return interests, nil
}

func (a *queryResolver) ListEventTypes(ctx context.Context, offset int) ([]db.EventType, error) {
	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return nil, err
	}
	eventTypes, err := arango.ListEventTypes(offset)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return eventTypes, nil
}

func (a *queryResolver) ListPositions(ctx context.Context, offset int) ([]db.Position, error) {
	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return nil, err
	}

	positions, err := arango.ListPositions(offset)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return positions, nil
}

func (a *queryResolver) ListRegions(ctx context.Context, offset int) ([]db.Region, error) {
	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return nil, err
	}

	regions, err := arango.ListRegions(offset)
	if err != nil {
		return nil, err
	}

	return regions, nil
}

func (a *mutationResolver) CreateNewRegion(ctx context.Context, name string) (db.Response, error) {
	res, err := db.CreateNewRegion(name)
	if err != nil {
		log.Error(err)
		return res, err
	}
	return res, nil
}

func (a *queryResolver) GetRegionByID(ctx context.Context, id int) (db.Region, error) {
	res, err := db.GetRegionByID(id)
	if err != nil {
		log.Error(err)
		return db.Region{}, err
	}

	return res, err
}

func (a *mutationResolver) UpdateRegionByID(ctx context.Context, id int, name string) (db.Response, error) {
	res, err := db.UpdateRegionByID(id, name)
	if err != nil {
		log.Error(err)
		return res, err
	}
	return res, nil
}

func (a *queryResolver) GetPayoutDetails(ctx context.Context, offset int) ([]db.Payout, error) {
	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return nil, err
	}

	res, err := arango.PayoutDetails(offset)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, nil
}

func (a *queryResolver) UpcomingPayouts(ctx context.Context, offset, limit int) ([]db.Payout, error) {
	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return nil, err
	}

	res, err := arango.UpcomingPayouts(offset, limit)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, nil
}

func (a *queryResolver) GetUserProfilesCount(ctx context.Context) (int,error) {
	res, err := db.GetUserProfilesCount()
	if err != nil {
		log.Error(err)
		return 0, err
	}

	return res, nil
}

func (a *queryResolver) GetOrganizerProfilesCount(ctx context.Context) (int, error) {
	res, err := db.GetOrganizerProfilesCount()
	if err != nil {
		log.Error(err)
		return 0, err
	}

	return res, nil
}

func (a *queryResolver) GetTotalEventsCreated(ctx context.Context) (int, error) {
	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return 0, err
	}

	res, err := arango.GetTotalEventsCreated()
	if err != nil {
		log.Error(err)
		return 0, err
	}

	return res, nil
}

func (a *queryResolver) GetRecentEvents(ctx context.Context) ([]db.Event, error) {
	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return nil, err
	}

	recentEvents, err := arango.GetRecentEvents()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return recentEvents, nil
}

func (a *mutationResolver) MarkAsPaidByEventId(ctx context.Context, eventId string) (db.Response, error){
	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return db.Response{}, err
	}

	res, err := arango.MarkAsPaidByEventID(eventId)
	if err != nil {
		log.Error(err)
		return res, err
	}
	return res, nil
}

func (a *queryResolver) TotalRevenueGenerated(ctx context.Context) (string, error) {
	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return "", err
	}
	return arango.TotalRevenueGenerated()
}

func (a *queryResolver) GetEventsOverview(ctx context.Context) (db.EventsOverview, error) {
	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return db.EventsOverview{}, err
	}
	return arango.EventReportsOverview()
}

func (a *queryResolver) GetEventSpecificReport(ctx context.Context, id string) (db.Event, error) {
	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return db.Event{}, err
	}
	return arango.EventSpecificReport(id)
}

func (a *queryResolver) ExhibitorEnquiriesCount(ctx context.Context, eventID string) (int, error) {
	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return 0, err
	}

	count, err := arango.ExhibitorEnquiriesCount(eventID)
	if err != nil {
		log.Info(err)
		return 0, err
	}

	return count, nil
}

func (a *queryResolver) SponsorEnquiriesCount(ctx context.Context, eventID string) (int, error) {
	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return 0, err
	}

	count, err := arango.SponsorEnquiriesCount(eventID)
	if err != nil {
		log.Info(err)
		return 0, err
	}

	return count, nil
}

func (a *queryResolver) TotalBrochureRequests(ctx context.Context, eventID string) (int, error) {
	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return 0, err
	}

	count, err := arango.TotalBrochureRequests(eventID)
	if err != nil {
		log.Info(err)
		return 0, err
	}

	return count, nil
}

func (a *queryResolver) GetGlobalEtCommissionRate(ctx context.Context) (float64, error) {
	etShare, err := db.GetGlobalEtCommissionRate()
	if err != nil {
		log.Info(err)
		return 0, err
	}

	return etShare, nil
}

func (a *queryResolver) GetEventPageHitsByID(ctx context.Context, eventId string) (int, error) {
	total, err := analytics.GetPageViewsByEventID(eventId)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (a *mutationResolver) UpdateGlobalEtCommissionRate(ctx context.Context, etCommission float64) (db.Response, error) {
	res, err := db.UpdateGlobalEtCommissionRate(etCommission)
	if err != nil {
		log.Info(err)
		return db.Response{}, err
	}

	return res, nil
}

func (a *queryResolver) GetOrganizersReport(ctx context.Context, offset int) ([]db.OrganizerProfile, error) {
	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return nil, err
	}

	res, err := arango.OrganizersReport(offset)
	if err != nil {
		log.Info(err)
		return nil, err
	}

	return res, nil
}

func (a *mutationResolver) ChangeOrganizersStatus(ctx context.Context, organizerId string, status int, value bool) (db.Response, error){
	_, err := db.ChangeOrganizerStatus(organizerId, status)
	if err != nil {
		return db.Response{}, err
	}

	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return db.Response{}, err
	}
	eventID, err := arango.ChangeOrganizersEventsStatus(organizerId, value)
	if err != nil {
		return db.Response{}, err
	}

	es, err := elastic.NewElasticSearch(context.Background())
	if err != nil {
		return db.Response{}, err
	}

	ch := make(chan bool, len(eventID))


	if value {

		for _, v := range eventID {
			go func(eID string){
				es.DeactivateEventByID(eID)
				ch <- true
			}(v)
		}

		for i := 0; i < len(eventID); i++ {
			<- ch
		}

	} else {

		for _, v := range eventID {
			go func(eID string){
				es.ActivateEventByID(eID)
				ch <- true
			}(v)
		}

		for i := 0; i < len(eventID); i++ {
			<- ch
		}
	}

	return db.Response{Status:true, Message:"Success."}, nil
}

func (a *mutationResolver) DeactivateEventByID(ctx context.Context, eventID string) (db.Response, error) {
	es, err := elastic.NewElasticSearch(context.Background())
	if err != nil {
		return db.Response{}, err
	}

	err = es.DeactivateEventByID(eventID)
	if err != nil {
		return db.Response{}, err
	}

	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return db.Response{}, err
	}

	res, err := arango.DeactivateEventByID(eventID)
	if err != nil {
		return db.Response{}, err
	}

	return res, nil
}

func (a *mutationResolver) ActivateEventByID(ctx context.Context, eventID string) (db.Response, error) {
	es, err := elastic.NewElasticSearch(context.Background())
	if err != nil {
		return db.Response{}, err
	}

	err = es.ActivateEventByID(eventID)
	if err != nil {
		return db.Response{}, err
	}

	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return db.Response{}, err
	}

	res, err := arango.ActivateEventByID(eventID)
	if err != nil {
		return db.Response{}, err
	}

	return res, nil
}

func (a *queryResolver) TotalEtEarnings(ctx context.Context) (string, error) {
	arango, err := db.NewArangoDB(context.Background())
	if err != nil {
		log.Error(err)
		return "", err
	}

	total, err := arango.GetTotalEtEarnings()
	if err != nil {
		return "", err
	}

	return total, nil
}