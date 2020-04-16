package database

import (
	"testing"
	"encoding/json"
	"os"
	"context"
	"github.com/spf13/viper"
	"fmt"
)

func TestArangoDB_FetchEvents(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	offset := 1
	arangoDb, err := NewArangoDB(context.Background())
	events, err := arangoDb.FetchEvents(offset)
	if err != nil {
		t.Logf("err: %v", err)
	}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent(" ", " ")
	e.Encode(&events)
}

func TestArangoDB_TotalRevenueGenerated(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	arango, err := NewArangoDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	res, _ := arango.TotalRevenueGenerated()

	fmt.Println(res)
}

func TestArangoDB_ApproveOrganizer(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	arango, err := NewArangoDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	res, err := arango.ApproveOrganizer("05e3e92f-ef8b-4891-9f7e-b5be8c5d2e74")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(res)
}

func TestArangoDB_GetTicketList(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	offset := 0
	arangoDb, err := NewArangoDB(context.Background())
	tickets, err:=	 arangoDb.GetTicketList(offset)
	if err != nil {
		t.Logf("err: %v", err)
	}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent(" ", " ")
	e.Encode(&tickets)
}

func TestArangoDB_ListEventCategory(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")
	arango, err := NewArangoDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	categories, err := arango.ListEventCategory(0)
	if err != nil {
		t.Fatal(err)
	}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent(" ", " ")
	e.Encode(categories)
}

func TestArangoDB_ListEventTypes(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	arango, err := NewArangoDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	eventTypes, err := arango.ListEventTypes(0)
	if err != nil {
		t.Fatal(err)
	}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent(" ", " ")
	e.Encode(eventTypes)
}

func TestArangoDB_ListRegions(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	arango, err := NewArangoDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	region, err := arango.ListRegions(0)
	if err != nil {
		t.Fatal(err)
	}
	e:= json.NewEncoder(os.Stdout)
	e.SetIndent(" ", " ")
	e.Encode(region)
}

func TestArangoDB_GetTotalEventsCreated(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	arango, err := NewArangoDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	res, err := arango.GetTotalEventsCreated()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res)
}

func TestArangoDB_GetRecentEvents(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	arango, err := NewArangoDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	res, err := arango.GetRecentEvents()
	if err != nil {
		t.Fatal(err)
	}

	e:= json.NewEncoder(os.Stdout)
	e.SetIndent(" ", " ")
	e.Encode(res)
}

func TestArangoDB_GetTOpEvents(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	arango, err := NewArangoDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	res, err := arango.GetTopEvents("week")
	if err != nil {
		t.Fatal(err)
	}

	e:= json.NewEncoder(os.Stdout)
	e.SetIndent(" ", " ")
	e.Encode(res)
}

func TestArangoDB_GetEventNameByEventID(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	arango, err := NewArangoDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	mp := map[string]int {"bem7jmrohqi0009mpv90":29, "beksovd0p6eg00f1nu7g":19, "beoqdroa6f0g009d1be0":19, "belgsed0p6eg00f1nua0":18, "beog498a6f0g009d1bdg":45}

	res, err := arango.GetEventNameByEventID(mp)
	if err != nil {
		t.Fatal(err)
	}

	e:= json.NewEncoder(os.Stdout)
	e.SetIndent(" ", " ")
	e.Encode(res)
}

func TestArangoDB_UpcomingPayouts(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	arango, err := NewArangoDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	res, err := arango.UpcomingPayouts(0, 50)
	if err != nil {
		t.Fatal(err)
	}

	e:= json.NewEncoder(os.Stdout)
	e.SetIndent(" ", " ")
	e.Encode(res)
}

func TestArangoDB_PayoutDetails(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	arango, err := NewArangoDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	res, err := arango.PayoutDetails(0)
	if err != nil {
		t.Fatal(err)
	}

	e:= json.NewEncoder(os.Stdout)
	e.SetIndent(" ", " ")
	e.Encode(res)
}

