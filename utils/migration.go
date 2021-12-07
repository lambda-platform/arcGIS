package utils

import (
	"encoding/json"
	"github.com/lambda-platform/lambda/DB"
	"github.com/lambda-platform/lambda/config"
	arcGISModels "github.com/lambda-platform/arcGIS/models"
    puzzleModels "github.com/lambda-platform/lambda/models"
	"os"
	"fmt"
)

func AutoMigrateSeed() {
	db := DB.DB

	if config.Config.App.Seed == "true" {

		var vbs []puzzleModels.VBSchema
		DB.DB.Where("name = ?", "Газрын зургийн сервис холболт").Find(&vbs)

		if len(vbs) <= 0 {
			AbsolutePath := AbsolutePath()
			db.AutoMigrate(
				&arcGISModels.GISConnection{},
				&arcGISModels.GISBaseMaps{},
				&arcGISModels.GISCategory{},
				&arcGISModels.GISLayers{},
				&arcGISModels.GISLegends{},
			)
			var vbs2 []puzzleModels.VBSchema

			dataFile2, err2 := os.Open(AbsolutePath + "initialData/vb_schemas.json")
			defer dataFile2.Close()
			if err2 != nil {
				fmt.Println("PUZZLE SEED ERROR")
			}
			jsonParser2 := json.NewDecoder(dataFile2)
			err := jsonParser2.Decode(&vbs2)
			if err != nil {
				fmt.Println(err)
				fmt.Println("PUZZLE SEED DATA ERROR")
			}
			//fmt.Println(len(vbs))

			for _, vb := range vbs2 {

				DB.DB.Create(&vb)

			}
		}
	}

}
