/*
GET /info
	Peer info
POST /subscribe
	Request subscription
*/

/*
TODO:

GET /
	Read posts
POST /
	Post comment
*/
package public

import (
	"github.com/gorilla/mux"
	"github.com/wybiral/pub/internal/app"
	"github.com/wybiral/pub/pkg/utils"
	"io/ioutil"
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
	r.HandleFunc("/info", api.infoGetHandler).Methods("GET")
	r.HandleFunc("/subscribe", api.subscribeHandler).Methods("POST")
	// Create listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	// Start onion
	onionKeyType := app.Self.OnionKeyType
	onionKeyContent := app.Self.PrivateOnionKey
	err = app.Tor.StartOnion(port, onionKeyType, onionKeyContent)
	if err != nil {
		log.Fatal(err)
	}
	// Print onion address
	log.Println(app.Self.Onion)
	// Serve routes on listener
	err = http.Serve(listener, r)
	if err != nil {
		log.Fatal(err)
	}
}

// Return JSON encoded identity info for peers.
func (api *Api) infoGetHandler(w http.ResponseWriter, r *http.Request) {
	app := api.app
	utils.JsonResponse(w, app.Self)
}

// Handle a subscribe request (currently accepts all subscriptions).
func (api *Api) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	app := api.app
	log.Println("/subscribe")
	auth, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	onion := r.Header.Get("Peer")
	peer, err := app.Self.SubscribeAccept(app.Tor.Client, onion, auth)
	if err != nil {
		log.Println(err)
		utils.JsonError(w, err.Error())
		return
	}
	utils.JsonResponse(w, peer)
}
