package handlers

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/grokify/html-strip-tags-go"
	"github.com/lambda-platform/arcGIS/models"
	"github.com/lambda-platform/lambda/DB"
	agentUtils "github.com/lambda-platform/lambda/agent/utils"
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
	"github.com/lambda-platform/lambda/dataform"
	"github.com/lambda-platform/lambda/datagrid"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func Token(c *fiber.Ctx) error {

	token := GetArcGISToken(c)

	configFile, err := os.Open("public/GIS/GISDATA.json")
	defer configFile.Close()
	byteValue, _ := ioutil.ReadAll(configFile)
	if err != nil {
		fmt.Println("GISDATA.json CONFIG FILE NOT FOUND")
	}
	User, err := agentUtils.AuthUserObject(c)

	if err != nil {
		c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":  err.Error(),
			"status": false,
		})
	}
	GISDATA := map[string]interface{}{}

	json.Unmarshal(byteValue, &GISDATA)
	return c.JSON(map[string]interface{}{
		"token":   token,
		"GISDATA": GISDATA,
		"User":    User,
	})
}
func GetArcGISToken(c *fiber.Ctx) models.ArcGisResponse {

	referer := os.Getenv("ARCGIS_REFER")
	server := os.Getenv("ARCGIS_SERVER")
	username := os.Getenv("ARCGIS_USER")
	password := os.Getenv("ARCGIS_USERPASSWORD")

	url := server + "/arcgis/tokens/generateToken"
	payload := strings.NewReader("username=" + username + "&password=" + password + "&client=referer&referer=" + referer + "&f=json&expiration=120")

	if referer == "requestip" {
		clientIP := c.IP()
		payload = strings.NewReader("username=" + username + "&password=" + password + "&client=ip&ip=" + clientIP + "&f=json&expiration=120")
	}

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	response := models.ArcGisResponse{}
	json.Unmarshal(body, &response)

	return response
}
func FillArcGISData(c *fiber.Ctx, GetGridMODEL func(schema_id string) datagrid.Datagrid, GetMODEL func(schema_id string) dataform.Dataform) error {

	var connections []models.GISConnection
	var result []models.FillResult
	DB.DB.Find(&connections)
	for _, connectionPre := range connections {
		fmt.Println(connectionPre.LocalForm)

		var connection Connection

		json.Unmarshal([]byte(connectionPre.Connection), &connection)

		grid := GetGridMODEL(strconv.Itoa(connectionPre.LocalGrid))
		rows := []models.GISROW{}
		DB.DB.Table(grid.MainTable).Select(fmt.Sprintf("%s as id, %s as object_id", grid.Identity, connection.ObjectIdField)).Where(connection.ObjectIdField+" <= ?", 0).Or(connection.ObjectIdField + " IS NULL").Find(&rows)

		for _, row := range rows {
			form := GetMODEL(strconv.Itoa(connectionPre.LocalForm))

			DB.DB.Where(form.Identity+" = ?", row.ID).Find(form.Model)

			data := make(map[string]interface{})
			dataPre, _ := json.Marshal(form.Model)
			json.Unmarshal(dataPre, &data)

			ArcGISSaveAction(data, connection, connectionPre, form.Model, form.Identity, strconv.Itoa(row.ID), form.Table)
		}

		result = append(result, models.FillResult{
			Layer: connectionPre.Title,
			Count: len(rows),
		})
		//return c.JSON(http.StatusOK, result)
	}
	return c.JSON(result)
}
func isNil(v interface{}) bool {
	return v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}
