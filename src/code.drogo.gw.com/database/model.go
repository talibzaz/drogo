package database

type Response struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Token string   `json:"token"`
}

type Category struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	NoOfEvents  int     `json:"noOfEvents"`
	ImageData   string  `json:"imageData"`
	ImageSize	string	`json:"imageSize"`
}

type Attendee struct {
	Attendee		[]string	`json:"attendee"`
	EventName		string		`json:"eventName"`
	OrganizerName	string		`json:"organizerName"`
}

type Event struct {
	Id				string	`json:"eventId"`
	Name 	  		string	`json:"name"`
	StartDate	 	string 	`json:"startDate"`
	StartTime 		string	`json:"startTime"`
	EndDate			string  `json:"endDate"`
	EndTime			string  `json: "endTime"`
	Description		string	`json:"description"`
	Location		string	`json:"location"`
	OrganizerName	string	`json:"organizerName"`
	OrganizerId 	string	`json:"organizerId"`
	UserName 		string	`json:"userName"`
	Status    		string	`json:"status"`
	PublishDate 	string	`json:"publishDate"`
	Types			[]string`json:"types"`
	Categories		[]string`json:"categories"`
	SoldAmount  	float64	`json:"soldAmount"`
	TicketsSold 	int		`json:"ticketsSold"`
	ImageData 		string	`json:"imageData"`
	IsFeatured		bool	`json:"isFeatured"`
	OrganizerID 	string 	`json:"organizerID"`
	VenueCity 		string 	`json:"venueCity"`
	Revenue			float64	`json:"revenue"`
	TotalRevenue    string
	TicketLiveDays  int		`json:"ticketLiveDays"`
	PageHits		int
	Deactivated     bool	`json:"deactivated"`
	Currency        string  `json:"currency"`
}

type Ticket struct {
	OrderNo		string  `json:"orderNumber"`
	PurchasedBy string	`json:"purchasedBy"`
	NoOfTickets	int		`json:"noOfTickets"`
	EventId 	string	`json:"eventId"`
	EventName	string	`json:"eventName"`
}

type OrganizerData struct {
	OrganizerProfile  OrganizerProfile `json:"organizer_profile"`
	UserProfile       User             `json:"user_profile"`
	AgreementId 	  string 		   `json:"agreement_upload_id"`
	UploadId		  string		   `json:"upload_id"`
}

type OrganizerProfile struct {
	ID              	string 		`json:"id"`
	Name            	string 		`json:"name"`
	UserFirstName 		string 		`json:"user_first_name"`
	Description     	string 		`json:"description"`
	Website         	string 		`json:"website"`
	Status          	string 		`json:"status"`
	BankId         		string 		`json:"bank_id"`
	BillingId      		string 		`json:"billing_id"`
	UploadId       		string 		`json:"upload_id"`
	UserId         		string 		`json:"user_id"`
	AgreementUploadId   string 		`json:"agreement_upload_id"`
	IsActive 			int    		`json:"is_active"`
	EtEarnings			float64	   	`json:"etEarnings"`
	Revenue				float64    	`json:"revenue"`
	TotalEtEarning      string
	RevenueGenerated    string
	EtCommission		float64		`json:"et_commision_rate"`
	EventsCount			int			`json:"eventsCount"`
	EtNotes				string		`json:"et_notes"`
}

//type Report struct {
//	OrganizerId		string
//	NoOfEvents		int
//	EtEarings		int
//	Revenue			int
//}

type User struct {
	UserId        string  `json:"user_id"`
	UserName      string  `json:"user_name"`
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	Email         string  `json:"blog_url"`
	Organization  string `json:"organization"`
	Location      *string `json:"location"`
	IsOrganizer   bool    `json:"isOrganizer"`
	Salutation    string  `json:"salutation"`
	MobileNumber  string  `json:"mobile_number"`
	ImageUploadId *string `json:"image_upload_id"`
}

type Position struct {
	Id 			 int	 `json:"id"`
	Name         string  `json:"name"`
	NoOfEvents 	 int	 `json:"no_of_events"`
	Description  *string `json:"description"`
}



type EventType struct {
	Id           int	 `json:"id"`
	Name         string  `json:"name"`
	Description  *string `json:"description"`
	NoOfEvents 	 int 	 `json:"no_of_events"`
}

type Region struct {
	Id           int	 `json:"id"`
	Name         string  `json:"name"`
	NoOfEvents 	 int 	 `json:"no_of_events"`
}

type Interest struct {
	Id				int      `json:"id"`
	Name 	  		string   `json:"name"`
	NoOfEvents    	int		 `json:"no_of_events"`
	Description 	*string   `json:"description"`
}

type Payout struct {
	Id 				string  	`json:"id"`
	Name			string		`json:"name"`
	EndDate 		string		`json:"endDate"`
	Sold 			int			`json:"sold"`
	SaleAmount  	float64		`json:"saleAmount"`
	TotalSaleAmount string
	TaxAmount       float64 	`json:"taxAmount"`
	TotalTaxAmount  string
	EtShare 		float64		`json:"etShare"`
	TotalEtShare    string
	PayoutAmount 	float64		`json:"payoutAmount"`
	TotalPayoutAmount string
	Status			string		`json:"status"`
}

type EventsOverview struct {
	EventsHeld int	`json:"eventsHeld"`
	TicketsSold int	`json:"ticketsSold"`
	Revenue float64	`json:"revenue"`
	TotalRevenue string
}