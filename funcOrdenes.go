package main
import(
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	//"github.com/gorilla/mux"
	"strconv"
	"encoding/json"
	//"io/ioutil"
	"time"
)

/*func toDoc(v []interface{}) (doc *bson.D, err error) {
    data, err := bson.Marshal(v)
    if err != nil {
        return
    }

    err = bson.Unmarshal(data, &doc)
    return
}*/

func getAllOrdenes(w http.ResponseWriter, r *http.Request){
	colSupl:=Client.Database("ordenesCompra").Collection("OrdenesCompra");
	res,err:=colSupl.Find(context.TODO(),bson.D{});
	var resultados []bson.M;
	if err = res.All(context.TODO(), &resultados); err != nil {
		error,_:=json.Marshal(err);
		fmt.Fprintf(w,"{\"error\":\""+string(error)+"\"}");
	}
	json.NewEncoder(w).Encode(resultados)
}

func consumoDiario(codArt string) float64{
	colMovs:=Client.Database("ordenesCompra").Collection("MovimientoInventario");
	matchStage:=bson.D{
		{"$match",bson.D{
			{"codigoarticulo",codArt},
			{"tipomovimiento",bson.D{
				{"$regex","/(salida)/i"}}}}}}
	groupStage:=bson.D{
		{"$group",bson.D{
			{"_id",bson.D{
				{"codigoarticulo","$codigoarticulo"},
				{"fechaConsumo",bson.D{
					{"$dateToString",bson.D{
						{"format","%Y-%m-%d"},
						{"date","$fecha"}}}}}}},
			{"consumoDelDia",bson.D{
				{"$sum","$cantidad"}}}}}}
	groupStage2:=bson.D{{"$group",bson.D{{"_id","$_id.codigoarticulo"},{"consumoDiario",bson.D{{"$avg","$consumoDelDia"}}}}}}
	opts := options.Aggregate().SetMaxTime(2 * time.Second)
	res,err:=colMovs.Aggregate(context.TODO(),mongo.Pipeline{matchStage,groupStage,groupStage2},opts);
	if err!=nil{
		fmt.Printf("Error: %+v",err);
	}
	var result []bson.M
	err=res.All(context.TODO(),&result)
	if err!=nil{
		fmt.Printf("%+v",err)
	}else{
		fmt.Printf("%+v",result)
	}

	if err!=nil{
		fmt.Printf("Error");
	}else{
		var consumo float64=0
		if len(result)>0{
			consumo=result[0]["consumoDiario"].(float64);
		}
		return consumo;
	}
	return 0;
}

