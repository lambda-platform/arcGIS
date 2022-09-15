package arcGIS

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lambda-platform/arcGIS/handlers"
	"github.com/lambda-platform/arcGIS/middleware"
	"github.com/lambda-platform/arcGIS/utils"
	"github.com/lambda-platform/lambda/agent/agentMW"
	"github.com/lambda-platform/lambda/config"
	"github.com/lambda-platform/lambda/dataform"
	"github.com/lambda-platform/lambda/datagrid"
)

func Set(e *fiber.App, GetGridMODEL func(schema_id string) datagrid.Datagrid, GetMODEL func(schema_id string) dataform.Dataform) {
	if config.Config.App.Migrate == "true" {
		utils.AutoMigrateSeed()
	}

	g := e.Group("/arcgis")

	g.Get("/fill", func(c *fiber.Ctx) error {
		return handlers.FillArcGISData(c, GetGridMODEL, GetMODEL)
	})
	g.Get("/token", agentMW.IsLoggedIn(), handlers.Token)

	g.Post("/form-fields", handlers.FormFields)

}

func MW(GetGridMODEL func(schema_id string) datagrid.Datagrid, GetMODEL func(schema_id string) dataform.Dataform) fiber.Handler {

	return func(c *fiber.Ctx) error {
		return middleware.BodyDump(c, GetGridMODEL, GetMODEL)
	}

}
