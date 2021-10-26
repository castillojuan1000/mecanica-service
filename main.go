package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
)

type Customer struct {
	gorm.Model

	FirstName string
	LastName  string
	Phone     string `gorm:"typevarchar(100);unique_index"`
	Cars      []Car  `gorm:"constraint:OnDelete:CASCADE;"`
}

type Car struct {
	gorm.Model

	Make       string
	Modelo     string
	Color      string
	VinNumber  string     `gorm:"typevarchar(100);unique_index"`
	Services   []*Service `gorm:"constraint:OnDelete:CASCADE;"`
	CustomerId int
}

type Service struct {
	gorm.Model

	Comment string
	Miles   string
	CarId   int
}

var db *gorm.DB
var err error

func main() {

	e := godotenv.Load() //Load .env file
	if e != nil {
		fmt.Print(e)
	}

	//Loading env variables
	dialect := os.Getenv("DIALECT")
	host := os.Getenv("HOST")
	dbPort := os.Getenv("DBPORT")
	user := os.Getenv("USER")
	dbName := os.Getenv("NAME")
	password := os.Getenv("PASSWORD")

	// retrieve the url from hereku
	// dbURLHeroku := os.Getenv("DATABASE_URL")

	//connect to db postgres
	dbURI := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s port=%s", host, user, dbName, password, dbPort)

	// openning connection to DB
	db, err = gorm.Open(dialect, dbURI)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Succesfully connected to DB")
	}

	if err != nil {
		fmt.Println(err)
	}

	//close connection do db when main function finishes
	defer db.Close()

	//Make migration to the db
	db.Debug().AutoMigrate(&Customer{})
	db.Debug().AutoMigrate(&Car{})
	db.Debug().AutoMigrate(&Service{})
	//api routes
	router := mux.NewRouter()

	//customers
	router.HandleFunc("/customers", getCustomers).Methods("GET", "OPTIONS")
	router.HandleFunc("/customer/{id}", getCustomerById).Methods("GET", "OPTIONS") //and get their cars as well
	router.HandleFunc("/create/customer", createCustomer).Methods("POST", "OPTIONS")
	router.HandleFunc("/delete/customer/{id}", deleteCustomer).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/update/customer/{id}", updateCustomer).Methods("PUT", "OPTIONS")

	//cars
	router.HandleFunc("/cars", getCars).Methods("GET", "OPTIONS")
	router.HandleFunc("/car/{id}", getCar).Methods("GET", "OPTIONS")
	router.HandleFunc("/create/car", createCar).Methods("POST", "OPTIONS")
	router.HandleFunc("/delete/car/{id}", deleteCar).Methods("DELETE", "OPTIONS")

	//Maintanences
	router.HandleFunc("/create/service", createService).Methods("POST", "OPTIONS")
	router.HandleFunc("/delete/service", deleteService).Methods("DELETE", "OPTIONS")

	log.Fatal(http.ListenAndServe(":8080", router))
}

//Handle CORS
func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

//api controllers

//get get all customers
func getCustomers(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	var customers []Customer
	db.Find(&customers)
	json.NewEncoder(w).Encode(&customers)
}

// get a customer and cars
func getCustomerById(w http.ResponseWriter, r *http.Request) {

	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	params := mux.Vars(r)
	id := params["id"]

	var customer Customer
	var cars []Car

	db.Where("id = ?", id).Find(&customer)
	db.Model(&customer).Related(&cars)

	customer.Cars = cars
	json.NewEncoder(w).Encode(&customer)
}

//create new customer
func createCustomer(w http.ResponseWriter, r *http.Request) {
	var customer Customer

	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	json.NewDecoder(r.Body).Decode(&customer)

	createdPerson := db.Create(&customer)
	err = createdPerson.Error

	if err != nil {
		json.NewEncoder(w).Encode(err)
	} else {
		json.NewEncoder(w).Encode(&customer)
	}

}

//delete customer
func deleteCustomer(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	var customer Customer
	var cars []Car

	db.First(&customer, params["id"])
	db.Model(&customer).Related(&cars)

	deleteServicesSqlStatement := `
		DELETE FROM services
		WHERE car_id = $1;`

	for i, car := range cars {
		fmt.Println(i, car.ID)
		err := db.Exec(deleteServicesSqlStatement, car.ID).Error
		if err != nil {
			fmt.Println(err)
		}

	}

	DeleteCarsSqlStatement := `
		DELETE FROM cars
		WHERE customer_id = $1;`

	err := db.Exec(DeleteCarsSqlStatement, params["id"]).Error
	if err != nil {
		fmt.Println(err)
	}

	db.Debug().Unscoped().Delete(&customer)
	json.NewEncoder(w).Encode(&customer)

}

//edit customer
func updateCustomer(w http.ResponseWriter, r *http.Request) {
	//Allow CORS here By * or specific origin

	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	r.Close = true
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	var customer Customer
	db.First(&customer, params["id"])

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&customer); err != nil {
		fmt.Println(err)
		return
	}
	defer r.Body.Close()

	if err := db.Save(&customer).Error; err != nil {
		fmt.Println(err)
		return
	}

	json.NewEncoder(w).Encode(&customer)

}

//get cars
func getCars(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	params := mux.Vars(r)

	var customer Customer
	var cars []Car

	db.First(&customer, params["id"])
	db.Model(&customer).Related(&cars)

	customer.Cars = cars
	json.NewEncoder(w).Encode(&customer)
}

//get a car
func getCar(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}
	params := mux.Vars(r)
	var car Car
	var services []*Service

	db.First(&car, params["id"])
	db.Model(&car).Related(&services)

	car.Services = services
	json.NewEncoder(w).Encode(&car)
}

//create  a car
func createCar(w http.ResponseWriter, r *http.Request) {
	var car Car

	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	json.NewDecoder(r.Body).Decode(&car)
	createdCar := db.Create(&car)
	err = createdCar.Error

	if err != nil {
		json.NewEncoder(w).Encode(err)
	} else {
		json.NewEncoder(w).Encode(&car)
	}

}

//delete car
func deleteCar(w http.ResponseWriter, r *http.Request) {
	//handle CORS
	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	params := mux.Vars(r)

	var car Car
	db.First(&car, params["id"])

	sqlStatement := `
		DELETE FROM services
		WHERE car_id = $1;`

	err := db.Exec(sqlStatement, params["id"]).Error

	if err != nil {
		fmt.Println(err)
	}

	db.Debug().Unscoped().Delete(&car)
	json.NewEncoder(w).Encode(&car)
}

//delete Service
func deleteService(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var service Service

	db.First(&service, id)
	db.Delete(&service)

	json.NewEncoder(w).Encode(&service)
}

//create new service
func createService(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	var maintenance Service

	json.NewDecoder(r.Body).Decode(&maintenance)
	createdMaintanence := db.Create(&maintenance)
	err = createdMaintanence.Error

	if err != nil {
		json.NewEncoder(w).Encode(err)
	} else {
		json.NewEncoder(w).Encode(&maintenance)
	}

}
