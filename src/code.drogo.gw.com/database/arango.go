package database

import (
	"github.com/arangodb/go-driver/http"
	arango "github.com/arangodb/go-driver"
	"github.com/spf13/viper"
	"context"
	log "github.com/sirupsen/logrus"
	"strings"
	"errors"
	"fmt"
)

type ArangoDB struct {
	client arango.Client
	conn   arango.Connection
	ctx    context.Context
}
func NewArangoDB(ctx context.Context) (*ArangoDB, error) {

	conn, err := getConnection()
	if err != nil {
		return nil, err
	}

	client, err := arango.NewClient(arango.ClientConfig{
		Connection: conn,
		Authentication: arango.
			BasicAuthentication(viper.GetString("ARANGO_USERNAME"), viper.GetString("ARANGO_PASSWORD")),
	})

	if err != nil {
		return nil, err
	}

	return &ArangoDB{client: client, conn: conn, ctx: ctx}, nil
}

func (a *ArangoDB) Database(db string) (arango.Database, error) {
	if len(db) == 0 {
		db = viper.GetString("ARANGO_DB")
	}
	return a.client.Database(a.ctx, db)
}

func getConnection() (arango.Connection, error) {
	return http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{viper.GetString("ARANGO_HOST")},
	})
}

func (a *ArangoDB) ApproveOrganizer(orgID string) (Response, error){
	errSql := make(chan error, 1)
	errArango := make(chan error, 1)

	go func() {
		stmt, err := DB.Prepare("UPDATE organizer_profiles " +
			"SET status = 'APPROVED', is_active = 1 " +
			"WHERE id = ?")
		if err != nil {
			log.Error(err)
			errSql <- err
			close(errSql)
			return
		}

		res, err := stmt.Exec(orgID)
		if err != nil {
			log.Error(err)
			errSql <- err
			close(errSql)
			return
		}

		rowCount, err := res.RowsAffected()
		if rowCount < 1 {
			log.Error(errors.New("no data with given ID"))
			errSql <- errors.New("no data with given ID")
			close(errSql)
			return
		}

		if err != nil {
			log.Error(err)
			errSql <- err
			close(errSql)
			return
		}
		errSql <- nil
		close(errSql)
	}()

	go func() {
		db, err := a.Database("")
		if err != nil {
			log.Error(err)
			errArango <- err
			close(errArango)
			return
		}

		query := `
		for e in events
		filter e.organizer.id == @orgId
		update e with {
    		organizer: {status: "APPROVED"}
		} in events`

		bindVars := map[string]interface{}{"orgId": orgID}

		_, err = db.Query(a.ctx, query, bindVars)
		if err != nil {
			log.Error(err)
			errArango <- err
			close(errArango)
			return
		}

		errArango <- nil
		close(errArango)
	}()

	if err := <-errSql; err != nil {
		return Response{}, err
	}
	if err := <-errArango; err != nil {
		return Response{}, err
	}

	return Response{
		Status: true,
		Message: "Organizer updated successfully",
	}, nil
}

