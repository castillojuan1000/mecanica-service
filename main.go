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
)

type Customer struct {
	gorm.Model

	FirstName string
	LastName  string
	Phone     string `gorm:"typevarchar(100);unique_index"`
	Cars      []Car
}

type Car struct {
	gorm.Model

	Make         string
	Modelo       string
	VinNumber    string `gorm:"typevarchar(100);unique_index"`
	Maintenances []Maintenance
	CustomerId   int
}

type Maintenance struct {
	gorm.Model

	Note  string
	CarId int
}

var db *gorm.DB
var err error

func main() {
	//Loading env variables
	dialect := os.Getenv("DIALECT")
	host := os.Getenv("HOST")
	dbPort := os.Getenv("DBPORT")
	user := os.Getenv("USER")
	dbName := os.Getenv("NAME")
	password := os.Getenv("PASSWORD")

	//connect to db postgres
	dbURI := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s port=%s", host, user, dbName, password, dbPort)

	//openning connection to DB
	db, err = gorm.Open(dialect, dbURI)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Succesfully connected to DB")
	}

	//close connection do db when main function finishes
	defer db.Close()

	//Make migration to the db
	db.AutoMigrate(&Customer{})
	db.AutoMigrate(&Car{})
	db.AutoMigrate(&Maintenance{})

	//api routes
	router := mux.NewRouter()

	//customers
	router.HandleFunc("/customers", getCustomers).Methods("GET", "OPTIONS")
	router.HandleFunc("/customer/{id}", getCustomerById).Methods("GET") //and get their cars as well
	router.HandleFunc("/create/customer", createCustomer).Methods("POST", "OPTIONS")
	router.HandleFunc("/delete/customer/{id}", deleteCustomer).Methods("DELETE")
	router.HandleFunc("/customers/{firstName}/{lastName}", getCustomerByFullName).Methods("GET")
	router.HandleFunc("/customers/{phone}", getCustomerByPhoneNumber).Methods("GET")

	//cars
	router.HandleFunc("/cars", getCars).Methods("GET")
	router.HandleFunc("/car/{id}", getCar).Methods("GET")
	router.HandleFunc("/create/car", createCar).Methods("POST")
	router.HandleFunc("/delete/car/{id}", deleteCar).Methods("DELETE")

	//Maintanences
	router.HandleFunc("/create/maintanence", createMaintanence).Methods("POST")
	router.HandleFunc("/delete/maintanence", deleteMaintenance).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8080", router))

}

//api controllers

//get get all customers
func getCustomers(w http.ResponseWriter, r *http.Request) {
	//Allow CORS here By * or specific origin
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	var customers []Customer
	db.Find(&customers)
	json.NewEncoder(w).Encode(&customers)
}

//get a customer
func getCustomerById(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var customer Customer
	var cars []Car

	db.Where("id = ?", id).Find(&customer)
	db.Model(&customer).Related(&cars)

	customer.Cars = cars
	json.NewEncoder(w).Encode(&customer)
}

//get customer by phone number
func getCustomerByFullName(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	firstName := params["firstName"]
	lastName := params["lastName"]

	var customer Customer
	// var car Car
	var cars []Car
	// var maintenances []Maintenance

	db.Where("first_name = ? AND last_name = ?", firstName, lastName).Find(&customer)
	db.Model(&customer).Related(&cars)
	// db.Model(&car).Related(&maintenances)

	// car.Maintenances = maintenances
	customer.Cars = cars
	json.NewEncoder(w).Encode(&customer)
}

func getCustomerByPhoneNumber(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	phone := params["phone"]

	var customer Customer
	var cars []Car

	db.Where("phone = ?", phone).Find(&customer)
	db.Model(&customer).Related(&cars)

	fmt.Println("{}", customer)
	customer.Cars = cars
	json.NewEncoder(w).Encode(&customer)
}

//create new customer
func createCustomer(w http.ResponseWriter, r *http.Request) {
	var customer Customer

	//Allow CORS here By * or specific origin
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

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

	var customer Customer
	db.First(&customer, params["id"])
	db.Delete(&customer)

	json.NewEncoder(w).Encode(&customer)
}

//get cars
func getCars(w http.ResponseWriter, r *http.Request) {
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
	params := mux.Vars(r)
	var car Car
	var maintenances []Maintenance

	db.First(&car, params["id"])
	db.Model(&car).Related(&maintenances)

	car.Maintenances = maintenances
	json.NewEncoder(w).Encode(&car)
}

//create  a car
func createCar(w http.ResponseWriter, r *http.Request) {
	var car Car

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
	params := mux.Vars(r)

	var car Car
	db.First(&car, params["id"])
	db.Delete(&car)

	json.NewEncoder(w).Encode(&car)
}

//Maintenance controllers

//delete Maintenance
func deleteMaintenance(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var maintenance Maintenance

	db.First(&maintenance, id)
	db.Delete(&maintenance)

	json.NewEncoder(w).Encode(&maintenance)
}

//create new maintanence
func createMaintanence(w http.ResponseWriter, r *http.Request) {
	var maintenance Maintenance

	json.NewDecoder(r.Body).Decode(&maintenance)
	createdMaintanence := db.Create(&maintenance)
	err = createdMaintanence.Error

	if err != nil {
		json.NewEncoder(w).Encode(err)
	} else {
		json.NewEncoder(w).Encode(&maintenance)
	}

}
