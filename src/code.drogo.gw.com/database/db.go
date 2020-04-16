package database

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"github.com/spf13/viper"
	log "github.com/sirupsen/logrus"
	"fmt"
	"strings"
	"crypto/sha256"
	"strconv"
	"code.drogo.gw.com/util"
	"errors"
	"time"
)

var DB *sql.DB

func init() {
	viper.Set("DB_DATA_SOURCE", "root:root@tcp(159.65.153.232:3306)/eventackle")
	mustOpen()
}

// mustOpen creates a DB object. It panics if it fails!
func mustOpen() {
	var err error
	DB, err = sql.Open("mysql", viper.GetString("DB_DATA_SOURCE"))
	if err != nil {
		panic(err)
	}
	// try ping
	if err := ping(); err != nil {
		panic(err)
	}
	log.Error("PONG")
}

func ping() error {
	return DB.Ping()
}

func OrganizerProfileList(status string, offset int) ([]OrganizerProfile, error) {
	var query string

	if strings.ToLower(status) == "pending" {
		query = "SELECT organizer_profiles.id, organizer_profiles.name, organizer_profiles.website, user_profiles.first_name, organizer_profiles.status " +
			"FROM organizer_profiles " +
			"JOIN user_profiles " +
			"ON user_profiles.user_id = organizer_profiles.user_id " +
			"AND organizer_profiles.status = 'PENDING' " +
			"LIMIT 11 OFFSET ?"
	} else if strings.ToLower(status) == "all" {
		query = "SELECT organizer_profiles.id, organizer_profiles.name, organizer_profiles.website, user_profiles.first_name, organizer_profiles.status " +
			"FROM organizer_profiles " +
			"JOIN user_profiles " +
			"ON user_profiles.user_id = organizer_profiles.user_id " +
			"LIMIT 11 OFFSET ?"
	} else {
		return nil, errors.New("status should have values 'ALL' or 'PENDING' ")
	}

	rows, err := DB.Query(query, offset*10)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer rows.Close()

	var org OrganizerProfile

	organizationProfiles := make([]OrganizerProfile, 0)

	for rows.Next() {
		err := rows.Scan(&org.ID, &org.Name, &org.Website, &org.UserFirstName, &org.Status)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		organizationProfiles = append(organizationProfiles, org)
	}

	return organizationProfiles, rows.Err()
}

func GetApprovedOrganizersList() ([]OrganizerProfile, error) {
	rows, err := DB.Query("SELECT id, name  from organizer_profiles where status = 'APPROVED' ")
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer rows.Close()

	organizer := make([]OrganizerProfile, 0)

	for rows.Next() {
		var o OrganizerProfile
		err := rows.Scan(&o.ID, &o.Name)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		organizer = append(organizer, o)
	}
	err = rows.Err()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return organizer, nil
}

func GetOrganizerByID(id string) (OrganizerData, error) {

	var op OrganizerData
	var description, website, salution, mobile, email, agmtUploadId, imgUploadId, etNotes sql.NullString
	var etCommission sql.NullFloat64

	err := DB.QueryRow("SELECT organizer_profiles.name, organizer_profiles.description, " +
		"organizer_profiles.website, user_profiles.user_id, user_profiles.first_name, user_profiles.last_name, " +
			"user_profiles.salutation, user_profiles.mobile_number, " +
			"users.email, organizer_profiles.is_active, organizer_profiles.agreement_upload_id, " +
			"organizer_profiles.upload_id, organizer_profiles.et_commision_rate, organizer_profiles.et_notes " +
		"FROM organizer_profiles " +
		"INNER JOIN user_profiles " +
			"ON organizer_profiles.user_id = user_profiles.user_id " +
		"INNER JOIN users " +
			"ON organizer_profiles.user_id = users.id " +
		"AND organizer_profiles.id = ?" +
		"", id).Scan(&op.OrganizerProfile.Name, &description, &website, &op.UserProfile.UserId , &op.UserProfile.FirstName, &op.UserProfile.LastName, &salution, &mobile, &email, &op.OrganizerProfile.IsActive, &agmtUploadId, &imgUploadId, &etCommission, &etNotes)
	if err != nil {
		log.Error(err)
		return op, err
	}
	if description.Valid {
		op.OrganizerProfile.Description = description.String
	}
	if website.Valid {
		op.OrganizerProfile.Website = website.String
	}
	if salution.Valid {
		op.UserProfile.Salutation = salution.String
	}
	if mobile.Valid {
		op.UserProfile.MobileNumber = mobile.String
	}
	if email.Valid {
		op.UserProfile.Email = email.String
	}
	if etCommission.Valid {
		op.OrganizerProfile.EtCommission = etCommission.Float64
	} else {
		//set to Zero so that this value will be checked in front-end and will be considered as global.
		op.OrganizerProfile.EtCommission = 0
	}
	if agmtUploadId.Valid {
		op.AgreementId = "https://minio.eventackle.com/uploads/" + agmtUploadId.String
	}
	if imgUploadId.Valid {
		op.UploadId = "https://minio.eventackle.com/uploads/" + imgUploadId.String
	}
	if etNotes.Valid {
		op.OrganizerProfile.EtNotes = etNotes.String
	}

	op.UserProfile.UserName = op.UserProfile.FirstName + " " + op.UserProfile.LastName

	return op, nil
}

