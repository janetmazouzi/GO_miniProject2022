package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Product struct {
	ID                  int64
	Nombre              string
	Cantidad_Disponible int64
	Precio_Unitario     int64
}
type NewProduct struct {
	Nombre              string
	Cantidad_Disponible int64
	Precio_Unitario     int64
}
type Usuario struct {
	ID       int64
	Nombre   string
	Password string
}

type Detalle struct {
	ID_Compra   int64
	ID_Producto int64
	Cantidad    int64
	Fecha       time.Time
}

type Estadisticas struct {
	Producto_Mas_Vendido     int64
	Producto_Menos_Vendido   int64
	Producto_Mayor_Ganancias int64
	Producto_Menor_Ganancia  int64
}

type InicioSesion struct {
	Acceso_Valido bool
}

type Compra struct {
	ID_Producto int64
	Cantidad    int64
}
type CompraDatos struct {
	Cantidad int64
	Costo    int64
}

type Compras struct {
	ID_Cliente int64
	Productos  []Compra
}

type crearProducto struct {
	ID_Compra int64
}

type Delete struct {
	ID_Producto int64
}

type Agregar_Producto struct {
	ID_Producto int64
}

type session struct {
	ID       int64
	Password string
}

func main() {
	var err error
	db, err = sql.Open("mysql", "root:dfghdfgh@tcp(127.0.0.1:3306)/tarea_1_sd?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}

	router := mux.NewRouter()

	router.HandleFunc("/api/clientes/iniciar_sesion", ini_sesion).Methods("POST")
	router.HandleFunc("/api/compras", nueva_compra).Methods("POST")
	router.HandleFunc("/api/productos", agrega_producto).Methods("POST")
	router.HandleFunc("/api/productos", lista_productos).Methods("GET")
	router.HandleFunc("/api/productos/{id}", modifica_producto).Methods("PUT")
	router.HandleFunc("/api/productos/{id}", quita_producto).Methods("DELETE")
	router.HandleFunc("/api/estadisticas", mostra_estadisticas).Methods("GET")

	//Auxiliares
	router.HandleFunc("/api/compras/{id}", get_compra).Methods("GET")

	//aqui el server espera alguna consulta del main.go
	fmt.Println("Waiting")
	http.ListenAndServe(":5000", router)

}

func ini_sesion(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)

	var result session
	json.Unmarshal([]byte(reqBody), &result)

	var user Usuario
	var sesion InicioSesion
	row := db.QueryRow("SELECT * FROM cliente WHERE id_cliente = ?", result.ID)
	if err := row.Scan(&user.ID, &user.Nombre, &user.Password); err != nil {
		sesion.Acceso_Valido = false
		json.NewEncoder(w).Encode(sesion)
		return
	} else {
		if result.ID == user.ID && result.Password == user.Password {
			sesion.Acceso_Valido = true
			json.NewEncoder(w).Encode(sesion)
			return
		}
	}
	sesion.Acceso_Valido = false
	json.NewEncoder(w).Encode(sesion)
}

func nueva_compra(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var result Compras
	json.Unmarshal([]byte(reqBody), &result)

	resultCompra, err := db.Exec(`INSERT INTO compra(id_cliente) VALUES (?)`, result.ID_Cliente)
	if err != nil { // nil means there's no error
		panic(err.Error())
	}
	idCompra, err := resultCompra.LastInsertId()
	if err != nil {
		panic(err.Error())
	}

	var numProd int64
	numProd = 0
	var costoTotal int64
	costoTotal = 0

	for i := 0; i < len(result.Productos); i++ {
		var prod Product
		row := db.QueryRow("SELECT * FROM producto WHERE id_producto = ?", result.Productos[i].ID_Producto)
		if err := row.Scan(&prod.ID, &prod.Nombre, &prod.Cantidad_Disponible, &prod.Precio_Unitario); err != nil {
			log.Fatal(err)
		}

		var diff int64
		diff = prod.Cantidad_Disponible - result.Productos[i].Cantidad

		if diff >= 0 {
			db.Exec(`INSERT INTO detalle(id_compra, id_producto, cantidad) VALUES (?, ?, ?)`,
				idCompra, result.Productos[i].ID_Producto, result.Productos[i].Cantidad)

			db.Exec(`UPDATE producto SET cantidad_disponible = ? WHERE id_producto = ?`,
				diff, result.Productos[i].ID_Producto)

			numProd = numProd + result.Productos[i].Cantidad
			costoTotal = costoTotal + result.Productos[i].Cantidad*prod.Precio_Unitario
		}

	}

	json.NewEncoder(w).Encode(idCompra)

}

func agrega_producto(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	reqBody, _ := ioutil.ReadAll(r.Body)

	var res2 NewProduct

	json.Unmarshal([]byte(reqBody), &res2)
	result, err := db.Exec("INSERT INTO producto(nombre,cantidad_disponible,precio_unitario) VALUES (?,?,?)", res2.Nombre, res2.Cantidad_Disponible, res2.Precio_Unitario)

	if err != nil {
		panic(err.Error())
	}
	id, err := result.LastInsertId()
	if err != nil {
		panic(err)
	}
	var retorno Agregar_Producto
	retorno.ID_Producto = id
	json.NewEncoder(w).Encode(retorno)

}

func lista_productos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var productos []Product
	result, err := db.Query("SELECT id_producto, nombre,cantidad_disponible,precio_unitario from producto")
	if err != nil {
		panic(err.Error())
	}
	defer result.Close()
	for result.Next() {
		var product Product
		err := result.Scan(&product.ID, &product.Nombre, &product.Cantidad_Disponible, &product.Precio_Unitario)
		if err != nil {
			panic(err.Error())
		}
		productos = append(productos, product)
	}
	json.NewEncoder(w).Encode(productos)
}

