package postgres

import (
	"fmt"
	"path"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	custom "github.com/webitel/custom/data"
	customrel "github.com/webitel/custom/reflect"
)

// schema constants
const (
	// schema
	schemaCustom = "custom"
	// tables
	tableType  = "dataset"
	tableField = "field"
	// aliases
	aliasType  = "t" // "ds"
	aliasField = "f" // "fd"
	// common
	columnStatus = "err"
	// common
	columnDc        = "dc"
	columnId        = "id"
	columnVer       = "ver"
	columnName      = "name"
	columnTitle     = "title"
	columnUsage     = "usage"
	columnCreatedAt = "created_at"
	columnCreatedBy = "created_by"
	columnUpdatedAt = "updated_at"
	columnUpdatedBy = "updated_by"
	// dataset
	columnTypeId           = columnId
	columnTypeDir          = "dir"
	columnTypeName         = "scope" // UK // columnName
	columnTypePath         = "path"  // pseudo
	columnTypeTitle        = columnTitle
	columnTypeUsage        = columnUsage
	columnTypeExtendable   = "extendable"
	columnTypeFieldPrimary = "primary"
	columnTypeFieldDisplay = "display"
	// field
	columnFieldOf          = "of"       // PK
	columnFieldNum         = "num"      // position
	columnFieldName        = columnName // PK
	columnFieldTitle       = columnTitle
	columnFieldUsage       = columnUsage
	columnFieldTypeKind    = "kind" // NOTNULL ; kind of data type
	columnFieldTypeList    = "list" // NULL ; kind of list element ; ( kind = 'list' )
	columnFieldTypeLookup  = "rel"  // NULL ; lookup dataset id ; coalesce(list, kind) = 'lookup'
	columnFieldTypeSpec    = "type" // JSONB ; data type specification
	columnFieldDataDefault = "default"
	columnFieldDataAlways  = "always"
	columnFieldIsReadonly  = "readonly"
	columnFieldIsRequired  = "required"
	columnFieldIsDisabled  = "disabled"
	columnFieldIsHidden    = "hidden"
)

const (
	schemaDir = "directory"
	tableAuth = "wbt_auth"
)

// [1] - identifier::name
// [2] - schema.identifier
// [3] - database.schema.identifier
type sqlident []string

func (rel sqlident) String() string {
	return strings.Join([]string(rel), ".")
}

func (rel sqlident) Name() string {
	parts := []string(rel)
	if n := len(parts); n > 0 {
		return parts[n-1]
	}
	return ""
}

func (rel sqlident) Schema() string {
	parts := []string(rel)
	if n := len(parts); n > 1 {
		return parts[n-2]
	}
	return ""
}

// JUST return {column} name FOR SELECT
// MAY JOIN related table(s) if needed ...
//
// DO NOT include query.Column(..) inside this method,
// just return {sqlident} column relation name instead !
type columnQuery func(query SelectQ, left string, join names) (SelectQ, sqlident)

func customColumnName(name string) columnQuery {
	// // since field(s).name MAY be complex, like:
	// // - agents.display    ; "user.name"
	// // - contacts.display  ; "name.common_name"
	// // we just trim nested object(s) path
	// // name = lookupColumnName(name)
	// if dot := strings.LastIndexByte(name, '.'); dot >= 0 {
	// 	name = name[dot+1:]
	// }
	return func(query SelectQ, left string, _ names) (SelectQ, sqlident) {
		return query, sqlident{left, name}
	}
}

// FROM [directory.wbt_auth] AS [left]
func customRoleName(query SelectQ, left string, _ names) (SelectQ, sqlident) {
	return query, sqlident{fmt.Sprintf(
		"COALESCE(%[1]s.name,(%[1]s.auth)::::text,'[deleted]')", left,
	)}
}

// FROM [directory.wbt_user] AS [left]
func customUserName(query SelectQ, left string, _ names) (SelectQ, sqlident) {
	return query, sqlident{fmt.Sprintf(
		"COALESCE(%[1]s.name,(%[1]s.username)::::text,'[deleted]')", left,
	)}
}