func ArcGISSaveAction(data map[string]interface{}, connection Connection, connectionPre models.GISConnection, Model interface{}, Identity string, rowId string, table string) map[string]interface{} {

	objectIDSTRING := fmt.Sprintf("%v", data[connection.ObjectIdField])

	objectID, _ := strconv.ParseInt(objectIDSTRING, 10, 64)

	attributes := ""

	if isNil(data[connection.GeoJsonField]) == false {
		geoData := data[connection.GeoJsonField].(string)

		if geoData != "" {
			for _, attribute := range connection.Connections {

				//fmt.Println(attribute["attribute"])

				if data[attribute["field"]] != nil {

					fieldValue := ""
					if reflect.TypeOf(data[attribute["field"]]).String() == "time.Time" {
						fieldValue = "\"" + reflect.ValueOf(data[attribute["field"]]).Interface().(time.Time).Format("2006-01-02") + "\""

					} else if reflect.TypeOf(data[attribute["field"]]).String() == "float64" {

						fieldValue = fmt.Sprintf("%.8f", data[attribute["field"]])

					} else if reflect.TypeOf(data[attribute["field"]]).String() == "float32" {

						fieldValue = fmt.Sprintf("%.8f", data[attribute["field"]])

					} else if reflect.TypeOf(data[attribute["field"]]).String() == "string" {

						var re = regexp.MustCompile(`"`)
						fieldValue = re.ReplaceAllString(reflect.ValueOf(data[attribute["field"]]).Interface().(string), `\"`)

						fieldValue = strip.StripTags(fieldValue)
						fieldValue = "\"" + fieldValue + "\""

					} else {
						fieldValue = "\"" + fmt.Sprintf("%v", data[attribute["field"]]) + "\""
					}
					var re = regexp.MustCompile(`\r?\n`)
					fieldValue = re.ReplaceAllString(fieldValue, `\n`)

					if attribute["attribute"] == "coord_z" {

						if fieldValue == "" || fieldValue == "\""+"\"" {
							fieldValue = "0"
						}

					}
					if fieldValue == "" {
						fieldValue = "\"" + "\""
					}

					if attributes == "" {

						attributes = "\"" + attribute["attribute"] + "\"" + ":" + fieldValue
					} else {

						attributes = attributes + "," + "\"" + attribute["attribute"] + "\"" + ":" + fieldValue
					}
				}

			}

			if objectID >= 1 {
				if attributes == "" {
					attributes = "\"OBJECTID\":" + fmt.Sprintf("%v", objectID)
				} else {
					attributes = attributes + ",\"OBJECTID\":" + fmt.Sprintf("%v", objectID)
				}
			}

			if connection.LayerType == "Point" {

				var coordinate Coordinate

				json.Unmarshal([]byte(geoData), &coordinate)

				err := json.Unmarshal([]byte(geoData), &coordinate)

				if err != nil {

					var coordinateFloat CoordinateFloat
					err2 := json.Unmarshal([]byte(geoData), &coordinateFloat)
					if err2 != nil {
						return data
					}

					features := ""
					if data["coord_z"] != nil {

						if data["coord_z"] == "" || data["coord_z"] == nil {
							data["coord_z"] = 0
						}
						if _, err := strconv.ParseInt(fmt.Sprintf("%v", data["coord_z"]), 10, 64); err != nil {
							data["coord_z"] = 0
						}
						features = fmt.Sprintf("[{\"geometry\":{\"x\":%v,\"y\":%v,\"z\":%v,\"spatialReference\":{\"wkid\":4326}},\"attributes\":{%v}}]", fmt.Sprintf("%v", coordinateFloat.Lng), fmt.Sprintf("%v", coordinateFloat.Lat), fmt.Sprintf("%v", data["coord_z"]), attributes)
					} else {
						features = fmt.Sprintf("[{\"geometry\":{\"x\":%v,\"y\":%v,\"spatialReference\":{\"wkid\":4326}},\"attributes\":{%v}}]", fmt.Sprintf("%v", coordinateFloat.Lng), fmt.Sprintf("%v", coordinateFloat.Lat), attributes)
					}

					layers := []string{}
					json.Unmarshal([]byte(connectionPre.Layer), &layers)

					if len(layers) >= 1 {

						layer_url := layers[len(layers)-1]
						objectIDNew := SaveArcGIS(layer_url, features, objectID)

						if objectIDNew >= 1 {
							//data[connection.ObjectIdField] = objectIDNew
							//data_, _ := json.Marshal(data)
							//json.Unmarshal(data_, Model)
							dataGIS := map[string]interface{}{}
							dataGIS[connection.ObjectIdField] = objectIDNew
							DB.DB.Table(table).Where(Identity+" = ?", rowId).Updates(dataGIS)

						}
					}
				} else {

					features := ""
					if data["coord_z"] != nil {
						if data["coord_z"] == "" || data["coord_z"] == nil {
							data["coord_z"] = 0
						}
						if _, err := strconv.ParseInt(fmt.Sprintf("%v", data["coord_z"]), 10, 64); err != nil {
							data["coord_z"] = 0
						}
						features = fmt.Sprintf("[{\"geometry\":{\"x\":%v,\"y\":%v,\"z\":%v,\"spatialReference\":{\"wkid\":4326}},\"attributes\":{%v}}]", fmt.Sprintf("%v", coordinate.Lng), fmt.Sprintf("%v", coordinate.Lat), fmt.Sprintf("%v", data["coord_z"]), attributes)
					} else {
						features = fmt.Sprintf("[{\"geometry\":{\"x\":%v,\"y\":%v,\"z\":0\"spatialReference\":{\"wkid\":4326}},\"attributes\":{%v}}]", fmt.Sprintf("%v", coordinate.Lng), fmt.Sprintf("%v", coordinate.Lat), attributes)
					}

					layers := []string{}
					json.Unmarshal([]byte(connectionPre.Layer), &layers)

					if len(layers) >= 1 {

						layer_url := layers[len(layers)-1]
						objectIDNew := SaveArcGIS(layer_url, features, objectID)

						if objectIDNew >= 1 {
							dataGIS := map[string]interface{}{}
							dataGIS[connection.ObjectIdField] = objectIDNew
							DB.DB.Table(table).Where(Identity+" = ?", rowId).Updates(dataGIS)

						}
					}
				}

			} else if connection.LayerType == "Polygon" || connection.LayerType == "Line" {
				var geoJson GeoJSon

				err := json.Unmarshal([]byte(geoData), &geoJson)
				if err != nil {
					return data
				}
				if len(geoJson.Features) >= 1 {
					rings := [][][]float64{}
					checkRings := []string{}
					for _, geoElement := range geoJson.Features {

						preString, _ := json.Marshal(geoElement.Geometry.Coordinates[0])
						_, found := Find(checkRings, string(preString))
						if !found {
							rings = append(rings, geoElement.Geometry.Coordinates[0])
							checkRings = append(checkRings, string(preString))
						}

					}
					Coordinates, _ := json.Marshal(rings)

					features := fmt.Sprintf("[{\"geometry\":{\"rings\":%v,\"spatialReference\":{\"wkid\":4326}},\"attributes\":{%v}}]", string(Coordinates), attributes)

					layers := []string{}
					json.Unmarshal([]byte(connectionPre.Layer), &layers)

					if len(layers) >= 1 {

						layer_url := layers[len(layers)-1]
						objectIDNew := SaveArcGIS(layer_url, features, objectID)

						if objectIDNew >= 1 {
							dataGIS := map[string]interface{}{}
							dataGIS[connection.ObjectIdField] = objectIDNew
							DB.DB.Table(table).Where(Identity+" = ?", rowId).Updates(dataGIS)

						}
					}
				}

			}
		}
	}

	return data

}
func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func SAVEGIS(rawData []byte, schemaId int64, action string, rowId string, GetMODEL func(schema_id string) dataform.Dataform) {

	dataJson := new(map[string]interface{})

	json.Unmarshal(rawData, dataJson)
	form := GetMODEL(strconv.Itoa(int(schemaId)))

	if action == "store" || action == "update" {
		(*dataJson)[form.Identity] = rowId

		ArcGISAfterSave(*dataJson, schemaId, form.Model, form.Identity, rowId, form.Table)

	}

}
func DELTEGIS(rawData []byte, schemaId int64, action string, rowId string, GetGridMODEL func(schema_id string) datagrid.Datagrid) {

	dataJson := new(map[string]interface{})

	json.Unmarshal(rawData, dataJson)

	grid := GetGridMODEL(strconv.Itoa(int(schemaId)))

	if action == "delete" {
		ArcGISAfterDelete(grid.MainModel, nil, rowId, strconv.Itoa(int(schemaId)), grid.Identity)
	}

}
func FormFields(c *fiber.Ctx) error {

	data := new(Data)
	if err := c.BodyParser(data); err != nil {

		return c.JSON(map[string]interface{}{
			"status": false,
		})
	}

	form_id := data.Model.LocalForm

	var schema Schema

	DB.DB.Table("vb_schemas").Where("id = ?", form_id).Find(&schema)

	var schemaData SchemaData

	json.Unmarshal([]byte(schema.Schema), &schemaData)

	fields := []map[string]string{}

	for _, field := range schemaData.Schema {

		fields = append(fields, map[string]string{
			"value": field["model"],
			"label": field["model"],
		})
	}

	Schemas := []interface{}{}

	connection := Field{
		Field: "connection",
		//Value: 1,
		Props: map[string]interface{}{
			"options": fields,
		},
	}
	Schemas = append(Schemas, connection)
	return c.JSON(map[string]interface{}{
		"status": true,
		"schema": Schemas,
	})

}
func ArcGISAfterSave(data map[string]interface{}, schemaId int64, Model interface{}, Identity string, rowId string, table string) map[string]interface{} {

	if schemaId >= 1 {

		formID := schemaId

		var connectionPre models.GISConnection
		DB.DB.Where("local_form = ?", formID).Find(&connectionPre)

		if connectionPre.Connection != "" {

			var connection Connection

			json.Unmarshal([]byte(connectionPre.Connection), &connection)

			return ArcGISSaveAction(data, connection, connectionPre, Model, Identity, rowId, table)

		}
	}

	return data

}
func ArcGISAfterDelete(Model interface{}, data []map[string]interface{}, id string, gridID string, Identity string) []map[string]interface{} {

	var connectionPre models.GISConnection
	DB.DB.Where("local_grid = ?", gridID).Find(&connectionPre)

	if connectionPre.Connection != "" {

		DB.DB.Unscoped().Where("id = ?", id).Find(Model)
		deletedData := map[string]interface{}{}

		data__, _ := json.Marshal(Model)
		json.Unmarshal(data__, &deletedData)

		var connection Connection

		json.Unmarshal([]byte(connectionPre.Connection), &connection)

		objectIDSTRING := fmt.Sprintf("%v", deletedData[connection.ObjectIdField])

		objectID, _ := strconv.ParseInt(objectIDSTRING, 10, 64)

		if objectID >= 1 {

			layers := []string{}
			json.Unmarshal([]byte(connectionPre.Layer), &layers)

			if len(layers) >= 1 {
				layer_url := layers[len(layers)-1]
				DeleteArcGIS(layer_url, objectID)
				//fmt.Println("objectID", objectID, "objectID")
			}

		}

	}

	return data

}
func GetArcGISTokenForBackEnd() models.ArcGisResponse {

	server := os.Getenv("ARCGIS_SERVER")
	username := os.Getenv("ARCGIS_USER")
	password := os.Getenv("ARCGIS_USERPASSWORD")

	url := server + "/arcgis/tokens/generateToken"

	payload := strings.NewReader("username=" + username + "&password=" + password + "&client=requestip&f=json&expiration=120")

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	response := models.ArcGisResponse{}
	json.Unmarshal(body, &response)

	return response
}

