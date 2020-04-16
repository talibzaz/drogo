package analytics

import (
	"testing"
	"fmt"
)

func TestPageViews(t *testing.T) {
	res, err := EventViews("month", "eventPage")
	if err != nil {
		fmt.Println("err", err)
		t.Fatal(err)
	}

	fmt.Println(res)
}

func TestGetPageViewByEventID(t *testing.T) {
	res, err := GetPageViewsByEventID("beq5beido6kg00dj7lmg")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(res)
}