// FROM [call_center.cc_agent] AS [left]
// LEFT JOIN [directory.wbt_user] AS "ua"
func customAgentName(query sq.SelectBuilder, left string, join names) (SelectQ, sqlident) {
	right := (left + "ua") // LEFT JOIN [directory.wbt_user] AS [alias]
	if _, ok := join[right]; !ok {
		joinUser := JOIN{
			Kind:   "LEFT JOIN",
			Source: sqlident{"directory", "wbt_user"}.String(),
			Alias:  right,
			Pred:   fmt.Sprintf("%s.user_id = %s.id", left, right),
		}
		query = query.JoinClause(&joinUser)
	}
	return customUserName(query, right, join)
}

// customTable map
type customTable struct {
	rel sqlident    // schema.table
	dc  string      // [primary] simple column name
	dn  columnQuery // [display] complex column name
	// dn  string   // [display] column name
}

var knownTableMap = map[string]customTable{ // string
	"roles":                          {rel: sqlident{"directory", "wbt_auth"}, dc: "dc", dn: customRoleName},                // "directory.wbt_auth",
	"users":                          {rel: sqlident{"directory", "wbt_user"}, dc: "dc", dn: customUserName},                // "directory.wbt_user",
	"cases":                          {rel: sqlident{"cases", "case"}, dc: "dc", dn: customColumnName("name")},              // "cases.case",
	"cases/priorities":               {rel: sqlident{"cases", "priority"}, dc: "dc", dn: customColumnName("name")},          // "cases.priority",
	"contacts":                       {rel: sqlident{"contacts", "contact"}, dc: "dc", dn: customColumnName("common_name")}, // "contacts.contact",
	"calendars":                      {rel: sqlident{"flow", "calendar"}, dc: "domain_id", dn: customColumnName("name")},    // flow.calendar
	"call_center/list":               {rel: sqlident{"call_center", "cc_list"}, dc: "domain_id", dn: customColumnName("name")},
	"call_center/agents":             {rel: sqlident{"call_center", "cc_agent"}, dc: "domain_id", dn: customAgentName},
	"call_center/queues":             {rel: sqlident{"call_center", "cc_queue"}, dc: "domain_id", dn: customColumnName("name")},
	"call_center/communication_type": {rel: sqlident{"call_center", "cc_communication"}, dc: "domain_id", dn: customColumnName("name")},
}

func init() {
	// also register UNIQUE type name without base directory
	for pkg, rel := range knownTableMap {
		dir, name := path.Split(pkg)
		if dir != "" && name != "" {
			if _, ok := knownTableMap[name]; ok {
				panic("custom: duplicate register known type table")
			}
			knownTableMap[name] = rel
		}
	}
}

// returns [schema.table] for given custom ( dc + pkg ) type identity
func customDatasetTable(of customrel.DatasetDescriptor) customTable {
	pkg := of.Path() // strings.ToLower(pkg)
	if of.Dc() < 1 {
		// [ GLOBAL ]
		regclass, known := knownTableMap[pkg]
		if known {
			return regclass // well-known type table relation
		}
	}
	// format custom type table relation
	return customTable{
		// [ CUSTOM ]
		rel: sqlident{
			schemaCustom,
			fmt.Sprintf("d%d_%s", of.Dc(), of.Name()),
		},
		// default
		dc: columnDc, // "dc"
		dn: customColumnName(
			of.Display().Name(),
		),
	}
}

