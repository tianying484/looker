package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tianying484/looker/v4/rtl"
	v4 "github.com/tianying484/looker/v4/sdk/v4"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func printAllUsers(sdk *v4.LookerSDK) {
	// List all users in Looker
	users, err := sdk.AllUsers(v4.RequestAllUsers{}, nil)
	check(err)

	println("-------------------------")
	// Iterate the users and print basic user info
	for _, u := range users {
		fmt.Printf("user: %s:%s:%s\n", *u.FirstName, *u.LastName, *u.Email)
	}
	println("-------------------------")
}

func printAllProjects(sdk *v4.LookerSDK) {
	projects, err := sdk.AllProjects("", nil)
	check(err)
	for _, proj := range projects {
		fmt.Printf("Project: %s %s %s\n", *proj.Name, *proj.Id, *proj.GitRemoteUrl)
	}
}

func printAboutMe(sdk *v4.LookerSDK) {

	me, err := sdk.Me("", nil)
	check(err)

	fmt.Printf("You are %s\n", *(me.Email))

	// Search for this user by their e-mail
	users, err := sdk.SearchUsers(v4.RequestSearchUsers{Email: me.Email}, nil)
	if err != nil {
		fmt.Printf("Error getting myself %v\n", err)
	}
	if len(users) != 1 {
		fmt.Printf("Found %d users with my email expected 1\n")
	}
}

func printQuery(sdk *v4.LookerSDK) {
	var (
		lookerId = "737"
		limit    = int64(100)
		cache    = false
	)

	looker, err := sdk.Look(lookerId, "", nil)
	if err != nil {
		panic(err)
	}

	query, err := sdk.CreateQuery(v4.WriteQuery{
		Model:         looker.Query.Model,
		View:          looker.Query.View,
		Fields:        looker.Query.Fields,
		Pivots:        looker.Query.Pivots,
		FillFields:    looker.Query.FillFields,
		Filters:       looker.Query.Filters,
		Sorts:         looker.Query.Sorts,
		Limit:         looker.Query.Limit,
		ColumnLimit:   looker.Query.ColumnLimit,
		Total:         looker.Query.Total,
		RowTotal:      looker.Query.RowTotal,
		Subtotals:     looker.Query.Subtotals,
		DynamicFields: looker.Query.DynamicFields,
		QueryTimezone: looker.Query.QueryTimezone,
	}, "", nil)
	if err != nil {
		panic(err)
	}

	result, err := sdk.RunQuery(v4.RequestRunQuery{
		QueryId:      *query.Id,
		ResultFormat: "csv",
		Limit:        &limit,
		Cache:        &cache,
	}, nil)
	if err != nil {
		panic(err)
	}
	log.Println(result)
}

func main() {
	// Default config file location
	lookerIniPath := "/opt/looker.ini"
	if len(os.Args) > 1 {
		// If first argument exists then it is the config file
		lookerIniPath = os.Args[1]
	}

	// Read settings from ini file
	cfg, err := rtl.NewSettingsFromFile(lookerIniPath, nil)
	check(err)

	// New instance of LookerSDK
	sdk := v4.NewLookerSDK(rtl.NewAuthSession(cfg))

	printAllProjects(sdk)

	printAllUsers(sdk)

	printAboutMe(sdk)

	printQuery(sdk)
}
