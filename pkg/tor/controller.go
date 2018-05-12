package tor

import (
	"fmt"
	"github.com/wybiral/torgo"
)

// Return new Tor controller.
func NewController(host string, port int) (*torgo.Controller, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	return torgo.NewController(addr)
}
