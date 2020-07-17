package main
import(
	"context"
	"fmt"
	"log"
	"time"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	//"go.mongodb.org/mongo-driver/mongo/readpref"
	//"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"github.com/gorilla/mux"
	//"io/ioutil"
	//"strconv"
	//"encoding/json"
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
	route.HandleFunc("/",func(w http.ResponseWriter, r *http.Request){
		fmt.Fprintf(w,"{bret:\"Si.\"}");
	}).Methods("GET");
	route.HandleFunc("/insertArtSupl",crearArticuloSuplidor).Methods("POST");
	route.HandleFunc("/deleteArtSupl/{codigoSupl}",borrarArticuloSuplidor).Methods("DELETE")
	http.ListenAndServe(":8089", route)

}



