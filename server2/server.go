package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Despacho struct {
	Id_despacho int64
	Estado      string
	Id_compra   int64
}

func main() {
	var err error
	db, err = sql.Open("mysql", "root:dfghdfgh@tcp(127.0.0.1:3306)/db_despachos?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
		fmt.Println("error")
	}

	router := mux.NewRouter()

	router.HandleFunc("/api/clientes/estado_despacho/{id}", get_estado).Methods("GET")

	//aqui el server espera alguna consulta del main.go
	fmt.Println("Waiting")
	http.ListenAndServe(":5000", router)

}

func get_estado(w http.ResponseWriter, r *http.Request) {
	fmt.Println("get estado")
	vars := mux.Vars(r)
	key := vars["id"]
	intVar, err := strconv.Atoi(key)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")

	rows := db.QueryRow("SELECT * FROM despacho WHERE id_despacho = ? ", intVar)

	var despacho Despacho

	if err := rows.Scan(&despacho.Id_despacho, &despacho.Estado, &despacho.Id_despacho); err != nil {
		log.Fatal(err)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("dd")
	fmt.Println(despacho.Id_compra)

	json.NewEncoder(w).Encode(despacho)
}
