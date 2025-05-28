package main

import (
	"log"

	"github.com/PerryYao-GitHub/gorder/common/config"
	"github.com/spf13/viper"
)

func init() {
	if err := config.NewViperConfig(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	log.Print(viper.Get("order"))
	// log.Println("Listening :8082")
	// mux := http.NewServeMux()
	// mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
	// 	log.Printf("%v", r.RequestURI)
	// 	_, _ = io.WriteString(w, "pong")
	// })

	// if err := http.ListenAndServe(":8082", mux); err != nil {
	// 	log.Fatal(err)
	// }
}
