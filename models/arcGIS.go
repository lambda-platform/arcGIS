package models


type ArcgisConnection struct {
	Connection string `gorm:"column:connection;type:TEXT" json:"connection"`
	ID         int    `gorm:"column:id;primary_key" json:"id"`
	Layer      string `gorm:"column:layer;type:TEXT" json:"layer"`
	LocalForm  int `gorm:"column:local_form" json:"local_form"`
	LocalGrid  int `gorm:"column:local_grid" json:"local_grid"`
	Title      string `gorm:"column:title" json:"title"`
}
type FillResult struct {
    Layer      string `gorm:"column:layer;type:TEXT" json:"layer"`
    Count      int `gorm:"column:count" json:"count"`
}
type GISROW struct {
    ID         int    `gorm:"column:id;primary_key" json:"id"`
    ObjectID  int `gorm:"column:object_id" json:"object_id"`
}
//  TableName sets the insert table name for this struct type
func (a *ArcgisConnection) TableName() string {
	return "arcgis_connection"
}
type ArcGisResponse struct {
	Token   string `json:"token"`
	Expires int64  `json:"expires"`
}
