package main

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/util"
)

func main() {
	key := "account:c42f8679-add8-4379-8166-28215bfef8f6"
	pat := "account:*"

	fmt.Printf("keyMatch(%q, %q)  = %v\n", key, pat, util.KeyMatch(key, pat))
	fmt.Printf("keyMatch2(%q, %q) = %v\n", key, pat, util.KeyMatch2(key, pat))

	// Full model test with in-memory policies:
	e, err := casbin.NewEnforcer("configs/casbin/model.conf")
	if err != nil {
		panic(err)
	}

	_, _ = e.AddPolicy("account_admin", "account:*", "team", "create")
	_, _ = e.AddGroupingPolicy("ee9bed71-725e-499e-8646-17f6729fe0ae", "account_admin", "account:c42f8679-add8-4379-8166-28215bfef8f6")

	res, err := e.Enforce(
		"ee9bed71-725e-499e-8646-17f6729fe0ae",
		"account:c42f8679-add8-4379-8166-28215bfef8f6",
		"team",
		"create",
	)
	fmt.Printf("Enforce (in-memory test) = %v (err=%v)\n", res, err)
}