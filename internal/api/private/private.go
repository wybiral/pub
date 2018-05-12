/*
GET /peers
	Get peer list
GET /subscribe/{onion id}
	Make subscribe request to {onion id}
*/

/*
TODO:

POST /
	Publish article
GET /
	Get recent timeline
*/

package private

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/wybiral/pub/internal/app"
	"github.com/wybiral/pub/pkg/utils"
	"log"
	"net"
	"net/http"
)

type Api struct {
	app *app.App
}

func StartApi(app *app.App) {
	api := Api{
		app: app,
	}
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/peers", api.peersHandler).Methods("GET")
	r.HandleFunc("/subscribe/{onion}", api.subscribeHandler).Methods("GET")
	// Create listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	// Print private port
	log.Println("Private interface:", addr)
	// Serve routes on listener
	err = http.Serve(listener, r)
	if err != nil {
		log.Fatal(err)
	}
}

// Returns JSON encoded list of peers.
func (api *Api) peersHandler(w http.ResponseWriter, r *http.Request) {
	app := api.app
	peers, err := app.Model.GetPeers()
	if err != nil {
		utils.JsonError(w, err.Error())
		return
	}
	utils.JsonResponse(w, peers)
}

// Make a subscribe request to a peer by onion.
func (api *Api) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	app := api.app
	vars := mux.Vars(r)
	onion := vars["onion"]
	peer, err := app.Self.SubscribeRequest(app.Tor.Client, onion)
	if err != nil {
		utils.JsonError(w, err.Error())
		return
	}
	utils.JsonResponse(w, peer)
}
