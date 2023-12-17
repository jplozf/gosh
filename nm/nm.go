// ****************************************************************************
//
//	 _____ _____ _____ _____
//	|   __|     |   __|  |  |
//	|  |  |  |  |__   |     |
//	|_____|_____|_____|__|__|
//
// ****************************************************************************
// G O S H   -   Copyright © JPL 2023
// ****************************************************************************
package nm

// ****************************************************************************
// nm is the Network Manager module
// ****************************************************************************

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

// ****************************************************************************
// ServeHTTP()
// ****************************************************************************
func ServeHTTP() {
	// Simple static webserver:
	pathPtr := flag.String("d", ".", "Directory path to serve")
	portPtr := flag.String("p", "8080", "Port to listen on")
	interfacePtr := flag.String("i", "", "Interface to listen on (default all)")
	flag.Parse()
	log.Printf("Serving %s on %s:%s", *pathPtr, *interfacePtr, *portPtr)
	err := http.ListenAndServe(fmt.Sprintf("%s:%s", *interfacePtr, *portPtr), http.FileServer(http.Dir(*pathPtr)))
	if err != nil {
		log.Fatal(err)
	}
}
