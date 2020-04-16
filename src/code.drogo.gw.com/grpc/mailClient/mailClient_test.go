package mailClient

import (
	"testing"
	"fmt"
)

func TestApproveOrganizerMail(t *testing.T) {
	res, err := ApproveOrganizerMail("Talib", "talb_m4@yahoo.co.in")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(res)
}