func CategoryAfterInsert(data interface{}) {

	GenerateGISDATA()

}
func GenerateGISDATA() {
	categories := []GISCategory{}
	baseMaps := []models.GISBaseMaps{}

	DB.DB.Order("menu_order").Where("active = ?", 1).Find(&categories)
	DB.DB.Find(&baseMaps)

	for i, _ := range categories {
		Children := []GISLayers{}
		DB.DB.Where("category_id = ?", categories[i].ID).Where("active = ?", 1).Order("layer_order").Find(&Children)

		for c, _ := range Children {
			Legends := []models.GISLegends{}
			DB.DB.Where("layer_id = ?", Children[c].ID).Find(&Legends)
			Children[c].Legends = Legends

			layerUrl := []string{}

			json.Unmarshal([]byte(Children[c].LayerURL), &layerUrl)

			Children[c].LayerURL = layerUrl[len(layerUrl)-1]
		}
		categories[i].Children = Children
	}

	geoData := map[string]interface{}{
		"baseMaps":   baseMaps,
		"categories": categories,
	}

	file, _ := json.MarshalIndent(geoData, "", " ")

	_ = ioutil.WriteFile("public/GIS/GISDATA.json", file, 0755)

	fmt.Println("SAVED")
}

func CategoryAfterDelete(data interface{}, datagrid datagrid.Datagrid, query *gorm.DB, c *fiber.Ctx) (interface{}, *gorm.DB, bool, bool) {

	GenerateGISDATA()

	return []interface{}{}, query, false, false

}

