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
	//	"github.com/gorilla/schema"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
)

func serveSingle(pattern string, filename string) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filename)
	})
}

func main() {
	r := mux.NewRouter().StrictSlash(false)
	r.HandleFunc("/", HomeHandler)

	//	r.HandleFunc("/bus", HomeHandler)
	//	r.HandleFunc("/bus/{bus}/{route}", HomeHandler)
	NextBus := r.PathPrefix("/bus").Subrouter()
	NextBus.Methods("GET").HandlerFunc(HomeHandler)

	// Normal resources
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

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
	/*tmp, _ := ioutil.ReadFile("./static/index.html")
	tmpl := string(tmp)

	type BusReqFrom struct {
		BusNo  string `schema:"BusNo"`
		StopNo string `schema:"StopNo"`
	}
	*/
	err := r.ParseForm()

	if err != nil {
		// Handle error
	}
	/*
		decoder := schema.NewDecoder()
		// r.PostForm is a map of our POST form values
		form := new(BusReqFrom)
		err = decoder.Decode(form, r.PostForm)

		if err != nil {
			// Handle error
		}

		// Do something with person.Name or person.Phone
		fmt.Println("Getting:", form.BusNo, form.StopNo)
		fmt.Println("Getting:", r.Form["BusNo"][0], r.Form["StopNo"][0])
		fmt.Println(r.Form) // print information on server side.
	*/
	fmt.Println(r.Form) // print information on server side.
	bus := r.Form["BusNo"][0]
	stop := r.Form["StopNo"][0]

	data := NextBusAt(bus, stop)

	type DispTrip struct {
		Type      string
		Label     string
		Dst       string
		Eta       string
		GpsStatus string
		GpsLat    string
		GpsLong   string
		Speed     string
	}

	type NextTrips struct {
		BusNo  string
		StopNo string
		Trips  []*DispTrip
	}

	lsTrips := []*DispTrip{}
	//fmt.Println("RAW DATA:", data)
	for route := range data.Body.Response.Result.Route.RouteDirection {
		for trip := range data.Body.Response.Result.Route.RouteDirection[route].Trips.Trip {
			// Create trips and append to trips[]
			cur_trip := DispTrip{}
			cur_trip.Label = data.Body.Response.Result.Route.RouteDirection[route].RouteNo
			cur_trip.Type = data.Body.Response.Result.Route.RouteDirection[route].Trips.Trip[trip].BusType
			fmt.Println("DBG:", data.Body.Response.Result.Route.RouteDirection[route].Trips.Trip[trip].TripDestination)
			cur_trip.Dst = data.Body.Response.Result.Route.RouteDirection[route].Trips.Trip[trip].TripDestination
			cur_trip.Speed = data.Body.Response.Result.Route.RouteDirection[route].Trips.Trip[trip].GPSSpeed
			cur_trip.GpsLat = data.Body.Response.Result.Route.RouteDirection[route].Trips.Trip[trip].Latitude
			cur_trip.GpsLong = data.Body.Response.Result.Route.RouteDirection[route].Trips.Trip[trip].Longitude
			cur_trip.Eta = data.Body.Response.Result.Route.RouteDirection[route].Trips.Trip[trip].TripAdjustedScheduleTime
			lsTrips = append(lsTrips, &cur_trip)

		}
	}

	trips := NextTrips{}
	trips.Trips = lsTrips
	trips.BusNo = bus
	trips.StopNo = stop

	/*fmt.Fprintln(rw, "DBGHBus triupos:", lsTrips)
	fmt.Fprintln(rw, "DBGHBus label:", len(trips.Trips))
	for i := range trips.Trips {
			fmt.Fprintln(rw, "Bus label:", trips.Trips[i].Dst)
	}
	ct := trips.Trips[0]
	*/
	index, _ := ioutil.ReadFile("./static/index.html")
	t := template.New("NextTrips template")
	t, err = t.Parse(string(index))
	//t, err = t.ParseFiles("./static/index.html")
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}

	t.Execute(rw, trips)
	//fmt.Fprintln(rw, ct)

	//fmt.Fprintln(rw, t)
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
	fmt.Println("From OC RAW XML: %s", string(body))

	// OC transpo unmarshal test

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
	TripAdjustedScheduleTime string `xml:"AdjustedScheduleTime"`
	TripAdjustmentAge        string `xml:"djustmentAge"`
	TripLastTripOfSchedule   string `xml:"LastTripOfSchedule"`
	BusType                  string `xml:"BusType"`
	Latitude                 string `xml:"Latitude"`
	Longitude                string `xml:"Longitude"`
	GPSSpeed                 string `xml:"GPSSpeed"`
}

func Unmarshal_Soap_OC_Transpo(contents *SoapOcTranspoEnvelope) {

	fmt.Println("!!!!!!!!!DEBUG")
	//fmt.Println("RAW DATA:", contents)
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
