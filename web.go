/*
Travvik-go

Author: Jaime Yu

License: GPLv3
Copyright 2015
*/
package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
)

func main() {
	r := mux.NewRouter().StrictSlash(false)
	r.HandleFunc("/", HomeHandler)

	// Posts singular
	// Get a JSON output of the bus times
	NextBusTimes := r.PathPrefix("/json/bus/{busno}/{stopno}").Subrouter()
	NextBusTimes.Methods("GET").HandlerFunc(JsonGetHandler)

	PebbleBusTimes := r.PathPrefix("/json/pebble/bus/{busno}/{stopno}").Subrouter()
	PebbleBusTimes.Methods("GET").HandlerFunc(PebbleJsonGetHandler)

	fmt.Println("Starting server on :8080")

	bind := fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT"))
	http.ListenAndServe(bind, r)
}

func HomeHandler(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(rw, "Home")
}
func PebbleJsonGetHandler(rw http.ResponseWriter, r *http.Request) {
	bus := mux.Vars(r)["busno"]
	stop := mux.Vars(r)["stopno"]
	fmt.Fprintln(rw, "Showing bus: at stop", bus, stop)
	data := NextBusAt(bus, stop)

	type NextBus struct {
		Busno  string `json:"busno"`
		Stopno string `json:"stopno"`
		Label  string `json:"label"`
		//	nexttrips []NextTrips
	}
	type NextTrips struct {
		Label    string
		Dest     string
		Gpsspeed string
		Lat      string
		Long     string
		Bustype  string
		Eta      string
	}

	pebble := NextBus{}
	pebble.Busno = bus
	pebble.Stopno = stop
	pebble.Label = data.Body.Response.Result.StopLabel

	fmt.Println("RAW pebble:", pebble)

	jsondata, err := json.Marshal(pebble)
	fmt.Println("RAW pebble:", err)

	fmt.Fprintln(rw, string(jsondata))
}
func JsonGetHandler(rw http.ResponseWriter, r *http.Request) {
	bus := mux.Vars(r)["busno"]
	stop := mux.Vars(r)["stopno"]
	fmt.Fprintln(rw, "Showing bus: at stop", bus, stop)
	data := NextBusAt(bus, stop)
	jsondata, _ := json.Marshal(data)
	fmt.Fprintln(rw, string(jsondata))
}

func NextBusAt(bus string, stop string) *SoapOcTranspoEnvelope {

	// Get our credentials
	proctime := time.Now()
	cred := OCT_credentials{
		ID:  os.Getenv("OCTID"),
		KEY: os.Getenv("OCTKEY"),
	}

	//DEBUG
	//xml_string, _ := ioutil.ReadFile("nexttrip.xml")
	// Make the POST request
	soapresp, _ := http.PostForm("https://api.octranspo1.com/v1.2/GetNextTripsForStop",
		url.Values{"appID": {cred.ID},
			"apiKey":  {cred.KEY},
			"routeNo": {bus},
			"stopNo":  {stop},
		})
	//fmt.Println("%#v", soapresp)

	// Pull the data out of the response
	defer soapresp.Body.Close()
	body, _ := ioutil.ReadAll(soapresp.Body)
	fmt.Println("From OC: %s", string(body))

	// OC transpo unmarshal test
	//data := Unmarshal_Soap_OC_Transpo([]byte(body))

	endproctime := time.Now()
	end := endproctime.Sub(proctime)
	fmt.Println("Time spent: ", end)
	contents := &SoapOcTranspoEnvelope{}
	xml.Unmarshal([]byte(body), contents)
	Unmarshal_Soap_OC_Transpo(contents)

	return contents

}

/* Credentials JSON Struct */
type OCT_credentials struct {
	ID  string `json:"ID"`
	KEY string `json:"KEY"`
}

/* Types for OC Transpo API V1.2 */

type SoapOcTranspoEnvelope struct {
	XMLName xml.Name
	Body    SoapBody
}

type SoapBody struct {
	XMLName  xml.Name
	Response SoapCompleteResponse `xml:"GetNextTripsForStopResponse"`
}

type SoapCompleteResponse struct {
	XMLName xml.Name
	Result  GetNextTripsForStopResponse `xml:"GetNextTripsForStopResult"`
}

