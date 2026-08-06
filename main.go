
package main

import (
	"fmt"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to WEB303 Microservices Lab I am Sonam Wangmo")
}

func main() {
	http.HandleFunc("/", homeHandler)
	fmt.Println("Server is running on port 8081...")

	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		fmt.Println(err)
	}
}