type GISCategory struct {
	Active     int         `gorm:"column:active" json:"active"`
	CreatedAt  *time.Time  `gorm:"column:created_at" json:"created_at"`
	Icon       string      `gorm:"column:icon" json:"icon"`
	ID         int         `gorm:"column:id;primary_key" json:"id"`
	LayerOrder int         `gorm:"column:layer_order" json:"layer_order"`
	MenuOrder  int         `gorm:"column:menu_order" json:"menu_order"`
	Name       string      `gorm:"column:name" json:"name"`
	Show       int         `gorm:"column:show" json:"show"`
	UpdatedAt  *time.Time  `gorm:"column:updated_at" json:"updated_at"`
	Children   []GISLayers `gorm:"-" json:"children"`
}

func (a *GISCategory) TableName() string {
	return "gis_category"
}

type GISLayers struct {
	Active       int                 `gorm:"column:active" json:"active"`
	CategoryID   int                 `gorm:"column:category_id" json:"category_id"`
	CheckInluded int                 `gorm:"column:check_inluded" json:"check_inluded"`
	CreatedAt    *time.Time          `gorm:"column:created_at" json:"created_at"`
	Export       int                 `gorm:"column:export" json:"export"`
	ID           int                 `gorm:"column:id;primary_key" json:"id"`
	InfoTemplate string              `gorm:"column:info_template" json:"info_template"`
	LayerOrder   int                 `gorm:"column:layer_order" json:"layer_order"`
	LayerType    string              `gorm:"column:layer_type" json:"layer_type"`
	LayerURL     string              `gorm:"column:layer_url" json:"layer_url"`
	MenuOrder    int                 `gorm:"column:menu_order" json:"menu_order"`
	Name         string              `gorm:"column:name" json:"name"`
	PopupFields  string              `gorm:"column:popup_fields" json:"popup_fields"`
	SearchFields string              `gorm:"column:search_fields" json:"search_fields"`
	SearchInfo   string              `gorm:"column:search_info" json:"search_info"`
	Show         int                 `gorm:"column:show" json:"show"`
	StyleField   string              `gorm:"column:style_field" json:"style_field"`
	UserRoles    string              `gorm:"column:user_roles" json:"user_roles"`
	UpdatedAt    *time.Time          `gorm:"column:updated_at" json:"updated_at"`
	Legends      []models.GISLegends `gorm:"-" json:"legends"`
}