func TestArangoDB_EventReportsOverview(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	arango, err := NewArangoDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	res, err := arango.EventReportsOverview()
	if err != nil {
		t.Fatal(err)
	}

	e:= json.NewEncoder(os.Stdout)
	e.SetIndent(" ", " ")
	e.Encode(res)

}
func TestArangoDB_EventSpecificReport(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	arango, err := NewArangoDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	res, err := arango.EventSpecificReport("bdlup2gaqtk000di2s60")
	if err != nil {
		t.Fatal(err)
	}

	e:= json.NewEncoder(os.Stdout)
	e.SetIndent(" ", " ")
	e.Encode(res)
}

func TestArangoDB_UpdateEventEtCommissionRate(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	arango, err := NewArangoDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	res, err := arango.UpdateEventEtCommissionRate("qwertyui", -1)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(res)
}

//func TestArangoDB_ListOrganizerReport(t *testing.T) {
//	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
//	viper.Set("ARANGO_DB", "eventackle")
//	viper.Set("ARANGO_USERNAME", "root")
//	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")
//
//	arango, err := NewArangoDB(context.Background())
//	if err != nil {
//		t.Fatal(err)
//	}
//	_, err = arango.ListOrganizerReport(0)
//	if err != nil {
//		t.Fatal(err)
//	}
	//
	//e:= json.NewEncoder(os.Stdout)
	//e.SetIndent(" ", " ")
	//e.Encode(res)
//}

//func TestArangoDB_GetEventList(t *testing.T) {
//	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
//	viper.Set("ARANGO_DB", "eventackle")
//	viper.Set("ARANGO_USERNAME", "root")
//	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")
//
//	arango, err := NewArangoDB(context.Background())
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	res, err := arango.GetEventList("all")
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	e:= json.NewEncoder(os.Stdout)
//	e.SetIndent(" ", " ")
//	e.Encode(res)
//}

func TestArangoDB_MarkAsPaidByEventID(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	arango, err := NewArangoDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	res,err := arango.MarkAsPaidByEventID("beog498a6f0g009d1bdg")
	if err != nil {
		t.Fatal(err)
	}

	e:= json.NewEncoder(os.Stdout)
	e.SetIndent(" ", " ")
	e.Encode(res)
}

func TestArangoDB_GetEventAttendees(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	arango, err := NewArangoDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	res, err := arango.GetEventAttendees("bdqk4f4mandg00a9ugk0", "00447c98-c3c0-4ed1-8738-7da167230b6d", 0)
	if err != nil {
		t.Fatal(err)
	}

	e:= json.NewEncoder(os.Stdout)
	e.SetIndent(" ", " ")
	e.Encode(res)
}

func TestArangoDB_GetTotalEtEarnings(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	arango, err := NewArangoDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	res, err := arango.GetTotalEtEarnings()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(res)
}

func TestArangoDB_ExhibitorEnquiriesCount(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	arango, err := NewArangoDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	res, _ := arango.ExhibitorEnquiriesCount("be56v7ijbmo0009qnd4g")

	fmt.Println(res)
}

func TestArangoDB_TotalBrochureRequests(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	arango, err := NewArangoDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	res, _ := arango.TotalBrochureRequests("bdqk4f4mandg00a9ugk0")

	fmt.Println(res)
}

func TestArangoDB_ListOrganizersReport(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	arango, err := NewArangoDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	res, err := arango.OrganizersReport(0)
	if err != nil {
		t.Fatal(err)
	}

	e := json.NewEncoder(os.Stdout)
	e.SetIndent(" ", " ")
	e.Encode(res)
}

func TestChangeOrganizerStatus(t *testing.T) {
	viper.Set("ARANGO_HOST", "http://139.59.85.55:8529")
	viper.Set("ARANGO_DB", "eventackle")
	viper.Set("ARANGO_USERNAME", "root")
	viper.Set("ARANGO_PASSWORD", "qF3mKQcu7zyzBYly")

	arango, err := NewArangoDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	res, err := arango.ChangeOrganizersEventsStatus("08d0604c-d790-48ae-8c35-3d021202072d", true)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(res)
}