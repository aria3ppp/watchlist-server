package repo

import (
	"fmt"
	"reflect"

	"github.com/aria3ppp/watchlist-server/internal/models"
)

var (
	RawSqlWhereTimeWatchedIsNull = fmt.Sprintf(
		"AND %s IS NULL",
		models.WatchfilmTableColumns.TimeWatched,
	)
	RawSqlWhereTimeWatchedIsNotNull = fmt.Sprintf(
		"AND %s IS NOT NULL",
		models.WatchfilmTableColumns.TimeWatched,
	)
	RawSqlWhereTimeWatchedEmptyClause = ""
)

var (
	watchfilmGetAllQuery = fmt.Sprintf(
		`SELECT %[1]s, %[2]s
		FROM %[3]s INNER JOIN %[4]s ON %[5]s = %[6]s
		WHERE %[7]s = $1 %[8]s
		ORDER BY %[9]s %[10]s
		OFFSET $2 LIMIT $3;`,
		/*1*/ columnsList(models.WatchfilmTableColumns),
		/*2*/ columnsList(models.FilmTableColumns),
		/*3*/ models.TableNames.Watchfilms,
		/*4*/ models.TableNames.Films,
		/*5*/ models.WatchfilmTableColumns.FilmID,
		/*6*/ models.FilmTableColumns.ID,
		/*7*/ models.WatchfilmTableColumns.UserID,
		/*8: extended where clause*/ "%s",
		/*9*/ models.WatchfilmTableColumns.TimeAdded,
		/*10: sort order*/ "%s",
	)

	watchfilmCountQuery = fmt.Sprintf(
		`SELECT COUNT(*) FROM %[1]s WHERE %[2]s = $1 %[3]s;`,
		/*1*/ models.TableNames.Watchfilms,
		/*2*/ models.WatchfilmTableColumns.UserID,
		/*3: extended where clause*/ "%s",
	)

	tokenGetQuery = fmt.Sprintf(
		`WITH user_tokens_xyz AS (
			SELECT %[1]s FROM %[2]s WHERE %[3]s = $1 AND %[4]s > CURRENT_TIMESTAMP
		)
		SELECT %[1]s FROM user_tokens_xyz WHERE %[5]s = crypt($2, %[5]s);`,
		/*1*/ columnsList(models.TokenColumns),
		/*2*/ models.TableNames.Tokens,
		/*3*/ models.TokenColumns.UserID,
		/*4*/ models.TokenColumns.ExpiresAt,
		/*5*/ models.TokenColumns.TokenHash,
	)
)

func columnsList(tableColumnsStruct any) string {
	v := reflect.ValueOf(tableColumnsStruct)
	columns := ``
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldValue, isString := field.Interface().(string)
		if !isString {
			panic("columnsList: all fields must be of type string")
		}
		if i != 0 {
			columns += `, `
		}
		columns += fmt.Sprintf(`%[1]s AS "%[1]s"`, fieldValue)
	}
	return columns
}