func (a *GISLayers) TableName() string {
	return "gis_layers"
}

type Adds struct {
	Adds  string `url:"adds"`
	f     string `url:"f"`
	Token string `url:"token"`
}

type Updateds struct {
	Updates string `url:"updates"`
	f       string `url:"f"`
	Token   string `url:"token"`
}

func SaveArcGIS(reQuestUrl string, features string, objectID int64) int64 {

	reQuestUrl = reQuestUrl + "/applyEdits"

	token := GetArcGISTokenForBackEnd()

	action := "adds"

	if objectID >= 1 {
		action = "updates"
	}

	//fmt.Println(features)
	//sendData := fmt.Sprintf("%s=%s&f=pjson&token=%s", action, url.QueryEscape(features), "4bTnpNQ8yD7OHh8mzWT_IFM9LMgv2-h_ZxRjZRCLSQPH17_MgcsTsnUAujF4Hnwo")
	sendData := fmt.Sprintf("%s=%s&f=pjson&token=%s", action, features, token.Token)
	req, _ := http.NewRequest("POST", reQuestUrl, bytes.NewBufferString(sendData))

	req.Header.Add("content-type", "application/x-www-form-urlencoded; param=value")
	req.Header.Add("cache-control", "no-cache")

	//http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
	}
	client := &http.Client{Transport: transCfg}

	res, error_error := client.Do(req)
	//fmt.Println("====================================")
	//fmt.Println("CLIENT====================================")
	fmt.Println(error_error)
	//fmt.Println("====================================")
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	var result ArcGISResult
	json.Unmarshal([]byte(body), &result)

	if len(result.AddResults) >= 1 {
		if result.AddResults[0].Success == true {
			return result.AddResults[0].ObjectId
		} else {
			fmt.Println(string(body))
			return 0
		}

	} else if len(result.UpdateResults) >= 1 {
		if result.UpdateResults[0].Success == true {
			return result.UpdateResults[0].ObjectId
		} else {
			fmt.Println(string(body))
			return 0
		}
	} else {
		fmt.Println(string(body))
		return 0
	}
}
func DeleteArcGIS(reQuestUrl string, objectID int64) {

	reQuestUrl = reQuestUrl + "/deleteFeatures"

	data := url.Values{}

	if objectID >= 1 {
		token := GetArcGISTokenForBackEnd()
		data.Set("token", token.Token)
		data.Add("f", "pjson")
		data.Add("objectIds", fmt.Sprintf("%v", objectID))

		req, _ := http.NewRequest("POST", reQuestUrl, bytes.NewBufferString(data.Encode()))

		req.Header.Add("content-type", "application/x-www-form-urlencoded; param=value")
		req.Header.Add("cache-control", "no-cache")

		res, _ := http.DefaultClient.Do(req)

		defer res.Body.Close()
		//body, _ := ioutil.ReadAll(res.Body)

		//fmt.Println(string(body))

	}

}

