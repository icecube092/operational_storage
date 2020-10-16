package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	str := []byte(`{"Key": "Hello", "Value": "World", "Duration": 10}`)
	client := &http.Client{}
	req, err := http.NewRequest(
		"GET", "http://localhost:4545", bytes.NewBuffer(str))
	req.Header.Add("Content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.Status)
	fmt.Println(resp.Header)
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)
}