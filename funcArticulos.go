package main
import(
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"github.com/gorilla/mux"
	"encoding/json"
)

func getAllArticulo(w http.ResponseWriter, r *http.Request){
	colArt:=Client.Database("ordenesCompra").Collection("Articulo");
	codArt:=mux.Vars(r)["codigoArticulo"]
	filtro:=bson.D{{}}
	if codArt!=""{
		filtro=bson.D{{"codigoarticulo",codArt}}
	}
	res,err:=colArt.Find(context.TODO(),filtro);
	var resultados []bson.M;
	if err = res.All(context.TODO(), &resultados); err != nil {
		error,_:=json.Marshal(err);
		fmt.Fprintf(w,"{\"error\":\""+string(error)+"\"}");
	}
	json.NewEncoder(w).Encode(resultados)
}

func createArticulo(w http.ResponseWriter, r *http.Request){
	art:=new(Articulo)
	err:=json.NewDecoder(r.Body).Decode(art)
	if err!=nil{
		w.WriteHeader(http.StatusInternalServerError);
		json.NewEncoder(w).Encode(err.Error());
	}
	colArt:=Client.Database("ordenesCompra").Collection("Articulo");
	fmt.Printf("%+v",art);
	res,err:=colArt.InsertOne(context.TODO(),art);
	if err!=nil{
		w.WriteHeader(http.StatusInternalServerError);
		json.NewEncoder(w).Encode(err.Error());
	}else{
		json.NewEncoder(w).Encode(res)
	}
}

func deleteArticulo(w http.ResponseWriter, r *http.Request){
	opts := options.Delete().SetCollation(&options.Collation{
		Locale:    "en_US",
		Strength:  1,
		CaseLevel: false,
	})
	colSuplidor:=Client.Database("ordenesCompra").Collection("Articulo");
	res,err:=colSuplidor.DeleteOne(context.TODO(),bson.D{{"codigoArticulo",mux.Vars(r)["codigoArt"]}},opts);
	fmt.Printf("%+v",res);
	if err!=nil{
		error,_:=json.Marshal(err);
		fmt.Fprintf(w,"{error:"+string(error)+"}");
	}else{
		json.NewEncoder(w).Encode(res)
	}
	
}
