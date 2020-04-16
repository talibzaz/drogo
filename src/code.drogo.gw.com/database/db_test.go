package database

import (
	"testing"
	"encoding/json"
	"os"
	"fmt"
	"github.com/spf13/viper"
	"context"
)

func TestValidateUser(t *testing.T) {
	var testTable = []struct {
		email string
		password string
		expected bool
	}{
		{"Admin@eventackle.com", "qwertyu", true},
		{"ADMIN@EVENTACKLE.COM", "qwertyu", true},
		{"admin@eventackle.com", "Qwertyu", false},
		{"", "qwertyu", false},
		{"admin@eventackle.com", "", false},
		{"", "", false},
		{"salfi@mail.com", "qwertyu", false},

	}
	for _, test := range testTable {
		result , _:= ValidateUser(test.email, test.password)
		if result.Status != test.expected {
			t.Errorf("Expected result for %s and %s is %t, but got %t",test.email, test.password, test.expected, result)
		}
	}
}

func TestFetchCategoryById(t *testing.T) {
	cat, err := FetchCategoryById("1")
	if err != nil {
		t.Log(err)
	}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent(" ", " ")
	e.Encode(&cat)
}

func TestAddCategory(t *testing.T) {
	var tesTable = []struct{
		category Category
		respStatus bool
	}{
		{Category{Name:"", Description:"", ImageSize:"", ImageData:"",},false},
		{Category{Name:"Some Other Name", Description:"some description", ImageSize:"", ImageData:""}, true},
		{Category{Name:"Some Other Name", Description:"some description", ImageSize:"", ImageData:""}, false},
	}
	for _, test := range tesTable{
		res := AddCategory(test.category)
		if res.Status != test.respStatus {
			t.Errorf("expected %s but got %s", test.respStatus, res.Status)
		}
	}
}

func TestUpdateCategory(t *testing.T) {
	var tesTable = []struct{
		category Category
		respStatus bool
	}{
		{Category{Name:"Mining", Description:"some other description",ImageSize:"2*2", ImageData:""}, true},
	}
	for _, test := range tesTable{
		res := UpdateCategory(test.category, "13")
		if res.Status != test.respStatus {
			t.Errorf("expected %s but got %s", test.respStatus, res.Status)
		}
	}
}

func TestGetUserList(t *testing.T) {
	res, err := GetUserList(0)
	if err != nil {
		t.Fatal(err)
	}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent(" ", " ")
	e.Encode(&res)
}

func TestGetUserProfilesCount(t *testing.T) {
	res, err := GetUserProfilesCount()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(res)
}

func TestGetOrganizerProfilesCount(t *testing.T) {
	res, err := GetOrganizerProfilesCount()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res)
}

func TestGetOrganizerByID(t *testing.T) {
	res, err := GetOrganizerByID("533e9b32-d7f5-4ed4-be72-914c657eb269")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(res)
}

func TestArangoDB_RejectOrganizerByID(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	arango, err := NewArangoDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	res, err := arango.RejectOrganizerByID("479cc1fe-4dda-419b-8f69-972e68a35b78", "random", "")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(res)
}