func createOrdenes(w http.ResponseWriter, r *http.Request){
	var arts []map[string]interface{}
	errJson:=json.NewDecoder(r.Body).Decode(&arts);
	if errJson!=nil{
		w.WriteHeader(http.StatusInternalServerError);
		json.NewEncoder(w).Encode(errJson.Error());
	}
	colArts:=Client.Database("ordenesCompra").Collection("Articulo");
	var codigoArt string;
	var fechaRequerida time.Time;
	var cantRequerida int;
	opts := options.Aggregate().SetMaxTime(2 * time.Second)

	matchStage:=bson.D{{}}
	lookupStage:=bson.D{{"$lookup",bson.D{
		{"from","ArticuloSuplidor"},
		{"localField","codigoarticulo"},
		{"foreignField","codigoarticulo"},
		{"as","suplidores"}}}}
	unwindStage:=bson.D{{"$unwind","$suplidores"}}
	setStage:=bson.D{}
	possibleStage:=bson.D{{"$set",bson.D{{"posible",bson.D{{"$lte",bson.A{"$tiempoentrega","$tiempoentrega2"}}}}}}}
	matchPossible:=bson.D{{"$match",bson.D{{"posible",true}}}}
	sortStage:=bson.D{{"$sort",bson.D{{"precioarticulo",1}}}}
	limitSage:=bson.D{{"$limit",1}}
	projectStage:=bson.D{}
	setMontoTotalStage:=bson.D{{"$set",bson.D{{"MontoTotal",bson.D{{"$multiply",bson.A{"$PrecioArticulo","$CantidadOrdenada"}}}}}}}
	cantOrdenes,_:=Client.Database("ordenesCompra").Collection("OrdenesCompra").CountDocuments(context.TODO(),bson.D{});
	var ordenes []OrdenCompra
	for _, articulo:=range arts{
		codigoArt=articulo["codigoArticulo"].(string);
		fechaRequerida,_=time.Parse(time.RFC3339,articulo["fechaRequerida"].(string));
		cantRequerida=int(articulo["cantRequerida"].(float64));
		codigoOrden:=int(cantOrdenes)+int(1)+int(len(ordenes))
		fmt.Printf("CodigoOrden: %d",codigoOrden)
		cantDiaria:=consumoDiario(codigoArt)
		matchStage=bson.D{{"$match",bson.D{{"codigoarticulo",codigoArt}}}};
		setStage=bson.D{{"$set",bson.D{
			{"tiempoentrega","$suplidores.tiempoentrega"},
			{"tiempoentrega2",bson.D{{"$divide",bson.A{bson.D{{"$subtract",bson.A{fechaRequerida,time.Now()}}},1000*60*60*24}}}},
			{"precioarticulo","$suplidores.precioarticulo"}}}}
		projectStage=bson.D{
			{"$project",bson.D{
				{"CodigoOrdenCompra","ORD"+strconv.Itoa(codigoOrden)},
				{"FechaRequerida",fechaRequerida},
				{"FechaGenerada",time.Now()},
				{"FechaAOrdenar",bson.D{
					{"$subtract",bson.A{
						fechaRequerida,
						bson.D{
							{"$multiply",bson.A{
								"$suplidores.tiempoentrega",1000*3600*24}}}}}}},
				{"CodigoSuplidor","$suplidores.codigosuplidor"},
				{"CodigoArticulo","$codigoarticulo"},
				{"CantidadOrdenada",bson.D{
					{"$subtract",bson.A{
						float64(cantRequerida)-((time.Since(fechaRequerida).Hours()/float64(24))*cantDiaria),"$balanceactual"}}}},
				{"UnidadCompra","$unidadcompra"},
				{"PrecioArticulo","$suplidores.precioarticulo"}}}}
		res,err:=colArts.Aggregate(context.TODO(),mongo.Pipeline{matchStage,lookupStage,unwindStage,setStage,possibleStage,matchPossible,sortStage,limitSage,projectStage,setMontoTotalStage},opts)
		if err!=nil{
			fmt.Printf("error Aggregate: %+v",err)
			w.WriteHeader(http.StatusInternalServerError);
			json.NewEncoder(w).Encode(err);
		}
		fmt.Printf("%+v\n",res)
		res.Next(context.TODO())
		results:=new(OrdenCompra)
		if err = res.Decode(results); err != nil{
			fmt.Printf("Orden Vacia: %+v\n",err)
		}else{
			ordenes=append(ordenes,*results)
		}
	}
	colOrdenes:=Client.Database("ordenesCompra").Collection("OrdenesCompra");
	
	//json.NewEncoder(w).Encode(ordenes)
	errCount:=0
	for _,orden:=range ordenes{
		_,err:=colOrdenes.InsertOne(context.TODO(),orden)
		if err!=nil{
			w.WriteHeader(http.StatusInternalServerError);
			json.NewEncoder(w).Encode(err.Error());
			errCount++
			break;
		}else{
			fmt.Printf("Inserted order: %s",orden.CodigoOrdenCompra)
		}
	}
	if errCount==0{
		w.WriteHeader(http.StatusCreated);
		json.NewEncoder(w).Encode(ordenes)
		
	}
	

	


	/*res,err:=colSupl.InsertOne(context.TODO(),supl);
	if err!=nil{
		w.WriteHeader(http.StatusInternalServerError);
		json.NewEncoder(w).Encode(err.Error());
	}else{
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(res)
	}*/
}
