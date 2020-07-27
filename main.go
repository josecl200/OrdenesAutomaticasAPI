package main
import(
	"context"
	"fmt"
	"log"
	"time"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"github.com/gorilla/mux"
)


var Client *mongo.Client

func ConnectDatabase() {
	log.Println("Database connecting...")
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	Client = client
	if err != nil {
		log.Fatal(err)
	}
	err = Client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Database Connected.")
}

func main(){
	route := mux.NewRouter();
	ConnectDatabase();
	route.Use(jsonHead);
	route.HandleFunc("/",func(w http.ResponseWriter, r *http.Request){
		fmt.Fprintf(w,"{\"Activo\":\"Si.\"}");
	}).Methods("GET");
	artSuplRoutes:=route.PathPrefix("/artSupl").Subrouter();
	artSuplRoutes.HandleFunc("/",getAllArticuloSuplidor).Methods("GET");
	artSuplRoutes.HandleFunc("/",createArticuloSuplidor).Methods("POST");
	artSuplRoutes.HandleFunc("/delete/{codigoSupl}",deleteArticuloSuplidor).Methods("DELETE");
	artRoutes:=route.PathPrefix("/articulo").Subrouter();
	artRoutes.HandleFunc("/",getAllArticulo).Methods("GET");
	artRoutes.HandleFunc("/{codigoArticulo}",getAllArticulo).Methods("GET");
	artRoutes.HandleFunc("/",createArticulo).Methods("POST");
	artRoutes.HandleFunc("/delete/{codigoArt}",deleteArticulo).Methods("DELETE");
	suplRoutes:=route.PathPrefix("/suplidor").Subrouter();
	suplRoutes.HandleFunc("/",getAllSuplidores).Methods("GET");
	suplRoutes.HandleFunc("/",createSuplidor).Methods("POST");
	suplRoutes.HandleFunc("/delete/{codigoSuplidor}",deleteArticulo).Methods("DELETE");
	movRoutes:=route.PathPrefix("/moviemiento").Subrouter();
	movRoutes.HandleFunc("/",getAllMovimientos).Methods("GET");
	movRoutes.HandleFunc("/{codigoArticulo}",getAllMovimientos).Methods("GET");
	movRoutes.HandleFunc("/",createMovimiento).Methods("POST");
	orderRoutes:=route.PathPrefix("/orden").Subrouter();
	orderRoutes.HandleFunc("/",getAllOrdenes).Methods("GET")
	orderRoutes.HandleFunc("/",createOrdenes).Methods("POST");
	
	http.ListenAndServe(":8089", route)
}

func jsonHead(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Add("Content-Type", "application/json")
        next.ServeHTTP(w, r)
    })
}

