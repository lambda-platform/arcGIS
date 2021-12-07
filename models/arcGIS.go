package models

import "time"

type GISConnection struct {
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
func (a *GISConnection) TableName() string {
	return "gis_connection"
}
type ArcGisResponse struct {
	Token   string `json:"token"`
	Expires int64  `json:"expires"`
}


type GISCategory struct {
    Active     int        `gorm:"column:active" json:"active"`
    CreatedAt  *time.Time `gorm:"column:created_at" json:"created_at"`
    Icon       string     `gorm:"column:icon" json:"icon"`
    ID         int        `gorm:"column:id;primary_key" json:"id"`
    LayerOrder int        `gorm:"column:layer_order" json:"layer_order"`
    MenuOrder  int        `gorm:"column:menu_order" json:"menu_order"`
    Name       string     `gorm:"column:name" json:"name"`
    Show       int        `gorm:"column:show" json:"show"`
    UpdatedAt  *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (a *GISCategory) TableName() string {
    return "gis_category"
}


type GISLayers struct {
    Active       int        `gorm:"column:active" json:"active"`
    CategoryID   int        `gorm:"column:category_id" json:"category_id"`
    CheckInluded int        `gorm:"column:check_inluded" json:"check_inluded"`
    CreatedAt    *time.Time `gorm:"column:created_at" json:"created_at"`
    Export       int        `gorm:"column:export" json:"export"`
    ID           int        `gorm:"column:id;primary_key" json:"id"`
    InfoTemplate string     `gorm:"column:info_template" json:"info_template"`
    LayerOrder   int        `gorm:"column:layer_order" json:"layer_order"`
    LayerType    string     `gorm:"column:layer_type" json:"layer_type"`
    LayerURL     string     `gorm:"column:layer_url" json:"layer_url"`
    MenuOrder    int        `gorm:"column:menu_order" json:"menu_order"`
    Name         string     `gorm:"column:name" json:"name"`
    PopupFields  string     `gorm:"column:popup_fields" json:"popup_fields"`
    SearchFields string     `gorm:"column:search_fields" json:"search_fields"`
    SearchInfo   string     `gorm:"column:search_info" json:"search_info"`
    Show         int        `gorm:"column:show" json:"show"`
    StyleField   string     `gorm:"column:style_field" json:"style_field"`
    UpdatedAt    *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (a *GISLayers) TableName() string {
    return "gis_layers"
}


type GISLegends struct {
    BorderColor string `gorm:"column:border_color" json:"border_color"`
    ElementType string `gorm:"column:element_type" json:"element_type"`
    FillColor   string `gorm:"column:fill_color" json:"fill_color"`
    Icon        string `gorm:"column:icon" json:"icon"`
    ID          int    `gorm:"column:id;primary_key" json:"id"`
    LayerID     int    `gorm:"column:layer_id" json:"layer_id"`
    LineType    string `gorm:"column:line_type" json:"line_type"`
    StyleType   string `gorm:"column:style_type" json:"style_type"`
    StyleValue  string `gorm:"column:style_value" json:"style_value"`
    Title       string `gorm:"column:title" json:"title"`
}

func (a *GISLegends) TableName() string {
    return "gis_legends"
}


type GISBaseMaps struct {
    ID        int    `gorm:"column:id;primary_key" json:"id"`
    Image     string `gorm:"column:image" json:"image"`
    IsDynamic int `gorm:"column:is_dynamic" json:"is_dynamic"`
    LayerName string `gorm:"column:layerName" json:"layerName"`
    MaxZoom   int `gorm:"column:maxZoom" json:"maxZoom"`
    MinZoom   int `gorm:"column:minZoom" json:"minZoom"`
    Show      int `gorm:"column:show" json:"show"`
    URL       string `gorm:"column:url" json:"url"`
}

func (b *GISBaseMaps) TableName() string {
    return "gis_base_maps"
}
