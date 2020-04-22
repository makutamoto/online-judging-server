package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/docker/docker/client"
)

func server(cli *client.Client) {
	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			bytes, err := ioutil.ReadAll(r.Body)
			if err != err {
				fmt.Println(err)
			}
			ctx := context.Background()
			go runContainer(ctx, cli, string(bytes))
			fmt.Fprintln(w, "OK")
		}
	})
	log.Fatal(http.ListenAndServe(":7867", nil))
}
