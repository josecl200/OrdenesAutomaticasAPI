package main
import(
	"time"
)
type MovimientoInventario struct
{ 
	CodigoMovimiento string `bson:"codigomovimiento"`
	TipoMovimiento string `bson:"tipomovimiento"`
	CodigoArticulo string `bson:"codigoarticulo"`
	Cantidad int32 `bson:"cantidad"`
	Fecha time.Time
	Unidad string `bson:"unidad"`//lb,caja,etc
}

type ArticuloSuplidor struct{
	CodigoArticulo string
	CodigoSuplidor string
	TiempoEntrega int32
	PrecioArticulo float64
}

type Articulo struct{
	CodigoArticulo string
	Descripcion string
	BalanceActual int32
	UnidadCompra string
}

type Suplidor struct{
	CodigoSuplidor string
	NombreSuplidor string
}

type OrdenCompra struct
{
	CodigoOrdenCompra string
	FechaRequerida time.Time
	FechaGenerada time.Time
	FechaAOrdenar time.Time
	CodigoSuplidor string
	CodigoArticulo string
	CantidadOrdenada int32
	UnidadCompra string
	PrecioArticulo float64
	MontoTotal float64
}
