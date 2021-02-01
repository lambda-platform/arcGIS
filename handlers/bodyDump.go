package handlers

import (
	"encoding/json"
	"strconv"
	"github.com/labstack/echo/v4"
)
type crudResponse struct {
	Data struct{
		ID        int     `gorm:"column:id;" json:"id"`
	} `json:"data"`
}

func BodyDump(c echo.Context, reqBody, resBody []byte, GetMODEL func(schema_id string) (string, interface{}), GetGridMODEL func(schema_id string) (interface{}, interface{}, string, string, interface{}, string)) {

	action := c.Param("action")
	if(action == "store" || action == "update" || action == "delete" || action == "edit"){
		RowId := ""
		schemaId, _ := strconv.ParseInt(c.Param("schemaId"), 10, 64)
		if(action == "store"){

			var response crudResponse
			if err := json.Unmarshal(resBody, &response); err != nil {
				panic(err)
			}
			RowId = strconv.Itoa(response.Data.ID)

		} else {
			RowId = c.Param("id")
		}


		if(action == "store" || action == "update"){
			SAVEGIS(reqBody, schemaId, action, RowId, GetMODEL)
		}else if(action == "delete"){
			DELTEGIS(reqBody, schemaId, action, RowId, GetGridMODEL)
		}

	}
	return

}
