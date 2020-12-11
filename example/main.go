package main

import (
    "encoding/json"
    "github.com/tianying484/looker/rtl"
    v4 "github.com/tianying484/looker/sdk/v4"
    "log"
)

func main() {
    var (
        err          error
        lookerId     int64  = 684
        limit        int64  = 5000
        resultFormat string = "csv"
        lookerIni    string = `./looker.ini`
    )

    // Read settings from ini file
    cfg, err := rtl.NewSettingsFromFile(lookerIni, nil)
    checkErr(err)

    // New instance of LookerSDK
    sdk := v4.NewLookerSDK(rtl.NewAuthSession(cfg))

    // get looker info
    looker, err := sdk.Look(lookerId, nil, nil)
    checkErr(err)

    // get looker filters
    filters, err := json.MarshalIndent(looker.Query.Filters, "  ", "  ")
    checkErr(err)
    log.Println("get looker filters", *looker.QueryId, string(filters))

    // create looker query request
    createQuery, err := sdk.CreateQuery(v4.WriteQuery{
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
    }, nil, nil)
    checkErr(err)
    log.Println("create query success, query_id=", *createQuery.Id)

    // run looker query
    csvText, err := sdk.RunQuery(v4.RequestRunQuery{
        QueryId:      *createQuery.Id,
        ResultFormat: resultFormat,
        Limit:        &limit,
    }, nil)
    checkErr(err)

    log.Println("run query result", csvText)

    // List all users in Looker
    users, err := sdk.AllUsers(v4.RequestAllUsers{}, nil)
    checkErr(err)

    log.Println("-------------------------")
    // Iterate the users and print basic user info
    for _, u := range users {
        log.Printf("user: %s:%s:%s\n", *u.FirstName, *u.LastName, *u.Email)
    }
    log.Println("-------------------------")
}

func checkErr(err error) {
    if err != nil {
        log.Fatal(err)
    }
}
