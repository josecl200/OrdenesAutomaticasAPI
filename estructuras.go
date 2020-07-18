package main
type MovimientoInventario struct{ 
	CodigoMovimiento string
	CodigoAlmacen string
	TipoMovimiento string
	CodigoArticulo string
	Cantidad int32
	Unidad string //lb,caja,etc
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
	codigoSuplidor string
	nombreSuplidor string
}

type ArticuloOrdenado struct{
	CodigoArticulo string
	CantidadOrdenada int32
	UnidadCompra string
	PrecioArticulo float64
}

type OrdenCompra struct{
	CodigoOrdenCompra string
	FechaRequerida datetime
	FechaGenerada datetime
	MontoTotal float64
	Articulos []ArticuloOrdenado
}