func (a *ArangoDB) RejectOrganizerByID(id string, reason string, desc string) (Response, error) {
	errSql := make(chan error, 1)
	errArango := make(chan error, 1)

	go func() {
		if strings.ToLower(reason) != "other" && desc == "" {
			stmt, err := DB.Prepare("UPDATE organizer_profiles " +
				"SET status = 'REJECTED', rejection_reason = ? " +
				"WHERE id = ? ")
			if err != nil {
				log.Error(err)
				errSql <- err
				close(errSql)
				return
			}
			defer stmt.Close()

			_, err = stmt.Exec(reason, id)
			if err != nil {
				log.Error(err)
				errSql <- err
				close(errSql)
				return
			}

			errSql	<- nil
			close(errSql)

		} else if strings.ToLower(reason) == "other" && desc != "" {
			stmt, err := DB.Prepare("UPDATE organizer_profiles " +
				"SET status = 'REJECTED', rejection_reason = 'Other', rejection_description = ? " +
				"WHERE id = ?")
			if err != nil {
				log.Error(err)
				errSql <- err
				close(errSql)
				return
			}
			defer stmt.Close()
			_, err = stmt.Exec(desc, id)
			if err != nil {
				log.Error(err)
				errSql <- err
				close(errSql)
				return
			}

			errSql	<- nil
			close(errSql)

		} else {
			errSql <- errors.New("reason / description cannot be empty")
			close(errSql)
			return
		}
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
    		organizer: {status: "REJECTED"}
		} in events`

		bindVars := map[string]interface{}{"orgId": id}

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
		Message: "Organizer rejected.",
	}, nil
}

func ChangeOrganizerStatus(organizerId string, status int) (Response, error){
	stmt, err := DB.Prepare("UPDATE organizer_profiles SET is_active = ? " +
		"WHERE id = ?")
	if err != nil {
		return Response{}, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(status, organizerId)
	if err != nil {
		return Response{}, err
	}

	return Response{Status:true, Message:"Organizer deactivated."}, nil
}

func UpdateOrganizer(org OrganizerData) (Response, error) {
	var query string
	var agreementName string
	var err error

	if org.OrganizerProfile.EtCommission >= 100 {
		return Response{Status: false, Message: "Et Commission cannot be greater or equal to 100"}, nil
	}
	//Agreement file name retrieved from minio that will be stored in db .
	if org.OrganizerProfile.AgreementUploadId != "" {
		query = "UPDATE organizer_profiles " +
			"SET description = ?, website = ?, agreement_upload_id = ?, " +
			"et_commision_rate = ?, et_notes = ? " +
			"WHERE id = ? "
		agreementName, err = util.StorePDF(org.OrganizerProfile.AgreementUploadId)
		if err != nil {
			agreementName = ""
			log.Println(err)
		}

		stmt, err := DB.Prepare(query)
		if err != nil {
			log.Error(err)
			return Response{}, err
		}
		defer stmt.Close()

		_, err = stmt.Exec(org.OrganizerProfile.Description, org.OrganizerProfile.Website, agreementName, org.OrganizerProfile.EtCommission, org.OrganizerProfile.EtNotes, org.OrganizerProfile.ID)
		if err != nil {
			log.Error(err)
			return Response{}, err
		}
	} else {
		query = "UPDATE organizer_profiles " +
			"SET description = ?, website = ?, et_commision_rate = ?, et_notes = ? " +
			"WHERE id = ? "
		stmt, err := DB.Prepare(query)
		if err != nil {
			log.Error(err)
			return Response{}, err
		}
		defer stmt.Close()

		_, err = stmt.Exec(org.OrganizerProfile.Description, org.OrganizerProfile.Website, org.OrganizerProfile.EtCommission, org.OrganizerProfile.EtNotes, org.OrganizerProfile.ID)
		if err != nil {
			log.Error(err)
			return Response{}, err
		}
	}

	stmt2, err := DB.Prepare("UPDATE user_profiles " +
		"SET first_name = ?, last_name = ?, salutation = ?, mobile_number = ? " +
		"WHERE user_id = ? ")
	if err != nil {
		log.Error(err)
		return Response{}, err
	}
	defer stmt2.Close()

	_, err = stmt2.Exec(org.UserProfile.FirstName, org.UserProfile.LastName, org.UserProfile.Salutation, org.UserProfile.MobileNumber, org.UserProfile.UserId)
	if err != nil {
		log.Error(err)
		return Response{}, err
	}

	stmt3, err := DB.Prepare("UPDATE users SET email = ? " +
		"WHERE id = ? ")
	if err != nil {
		log.Error(err)
		return Response{}, err
	}
	defer stmt3.Close()

	_, err = stmt3.Exec(org.UserProfile.Email, org.UserProfile.UserId)
	if err != nil {
		log.Error(err)
		return Response{}, err
	}

	return Response{
		Status: true,
		Message: "Organizer updated successfully.",
	}, nil
}

func GetUserList(offset int) ([]User, error) {
	rows, err := DB.Query("SELECT user_profiles.first_name, user_profiles.last_name, user_profiles.blog_url, organizer_profiles.name, user_profiles.address " +
		"FROM user_profiles " +
		"LEFT JOIN organizer_profiles " +
		"ON user_profiles.user_id = organizer_profiles.user_id " +
		"LIMIT 11 OFFSET ?", offset*10)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer rows.Close()
	var user User

	userList := make([]User, 0)

	var email, organization sql.NullString
	for rows.Next() {
		err := rows.Scan(&user.FirstName, &user.LastName, &email, &organization, &user.Location)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		if email.Valid {
			user.Email = email.String
		}
		user.UserName = user.FirstName + " " + user.LastName

		if organization.Valid {
			user.Organization = organization.String
			user.IsOrganizer = true
		} else {
			user.Organization = "-"
			user.IsOrganizer = false
		}
		userList = append(userList, user)
	}

	return userList, rows.Err()
}

func GetEventTypeByID(id int) (EventType, error) {
	var eventType EventType

	err := DB.QueryRow("SELECT name, description from lookup_event_types where id = ? ", id).Scan(&eventType.Name, &eventType.Description)

	return eventType, err
}

func CreateNewEventType(eventTypeName string, desc string) (Response, error){
	if eventTypeName == " " {
		return Response{}, errors.New("name cannot be empty")
	}

	var result string

	err := DB.QueryRow("SELECT name from lookup_event_types WHERE name = ?", eventTypeName).Scan(&result)
	if err == nil && result != "" {
		return Response{Status:false, Message:"Name already exists"}, errors.New("name already exists")
	}

	stmt, err := DB.Prepare("INSERT INTO lookup_event_types (name, created_at, description) VALUES (?, ?, ?)")
	if err != nil {
		return Response{}, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(eventTypeName, time.Now(), desc)

	if err != nil {
		log.Error(err)
		return Response{}, err
	}

	return Response{
		Status: true,
		Message: "Event type was created successfully.",
	}, nil
}

func UpdateEventType (id int, eventTypeName, desc string) (Response, error) {
	if eventTypeName == " " {
		return Response{}, errors.New("name cannot be empty")
	}

	var result string

	err := DB.QueryRow("SELECT name from lookup_event_types WHERE name = ? AND id != ?", eventTypeName, id).Scan(&result)
	if err == nil && result != "" {
		return Response{Status:false, Message:"Name already exists"}, errors.New("name already exists")
	}

	stmt, err := DB.Prepare("UPDATE lookup_event_types SET name = ?, description = ?, updated_at = ? WHERE id = ?")
	if err != nil {
		log.Error(err)
		return Response{}, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(eventTypeName, desc, time.Now(), id)

	if err != nil {
		log.Error(err)
		return Response{}, err
	}

	return Response{
		Status: true,
		Message: "Updated successfully",
	}, nil
}

// fetches job / position by id so that it can be updated....
func GetPositionByID(id int) (Position, error) {
	var position Position

		err := DB.QueryRow("SELECT name, description from lookup_attendees WHERE id = ?", id).Scan(&position.Name, &position.Description)

	return position, err
}

// creates new job / position in db...
func CreateNewJobPosition(name string, desc string) (Response, error) {
	if name == "" {
		return Response{}, errors.New("name cannot be empty")
	}

	var result string

	err := DB.QueryRow("SELECT name from lookup_attendees where name = ?",name).Scan(&result)
	if err == nil && result != "" {
		return Response{}, errors.New("name already exists")
	}

	stmt, err := DB.Prepare("INSERT INTO lookup_attendees (name, created_at, description) VALUES (?, ?, ?)")
	if err != nil {
		log.Error(err)
		return Response{}, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, time.Now(), desc)
	if err != nil {
		log.Error(err)
		return Response{}, err
	}

	return Response{
		Status: true,
		Message: "Position created successfully.",
	}, nil
}

// updates job / position in db..
func UpdateJobPositionByID(id int, name string, desc string) (Response, error) {
	if name == "" {
		return Response{}, errors.New("name cannot be empty")
	}

	var result string

	err := DB.QueryRow("SELECT name from lookup_attendees WHERE name = ? AND id != ?", name, id).Scan(&result)
	if err == nil && result != "" {
		return Response{Status:false, Message:"Name already exists"}, errors.New("name already exists")
	}

	stmt, err := DB.Prepare("UPDATE lookup_attendees SET name = ?, updated_at = ?, description = ? WHERE id = ?")
	if err != nil {
		log.Error(err)
		return Response{}, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, time.Now(), desc, id)
	if err != nil {
		log.Error(err)
		return Response{}, err
	}

	return Response{
		Status: true,
		Message: "Updated successfully",
	}, nil
}

// creates new area of interest..
func CreateAreaOfInterest(name string, desc string) (Response, error) {
	if name == "" || desc == "" {
		return Response{}, errors.New("name / description cannot be empty")
	}

	var result string

	err := DB.QueryRow("SELECT name from lookup_interests where name = ?",name).Scan(&result)
	if err == nil && result != "" {
		return Response{}, errors.New("name already exists")
	}

	stmt, err := DB.Prepare("INSERT INTO lookup_interests (name, created_at, description) VALUES (?, ?, ?)")
	if err != nil {
		return Response{}, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, time.Now(), desc)
	if err != nil {
		log.Error(err)
		return Response{}, err
	}

	return Response{
		Status: true,
		Message: "Interest created successfully.",
	}, nil
}

// updates area of interest using its id....
func UpdateAreaOfInterest(id int, name string, desc string) (Response, error) {
	if name == "" || desc == "" {
		return Response{}, errors.New("name / description cannot be empty")
	}

	var result string

	err := DB.QueryRow("SELECT name from lookup_interests WHERE name = ? AND id != ?", name, id).Scan(&result)
	if err == nil && result != "" {
		return Response{Status:false, Message:"Name already exists"}, errors.New("name already exists")
	}

	stmt, err := DB.Prepare("UPDATE lookup_interests SET name = ?, updated_at = ?, description = ? WHERE id = ?")
	if err != nil {
		return Response{}, err

	}
	defer stmt.Close()

	_, err = stmt.Exec(name, time.Now(), desc, id)
	if err != nil {
		log.Error(err)
		return Response{}, err
	}

	return Response{
		Status: true,
		Message: "Interest updated successfully.",
	}, nil
}

// fetches area of interest by id so that it can be updated...
func GetInterestByID(id int) (Interest, error) {
	var interest Interest

	err := DB.QueryRow("SELECT name, description from lookup_interests where id = ?", id).Scan(&interest.Name, &interest.Description)
	if err != nil {
		log.Error(err)
		return Interest{}, err
	}
	return interest, nil
}

// creates new region in db...
func CreateNewRegion(name string) (Response, error) {
	if name == "" {
		return Response{}, errors.New("name cannot be empty")
	}

	var result string

	err := DB.QueryRow("SELECT name from lookup_locations where name = ?", name).Scan(&result)
	if err == nil && result != "" {
		return Response{}, errors.New("region already exists")
	}

	stmt, err := DB.Prepare("INSERT INTO lookup_locations (name, created_at) VALUES (?, ?)")
	if err != nil {
		log.Error(err)
		return Response{}, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, time.Now())
	if err != nil {
		log.Error(err)
		return Response{}, err
	}
	return Response{
		Status: true,
		Message: "region created successfully",
	}, nil
}

// fetches the region by id so that it can be updated..
func GetRegionByID(id int) (Region, error) {
	var region Region

	err := DB.QueryRow("SELECT name from lookup_locations where id = ?", id).Scan(&region.Name)
	if err != nil {
		log.Error(err)
		return Region{}, err
	}
	return region, nil
}

// updates the region using its id..
func UpdateRegionByID(id int, name string) (Response, error) {
	if name == ""{
		return Response{}, errors.New("name cannot be empty")
	}

	var result string

	err := DB.QueryRow("SELECT name from lookup_locations where name = ? AND id != ?", name, id).Scan(&result)
	if err == nil  && result != "" {
		return Response{}, errors.New("region already added")
	}

	stmt, err := DB.Prepare("UPDATE lookup_locations SET name = ?, updated_at = ? where id= ?")
	if err != nil {
		log.Error(err)
		return Response{}, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, time.Now(), id)
	if err != nil {
		log.Error(err)
		return Response{}, err
	}

	return Response{
		Status: true,
		Message: "Updated successfully",
	}, nil
}

func ValidateUser(email, password string) (Response,error) {

	if email == "" || password == "" {
		return Response{
			Status:  false,
			Message: "email / password cannot be empty",
		}, nil
	}

	var dbEmail, dbPassword string

	err := DB.QueryRow("SELECT email, password FROM users WHERE email=? && role = 'ADMIN' && status = true", email).Scan(&dbEmail, &dbPassword)
	if err != nil {
		return Response{
			Status:false,
			Message:"authentication failed",}, err
	}
	hashedPassword := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
	if hashedPassword == dbPassword {
		return Response{Status:true, Message: "login success",}, nil
	}

	return Response{Status:false, Message:"incorrect password"}, nil
}

func FetchCategoryById(Id string) (Category, error) {

	var category Category
	row := DB.QueryRow("SELECT id, name, description, upload_id, image_size FROM lookup_categories WHERE id=?", Id)

	var name, uploadId, description, imageSize sql.NullString
	var id int

	err := row.Scan(&id, &name, &description, &uploadId, &imageSize)
	if err != nil {
		log.Error(err)
		return category, err
	}

	var imageData string

	if uploadId.Valid {
		imageData = "https://minio.eventackle.com/uploads/" + uploadId.String
	}

	category = Category{
		ID:          strconv.Itoa(id),
		Name:        name.String,
		Description: description.String,
		ImageSize:   imageSize.String,
		ImageData:   imageData,
	}

	return category, nil
}

func AddCategory(category Category) Response {
	if category.Name == "" || category.Description == "" {
		response := Response {
			Status:  false,
			Message: "name or description field empty",
		}
		return response
	}

	row := DB.QueryRow("SELECT name FROM lookup_categories WHERE name = ?", category.Name)
	var Name string
	row.Scan(&Name)
	if Name != "" {
		response := Response{
			Status:  false,
			Message: "category of given name already exists",
		}
		return response
	}

	var uploadId string
	uploadId, err := util.NewMinioClient().StoreImage(category.Name, category.ImageData)
	if err != nil {
		response := Response {
			Status:  false,
			Message: fmt.Sprintln(err),
		}
		return response
	}

	stmt, err := DB.Prepare("INSERT lookup_categories SET name=?, created_at=?, updated_at=?, description=?, upload_id=?, image_size=?")
	if err != nil {
		response := Response{
			Status:  false,
			Message: fmt.Sprintln(err),
		}
		return response
	}
	res, err := stmt.Exec(category.Name, time.Now(), time.Now(), category.Description, uploadId, category.ImageSize)
	if err != nil {
		response := Response{
			Status:  false,
			Message: fmt.Sprintln(err),
		}
		return response
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		response := Response{
			Status:  false,
			Message: fmt.Sprintln(err),
		}
		return response
	}

	response := Response{
		Status:  true,
		Message: fmt.Sprintf("%v row(s) created", rowsAffected),
	}
	return response
}

func UpdateCategory(category Category, id string) Response {
	if category.Name == "" || category.Description == "" {
		response := Response{
			Status:  false,
			Message: "name or description field empty",
		}
		return response
	}

	row := DB.QueryRow("SELECT name FROM lookup_categories WHERE id != ? AND name = ? ", id, category.Name)
	var Name sql.NullString
	row.Scan(&Name.String)
	if Name.Valid {
		response := Response{
			Status:  false,
			Message: "category of given name already exists",
		}
		return response
	}


	// Update category but keep the uploadId same if we have not got the new image
	if category.ImageData == "" {
		stmt, err := DB.Prepare("UPDATE lookup_categories SET name=?, description=?, updated_at=? WHERE id=?")
		if err != nil {
			return Response{
				Status:  false,
				Message: fmt.Sprintln(err),
			}
		}
		res, err := stmt.Exec(category.Name, category.Description, time.Now(), id)
		if err != nil {
			return Response{
				Status:  false,
				Message: fmt.Sprintln(err),
			}
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return Response{
				Status:  false,
				Message: fmt.Sprintln(err),
			}
		}
		return Response{
			Status:  true,
			Message: fmt.Sprintf("%v row(s) updated", rowsAffected),
		}
	}


	// Store the new image if we have got one
	var uploadId string
	uploadId, err := util.NewMinioClient().StoreImage(category.Name, category.ImageData)
	if err != nil {
		return Response {
			Status:  false,
			Message: fmt.Sprintln(err),
		}
	}

	// Fetch the name of previous image for later deleting that from db
	row = DB.QueryRow("SELECT upload_id FROM lookup_categories WHERE id=?", id)
	var previousUploadId sql.NullString
	err = row.Scan(&previousUploadId)
	if err != nil {
		log.Error(err)
		return Response {
			Status:  false,
			Message: fmt.Sprintln(err),
		}
	}

	// Update a particular category along with its uploadId
	stmt, err := DB.Prepare("UPDATE lookup_categories SET name=?, description=?, updated_at=?, upload_id=?, image_size=? WHERE id=?")
	if err != nil {
		return Response{
			Status:  false,
			Message: fmt.Sprintln(err),
		}
	}
	res, err := stmt.Exec(category.Name, category.Description, time.Now(), uploadId, category.ImageSize, id)
	if err != nil {
		return Response {
			Status:  false,
			Message: fmt.Sprintln(err),
		}
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return Response {
			Status:  false,
			Message: fmt.Sprintln(err),
		}
	}

	// Delete the previous image from minio if present
	if previousUploadId.Valid {
		err = util.NewMinioClient().DeleteImage(previousUploadId.String)
		if err != nil {
			return Response {
				Status:  false,
				Message: fmt.Sprintf("%v", err),
			}
		}
	}

	return Response {
		Status:  true,
		Message: fmt.Sprintf("%v row(s) updated", rowsAffected),
	}
}

//Used in home page to get user profiles count..
func GetUserProfilesCount() (int, error) {
	var val string

	err := DB.QueryRow("SELECT COUNT(id) from user_profiles").Scan(&val)
	if err != nil {
		log.Error(err)
		return 0, err
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		log.Error(err)
		return 0, err
	}

	return count, nil
}

//Used in home page to get organizer profiles count..
func GetOrganizerProfilesCount() (int, error) {
	var val string

	err := DB.QueryRow("SELECT COUNT(id) from organizer_profiles").Scan(&val)
	if err != nil {
		log.Error(err)
		return 0, err
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		log.Error(err)
		return 0, err
	}

	return count, nil
}

//Get the value of et commission from global settings..
func GetGlobalEtCommissionRate() (float64, error) {
	var etRate float64

	err := DB.QueryRow("SELECT et_commission_rate from global_settings").Scan(&etRate)
	if err != nil {
		log.Info(err)
		return 0, err
	}

	return etRate, nil
}

func UpdateGlobalEtCommissionRate(etCommission float64) (Response, error) {
	if etCommission >= 100 {
		return Response{Status:false, Message:"Et Commission cannot be greater than or equal to 100"}, nil
	}
	stmt, err := DB.Prepare("UPDATE global_settings SET et_commission_rate = ? ")
	if err != nil {
		log.Info(err)
		return Response{}, err
	}

	defer stmt.Close()

	_, err = stmt.Exec(etCommission)
	if err != nil {
		log.Info(err)
		return Response{}, err
	}

	return Response{Status: true, Message:"et commission updated in global settings"}, nil
}