func customSchemaError(err error) error {
	if err == nil {
		return nil
	}

	if e, is := err.(*pgconn.PgError); is {
		switch e.Code {
		case "23502": // not_null_violation
			// // Severity: "ERROR"
			// // Code: "23502"
			// // Message: "null value in column \"type_of\" of relation \"contact_phone\" violates not-null constraint"
			// // Detail: "Failing row contains (50, 1, 366, 2023-10-04 10:33:18.016654, 3, 2023-10-04 10:33:18.016654, 3, 0, t, t, 4, null, null, 444444)."
			// // SchemaName: "contacts"
			// // TableName: "contact_phone"
			// // ColumnName: "type_of"
			// // DataTypeName: ""
			// // ConstraintName: ""
			// switch re.TableName {
			// case contactEmailTable: // contact_email
			// 	switch re.ColumnName {
			// 	case "type_of":
			// 		err = model.ErrBadRequest(
			// 			"contacts.emails.type.required",
			// 			"contacts: emails( type: ); input: invalid or missing",
			// 		)
			// 	}
			// case contactPhoneTable: // contact_phone
			// 	switch re.ColumnName {
			// 	case "type_of":
			// 		err = model.ErrBadRequest(
			// 			"contacts.phones.type.required",
			// 			"contacts: phones( type: ); input: invalid or missing",
			// 		)
			// 	}
			// }
		case "23503": // foreign_key_violation
			// switch re.ConstraintName {
			// // case "wbt_device_user_fk":       // Key (user_id, dc) not present in wbt_auth
			// // 	ctx.Err = errors.BadRequest("app.user.not_found.error", "user: not found")
			// case "wbt_user_created_by_fk": // Key (created_by, dc) not present in wbt_auth
			// case "wbt_user_updated_by_fk": // Key (updated_by, dc) not present in wbt_auth
			// // TABLE contacts.contact_manager
			// // CONSTRAINT (user_id, dc)
			// // REFERENCES directory.wbt_user(id, dc)
			// case "contact_manager_user_fk":
			// 	err = model.ConflictError(
			// 		"contacts.managers.conflict",
			// 		"contacts: no such user; "+re.Detail,
			// 	)
			// // TABLE contacts.contact_timezone
			// // CONSTRAINT (timezone_id)
			// // REFERENCES flow.calendar_timezones(id)
			// case "contact_timezone_timezone_id_fk":
			// 	err = model.ConflictError(
			// 		"contacts.timezones.timezone.id.conflict",
			// 		"contacts: no such timezone; "+re.Detail,
			// 	)
			// case "contact_imclient_app_fk":
			// 	err = model.ConflictError(
			// 		"contacts.imclients.app.id.conflict",
			// 		"contacts: no such gateway; "+re.Detail,
			// 	)
			// case "contact_imclient_fk":
			// 	err = model.ConflictError(
			// 		"contacts.imclients.contact.id.conflict",
			// 		"contacts: no such contact; "+re.Detail,
			// 	)
			// }
		case "23505": // unique_violation: duplicate key value violates unique constraint
			switch e.ConstraintName {
			case "dataset_scope":
				// {detail:"Key (scope, dc)=(cities, 1) already exists."}
				err = custom.ConflictError(
					"custom.dataset.conflict",
					// "contacts: duplicate labels tag",
					"custom: dataset( repo: ); name: duplicate", // +e.Detail,
				)
				// case "contact_label_tag_unique":
				// 	// {detail:"Key (contact_id, tag)=(33, VIP) already exists."}
				// 	err = model.ConflictError(
				// 		"contacts.labels.conflict",
				// 		// "contacts: duplicate labels tag",
				// 		"contacts: duplicate labels tag; "+re.Detail,
				// 	)
				// // TABLE contacts.contact_timezone
				// // UNIQUE (contact_id, timezone_id)
				// case "contact_timezone_unique":
				// 	// {detail:"Key (contact_id, timezone_id)=(350, 416) already exists."}
				// 	err = model.ConflictError(
				// 		"contacts.timezones.timezone.id.conflict",
				// 		"contacts: timezone duplicate; "+re.Detail,
				// 	)
				// case "wbt_user_id_uindex", // directory.wbt_user(lower(username::text), dc)
				// 	"wbt_domain_auth_uindex",      // directory.wbt_auth(lower(auth::text), dc)
				// 	"wbt_user_dc_username_uindex": // directory.wbt_user(dc, lower(username::text))
				// 	// // NOTE: User -or- Role with the same .Name already exists !
				// 	// // NOTE: User, same as a Role, represents an administrative Unit(s) and must be unique within domain !
				// 	// err = model.ConflictError(`app.user.conflict.error`, `user: role '%s' already exists`, src.Username)
				// case "wbt_user_email_uindex":
				// 	// err = model.ConflictError(`app.user.email.conflict.error`, `email: address '%s' already registered`, src.Email)
				// case "wbt_user_extension_uindex":
				// 	// err = model.ConflictError(`app.user.extension.conflict.error`, `extension: number '%s' already registered`, src.Extension)
				// case "contact_imclient_pk2", "contact_imclient_user_id_unique":
				// 	err = model.ConflictError(
				// 		"contacts.imclients.user.id.conflict",
				// 		"contacts: this client already has a contact; "+re.Detail,
				// 	)
			}
		}
	}
	return err
}