type GetNextTripsForStopResponse struct {
	XMLName   xml.Name
	StopNo    string `xml:"StopNo"`
	StopLabel string `xml:"StopLabel"`
	Error     string `xml:"Error"`
	Route     Route  `xml:"Route"`
}

type Route struct {
	XMLName        xml.Name
	RouteDirection []RouteDirection `xml:"RouteDirection"`
}

type RouteDirection struct {
	XMLName    xml.Name
	RouteNo    string `xml:"RouteNo"`
	RouteLabel string `xml:"RouteLabel"`
	Direction  string `xml:"Direction"`
	ProcTime   string `xml:"RequestProcessingTime"`
	Trips      Trips  `xml:"Trips"`
}

type Trips struct {
	XMLName xml.Name
	Trip    []Trip
}

type Trip struct {
	XMLName                  xml.Name
	TripDestination          string `xml:"TripDestination"`
	TripStartTime            string `xml:"TripStartTime"`
	TripAdjustedScheduleTime string `xml:"TripAdjustedScheduleTime"`
	TripAdjustmentAge        string `xml:"TripAdjustmentAge"`
	TripLastTripOfSchedule   string `xml:"TripLastTripOfSchedule"`
	BusType                  string `xml:"BusType"`
	Latitude                 string `xml:"Latitude"`
	Longitude                string `xml:"Longitude"`
	GPSSpeed                 string `xml:"GPSSpeed"`
}

func Unmarshal_Soap_OC_Transpo(contents *SoapOcTranspoEnvelope) {

	fmt.Println("!!!!!!!!!DEBUG")
	fmt.Println("BODY:%#v", contents.Body.Response)
	fmt.Println("Result:%#v", contents.Body.Response.Result)
	fmt.Println("Route:%#v", contents.Body.Response.Result.Route)
	fmt.Println("!!!!!!!!!DEBUG")

	fmt.Println(contents.Body.Response.Result.StopNo)
	fmt.Println(contents.Body.Response.Result.StopLabel)
	fmt.Println(contents.Body.Response.Result.Error)
	fmt.Println("Number of directions:", len(contents.Body.Response.Result.Route.RouteDirection))
	sz := len(contents.Body.Response.Result.Route.RouteDirection)
	if sz == 0 {
		//fmt.Println(contents.Body.Response.Result.Route.RouteDirection.RouteNo)
		//fmt.Println(contents.Body.Response.Result.Route.RouteDirection[0].RouteNo)
		fmt.Println("ERROR, missing data")
	} else {

		for route := range contents.Body.Response.Result.Route.RouteDirection {
			fmt.Println(contents.Body.Response.Result.Route.RouteDirection[route].RouteNo)
			fmt.Println(contents.Body.Response.Result.Route.RouteDirection[route].RouteLabel)
			fmt.Println(contents.Body.Response.Result.Route.RouteDirection[route].Direction)
			fmt.Println(contents.Body.Response.Result.Route.RouteDirection[route].ProcTime)

			for trip := range contents.Body.Response.Result.Route.RouteDirection[route].Trips.Trip {
				fmt.Println(contents.Body.Response.Result.Route.RouteDirection[route].Trips.Trip[trip].TripDestination)
				fmt.Println(contents.Body.Response.Result.Route.RouteDirection[route].Trips.Trip[trip].TripStartTime)
				fmt.Println(contents.Body.Response.Result.Route.RouteDirection[route].Trips.Trip[trip].TripAdjustedScheduleTime)
				fmt.Println(contents.Body.Response.Result.Route.RouteDirection[route].Trips.Trip[trip].TripAdjustmentAge)
				fmt.Println(contents.Body.Response.Result.Route.RouteDirection[route].Trips.Trip[trip].TripLastTripOfSchedule)
				fmt.Println(contents.Body.Response.Result.Route.RouteDirection[route].Trips.Trip[trip].BusType)
				fmt.Println(contents.Body.Response.Result.Route.RouteDirection[route].Trips.Trip[trip].Latitude)
				fmt.Println(contents.Body.Response.Result.Route.RouteDirection[route].Trips.Trip[trip].Longitude)
				fmt.Println(contents.Body.Response.Result.Route.RouteDirection[route].Trips.Trip[trip].GPSSpeed)
			}
		}
	}

}