type ArcGISResult struct {
	AddResults    []ObjectResult `json:"addResults"`
	UpdateResults []ObjectResult `json:"updateResults"`
}
type ObjectResult struct {
	Success  bool  `json:"success"`
	ObjectId int64 `json:"objectId"`
}
type Connection struct {
	Connections   []map[string]string `json:"connections"`
	ObjectIdField string              `json:"objectIdField"`
	GeoJsonField  string              `json:"geoJsonField"`
	LayerType     string              `json:"layerType"`
}
type Coordinate struct {
	Lat string `json:"lat"`
	Lng string `json:"lng"`
}
type CoordinateFloat struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type GeoJSon struct {
	Type     string `json:"type"`
	Features []struct {
		Type       string `json:"type"`
		Properties struct {
			string `json:""`
		} `json:"properties"`
		Geometry struct {
			Type        string        `json:"type"`
			Coordinates [][][]float64 `json:"coordinates"`
		} `json:"geometry"`
	} `json:"features"`
}
type ArcGIS struct {
	Data interface{} `json:"data"`
}
type Schema struct {
	Schema string `json:"schema"`
}
type SchemaData struct {
	Schema []map[string]string `json:"schema"`
}
type Field struct {
	Field string                 `json:"field"`
	Props map[string]interface{} `json:"props"`
}
type Data struct {
	EditMode bool      `json:"editMode" form:"editMode"`
	Model    FormModel `json:"model" form:"model"`
}
type FormModel struct {
	LocalForm int `json:"local_form"`
}
