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
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func Set(e *fiber.App, GetGridMODEL func(schema_id string) datagrid.Datagrid, GetMODEL func(schema_id string) dataform.Dataform) {
	if config.Config.App.Migrate == "true" {
		utils.AutoMigrateSeed()
	}

	g := e.Group("/gis")

	g.Get("/fill", func(c *fiber.Ctx) error {
		return handlers.FillArcGISData(c, GetGridMODEL, GetMODEL)
	})
	g.Get("/token", agentMW.IsLoggedIn(), handlers.Token)

	g.Post("/form-fields", handlers.FormFields)

	target, _ := url.Parse("http://localhost:6080")

	// Create a reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(target)

	e.Use("/arcgis", func(c *fiber.Ctx) error {
		// Create a new request based on the original
		c.Response().Header.Set("Access-Control-Allow-Origin", "*")
		c.Request().Header.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, X-Requested-With, x-csrf-token, Accept-Language, Content-Length, Authorization, Accept-Encoding, Connection")
		c.Request().Header.Set("Access-Control-Allow-Methods", "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS")

		req := fasthttp.Request{}
		url := c.Request().URI()
		url.SetScheme(target.Scheme)
		url.SetHost(target.Host)
		req.SetRequestURI(url.String())
		req.Header.SetHost(target.Host)

		fasthttpadaptor.NewFastHTTPHandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Access-Control-Allow-Origin", "*")
			writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, X-Requested-With, x-csrf-token, Accept-Language, Content-Length, Authorization, Accept-Encoding, Connection")
			writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS")

			proxy.ServeHTTP(writer, request)

		})(c.Context())
		// Since we're writing directly to the ResponseWriter, return nil
		return nil
	})

}

func MW(GetGridMODEL func(schema_id string) datagrid.Datagrid, GetMODEL func(schema_id string) dataform.Dataform) fiber.Handler {

	return func(c *fiber.Ctx) error {
		return middleware.BodyDump(c, GetGridMODEL, GetMODEL)
	}

}
