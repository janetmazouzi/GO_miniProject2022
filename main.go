package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

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
	Password string
}

type Estadisticas struct {
	Producto_Mas_Vendido     int64
	Producto_Menos_Vendido   int64
	Producto_Mayor_Ganancias int64
	Producto_Menor_Ganancia  int64
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

type Despacho struct {
	Id_despacho int64
	Estado      string
	Id_compra   int64
}

func main() {

	fmt.Println("Bienvenido")

	var option int
	option = -1

	for option != 3 {
		fmt.Println("Opciones:")
		fmt.Println("1. Iniciar sesión como cliente")
		fmt.Println("2. Iniciar sesión como administrador")
		fmt.Println("3. Salir")
		fmt.Print("Ingrese una opción: ")
		fmt.Scanln(&option)
		switch option {
		case 1:
			var user Usuario
			fmt.Print("Ingrese su id: ")
			//fmt.Scanln(&user.ID)
			user.ID = 890
			fmt.Print("Ingrese su contraseña: ")
			//fmt.Scanln(&user.Password)
			user.Password = "password"
			b, err := json.Marshal(user)
			if err != nil {
				fmt.Println("error:", err)
			}
			request, error := http.NewRequest("POST", "http://localhost:5000/api/clientes/iniciar_sesion", bytes.NewBuffer(b))
			request.Header.Set("Content-Type", "application/json; charset=UTF-8")

			client := &http.Client{}
			response, error := client.Do(request)
			if error != nil {
				panic(error)
			}
			defer response.Body.Close()

			body, _ := ioutil.ReadAll(response.Body)
			var result map[string]bool
			json.Unmarshal([]byte(body), &result)

			if result["Acceso_Valido"] {
				fmt.Println("Inicio de sesión exitoso")

				var opcionCliente int
				opcionCliente = -1
				for opcionCliente != 4 {
					fmt.Println("\nOpciones:")
					fmt.Println("1. Ver lista de productos")
					fmt.Println("2. Hacer compra")
					fmt.Println("3. Consultar despacho")
					fmt.Println("4. Salir")
					fmt.Print("Ingrese una opción: ")
					fmt.Scanln(&opcionCliente)
					switch opcionCliente {
					case 1:
						seeProducts()

					case 2:
						buyProducts(int(user.ID))

					case 3:
						consultaDespacho()

					case 4:
						fmt.Println("Saliendo..")
						break
					default:
						fmt.Println("Ingrese una opción válida")
					}
				}
			} else {
				fmt.Println("Error, no hay ninguna coincidencia con los datos ingresados.")
			}

		case 2:
			var passwordAdmin int
			passwordAdmin = 0
			fmt.Print("Ingrese contraseña de administrador: ")

			fmt.Scanln(&passwordAdmin)
			if passwordAdmin == 1234 {
				fmt.Println("Inicio de sesión exitoso")
				var opcionAdmin int
				opcionAdmin = -1
				for opcionAdmin != 6 {
					fmt.Println("\nOpciones:")
					fmt.Println("1. Ver lista de productos")
					fmt.Println("2. Crear producto")
					fmt.Println("3. Eliminar producto")
					fmt.Println("4. Ver estadísticas")
					fmt.Println("5. Modificar producto")
					fmt.Println("6. Salir")
					fmt.Print("Ingrese una opción: ")
					fmt.Scanln(&opcionAdmin)
					switch opcionAdmin {
					case 1:
						seeProducts()
					case 2:
						addProduct()
					case 3:
						delProduct()
					case 4:
						seeStatistics()
					case 5:
						editProduct()
					case 6:
						fmt.Println("Saliendo...")
						break
					default:
						fmt.Println("Ingrese una opción válida")
					}
				}
			}
		case 3:
			fmt.Println("Saliendo...")
		default:
			fmt.Println("Ingrese una opción válida")
		}

	}
	fmt.Println("\nHasta luego!")
}

func seeProducts() {
	curl := exec.Command("curl", " http://localhost:5000/api/productos", "-H", "@{'Content-Type'='application/json'}", "http://localhost:5000/api/productos")
	out, err := curl.Output()
	if err != nil {
		fmt.Println("curl failed", err)
		return
	}
	var productos []map[string]interface{}
	err2 := json.Unmarshal([]byte(out), &productos)

	if err2 != nil {
		fmt.Println("JSON decode error!")
		return
	}
	if len(productos) == 0 {
		fmt.Println("No se ha añadido ningún producto")
	} else {
		for i := 0; i < len(productos); i++ {
			fmt.Printf("%v;%v;%v por unidad;%v disponibles\n", productos[i]["ID"], productos[i]["Nombre"], productos[i]["Precio_Unitario"], productos[i]["Cantidad_Disponible"])
		}
	}
}

func buyProducts(id int) {
	var cantidad int
	var req Compras
	var compras []Compra
	fmt.Print("Ingrese cantidad de productos a comprar: ")
	fmt.Scanln(&cantidad)
	count := 1
	for count <= cantidad {
		var id_cantidad string
		fmt.Printf("Ingrese producto %v par id-cantidad: ", count)
		fmt.Scanln(&id_cantidad)

		stringSplitted := strings.Split(id_cantidad, "-")
		id_producto, err := strconv.Atoi(stringSplitted[0])
		if err != nil {
			panic(err)
		}

		cantidad, err := strconv.Atoi(stringSplitted[1])
		if err != nil {
			panic(err)
		}

		var compra Compra
		compra.ID_Producto = int64(id_producto)
		compra.Cantidad = int64(cantidad)
		compras = append(compras, compra)

		count += 1
	}
	req.ID_Cliente = int64(id)
	req.Productos = compras
	b, err := json.Marshal(req)
	if err != nil {
		fmt.Println("error:", err)
	}
	request, error := http.NewRequest("POST", "http://localhost:5000/api/compras", bytes.NewBuffer(b))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		panic(error)
	}
	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	var result int64
	json.Unmarshal([]byte(body), &result) // Se entrega el id de la compra

	fmt.Println("Gracias por su compra!")
	printDatosCompra(result)
}

