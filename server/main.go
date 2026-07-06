// fakecloud is a pretend cloud provider for learning Terraform. It serves a
// JSON API for resources (VMs and tic-tac-toe games) and a live dashboard
// that visualizes the state Terraform manages.
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/pokgak/fakecloud/server/api"
	"github.com/pokgak/fakecloud/server/store"
	"github.com/pokgak/fakecloud/server/web"
)

func main() {
	addr := flag.String("addr", ":8000", "address to listen on")
	flag.Parse()

	mux := http.NewServeMux()
	api.New(store.New()).Register(mux)
	mux.Handle("/", web.Handler())

	log.Printf("fakecloud listening on %s — open http://localhost%s to watch your resources", *addr, *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
