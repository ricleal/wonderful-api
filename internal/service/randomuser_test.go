//go:build example
// +build example

package service_test

import (
	"wonderful/internal/service"
	"context"
	"fmt"
	"net/http"
)

func ExampleRandomUser() {
	ctx := context.Background()
	c := http.Client{}
	r, err := service.FetchRandomUsers(ctx, c)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(r.Results))
	// Output:
	// 5000
}
