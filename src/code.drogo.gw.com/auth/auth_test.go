package auth

import (
	"testing"
	"fmt"
)

func TestGenerateToken(t *testing.T) {
	res, err := GenerateToken()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(res)
}
