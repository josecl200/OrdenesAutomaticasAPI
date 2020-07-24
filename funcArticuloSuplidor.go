package main
import(
	"context"
	"fmt"
	//"log"
	//"time"
	//"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	//"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"github.com/gorilla/mux"
	//"io/ioutil"
	//"strconv"
	"encoding/json"
)

func getAllArticuloSuplidor(w http.ResponseWriter, r *http.Request){
	colSuplidor:=Client.Database("ordenesCompra").Collection("ArticuloSuplidor");
	res,err:=colSuplidor.Find(context.TODO(),bson.D{});
	var resultados []bson.M;
	if err = res.All(context.TODO(), &resultados); err != nil {
		error,_:=json.Marshal(err);
		fmt.Fprintf(w,"{\"error\":\""+string(error)+"\"}");
	}
	//	fmt.Fprintf(w,"%++v\n",resultados);
	json.NewEncoder(w).Encode(resultados)

}

func createArticuloSuplidor(w http.ResponseWriter, r *http.Request){
	artSupl:=new(ArticuloSuplidor)
	err:=json.NewDecoder(r.Body).Decode(artSupl)
	if err!=nil{
		w.WriteHeader(http.StatusInternalServerError);
		json.NewEncoder(w).Encode(err.Error());
	}
	colSuplidor:=Client.Database("ordenesCompra").Collection("ArticuloSuplidor");
	fmt.Printf("%+v",artSupl);
	res,err:=colSuplidor.InsertOne(context.TODO(),artSupl);
	if err!=nil{
		w.WriteHeader(http.StatusInternalServerError);
		json.NewEncoder(w).Encode(err.Error());
	}else{
		json.NewEncoder(w).Encode(res)
	}
}

func deleteArticuloSuplidor(w http.ResponseWriter, r *http.Request){
	opts := options.Delete().SetCollation(&options.Collation{
		Locale:    "en_US",
		Strength:  1,
		CaseLevel: false,
	})
	
	colSuplidor:=Client.Database("ordenesCompra").Collection("ArticuloSuplidor");
	res,err:=colSuplidor.DeleteOne(context.TODO(),bson.D{{"codigosuplidor",mux.Vars(r)["codigoSupl"]}},opts);
	fmt.Printf("%+v",res);
	if err!=nil{
		error,_:=json.Marshal(err);
		fmt.Fprintf(w,"{error:"+string(error)+"}");
	}else{
		json.NewEncoder(w).Encode(res)
	}
	
}
