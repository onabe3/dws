package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
)

func createContainer(w http.ResponseWriter, r *http.Request) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		http.Error(w, "Failed to initialize Docker client", http.StatusInternalServerError)
		return
	}

	// イメージ名として "ubuntu:latest" を使用してコンテナを作成
	config := &container.Config{
		Image: "ubuntu:latest",
		Cmd:   []string{"echo", "hello world"},
	}
	resp, err := cli.ContainerCreate(context.Background(), config, nil, nil, nil, "")
	if err != nil {
		http.Error(w, "Failed to create container", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Container %s created!", resp.ID)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/create-container", createContainer).Methods("GET")

	http.Handle("/", r)
	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
