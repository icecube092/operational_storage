package main

import (
	"fmt"
	"net/http"
)


var addr = "localhost:4545"

func main(){
	server := http.Server{}
	server.Addr = addr
	handler := &httpHandler{}
	handler.storage = make(map[string]stringLife)

	go checking(handler)
	server.Handler = handler
	defer func(){
		fmt.Println("Exit...")
		err := server.Close()
		if err != nil{
			fmt.Println(err)
		}
	}()
	fmt.Println("Listening...")
	err := server.ListenAndServe()
	if err != nil{
		fmt.Println(err)
	}
}



