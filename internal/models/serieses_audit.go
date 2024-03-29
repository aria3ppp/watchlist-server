// Code generated by SQLBoiler 4.13.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// SeriesesAudit is an object representing the database table.
type SeriesesAudit struct {
	ID            int         `db:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	Title         string      `db:"title" boil:"title" json:"title" toml:"title" yaml:"title"`
	Descriptions  null.String `db:"descriptions" boil:"descriptions" json:"descriptions,omitempty" toml:"descriptions" yaml:"descriptions,omitempty"`
	DateStarted   time.Time   `db:"date_started" boil:"date_started" json:"date_started" toml:"date_started" yaml:"date_started"`
	DateEnded     null.Time   `db:"date_ended" boil:"date_ended" json:"date_ended,omitempty" toml:"date_ended" yaml:"date_ended,omitempty"`
	Poster        null.String `db:"poster" boil:"poster" json:"poster,omitempty" toml:"poster" yaml:"poster,omitempty"`
	ContributedBy int         `db:"contributed_by" boil:"contributed_by" json:"contributed_by" toml:"contributed_by" yaml:"contributed_by"`
	ContributedAt time.Time   `db:"contributed_at" boil:"contributed_at" json:"contributed_at" toml:"contributed_at" yaml:"contributed_at"`
	Invalidation  null.String `db:"invalidation" boil:"invalidation" json:"invalidation,omitempty" toml:"invalidation" yaml:"invalidation,omitempty"`

	R *seriesesAuditR `db:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L seriesesAuditL  `db:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var SeriesesAuditColumns = struct {
	ID            string
	Title         string
	Descriptions  string
	DateStarted   string
	DateEnded     string
	Poster        string
	ContributedBy string
	ContributedAt string
	Invalidation  string
}{
	ID:            "id",
	Title:         "title",
	Descriptions:  "descriptions",
	DateStarted:   "date_started",
	DateEnded:     "date_ended",
	Poster:        "poster",
	ContributedBy: "contributed_by",
	ContributedAt: "contributed_at",
	Invalidation:  "invalidation",
}

var SeriesesAuditTableColumns = struct {
	ID            string
	Title         string
	Descriptions  string
	DateStarted   string
	DateEnded     string
	Poster        string
	ContributedBy string
	ContributedAt string
	Invalidation  string
}{
	ID:            "serieses_audit.id",
	Title:         "serieses_audit.title",
	Descriptions:  "serieses_audit.descriptions",
	DateStarted:   "serieses_audit.date_started",
	DateEnded:     "serieses_audit.date_ended",
	Poster:        "serieses_audit.poster",
	ContributedBy: "serieses_audit.contributed_by",
	ContributedAt: "serieses_audit.contributed_at",
	Invalidation:  "serieses_audit.invalidation",
}

// Generated where

var SeriesesAuditWhere = struct {
	ID            whereHelperint
	Title         whereHelperstring
	Descriptions  whereHelpernull_String
	DateStarted   whereHelpertime_Time
	DateEnded     whereHelpernull_Time
	Poster        whereHelpernull_String
	ContributedBy whereHelperint
	ContributedAt whereHelpertime_Time
	Invalidation  whereHelpernull_String
}{
	ID:            whereHelperint{field: "\"serieses_audit\".\"id\""},
	Title:         whereHelperstring{field: "\"serieses_audit\".\"title\""},
	Descriptions:  whereHelpernull_String{field: "\"serieses_audit\".\"descriptions\""},
	DateStarted:   whereHelpertime_Time{field: "\"serieses_audit\".\"date_started\""},
	DateEnded:     whereHelpernull_Time{field: "\"serieses_audit\".\"date_ended\""},
	Poster:        whereHelpernull_String{field: "\"serieses_audit\".\"poster\""},
	ContributedBy: whereHelperint{field: "\"serieses_audit\".\"contributed_by\""},
	ContributedAt: whereHelpertime_Time{field: "\"serieses_audit\".\"contributed_at\""},
	Invalidation:  whereHelpernull_String{field: "\"serieses_audit\".\"invalidation\""},
}

// SeriesesAuditRels is where relationship names are stored.
var SeriesesAuditRels = struct {
}{}

// seriesesAuditR is where relationships are stored.
type seriesesAuditR struct {
}

// NewStruct creates a new relationship struct
func (*seriesesAuditR) NewStruct() *seriesesAuditR {
	return &seriesesAuditR{}
}

// seriesesAuditL is where Load methods for each relationship are stored.
type seriesesAuditL struct{}

var (
	seriesesAuditAllColumns            = []string{"id", "title", "descriptions", "date_started", "date_ended", "poster", "contributed_by", "contributed_at", "invalidation"}
	seriesesAuditColumnsWithoutDefault = []string{"id", "title", "date_started", "contributed_by", "contributed_at"}
	seriesesAuditColumnsWithDefault    = []string{"descriptions", "date_ended", "poster", "invalidation"}
	seriesesAuditPrimaryKeyColumns     = []string{"id", "contributed_by", "contributed_at"}
	seriesesAuditGeneratedColumns      = []string{}
)

type (
	// SeriesesAuditSlice is an alias for a slice of pointers to SeriesesAudit.
	// This should almost always be used instead of []SeriesesAudit.
	SeriesesAuditSlice []*SeriesesAudit
	// SeriesesAuditHook is the signature for custom SeriesesAudit hook methods
	SeriesesAuditHook func(context.Context, boil.ContextExecutor, *SeriesesAudit) error

	seriesesAuditQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	seriesesAuditType                 = reflect.TypeOf(&SeriesesAudit{})
	seriesesAuditMapping              = queries.MakeStructMapping(seriesesAuditType)
	seriesesAuditPrimaryKeyMapping, _ = queries.BindMapping(seriesesAuditType, seriesesAuditMapping, seriesesAuditPrimaryKeyColumns)
	seriesesAuditInsertCacheMut       sync.RWMutex
	seriesesAuditInsertCache          = make(map[string]insertCache)
	seriesesAuditUpdateCacheMut       sync.RWMutex
	seriesesAuditUpdateCache          = make(map[string]updateCache)
	seriesesAuditUpsertCacheMut       sync.RWMutex
	seriesesAuditUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var seriesesAuditAfterSelectHooks []SeriesesAuditHook

var seriesesAuditBeforeInsertHooks []SeriesesAuditHook
var seriesesAuditAfterInsertHooks []SeriesesAuditHook

var seriesesAuditBeforeUpdateHooks []SeriesesAuditHook
var seriesesAuditAfterUpdateHooks []SeriesesAuditHook

var seriesesAuditBeforeDeleteHooks []SeriesesAuditHook
var seriesesAuditAfterDeleteHooks []SeriesesAuditHook

var seriesesAuditBeforeUpsertHooks []SeriesesAuditHook
var seriesesAuditAfterUpsertHooks []SeriesesAuditHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *SeriesesAudit) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range seriesesAuditAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *SeriesesAudit) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range seriesesAuditBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *SeriesesAudit) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range seriesesAuditAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *SeriesesAudit) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range seriesesAuditBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *SeriesesAudit) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range seriesesAuditAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *SeriesesAudit) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range seriesesAuditBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *SeriesesAudit) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range seriesesAuditAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *SeriesesAudit) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range seriesesAuditBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *SeriesesAudit) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range seriesesAuditAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddSeriesesAuditHook registers your hook function for all future operations.
func AddSeriesesAuditHook(hookPoint boil.HookPoint, seriesesAuditHook SeriesesAuditHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		seriesesAuditAfterSelectHooks = append(seriesesAuditAfterSelectHooks, seriesesAuditHook)
	case boil.BeforeInsertHook:
		seriesesAuditBeforeInsertHooks = append(seriesesAuditBeforeInsertHooks, seriesesAuditHook)
	case boil.AfterInsertHook:
		seriesesAuditAfterInsertHooks = append(seriesesAuditAfterInsertHooks, seriesesAuditHook)
	case boil.BeforeUpdateHook:
		seriesesAuditBeforeUpdateHooks = append(seriesesAuditBeforeUpdateHooks, seriesesAuditHook)
	case boil.AfterUpdateHook:
		seriesesAuditAfterUpdateHooks = append(seriesesAuditAfterUpdateHooks, seriesesAuditHook)
	case boil.BeforeDeleteHook:
		seriesesAuditBeforeDeleteHooks = append(seriesesAuditBeforeDeleteHooks, seriesesAuditHook)
	case boil.AfterDeleteHook:
		seriesesAuditAfterDeleteHooks = append(seriesesAuditAfterDeleteHooks, seriesesAuditHook)
	case boil.BeforeUpsertHook:
		seriesesAuditBeforeUpsertHooks = append(seriesesAuditBeforeUpsertHooks, seriesesAuditHook)
	case boil.AfterUpsertHook:
		seriesesAuditAfterUpsertHooks = append(seriesesAuditAfterUpsertHooks, seriesesAuditHook)
	}
}

// One returns a single seriesesAudit record from the query.
func (q seriesesAuditQuery) One(ctx context.Context, exec boil.ContextExecutor) (*SeriesesAudit, error) {
	o := &SeriesesAudit{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for serieses_audit")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all SeriesesAudit records from the query.
func (q seriesesAuditQuery) All(ctx context.Context, exec boil.ContextExecutor) (SeriesesAuditSlice, error) {
	var o []*SeriesesAudit

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to SeriesesAudit slice")
	}

	if len(seriesesAuditAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all SeriesesAudit records in the query.
func (q seriesesAuditQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count serieses_audit rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q seriesesAuditQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if serieses_audit exists")
	}

	return count > 0, nil
}

// SeriesesAudits retrieves all the records using an executor.
func SeriesesAudits(mods ...qm.QueryMod) seriesesAuditQuery {
	mods = append(mods, qm.From("\"serieses_audit\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"serieses_audit\".*"})
	}

	return seriesesAuditQuery{q}
}

// FindSeriesesAudit retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindSeriesesAudit(ctx context.Context, exec boil.ContextExecutor, iD int, contributedBy int, contributedAt time.Time, selectCols ...string) (*SeriesesAudit, error) {
	seriesesAuditObj := &SeriesesAudit{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"serieses_audit\" where \"id\"=$1 AND \"contributed_by\"=$2 AND \"contributed_at\"=$3", sel,
	)

	q := queries.Raw(query, iD, contributedBy, contributedAt)

	err := q.Bind(ctx, exec, seriesesAuditObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from serieses_audit")
	}

	if err = seriesesAuditObj.doAfterSelectHooks(ctx, exec); err != nil {
		return seriesesAuditObj, err
	}

	return seriesesAuditObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *SeriesesAudit) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no serieses_audit provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(seriesesAuditColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	seriesesAuditInsertCacheMut.RLock()
	cache, cached := seriesesAuditInsertCache[key]
	seriesesAuditInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			seriesesAuditAllColumns,
			seriesesAuditColumnsWithDefault,
			seriesesAuditColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(seriesesAuditType, seriesesAuditMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(seriesesAuditType, seriesesAuditMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"serieses_audit\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"serieses_audit\" %sDEFAULT VALUES%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			queryReturning = fmt.Sprintf(" RETURNING \"%s\"", strings.Join(returnColumns, "\",\""))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}

	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}

	if err != nil {
		return errors.Wrap(err, "models: unable to insert into serieses_audit")
	}

	if !cached {
		seriesesAuditInsertCacheMut.Lock()
		seriesesAuditInsertCache[key] = cache
		seriesesAuditInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the SeriesesAudit.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *SeriesesAudit) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	seriesesAuditUpdateCacheMut.RLock()
	cache, cached := seriesesAuditUpdateCache[key]
	seriesesAuditUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			seriesesAuditAllColumns,
			seriesesAuditPrimaryKeyColumns,
		)
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update serieses_audit, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"serieses_audit\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, seriesesAuditPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(seriesesAuditType, seriesesAuditMapping, append(wl, seriesesAuditPrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, values)
	}
	var result sql.Result
	result, err = exec.ExecContext(ctx, cache.query, values...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update serieses_audit row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for serieses_audit")
	}

	if !cached {
		seriesesAuditUpdateCacheMut.Lock()
		seriesesAuditUpdateCache[key] = cache
		seriesesAuditUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q seriesesAuditQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for serieses_audit")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for serieses_audit")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o SeriesesAuditSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("models: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), seriesesAuditPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"serieses_audit\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, seriesesAuditPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in seriesesAudit slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all seriesesAudit")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *SeriesesAudit) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no serieses_audit provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(seriesesAuditColumnsWithDefault, o)

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	if updateOnConflict {
		buf.WriteByte('t')
	} else {
		buf.WriteByte('f')
	}
	buf.WriteByte('.')
	for _, c := range conflictColumns {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(updateColumns.Kind))
	for _, c := range updateColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(insertColumns.Kind))
	for _, c := range insertColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	seriesesAuditUpsertCacheMut.RLock()
	cache, cached := seriesesAuditUpsertCache[key]
	seriesesAuditUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			seriesesAuditAllColumns,
			seriesesAuditColumnsWithDefault,
			seriesesAuditColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			seriesesAuditAllColumns,
			seriesesAuditPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert serieses_audit, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(seriesesAuditPrimaryKeyColumns))
			copy(conflict, seriesesAuditPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"serieses_audit\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(seriesesAuditType, seriesesAuditMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(seriesesAuditType, seriesesAuditMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}
	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(returns...)
		if errors.Is(err, sql.ErrNoRows) {
			err = nil // Postgres doesn't return anything when there's no update
		}
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}
	if err != nil {
		return errors.Wrap(err, "models: unable to upsert serieses_audit")
	}

	if !cached {
		seriesesAuditUpsertCacheMut.Lock()
		seriesesAuditUpsertCache[key] = cache
		seriesesAuditUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single SeriesesAudit record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *SeriesesAudit) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no SeriesesAudit provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), seriesesAuditPrimaryKeyMapping)
	sql := "DELETE FROM \"serieses_audit\" WHERE \"id\"=$1 AND \"contributed_by\"=$2 AND \"contributed_at\"=$3"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from serieses_audit")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for serieses_audit")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q seriesesAuditQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no seriesesAuditQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from serieses_audit")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for serieses_audit")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o SeriesesAuditSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(seriesesAuditBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), seriesesAuditPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"serieses_audit\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, seriesesAuditPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from seriesesAudit slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for serieses_audit")
	}

	if len(seriesesAuditAfterDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *SeriesesAudit) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindSeriesesAudit(ctx, exec, o.ID, o.ContributedBy, o.ContributedAt)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *SeriesesAuditSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := SeriesesAuditSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), seriesesAuditPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"serieses_audit\".* FROM \"serieses_audit\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, seriesesAuditPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in SeriesesAuditSlice")
	}

	*o = slice

	return nil
}

// SeriesesAuditExists checks if the SeriesesAudit row exists.
func SeriesesAuditExists(ctx context.Context, exec boil.ContextExecutor, iD int, contributedBy int, contributedAt time.Time) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"serieses_audit\" where \"id\"=$1 AND \"contributed_by\"=$2 AND \"contributed_at\"=$3 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD, contributedBy, contributedAt)
	}
	row := exec.QueryRowContext(ctx, sql, iD, contributedBy, contributedAt)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if serieses_audit exists")
	}

	return exists, nil
}
