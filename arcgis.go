package arcGIS

import (
    "github.com/lambda-platform/arcGIS/handlers"
    "github.com/lambda-platform/arcGIS/utils"
    //"lambda/modules/agent/agentMW"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    vpUtils "github.com/lambda-platform/lambda/config"
)

func Set(e *echo.Echo, GetMODEL func(schema_id string) (string, interface{}), GetGridMODEL func(schema_id string) (interface{}, interface{}, string, string, interface{}, string)) {
	if vpUtils.Config.App.Migrate == "true"{
		utils.AutoMigrateSeed()
	}

	g :=e.Group("/arcgis")

    g.GET("/fill", func(c echo.Context) error {
        return handlers.FillArcGISData(c, GetMODEL ,GetGridMODEL)
    })

	g.POST("/form-fields", handlers.FormFields)


}


func MW(GetMODEL func(schema_id string) (string, interface{}), GetGridMODEL func(schema_id string) (interface{}, interface{}, string, string, interface{}, string)) echo.MiddlewareFunc{

	return  middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte){
		 handlers.BodyDump(c, reqBody, resBody, GetMODEL, GetGridMODEL)
	})

}
