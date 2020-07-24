package main
import(
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"time"
	"github.com/gorilla/mux"
	"strconv"
	"encoding/json"
	//"io/ioutil"
	"strings"
)

func getAllMovimientos(w http.ResponseWriter, r *http.Request){
	colMovs:=Client.Database("ordenesCompra").Collection("MovimientoInventario");
	filtro:=bson.D{{}}
	codArt:=mux.Vars(r)["codigoArticulo"]
	if codArt!=""{
		filtro=bson.D{{"codigoarticulo",codArt}}
	}
	res,err:=colMovs.Find(context.TODO(),filtro);
	var resultados []bson.M;
	if err = res.All(context.TODO(), &resultados); err != nil {
		error,_:=json.Marshal(err);
		fmt.Fprintf(w,"{\"error\":\""+string(error)+"\"}");
	}
	json.NewEncoder(w).Encode(resultados)
}
func createMovimiento(w http.ResponseWriter, r *http.Request){
	movArt:=new(MovimientoInventario)
	errJson:=json.NewDecoder(r.Body).Decode(movArt);
	if errJson!=nil{
		w.WriteHeader(http.StatusInternalServerError);
		json.NewEncoder(w).Encode(errJson.Error());
	}
	cantMovs,err:=Client.Database("ordenesCompra").Collection("MovimientoInventario").CountDocuments(context.TODO(),bson.D{});
	if err!=nil{
		w.WriteHeader(http.StatusInternalServerError);
		json.NewEncoder(w).Encode(err.Error());
	}
	movArt.CodigoMovimiento=strconv.Itoa(int(cantMovs))
	colArt:=Client.Database("ordenesCompra").Collection("Articulo");
	colMovs:=Client.Database("ordenesCompra").Collection("MovimientoInventario");

	matchStage1:=bson.D{{"$match",bson.D{{"codigoarticulo",movArt.CodigoArticulo}}}}
	setStage:=bson.D{{"$set",bson.D{{"Cantidad",movArt.Cantidad},{"TipoMovimiento",movArt.TipoMovimiento},{"CodigoMovimiento",movArt.CodigoMovimiento}}}}
	projectStage:=bson.D{{"$project",bson.D{{"CodigoMovimiento",1},{"TipoMovimiento",1},{"codigoarticulo",1},{"Cantidad",1},{"unidad","$unidadcompra"},{"Fecha",time.Now()},{"Posible",bson.D{{"$cond",bson.A{bson.D{{"$in",bson.A{"$TipoMovimiento",bson.A{"/(Salida)/"}}}},bson.D{{"$cond",bson.A{bson.D{{"$gt",bson.A{bson.D{{"$subtract",bson.A{"$balanceactual","$Cantidad"}}},0}}},true,false}}},true}}}}}}}
	matchStage2:=bson.D{{"$match",bson.D{{"Posible",true}}}}
	//unsetStage:=bson.D{{"$unset","Posible"}}
	opts := options.Aggregate().SetMaxTime(2 * time.Second)
	res,err:=colArt.Aggregate(context.TODO(),mongo.Pipeline{matchStage1,setStage,projectStage,matchStage2},opts);

	if err!=nil{
		w.WriteHeader(http.StatusInternalServerError);
		json.NewEncoder(w).Encode(err.Error());
	}else{
		if strings.EqualFold(movArt.TipoMovimiento,"Salida"){
			movArt.Cantidad*=-1
		}
		res.Next(context.TODO())
		results:=new(MovimientoInventario)
		if err = res.Decode(results); err != nil{
			w.WriteHeader(http.StatusInternalServerError);
			json.NewEncoder(w).Encode(err);
		}else if results.CodigoArticulo==""{
			fmt.Printf("%+v",results)
			w.WriteHeader(http.StatusBadRequest);
			fmt.Fprintf(w,"{\"Error\":\"Operacion invalida\"}")
		}else{
			fmt.Printf("%+v",results)
			colMovs.InsertOne(context.TODO(),results)
			colArt.UpdateOne(context.TODO(),bson.D{{"codigoarticulo",movArt.CodigoArticulo}},bson.D{{"$inc",bson.D{{"balanceactual",movArt.Cantidad}}}})
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(results)
		}
	}
}