func modifica_producto(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	reqBody, _ := ioutil.ReadAll(r.Body)

	vars := mux.Vars(r)
	key := vars["id"]
	intVar, err := strconv.Atoi(key)
	if err != nil {
		log.Fatal(err)
	}

	var res2 NewProduct

	json.Unmarshal([]byte(reqBody), &res2)

	result1, err := db.Exec(`UPDATE producto SET nombre = ? WHERE id_producto = ?`, res2.Nombre, intVar)
	if err != nil {
		fmt.Println(result1)
		panic(err.Error())
	}
	result2, err := db.Exec(`UPDATE producto SET cantidad_disponible = ? WHERE id_producto = ?`, res2.Cantidad_Disponible, intVar)
	if err != nil {
		fmt.Println(result2)
		panic(err.Error())
	}
	result3, err := db.Exec(`UPDATE producto SET precio_unitario = ? WHERE id_producto = ?`, res2.Precio_Unitario, intVar)
	if err != nil {
		fmt.Println(result3)
		panic(err.Error())
	}

	json.NewEncoder(w).Encode(intVar)
}

func quita_producto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["id"]
	intVar, err := strconv.Atoi(key)
	if err != nil {
		log.Fatal(err)
	}

	if productByID(int64(intVar)) {
		result, err := db.Exec(`DELETE FROM producto WHERE id_producto = ?`, intVar)
		if err != nil {
			fmt.Println(result)
			log.Fatal(err)
		}
		var delete Delete
		delete.ID_Producto = int64(intVar)
		json.NewEncoder(w).Encode(delete)
	}
}

func productByID(id int64) bool {
	var prod Product
	row := db.QueryRow("SELECT * FROM producto WHERE id_producto = ?", id)
	if err := row.Scan(&prod.ID, &prod.Nombre, &prod.Cantidad_Disponible, &prod.Precio_Unitario); err != nil {
		if err == sql.ErrNoRows {
			return false
		}
	}
	return true
}

func mostra_estadisticas(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var compras []Detalle
	rows, err := db.Query("SELECT * FROM detalle")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var compra Detalle
		if err := rows.Scan(&compra.ID_Compra, &compra.ID_Producto, &compra.Cantidad, &compra.Fecha); err != nil {
			log.Fatal(err)
		}
		compras = append(compras, compra)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	if len(compras) > 0 {
		ventas := make(map[int]int)
		for i := 0; i < len(compras); i++ {
			if ventas[int(compras[i].ID_Producto)] == 0 {
				ventas[int(compras[i].ID_Producto)] = int(compras[i].Cantidad)
			} else {
				ventas[int(compras[i].ID_Producto)] += int(compras[i].Cantidad)
			}
		}

		ganancias := make(map[int]int)
		for key, value := range ventas {
			rows, err := db.Query("SELECT * FROM producto WHERE id_producto = ?", key)
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()
			var prod Product
			for rows.Next() {
				if err := rows.Scan(&prod.ID, &prod.Nombre, &prod.Cantidad_Disponible, &prod.Precio_Unitario); err != nil {
					log.Fatal(err)
				}
				ganancias[key] = value * int(prod.Precio_Unitario)
			}
		}

		maxVentas, numeroMaximoDeVentas := max(ventas)
		minVentas := min(ventas, numeroMaximoDeVentas)
		maxGanancias, gananciasMaximas := max(ganancias)
		minGanancias := min(ganancias, gananciasMaximas)

		var estadisticas Estadisticas
		estadisticas.Producto_Mas_Vendido = int64(maxVentas)
		estadisticas.Producto_Menos_Vendido = int64(minVentas)
		estadisticas.Producto_Mayor_Ganancias = int64(maxGanancias)
		estadisticas.Producto_Menor_Ganancia = int64(minGanancias)
		json.NewEncoder(w).Encode(estadisticas)

	}
}

func get_compra(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	key := vars["id"]
	intVar, err := strconv.Atoi(key)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")

	rows, err := db.Query("SELECT * FROM detalle WHERE id_compra = ? ", intVar)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var numProd int64
	numProd = 0
	var costoTotal int64
	costoTotal = 0
	for rows.Next() {
		var compra Detalle
		if err := rows.Scan(&compra.ID_Compra, &compra.ID_Producto, &compra.Cantidad, &compra.Fecha); err != nil {
			log.Fatal(err)
		}

		rows2, err := db.Query("SELECT * FROM producto WHERE id_producto = ? ", compra.ID_Producto)
		if err != nil {
			log.Fatal(err)
		}
		defer rows2.Close()

		for rows2.Next() {
			var articulo Product
			if err := rows2.Scan(&articulo.ID, &articulo.Nombre, &articulo.Cantidad_Disponible, &articulo.Precio_Unitario); err != nil {
				log.Fatal(err)
			}
			numProd = numProd + compra.Cantidad
			costoTotal = costoTotal + compra.Cantidad*articulo.Precio_Unitario
		}
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	var retorno CompraDatos
	retorno.Cantidad = numProd
	retorno.Costo = costoTotal
	json.NewEncoder(w).Encode(retorno)
}

func max(datos map[int]int) (int, int) {
	var maxValue int
	var maxKey int
	for key, value := range datos {
		if value > maxValue {
			maxValue = value
			maxKey = key
		}
	}
	return maxKey, maxValue
}

func min(datos map[int]int, max int) int {
	minValue := max
	var minKey int
	for key, value := range datos {
		if value <= minValue {
			minValue = value
			minKey = key
		}
	}
	return minKey
}
