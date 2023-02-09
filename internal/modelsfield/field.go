package modelsfield

import (
	"reflect"

	"github.com/aria3ppp/watchlist-server/internal/models"
)

func Exists(model, field string) bool {
	_, exists := modelFields[model][field]
	return exists
}

func init() {
	// check all models are set
	for model := range fieldMap(models.TableNames) {
		if _, exists := modelFields[model]; !exists {
			panic(
				"modelsfield: model '" + model + "' is not set: provide corresponding 'fieldMap' value.",
			)
		}
	}
}

var modelFields = map[string]map[string]struct{}{
	models.TableNames.Users:         fieldMap(models.UserColumns),
	models.TableNames.Tokens:        fieldMap(models.TokenColumns),
	models.TableNames.Films:         fieldMap(models.FilmColumns),
	models.TableNames.FilmsAudit:    fieldMap(models.FilmsAuditColumns),
	models.TableNames.Serieses:      fieldMap(models.SeriesColumns),
	models.TableNames.SeriesesAudit: fieldMap(models.SeriesesAuditColumns),
	models.TableNames.Watchfilms:    fieldMap(models.WatchfilmColumns),
}

func fieldMap(modelColumnsStruct any) map[string]struct{} {
	fields := map[string]struct{}{}
	v := reflect.ValueOf(modelColumnsStruct)
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldValue, isString := field.Interface().(string)
		if !isString {
			panic("fieldMap: all fields must be of type string")
		}
		fields[fieldValue] = struct{}{}
	}
	return fields
}
