package middleware

import (
    "encoding/json"
    "github.com/gofiber/fiber/v2"
    "github.com/lambda-platform/arcGIS/handlers"
    "github.com/lambda-platform/lambda/dataform"
    "github.com/lambda-platform/lambda/datagrid"
    "github.com/lambda-platform/lambda/utils"
    "strconv"
)

type crudResponse struct {
    Data struct {
        ID int `gorm:"column:id;" json:"id"`
    } `json:"data"`
}

func BodyDump(c *fiber.Ctx, GetGridMODEL func(schema_id string) datagrid.Datagrid, GetMODEL func(schema_id string) dataform.Dataform) error {
    if err := c.Next(); err != nil {
        return err
    }
    action := c.Params("action")
    if c.Path() == "/lambda/krud/delete/:schemaId/:id" {
        action = "delete"
    }

    if action == "store" || action == "update" || action == "delete" {

        reqBody := utils.GetBody(c)
        RowId := ""
        schemaId, _ := strconv.ParseInt(c.Params("schemaId"), 10, 64)
        if action == "store" {

            var response crudResponse

            if err := json.Unmarshal(c.Response().Body(), &response); err != nil {
                panic(err)
            }
            RowId = strconv.Itoa(response.Data.ID)

        } else {
            RowId = c.Params("id")
        }

        if action == "store" || action == "update" {
            handlers.SAVEGIS(reqBody, schemaId, action, RowId, GetMODEL)
        } else if action == "delete" {
            handlers.DELTEGIS(reqBody, schemaId, action, RowId, GetGridMODEL)
        }

    }

    return nil
}