func consultaDespacho() {
	var idDespacho int64
	fmt.Print("Ingrese el ID del despacho: ")
	fmt.Scanln(&idDespacho)

	url := "http://localhost:5000/api/clientes/estado_despacho" + strconv.Itoa(int(idDespacho))
	request, error := http.NewRequest("GET", url, nil)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		panic(error)
	}
	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	var despacho Despacho
	json.Unmarshal([]byte(body), &despacho)

	fmt.Println(despacho.Id_compra)
	fmt.Println("El estado del despacho es : " + despacho.Estado)

	/*if len(despacho) == 0 {
		fmt.Println("Este despacho no se encuentra en la base de datos")
	} else {
		fmt.Println("Producto eliminado exitosamente")
	}*/

}

func delProduct() {
	var idProducto int64

	fmt.Print("Ingrese el ID del producto: ")
	fmt.Scanln(&idProducto)
	url := "http://localhost:5000/api/productos/" + strconv.Itoa(int(idProducto))

	request, error := http.NewRequest("DELETE", url, nil)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		panic(error)
	}
	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	var result map[string]int64
	json.Unmarshal([]byte(body), &result)

	if len(result) == 0 {
		fmt.Println("Este producto no se encuentra en la base de datos")
	} else {
		fmt.Println("Producto eliminado exitosamente")
	}
}

func seeStatistics() {
	curl := exec.Command("curl", " http://localhost:5000/api/estadisticas", "-H", "@{'Content-Type'='application/json'}", "http://localhost:5000/api/estadisticas")
	out, err1 := curl.Output()
	if err1 != nil {
		fmt.Println("curl failed", err1)
		return
	}
	var estadisticas map[string]int
	err2 := json.Unmarshal([]byte(out), &estadisticas)

	if err2 != nil {
		fmt.Println("No se han realizado compras")
	} else {
		fmt.Printf("Producto más vendido: %v\n", estadisticas["Producto_Mas_Vendido"])
		fmt.Printf("Producto menos vendido: %v\n", estadisticas["Producto_Menos_Vendido"])
		fmt.Printf("Producto con más ganancias: %v\n", estadisticas["Producto_Mayor_Ganancias"])
		fmt.Printf("Producto con menos ganancias: %v\n", estadisticas["Producto_Menor_Ganancia"])
	}
}

func addProduct() {
	var newProd NewProduct

	fmt.Print("Ingrese el nombre: ")
	fmt.Scanln(&newProd.Nombre)
	fmt.Print("Ingrese la disponibilidad: ")
	fmt.Scanln(&newProd.Cantidad_Disponible)
	fmt.Print("Ingrese el precio unitario: ")
	fmt.Scanln(&newProd.Precio_Unitario)

	b, err := json.Marshal(newProd)

	if err != nil {
		fmt.Println("error:", err)
	}

	request, error := http.NewRequest("POST", "http://localhost:5000/api/productos", bytes.NewBuffer(b))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}

	response, error := client.Do(request)

	if error != nil {
		panic(error)
	}

	defer response.Body.Close()

	fmt.Printf("Producto ingresado correctamente! \n")
}

func editProduct() {
	var newProd NewProduct
	var idProducto int64

	fmt.Print("Ingrese la id: ")
	fmt.Scanln(&idProducto)
	fmt.Print("Ingrese el nombre: ")
	fmt.Scanln(&newProd.Nombre)
	fmt.Print("Ingrese la disponibilidad: ")
	fmt.Scanln(&newProd.Cantidad_Disponible)
	fmt.Print("Ingrese el precio unitario: ")
	fmt.Scanln(&newProd.Precio_Unitario)

	b, err := json.Marshal(newProd)
	if err != nil {
		fmt.Println("error:", err)
	}

	url := "http://localhost:5000/api/productos/" + strconv.Itoa(int(idProducto))
	request, error := http.NewRequest("PUT", url, bytes.NewBuffer(b))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		panic(error)
	}
	defer response.Body.Close()

	fmt.Printf("Producto modificado correctamente! \n")

}

func printDatosCompra(idCompra int64) {

	url := "http://localhost:5000/api/compras/" + strconv.Itoa(int(idCompra))
	request, error := http.NewRequest("GET", url, nil)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		panic(error)
	}
	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	var compra CompraDatos
	json.Unmarshal([]byte(body), &compra)

	fmt.Println("Cantidad de productos comprados: " + strconv.Itoa(int(compra.Cantidad)))
	fmt.Println("Monto total de la compra: " + strconv.Itoa(int(compra.Costo)))
}
