package main
import(
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	//"github.com/gorilla/mux"
	//"strconv"
	"encoding/json"
	//"io/ioutil"
	"time"
)

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
	matchStage:=bson.D{{"$match",bson.D{{"codigoarticulo",codArt},{"tipomovimiento",bson.D{{"$regex","/(salida)/i"}}}}}}
	groupStage:=bson.D{{"$group",bson.D{{"_id",bson.D{{"codigoarticulo","$codigoarticulo"},{"fechaConsumo",bson.D{{"$dateToString",bson.D{{"format","%Y-%m-%d"},{"date","$fecha"}}}}}}},{"consumoDelDia",bson.D{{"$sum","$cantidad"}}}}}}
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
	lookupStage:=bson.D{{"$lookup",bson.D{{"from","ArticuloSuplidor"},{"localField","codigoarticulo"},{"foreignField","codigoarticulo"},{"as","suplidores"}}}}
	unwindStage:=bson.D{{"$unwind","$suplidores"}}
	matchDateStage:=bson.D{}
	sortStage:=bson.D{{"$sort",bson.D{{"suplidores.precioarticulo",1}}}}
	limitSage:=bson.D{{"$limit",1}}
	projectStage:=bson.D{}
	cantOrdenes,_:=Client.Database("ordenesCompra").Collection("OrdenesCompra").CountDocuments(context.TODO(),bson.D{});
	var ordenes []OrdenCompra
	for _, articulo:=range arts{
		//fmt.Printf("articulo: %+v",articulo)
		codigoArt=articulo["codigoArticulo"].(string);
		fechaRequerida,_=time.Parse(time.RFC3339,articulo["fechaRequerida"].(string));
		cantRequerida=int(articulo["cantRequerida"].(float64));
		
		cantDiaria:=consumoDiario(codigoArt)
		matchStage=bson.D{{"$match",bson.D{{"codigoarticulo",codigoArt}}}};
		matchDateStage=bson.D{{"$match",bson.D{{"suplidores.tiempoentrega",bson.D{{"$lte",bson.D{{"$divide",bson.A{bson.D{{"$subtract",bson.A{time.Now(),fechaRequerida}}},1000*60*60*24*365}}}}}}}}}
		projectStage=bson.D{
			{"$project",bson.D{
				{"CodigoOrden",int(cantOrdenes)+int(1)+int(len(ordenes))},
				{"FechaRequerida",fechaRequerida},
				{"FechaGenerada",time.Now()},
				{"FechaAOrdenar",bson.D{
					{"$subtract",bson.A{
						fechaRequerida,
						bson.D{
							{"$multiply",bson.A{
								"$suplidores.tiempoentrega",1000*3600*24}}}}}}},
				{"CodigoSuplidor","$suplidores.codigosuplidor"},
				{"CodigoArticulo","codigoarticulo"},
				{"CantidadOrdenada",bson.D{
					{"$subtract",bson.A{
						float64(cantRequerida)-((time.Since(fechaRequerida).Hours()/float64(24))*cantDiaria),"$balanceactual"}}}},
				{"UnidadCompra","$unidadcompra"},
				{"PrecioArticulo","$suplidores.precioarticulo"},
				{"MontoTotal",bson.D{{"$multiply",bson.A{
					"$suplidores.precioarticulo",bson.D{
						{"$subtract",bson.A{
							float64(cantRequerida)-((time.Since(fechaRequerida).Hours()/24)*cantDiaria),"$balanceactual"}}}}}}}}}}
		res,err:=colArts.Aggregate(context.TODO(),mongo.Pipeline{matchStage,lookupStage,unwindStage,matchDateStage,sortStage,limitSage,projectStage},opts)
		if err!=nil{
			fmt.Printf("error Aggregate: %+v",err)
			w.WriteHeader(http.StatusInternalServerError);
			json.NewEncoder(w).Encode(err);
		}else{
			var salidaAgg []bson.M
			res.All(context.TODO(),salidaAgg)
			fmt.Printf("salida aggregate: %+v",salidaAgg)
		}
		//fmt.Printf("%+v",res)
		//res.Next(context.TODO())
		results:=new(OrdenCompra)
		if err = res.Decode(results); err != nil{
			fmt.Printf("error Decode: %+v",err)

			w.WriteHeader(http.StatusInternalServerError);
			json.NewEncoder(w).Encode(err);
		}else{
			ordenes=append(ordenes,*results)
		}
	}
	fmt.Printf("%+v", ordenes)	
	/*res,err:=colSupl.InsertOne(context.TODO(),supl);
	if err!=nil{
		w.WriteHeader(http.StatusInternalServerError);
		json.NewEncoder(w).Encode(err.Error());
	}else{
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(res)
	}*/
}
