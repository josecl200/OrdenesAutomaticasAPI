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
