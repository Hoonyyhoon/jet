package generator

import (
	"github.com/sub0Zero/go-sqlbuilder/generator/metadata"
	"path/filepath"
)

func generateDataModel(databaseInfo *metadata.DatabaseInfo, dirPath string) error {
	modelDirPath := filepath.Join(dirPath, databaseInfo.DatabaseName, databaseInfo.SchemaName, "model")

	err := ensureDirPath(modelDirPath)

	if err != nil {
		return err
	}

	for _, tableInfo := range databaseInfo.TableInfos {
		text, err := generateTemplate(DataModelTemplate, tableInfo)

		if err != nil {
			return err
		}

		err = saveGoFile(modelDirPath, tableInfo.Name, text)

		if err != nil {
			return err
		}
	}

	return nil
}