//Get Events for the reports purpose
func (a  *ArangoDB) ReportsEventList(id string)([]Event, error) {
	db, err := a.Database("")
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var query string
	var bindVars map[string]interface{}

	if strings.ToLower(id) == "all" {
		query = `
		for e in events
		return {
    		eventId: e.eventDetail.id,
    		name: e.eventDetail.name
		}`
		bindVars = nil
	} else {
		query = `
		for e in events
		filter e.organizer.id == @id
		return {
    		eventId: e.eventDetail.id,
    		name: e.eventDetail.name
		}`
		bindVars = map[string]interface{}{"id": id}
	}

	cursor, err := db.Query(a.ctx, query, bindVars)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer cursor.Close()

	events := make([]Event, 0)

	for {
		var e Event

		_, err := cursor.ReadDocument(a.ctx, &e)
		if err != nil {
			if arango.IsNoMoreDocuments(err) {
				break
			}
			log.Error(err)
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

//Update Event Et Commission Rate...
func  (a *ArangoDB)UpdateEventEtCommissionRate(eventId string, etCommission float64) (Response, error) {
	if etCommission >= 100 {
		return Response{ Status: false, Message: "Et Commission cannot be greater or equal to 100"}, nil
	}

	db, err := a.Database("")
	if err != nil {
		log.Error(err)
		return Response{}, err
	}

	query := `
	for e in events
	filter e.eventDetail.id == @eventId
	update e with {
    	eventDetail: {et_commission_rate: @etCommission}
	} in events`

	bindVars := map[string]interface{}{"eventId": eventId, "etCommission": etCommission}

	cursor, err := db.Query(a.ctx, query, bindVars)
	if err != nil {
		return Response{}, err
	}

	cursor.Close()

	return Response{Status:true, Message:"Updated Et Commission Rate in event."}, nil
}

// Shows up in Organizer's Profile
func (a *ArangoDB) OrganizersEventList(id string, offset int) ([]Event, error) {
	db, err := a.Database("")
	if err != nil {
		log.Error(err)
		return nil, err
	}

	query := `
		FOR e IN events
		FILTER e.organizer.id == @id
		LIMIT @offset,11
		RETURN {
			"name": e.eventDetail.name,
			"startDate": e.eventDetail.start_date,
			"venueCity": e.eventDetail.venue_city,
			"status": e.organizer.status
		}`

	bindVars := map[string]interface{}{"id": id, "offset":offset*10}

	cursor, err := db.Query(a.ctx, query, bindVars)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer cursor.Close()

	events := make([]Event, 0)

	for {
		var e Event

		_, err := cursor.ReadDocument(a.ctx, &e)
		if err != nil {
			if arango.IsNoMoreDocuments(err) {
				break
			}
			log.Error(err)
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (a *ArangoDB) ListOfInterests(offset int) ([]Interest, error) {

	sqlCh := make(chan []Interest)
	arangoCh := make(chan map[string]int)

	errSql := make(chan error, 1)
	errArango := make(chan error, 1)

	go func(){
		rows, err := DB.Query("SELECT id, name from lookup_interests LIMIT 11 OFFSET ?", offset*10)
		if err != nil {
			log.Errorf("error: %v",err)
			errSql <- err
			return
		}
		defer rows.Close()

		var interests []Interest

		for rows.Next() {
			var i Interest
			err = rows.Scan(&i.Id, &i.Name)
			if err != nil {
				log.Errorf("error: %v",err)
				errSql <- err
				return
			}
			interests = append(interests, i)
		}

		err = rows.Err()
		if err != nil {
			log.Errorf("error: %v",err)
			errSql <- err
			return
		}

		errSql <- nil
		sqlCh <- interests

		close(errSql)
		close(sqlCh)
	}()

	go func() {
		db, err := a.Database("")
		if err != nil {
			log.Errorf("error: %v",err)
			errArango <- err
			return
		}

		query := `
		RETURN FLATTEN (
   			FOR e in events
   			RETURN e.interests  
		)`

		cursor, err := db.Query(a.ctx, query, nil)
		if err != nil {
			log.Errorf("error: %v",err)
			errArango <- err
			return
		}
		defer cursor.Close()

		var arangoInterests []string

		for {
			_, err = cursor.ReadDocument(a.ctx, &arangoInterests)
			if err != nil {
				if arango.IsNoMoreDocuments(err) {
					break
				}
				log.Errorf("error: %v",err)
				errArango <- err
				return
			}
		}

		interests := frequencyCount(arangoInterests)

		errArango <- nil
		arangoCh <- interests

		close(arangoCh)
		close(errArango)
	}()

	if e1 := <- errSql; e1 != nil {
		return nil, e1
	}
	if e2 := <- errArango; e2 != nil {
		return nil, e2
	}

	sqlInterest, arangoInterest := <-sqlCh, <-arangoCh

	for i, val := range sqlInterest {
		_, exists := arangoInterest[val.Name]
		if exists {
			sqlInterest[i].NoOfEvents = arangoInterest[val.Name]
		}
	}

	return sqlInterest, nil
}

func (a *ArangoDB) ListEventTypes(offset int) ([]EventType, error) {

	sqlCh := make(chan []EventType)
	arangoCh := make(chan map[string]int)

	errSql := make(chan error, 1)
	errArango := make(chan error, 1)

	go func() {
		rows, err := DB.Query("Select id, name from lookup_event_types LIMIT 11 OFFSET ?", offset*10)
		if err != nil {
			log.Error(err)
			errSql <- err
			return
		}
		defer rows.Close()

		var eventTypes []EventType

		for rows.Next() {
			var e EventType
			err = rows.Scan(&e.Id, &e.Name)
			if err != nil {
				log.Error(err)
				errSql <- err
				return
			}
			eventTypes = append(eventTypes, e)
		}
		err = rows.Err()
		if err != nil {
			log.Error(err)
			errSql <- err
			return
		}

		errSql <- nil
		sqlCh <- eventTypes

		close(errSql)
		close(sqlCh)
	}()

	go func() {
		db, err := a.Database("")
		if err != nil {
			log.Error(err)
			errArango <- err
			return
		}

		query := `
		RETURN FLATTEN (
   			FOR e in events
   			RETURN e.eventTypes  
		)`

		cursor, err := db.Query(a.ctx, query, nil)
		if err != nil {
			log.Error(err)
			errArango <- err
			return
		}
		defer cursor.Close()

		var arangoEventTypes []string

		for {
			_, err = cursor.ReadDocument(a.ctx, &arangoEventTypes)
			if err != nil {
				if arango.IsNoMoreDocuments(err) {
					break
				}
				log.Error(err)
				errArango <- err
				return
			}
		}

		eventTypes := frequencyCount(arangoEventTypes)

		errArango <- nil
		arangoCh <- eventTypes

		close(arangoCh)
		close(errArango)
	}()

	if e1 := <- errSql; e1 != nil {
		return nil, e1
	}
	if e2 := <- errArango; e2 != nil {
		return nil, e2
	}


	sqlEventTypes, arangoEventTypes := <- sqlCh, <- arangoCh

	for i, val := range sqlEventTypes {
		_, exists := arangoEventTypes[val.Name]

		if exists {
			sqlEventTypes[i].NoOfEvents = arangoEventTypes[val.Name]
		}
	}
	return sqlEventTypes, nil
}

func (a *ArangoDB) ListPositions(offset int) ([]Position, error) {

	sqlCh := make(chan []Position)
	arangoCh := make(chan map[string]int)

	errSql := make(chan error, 1)
	errArango := make(chan error, 1)

	go func() {
		rows, err := DB.Query("Select id, name from lookup_attendees LIMIT 11 OFFSET ?", offset*10)
		if err != nil {
			log.Error(err)
			errSql <- err
			return
		}
		defer rows.Close()

		var positions []Position

		for rows.Next() {
			var p Position
			err = rows.Scan(&p.Id, &p.Name)
			if err != nil {
				log.Error(err)
				errSql <- err
				return
			}
			positions = append(positions, p)
		}
		err = rows.Err()
		if err != nil {
			log.Error(err)
			errSql <- err
			return
		}

		errSql <- nil
		sqlCh <- positions

		close(errSql)
		close(sqlCh)
	}()

	go func() {
		db, err := a.Database("")
		if err != nil {
			log.Error(err)
			errArango <- err
			return
		}

		query := `
		RETURN FLATTEN (
   			FOR e in events
   			RETURN e.attendees  
		)`

		cursor, err := db.Query(a.ctx, query, nil)
		if err != nil {
			log.Error(err)
			errArango <- err
			return
		}
		defer cursor.Close()

		var arangoPositions []string

		for {
			_, err = cursor.ReadDocument(a.ctx, &arangoPositions)
			if err != nil {
				if arango.IsNoMoreDocuments(err) {
					break
				}
				log.Error(err)
				errArango <- err
				return
			}
		}

		positions := frequencyCount(arangoPositions)

		errArango <- nil
		arangoCh <- positions

		close(arangoCh)
		close(errArango)
	}()

	if e1 := <- errSql; e1 != nil {
		return nil, e1
	}
	if e2 := <- errArango; e2 != nil {
		return nil, e2
	}

	sqlPositions, arangoPositions := <-sqlCh, <-arangoCh

	for i, val := range sqlPositions {
		_, exists := arangoPositions[val.Name]
		if exists {
			sqlPositions[i].NoOfEvents = arangoPositions[val.Name]
		}
	}

	return sqlPositions, nil
}

func (a*ArangoDB) ListRegions(offset int) ([]Region, error) {

	sqlCh := make(chan []Region)
	arangoCh := make(chan map[string]int)

	errSql := make(chan error, 1)
	errArango := make(chan error, 1)

	go func() {
		rows, err := DB.Query("Select id, name from lookup_locations LIMIT 11 OFFSET ?", offset*10)
		if err != nil {
			log.Error(err)
			errSql <- err
			return
		}
		defer rows.Close()

		var regions []Region

		for rows.Next() {
			var r Region
			err = rows.Scan(&r.Id, &r.Name)
			if err != nil {
				log.Error(err)
				errSql <- err
				return
			}
			regions = append(regions, r)
		}
		err = rows.Err()
		if err != nil {
			log.Error(err)
			errSql <- err
			return
		}

		errSql <- nil
		sqlCh <- regions

		close(errSql)
		close(sqlCh)
	}()

	go func() {
		db, err := a.Database("")
		if err != nil {
			log.Error(err)
			errArango <- err
			return
		}

		query := `
		RETURN FLATTEN (
   			FOR e in events
   			RETURN e.eventDetail.venue_region 
		)`

		cursor, err := db.Query(a.ctx, query, nil)
		if err != nil {
			log.Error(err)
			errArango <- err
			return
		}
		defer cursor.Close()

		var arangoRegions []string

		for {
			_, err = cursor.ReadDocument(a.ctx, &arangoRegions)
			if err != nil {
				if arango.IsNoMoreDocuments(err) {
					break
				}
				log.Error(err)
				errArango <- err
				return
			}
		}

		regions := frequencyCount(arangoRegions)

		errArango <- nil
		arangoCh <- regions

		close(arangoCh)
		close(errArango)
	}()

	if e1 := <- errSql; e1 != nil {
		return nil, e1
	}
	if e2 := <- errArango; e2 != nil {
		return nil, e2
	}

	sqlRegions, arangoRegions := <-sqlCh, <-arangoCh

	for i, val := range sqlRegions {
		_, exists := arangoRegions[val.Name]
		if exists {
			sqlRegions[i].NoOfEvents = arangoRegions[val.Name]
		}
	}

	return sqlRegions, nil
}

func (a *ArangoDB) PayoutDetails(offset int) ([]Payout, error) {
	db, err := a.Database("")
	if err != nil {
		log.Error(err)
		return nil, err
	}

	query := `
		FOR e IN events 
		FILTER e.eventDetail.end_date < DATE_ISO8601(DATE_ADD(DATE_NOW(), -1, "day")) && Length(
                    FOR t IN tickets
                    FILTER e.eventDetail.id == t.EventID && t.Payment == true
                    RETURN t.EventID
                    ) != 0

		LET etShare = (
		    FOR t IN tickets
		    FILTER e.eventDetail.id == t.EventID && t.Payment == true
          	RETURN t.ExchangeRate == 0 ? 0 : (t.AmountCharged - t.TaxAmount) / t.ExchangeRate * e.eventDetail.et_commission_rate * 0.01 
		)
		
		LET revenue = (
        RETURN SUM (
    		FOR t IN tickets 
    		FILTER t.EventID == e.eventDetail.id && t.Payment == true
    		RETURN t.ExchangeRate == 0 ? 0 : t.AmountCharged/t.ExchangeRate
		)   
        )[0]
        
        LET taxAmount = (
            FOR t IN tickets
            FILTER t.EventID == e.eventDetail.id && t.Payment == true
            RETURN t.TaxAmount / t.ExchangeRate
        )

		SORT e.eventDetail.end_date DESC

		LIMIT @offset,11

		RETURN {
   			id: e.eventDetail.id,
   			name: e.eventDetail.name,
   			endDate: e.eventDetail.end_date,
			sold: (
			    LET sold = (
                RETURN SUM(
                    FOR t IN tickets
                    FILTER e.eventDetail.id == t.EventID && t.Payment == true
                    RETURN t.NoOfAttendees
                    )
                )
                return sold
            )[0][0],
			saleAmount: revenue,
			taxAmount: SUM(taxAmount),
   			etShare: SUM(etShare),
   			payoutAmount: revenue  - sum(etShare),
   			status : e.payouts_status  == "PAID" ? "PAID" : "UNPAID"  
		}`

	bindVars := map[string]interface{}{"offset": offset*10}

	cursor, err := db.Query(a.ctx, query, bindVars)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer cursor.Close()

	payouts := make([]Payout, 0)

	for {
		var p Payout

		_, err := cursor.ReadDocument(a.ctx, &p)
		if err != nil {
			if arango.IsNoMoreDocuments(err) {
				break
			}
			log.Error(err)
			return nil, err
		}

		p.TotalEtShare = fmt.Sprintf("%.3f", p.EtShare)
		p.TotalPayoutAmount = fmt.Sprintf("%.3f", p.PayoutAmount)
		p.TotalSaleAmount = fmt.Sprintf("%.3f", p.SaleAmount)
		p.TotalTaxAmount = fmt.Sprintf("%.3f", p.TaxAmount)

		payouts = append(payouts, p)
	}

	return payouts, nil
}

func (a *ArangoDB) UpcomingPayouts(offset, limit int) ([]Payout, error) {
	db, err := a.Database("")
	if err != nil {
		log.Error(err)
		return nil, err
	}

	query := `
		FOR e IN events 
		FILTER e.eventDetail.end_date < DATE_ISO8601(DATE_ADD(DATE_NOW(), -1, "day")) && e.payouts_status != 'PAID' && Length(
                    FOR t IN tickets
                    FILTER e.eventDetail.id == t.EventID && t.Payment == true
                    RETURN t.EventID
                    ) != 0

		LET etShare = (
		    FOR t IN tickets
		    FILTER e.eventDetail.id == t.EventID && t.Payment == true
          	RETURN t.ExchangeRate == 0 ? 0 : (t.AmountCharged - t.TaxAmount) / t.ExchangeRate * e.eventDetail.et_commission_rate * 0.01 
		)
		
		LET revenue = (
        RETURN SUM (
    		FOR t IN tickets 
    		FILTER t.EventID == e.eventDetail.id && t.Payment == true
    		RETURN t.ExchangeRate == 0 ? 0 : t.AmountCharged/t.ExchangeRate
		)   
        )[0]

		SORT e.eventDetail.end_date DESC

		LIMIT @offset, @limit

		RETURN {
   			id: e.eventDetail.id,
   			name: e.eventDetail.name,
   			endDate: e.eventDetail.end_date,
   			payoutAmount: revenue  - sum(etShare)
		}`

	bindVars := map[string]interface{}{"offset": offset*10, "limit": limit}

	cursor, err := db.Query(a.ctx, query, bindVars)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer cursor.Close()

	payouts := make([]Payout, 0)

	for {
		var p Payout

		_, err := cursor.ReadDocument(a.ctx, &p)
		if err != nil {
			if arango.IsNoMoreDocuments(err) {
				break
			}
			log.Error(err)
			return nil, err
		}
		payouts = append(payouts, p)
	}

	return payouts, nil
}

func (a *ArangoDB) MarkAsPaidByEventID(eventId string) (Response, error) {
	db, err := a.Database("")
	if err != nil {
		log.Error(err)
		return Response{}, err
	}

	query := `
	for e in events
	filter e.eventDetail.id == @eventId
	update e with {
    	payouts_status: "PAID"
	} in events`

	bindVars := map[string]interface{}{"eventId" : eventId}

	cursor, err := db.Query(a.ctx, query, bindVars)
	if err != nil {
		log.Error(err)
		return Response{}, err
	}
	defer cursor.Close()

	return Response{
		Status: true,
		Message: "Successfully marked as paid.",
	}, nil
}

func (a *ArangoDB) FetchEvents(offset int) ([]Event, error) {
	db, err := a.Database("")
	if err != nil {
		log.Info(err)
		return nil, err
	}

	query := `FOR e IN events 
			LIMIT @offset, 11
			return {
				"eventId": e.eventDetail.id,
				"startDate": e.eventDetail.start_date,
				"name": e.eventDetail.name,
				"organizerName": e.organizer.name,
				"organizerId": e.organizer.id,
				"ticketsSold": e.ticket.sold,
				"status": e.status,
				"isFeatured": e.is_featured == 'no' ? false : true
			}`

	bindVars := map[string]interface{}{"offset": offset * 10}

	cursor, err := db.Query(a.ctx, query, bindVars)
	if err != nil {
		log.Info(err)
		return nil, err
	}
	defer cursor.Close()

	var events []Event

	for {
		var e Event

		_, err := cursor.ReadDocument(a.ctx, &e)
		if err != nil {
			if arango.IsNoMoreDocuments(err) {
				break
			}
			log.Info(err)
			return nil, err
		}

		events = append(events, e)
	}

return events, nil
}

func (a *ArangoDB) GetTicketList(offset int) ([]*Ticket, error) {
	var tickets []*Ticket
	db, err := a.Database("")
	if err != nil {
		return tickets, err
	}
	query := `FOR t IN tickets 
	LIMIT @offset, 11
	LET eventName = ( 
		for e in events 
		filter e.eventDetail.id == t.EventID
		return e.eventDetail.name 
	)
	return {
		"orderNumber": t.ID,
		"purchasedBy": t.Name,
		"noOfTickets": t.NoOfAttendees,
		"eventId": t.EventID,
		"eventName": FIRST(eventName)
    }`
	bindVars := map[string]interface{}{"offset": offset * 10}

	cursor, err := db.Query(a.ctx, query, bindVars)
	if err != nil {
		return tickets, err
	}
	defer cursor.Close()
	for {
		var e *Ticket
		_, err := cursor.ReadDocument(a.ctx, &e)
		if err != nil {
			if arango.IsNoMoreDocuments(err) {
				break
			}
			return tickets, err
		}
		tickets = append(tickets, e)
	}
	return tickets, nil
}

func (a *ArangoDB) FetchOngoingEvents(offset int) ([]Event, error) {

	sqlCh := make(chan map[string]string)
	arangoCh := make(chan []Event)
	errSql := make(chan error, 1)
	errArango := make(chan error, 1)

	go func() {
		db, err := a.Database("")
		if err != nil {
			errArango <- err
			return
		}
		query := `FOR e IN events
			FILTER e.eventDetail.start_date <= DATE_FORMAT(DATE_NOW(), "%yyyy-%mm-%dd")
				&&  e.eventDetail.end_date >= DATE_FORMAT(DATE_NOW(), "%yyyy-%mm-%dd")
			LIMIT @offset, 11
			return {
				"eventId": e.eventDetail.id,
				"startDate": e.eventDetail.start_date,
				"name": e.eventDetail.name,
				"organizerName": e.organizer.name,
				"organizerId": e.organizer.id,
				"status": e.status
			}`
		bindVars := map[string]interface{}{"offset": offset * 10}

		cursor, err := db.Query(a.ctx, query, bindVars)
		if err != nil {
			errArango <- err
			return
		}

		defer cursor.Close()
		var events []Event

		for {
			var e Event
			_, err := cursor.ReadDocument(a.ctx, &e)
			if err != nil {
				if arango.IsNoMoreDocuments(err) {
					break
				}
				errArango <- err
				return
			}
			events = append(events, e)
		}
		errArango <- nil
		arangoCh <- events
		close(arangoCh)
		close(errArango)
	}()

	go func() {
		userNames := make(map[string]string)

		rows, err := DB.Query("SELECT organizer_profiles.id, user_profiles.first_name, user_profiles.last_name " +
			"FROM organizer_profiles " +
			"INNER JOIN user_profiles " +
			"ON organizer_profiles.user_id = user_profiles.user_id")
		if err != nil {
			errSql <- err
			return
		}

		for rows.Next() {
			var organizerId, firstName, lastName string
			err := rows.Scan(&organizerId, &firstName, &lastName)
			if err != nil {
				errSql <- err
				return
			}
			userNames[organizerId] = firstName + " " + lastName
		}

		errSql <- nil
		sqlCh <- userNames
		close(errSql)
		close(sqlCh)
	}()

	if err := <-errArango; err != nil {
		return nil, err
	}

	if err := <-errSql; err != nil {
		return nil, err
	}

	userNames, events := <-sqlCh, <-arangoCh

	for i, _ := range events {
		events[i].UserName = userNames[events[i].OrganizerId]
	}
	return events, nil
}

func (a *ArangoDB) ListEventCategory(offset int) ([]Category, error) {

	sqlCh := make(chan []Category)
	arangoCh := make(chan map[string]int)
	errSql := make(chan error, 1)
	errArango := make(chan error, 1)

	go func() {

		rows, err := DB.Query("Select id, name from lookup_categories LIMIT 11 OFFSET ?", offset*10)
		if err != nil {
			errSql <- err
			return
		}

		defer rows.Close()

		var categories []Category

		for rows.Next() {
			var category Category
			err = rows.Scan(&category.ID, &category.Name)
			if err != nil {
				errSql <- err
				return
			}
			categories = append(categories, category)
		}

		err = rows.Err()
		if err != nil {
			errSql <- err
			return
		}
		errSql <- nil
		sqlCh <- categories
		close(errSql)
		close(sqlCh)
	}()

	go func() {

		db, err := a.Database("")
		if err != nil {
			errArango <- err
			return
		}

		query := `
				RETURN FLATTEN (
					FOR e in events
					RETURN (e.categories)
				)`

		cursor, err := db.Query(a.ctx, query, nil)
		if err != nil {
			errArango <- err
			return
		}
		defer cursor.Close()
		var arangoCategories []string

		for {
			_, err = cursor.ReadDocument(a.ctx, &arangoCategories)
			if err != nil {
				if arango.IsNoMoreDocuments(err) {
					break
				}
				errArango <- err
				return
			}
		}

		categories := frequencyCount(arangoCategories)
		errArango <- nil
		arangoCh <- categories
		close(arangoCh)
		close(errArango)

	}()

	if err := <-errSql; err != nil {
		return nil, err
	}
	if err := <-errArango; err != nil {
		return nil, err
	}
	sqlCategories, arangoCategories := <-sqlCh, <-arangoCh
	for i, val := range sqlCategories {
		_, exists := arangoCategories[val.Name]
		if exists {
			sqlCategories[i].NoOfEvents = arangoCategories[val.Name]
		}
	}
	return sqlCategories, nil
}

func(a *ArangoDB) TotalRevenueGenerated() (string, error) {
	db, err := a.Database("")
	if err != nil {
		return "", err
	}
	query := `
		RETURN SUM (
    		FOR t IN tickets
			FILTER t.Payment == true
    		RETURN t.ExchangeRate == 0 ? 0 : t.AmountCharged/t.ExchangeRate
		)`
	cursor, err := db.Query(a.ctx, query, nil)
	if err != nil {
		return "", err
	}
	defer cursor.Close()
		var revenueGenerated float64
		_, err = cursor.ReadDocument(a.ctx, &revenueGenerated)
		if err != nil {
			return "", err
		}
	return fmt.Sprintf("%.3f", revenueGenerated), nil
}

func (a *ArangoDB) GetTotalEventsCreated() (int, error) {
	db, err := a.Database("")
	if err != nil {
		log.Error(err)
		return 0, err
	}

	query := `
		RETURN LENGTH (events)`

	cursor, err := db.Query(a.ctx, query, nil)
	if err != nil {
		return 0, err
	}
	defer cursor.Close()

	var count int

	for {
		_, err = cursor.ReadDocument(a.ctx, &count)
		if err != nil {
			if arango.IsNoMoreDocuments(err) {
				break
			}
			log.Error(err)
			return 0, err
		}
	}
	return count, nil
}

func (a *ArangoDB) GetRecentEvents() ([]Event, error) {
	db, err := a.Database("")
	if err != nil {
		log.Error(err)
		return nil, err
	}

	query := `
		FOR e IN events
		SORT e.eventDetail.start_date DESC
		LIMIT 10
		RETURN {
    		name: e.eventDetail.name,
    		organizerName: e.organizer.name,
    		startDate: e.eventDetail.start_date
		}`

	cursor, err := db.Query(a.ctx, query, nil)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	events := make([]Event, 0)

	for {
		var e Event
		_, err = cursor.ReadDocument(a.ctx, &e)
		if err != nil {
			if arango.IsNoMoreDocuments(err) {
				break
			}
			log.Error(err)
			return nil, err
		}
		events = append(events, e)
	}

	return events, nil
}

func (a *ArangoDB) GetEventAttendees(eventId, organizerId string, offset int) ([]Attendee, error) {
	db, err := a.Database("")
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var query string
	var bindVars map[string]interface{}

	if strings.ToLower(eventId) != "all" && strings.ToLower(organizerId) != "all" {
		query = `
		for t in tickets
		filter t.Attendees != null && t.EventID == @eventId
		LIMIT @offset, 11
		return {
    		Attendee: t.Attendees[*].name,
    		EventName: (
        		let evtName = (
            		for e in events
            		filter e.eventDetail.id == @eventId
            		return e.eventDetail.name
        		)
        	return evtName
    		)[0][0],
    		OrganizerName: (
        		let orgName = (
            		for e in events
            		filter e.organizer.id == @organizerId
            		return e.organizer.name
        		)
			return orgName
    		)[0][0]
		}`

		bindVars = map[string]interface{}{"eventId": eventId, "organizerId": organizerId, "offset": offset*10}
	}

	if strings.ToLower(eventId) == "all" && strings.ToLower(organizerId) == "all" {
		query = `
		for t in tickets
			filter t.Attendees != null
			LIMIT @offset, 11
			return {
    			attendee: t.Attendees[*].name,
    			eventName: (
        			let evtName = (
            			for e in events
            				filter e.eventDetail.id == t.EventID
            				return e.eventDetail.name
        				)
        			return evtName
    			)[0][0],
    			organizerName: (
        			let orgName = (
            			for e in events
            				filter e.eventDetail.id == t.EventID
            				return e.organizer.name
						)
        			return orgName
    			)[0][0]
			}`

		bindVars = map[string]interface{}{"offset": offset*10}
	}

	if strings.ToLower(eventId) == "all" && strings.ToLower(organizerId) != "all" {
		query = `
		for t in tickets
		for e in events
		filter t.Attendees != null && t.EventID == e.eventDetail.id && e.organizer.id == @organizerId
		LIMIT @offset, 11
		
		return {
    		attendee: t.Attendees[*].name,
    		eventName: (
        		let evtName = (
            		filter e.eventDetail.id == t.EventID
            		return e.eventDetail.name
        		)
        		return evtName
    		)[0][0],
    		organizerName: (
        		let orgName = (
            		filter e.eventDetail.id == t.EventID
            		return e.organizer.name
        		)
        		return orgName
    		)[0][0]
		}`

		bindVars = map[string]interface{}{"organizerId": organizerId, "offset": offset*10}
	}

	if strings.ToLower(eventId) != "all" && strings.ToLower(organizerId) == "all" {
		query = `
		for t in tickets
		filter t.Attendees != null && t.EventID == @eventId
		return {
    		attendee: t.Attendees[*].name,
    		eventName: (
        		let evtName = (
            		for e in events
            		filter e.eventDetail.id == t.EventID
            		return e.eventDetail.name
        		)
        		return evtName
    		)[0][0],
    		organizerName: (
        		let orgName = (
            		for e in events
            		filter e.eventDetail.id == t.EventID
            		return e.organizer.name
        		)
        		return orgName
    		)[0][0]
		}`

		bindVars = map[string]interface{}{"eventId": eventId}
	}

	cursor, err := db.Query(a.ctx, query, bindVars)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var attendees []Attendee
	for {
		var at Attendee
		_, err := cursor.ReadDocument(a.ctx, &at)
		if err != nil {
			if arango.IsNoMoreDocuments(err) {
				break
			}
			log.Error(err)
			return nil, err
		}
		attendees = append(attendees, at)
	}

	return attendees, nil
}

func (a *ArangoDB) GetTotalEtEarnings() (string, error) {
	db, err := a.Database("")
	if err != nil {
		log.Error(err)
		return "", err
	}

	query := `
	return sum(
    	for e in events
        for t in tickets
        filter e.eventDetail.id == t.EventID && t.Payment == true
        return t.ExchangeRate == 0 ? 0 : (t.AmountCharged - t.TaxAmount) / t.ExchangeRate * e.eventDetail.et_commission_rate * 0.01 
    )`

	cursor, err := db.Query(a.ctx, query, nil)
	if err != nil {
		log.Error(err)
		return "", err
	}
	defer cursor.Close()

	var total float64

	_, err = cursor.ReadDocument(a.ctx, &total)
	if err != nil {
		log.Error(err)
		return "", err
	}

	return fmt.Sprintf("%.3f", total), nil
}

func (a *ArangoDB) OrganizersReport(offset int) ([]OrganizerProfile, error) {
	rows, err := DB.Query("Select id, name from organizer_profiles LIMIT 11 OFFSET ?", offset*10)
	if err != nil {
		log.Info(err)
		return nil, err
	}
	defer rows.Close()

	var ID []string
	var id []uint8
	var name string

	sqlOrganizers := make(map[string]string)

	for rows.Next() {

		err = rows.Scan(&id, &name)
		if err != nil {
			log.Info(err)
			return nil, err
		}
		ID = append(ID, string(id))

		sqlOrganizers[string(id)] = name
	}

	db, err := a.Database("")
	if err != nil {
		log.Error(err)
		return nil, err
	}

	operatedID := strings.Join(ID, "','")

	query := `
	for i in ['`+ operatedID +
		`']
	return {
    	id: i,
		name: (
    	    let name = (
    	        for e in events
    	        filter e.organizer.id == i
				LIMIT 1
    	        return e.organizer.name
    	    )
    	    return name
    	)[0][0],
    	eventsCount: (
      	let events = (
          	for e in events
          	filter e.organizer.id == i
          	return i
      	)
      	return length(events)
    	)[0],
    	etEarnings: (
      	let share= (
          	FOR e IN events
          	FOR t IN tickets
          	FILTER e.organizer.id == i && e.eventDetail.id == t.EventID && t.Payment == true
          	RETURN t.ExchangeRate == 0 ? 0 : (t.AmountCharged - t.TaxAmount) / t.ExchangeRate * e.eventDetail.et_commission_rate * 0.01 
    		)
		return sum(share)
    	)[0],
		revenue: (
		    LET revenue = (
		        RETURN SUM (
		        FOR e IN events
    		    FOR t IN tickets
    		    FILTER e.eventDetail.id == t.EventID && e.organizer.id == i && t.Payment == true
    		    RETURN t.ExchangeRate == 0 ? 0 : t.AmountCharged/t.ExchangeRate
		        )
            )
            RETURN revenue[0]
		)[0]
	}`

	cursor, err := db.Query(a.ctx, query, nil)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	var organizer []OrganizerProfile

	for {
		var o OrganizerProfile
		_, err := cursor.ReadDocument(a.ctx, &o)
		if err != nil {
			if arango.IsNoMoreDocuments(err) {
				break
			}
			log.Info(err)
			return nil, err
		}
		o.TotalEtEarning = fmt.Sprintf("%.3f", o.EtEarnings)
		o.RevenueGenerated = fmt.Sprintf("%.3f", o.Revenue)

		organizer = append(organizer, o)
	}

	for i, v := range organizer {
		if v.Name == "" {
			organizer[i].Name = sqlOrganizers[v.ID]
		}
	}

	return organizer, nil
}

func (a *ArangoDB) GetTopEvents(duration string) ([]Event, error) {
	db, err := a.Database("")
	if err != nil {
		log.Error(err)
		return nil, err
	}

	query := `
	RETURN SLICE (
		FOR e IN events
		FILTER DATE_ISO8601(e.ticket.end_date) >= DATE_ADD(DATE_NOW(), -1, "`+strings.ToLower(duration)+`")
    	LET revenue = (
			RETURN SUM (
    			FOR t IN tickets 
    			FILTER t.EventID == e.eventDetail.id && t.Payment == true
    			RETURN t.ExchangeRate == 0 ? 0 : t.AmountCharged/t.ExchangeRate
			)
    	)[0]
    	SORT revenue DESC
    	RETURN {
        	"revenue": revenue,
        	"name": e.eventDetail.name,
        	"eventId": e.eventDetail.id,
        	"ticketsSold": e.ticket.sold,
			"startDate": TO_STRING(DATE_TIMESTAMP(e.eventDetail.start_date)),
			"location": e.eventDetail.address
    	},0,5
	)`

	cursor, err := db.Query(a.ctx, query, nil)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	events := make([]Event, 0)

	for {
		_, err = cursor.ReadDocument(a.ctx, &events)
		if err != nil {
			if arango.IsNoMoreDocuments(err) {
				break
			}
			log.Error(err)
			return nil, err
		}
	}

	return events, nil
}


func (a *ArangoDB) EventReportsOverview() (EventsOverview, error) {
	var overView EventsOverview
	db, err := a.Database("")
	if err != nil {
		log.Error(err)
		return overView, err
	}

	query := `
	LET eventsHeld = (
		LENGTH (
			FOR e IN events
			FILTER DATE_FORMAT(DATE_NOW(), "%yyyy-%mm-%dd") > e.eventDetail.end_date
			RETURN e
    	)
	)
	LET ticketsSold = (
		SUM (
			FOR e IN events
			RETURN e.ticket.sold
		)
	)
	LET revenue = (
		RETURN SUM (
    		FOR t IN tickets
			FILTER t.Payment == true
    		RETURN t.ExchangeRate == 0 ? 0 : t.AmountCharged/t.ExchangeRate
		)
    )[0]
	RETURN {
    	"eventsHeld": eventsHeld,
    	"ticketsSold": ticketsSold,
    	"revenue": revenue
	}`

	cursor, err := db.Query(a.ctx, query, nil)
	if err != nil {
		return overView, err
	}
	defer cursor.Close()
	_, err = cursor.ReadDocument(a.ctx, &overView)
	if err != nil {
		if arango.IsNoMoreDocuments(err) {
			return overView, nil
		}
		log.Error(err)
		return overView, err
	}
	overView.TotalRevenue = fmt.Sprintf("%.3f", overView.Revenue)

	return overView, nil
}

func (a *ArangoDB) EventSpecificReport(id string)(Event, error) {
	db, err := a.Database("")
	if err != nil {
		log.Error(err)
		return Event{}, err
	}
	query := `
	FOR e IN events 
    FILTER e.eventDetail.id == @id
    LET days = (DATE_DIFF(e.ticket.start_date, e.ticket.end_date, "days"))
    LET ticketsSold = (e.ticket.sold)
    LET revenue = (
        RETURN SUM (
    		FOR t IN tickets 
    		FILTER t.EventID == e.eventDetail.id && t.Payment == true
    		RETURN t.ExchangeRate == 0 ? 0 : t.AmountCharged/t.ExchangeRate
		)   
    )[0]
	RETURN {
    	"name": e.eventDetail.name,
    	"ticketLiveDays": days,
    	"ticketsSold": ticketsSold,
    	"publishDate": e.eventDetail.created_on,
    	"revenue": revenue  
	}`
	bindVars := map[string]interface{}{"id": id}

	cursor, err := db.Query(a.ctx, query, bindVars)
	if err != nil {
		return Event{}, err
	}
	defer cursor.Close()
	var event Event
	_, err = cursor.ReadDocument(a.ctx, &event)
	if err != nil {
		if arango.IsNoMoreDocuments(err) {
			return event, nil
		}
		log.Error(err)
		return event, err
	}

	event.TotalRevenue = fmt.Sprintf("%.3f", event.Revenue)
	return event, nil
}

func (a *ArangoDB) ExhibitorEnquiriesCount(eventID string) (int, error) {
	db, err := a.Database("")
	if err != nil {
		log.Error(err)
		return 0, err
	}

	query := `
	RETURN LENGTH(
		FOR e IN enquiries
		FILTER e.eventId == @eventID AND e.enquiryType == "exhibitor"
		RETURN e)`

	bindVars := map[string]interface{}{"eventID": eventID}

	cursor, err := db.Query(a.ctx, query, bindVars)
	if err != nil {
		return 0, err
	}

	defer cursor.Close()

	var count int
	_, err = cursor.ReadDocument(a.ctx, &count)
	if err != nil {
		log.Info(err)
		return 0, err
	}

	return count, nil
}

func (a *ArangoDB) SponsorEnquiriesCount(eventID string) (int, error) {
	db, err := a.Database("")
	if err != nil {
		log.Error(err)
		return 0, err
	}

	query := `
	RETURN LENGTH(
		FOR e IN enquiries
		FILTER e.eventId == @eventID AND e.enquiryType == "sponsor"
		RETURN e)`

	bindVars := map[string]interface{}{"eventID": eventID}

	cursor, err := db.Query(a.ctx, query, bindVars)
	if err != nil {
		return 0, err
	}

	defer cursor.Close()

	var count int
	_, err = cursor.ReadDocument(a.ctx, &count)
	if err != nil {
		log.Info(err)
		return 0, err
	}

	return count, nil
}

func (a *ArangoDB) TotalBrochureRequests(eventID string) (int, error) {
	db, err := a.Database("")
	if err != nil {
		log.Error(err)
		return 0, err
	}

	query := `
		RETURN LENGTH(
			FOR b IN brochure_requests
			FILTER b.eventId == @eventID
			RETURN b)`

	bindVars := map[string]interface{}{"eventID": eventID}

	cursor, err := db.Query(a.ctx, query, bindVars)
	if err != nil {
		return 0, err
	}

	defer cursor.Close()

	var count int
	_, err = cursor.ReadDocument(a.ctx, &count)
	if err != nil {
		log.Info(err)
		return 0, err
	}

	return count, nil
}

func (a *ArangoDB) DeactivateOrganizersEvents(organizerId string) (Response, error) {
	db, err := a.Database("")
	if err != nil {
		log.Error(err)
		return Response{}, err
	}

	query := `FOR e IN events
	FILTER e.organizer.id == @id
	update e with {
		deactivated: true
	} in events`

	bindVars := map[string]interface{}{"id": organizerId}

	cursor, err := db.Query(a.ctx, query, bindVars)
	if err != nil {
		return Response{}, err
	}

	cursor.Close()

	return Response{Status:true, Message:"events with given organizer id deactivated."}, nil
}

func (a *ArangoDB) ChangeOrganizersEventsStatus(organizerId string, value bool) ([]string, error) {
	db, err := a.Database("")
	if err != nil {
		log.Error(err)
		return nil, err
	}

	query := `FOR e IN events
	FILTER e.organizer.id == @id
	update e with {
		deactivated: @status
	} in events
	return e.eventDetail.id`

	bindVars := map[string]interface{}{"id": organizerId, "status": value}

	cursor, err := db.Query(a.ctx, query, bindVars)
	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	eventId := make([]string, 0)
	for {
		var e string
		_, err = cursor.ReadDocument(a.ctx, &e)
		if err != nil {
			if arango.IsNoMoreDocuments(err) {
				break
			}
			log.Info(err)
			return nil, err
		}
		eventId = append(eventId, e)
	}

	return eventId, nil
}

func (a *ArangoDB) DeactivateEventByID(eventID string) (Response, error) {
	db, err := a.Database("")
	if err != nil {
		log.Error(err)
		return Response{}, err
	}

	query := `FOR e IN events
	FILTER e.eventDetail.id == @eventID
	update e with {
    	deactivated: true
	} in events
	`

	bindVars := map[string]interface{}{"eventID": eventID}

	cursor, err := db.Query(a.ctx, query, bindVars)
	if err != nil {
		return Response{}, err
	}

	cursor.Close()




	return Response{
		Status: true,
		Message: "event deactivated.",
	}, nil
}

func (a *ArangoDB) ActivateEventByID(eventID string) (Response, error) {
	db, err := a.Database("")
	if err != nil {
		log.Error(err)
		return Response{}, err
	}

	query := `FOR e IN events
	FILTER e.eventDetail.id == @eventID
	update e with {
    	deactivated: false
	} in events
	`

	bindVars := map[string]interface{}{"eventID": eventID}

	cursor, err := db.Query(a.ctx, query, bindVars)
	if err != nil {
		return Response{}, err
	}

	cursor.Close()

	return Response{
		Status: true,
		Message: "event activated.",
	}, nil
}

//Get Event Name by Event ID
func (a *ArangoDB) GetEventNameByEventID(events map[string]int) (map[string]int, error) {
	db, err := a.Database("")
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var IDs []string
	for k, _ := range events {
		IDs = append(IDs, k)
	}

	operatedID := strings.Join(IDs, "','")

	query := `
	for i in ['`+operatedID+`']
	let eName = (
            	for e in events
            	filter e.eventDetail.id == i
            	limit 1
            	return e.eventDetail.name 
        	)[0]
	return {
    	eventId: i,
    	name: eName == null? i: eName
	}`
	cursor, err := db.Query(a.ctx, query, nil)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	var event []Event

	for {
		var e Event
		_, err := cursor.ReadDocument(a.ctx, &e)
		if err != nil {
			if arango.IsNoMoreDocuments(err) {
				break
			}
			log.Info(err)
			return nil, err
		}
		event = append(event, e)
	}

	eventWithPageCount := make(map[string]int,0)

	for _, v := range event {
		value, exists :=  events[v.Id]
		if exists {
			eventWithPageCount[v.Name] = value
		}
	}
	return eventWithPageCount, nil
}

//get the frequency count of items in db..
func frequencyCount(list []string) map[string]int{
	itemFrequency := make(map[string]int)

	for _, item := range list {
		itemFrequency[item] += 1
	}
	return itemFrequency
}