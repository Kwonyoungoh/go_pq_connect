package main

import (
	"go_db_connect/myapp"
	"net/http"
)

func main()  {

	http.ListenAndServe(":2000", myapp.NewHttpHandler())
	
}