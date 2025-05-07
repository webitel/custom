package data

import (
	"fmt"

	custompb "github.com/webitel/proto/gen/custom"
	datatypb "github.com/webitel/proto/gen/custom/data"

	// custom "github.com/webitel/custom/data"
	customrel "github.com/webitel/custom/reflect"
	customreg "github.com/webitel/custom/registry"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// .well-known dataset fields ..
var (
	fieldSerialId = custompb.Field{
		Id:   "id",
		Name: "", // "id"
		Hint: "", // "id"
		Kind: customrel.INT64,
		Type: &custompb.Field_Int64{
			Int64: &datatypb.Int{
				Min: wrapperspb.Int64(1),
				Max: wrapperspb.Int64(1<<63 - 1),
				// Violation: map[string]string{
				// 	"min": "should be positive integer",
				// },
			},
		},
		// Value: &custompb.Field_Default{
		// 	Default: structpb.NewStringValue("$(nextval)"),
		// },
		Readonly: true,
		Required: true, // NOTNULL
		Disabled: false,
		Hidden:   true,
	}
	fieldCreatedAt = custompb.Field{
		Id:   "created_at",
		Name: "Creation at",
		Hint: "When created",
		Kind: customrel.DATETIME,
		// Type: &custompb.Field_Datetime{
		// 	Datetime: &datatype.Datetime{
		// 		Zone:   "UTC",
		// 		Epoch:  0,
		// 		Format: time.RFC3339,
		// 	},
		// },
		Value: &custompb.Field_Default{
			Default: structpb.NewStringValue("$(timestamp)"),
		},
		Readonly: true,
		Required: true, // NOTNULL
		Disabled: false,
		Hidden:   true,
	}
	fieldCreatedBy = custompb.Field{
		Id:   "created_by",
		Name: "Created by",
		Hint: "Who created. Author",
		Kind: customrel.LOOKUP,
		Type: &custompb.Field_Lookup{
			Lookup: &datatypb.Lookup{
				// Name: "",
				Path: "users",
				// Primary: "", // users.primary: "id"
				// Display: "", // users.primary: "name"
				// Query:     map[string]string{},
				// Violation: map[string]string{},
			},
		},
		Value: &custompb.Field_Default{
			Default: structpb.NewStringValue("$(user)"),
		},
		Readonly: true,
		Required: true, // NOTNULL
		Disabled: false,
		Hidden:   true,
	}
	fieldUpdatedAt = custompb.Field{
		Id:   "updated_at",
		Name: "Updated at",
		Hint: "Last modified",
		Kind: customrel.DATETIME,
		// Type: &custompb.Field_Datetime{
		// 	Datetime: &datatype.Datetime{
		// 		Zone:   "UTC",
		// 		Epoch:  0,
		// 		Format: time.RFC3339,
		// 	},
		// },
		Value: &custompb.Field_Always{
			Always: structpb.NewStringValue("$(timestamp)"),
		},
		Readonly: true,
		Required: true, // NOTNULL
		Disabled: false,
		Hidden:   true,
	}
	fieldUpdatedBy = custompb.Field{
		Id:   "updated_by",
		Name: "Updated by",
		Hint: "Who modified. Editor",
		Kind: customrel.LOOKUP,
		Type: &custompb.Field_Lookup{
			Lookup: &datatypb.Lookup{
				// Name: "",
				Path: "users",
				// Primary: "", // users.primary: "id"
				// Display: "", // users.primary: "name"
				// Query:     map[string]string{},
				// Violation: map[string]string{},
			},
		},
		Value: &custompb.Field_Always{
			Always: structpb.NewStringValue("$(user)"),
		},
		Readonly: true,
		Required: true, // NOTNULL
		Disabled: false,
		Hidden:   true,
	}
)

// .well-known dataset types ..
var (
	Users = custompb.Dataset{
		Name: "Users",
		Repo: "users",
		Path: "users",
		Fields: []*custompb.Field{
			// ----------------------- //
			&fieldSerialId,
			// ----------------------- //
			{
				Id:   "name",
				Name: "Common Name",
				Hint: "",
				Kind: customrel.STRING,
				Type: &custompb.Field_String_{
					String_: &datatypb.Text{
						MaxChars: 255, // Unicode
					},
				},
				Value:    nil,
				Readonly: false,
				Required: true,
				Disabled: false,
				Hidden:   false,
			},
			// {
			// 	Id:   "email",
			// 	Name: "Email",
			// 	Hint: "Email address",
			// 	Kind: types.STRING,
			// 	Type: &custompb.Field_String_{
			// 		String_: &datatype.Text{
			// 			MaxBytes: 255, // ASCII
			// 		},
			// 	},
			// 	Value:    nil,
			// 	Readonly: false,
			// 	Required: false,
			// 	Disabled: false,
			// 	Hidden:   false,
			// },
			// {
			// 	Id:   "username",
			// 	Name: "Username",
			// 	Hint: "Login",
			// 	Kind: types.STRING,
			// 	Type: &custompb.Field_String_{
			// 		String_: &datatype.Text{
			// 			MaxBytes: 64, // ASCII
			// 		},
			// 	},
			// 	Value:    nil,
			// 	Readonly: false,
			// 	Required: true,
			// 	Disabled: false,
			// 	Hidden:   false,
			// },
			// {
			// 	Id:   "extension",
			// 	Name: "Extension",
			// 	Hint: "Short dial number",
			// 	Type: &custompb.Field_String_{
			// 		String_: &datatype.Text{
			// 			MaxBytes: 6, // ASCII: DIGITS
			// 		},
			// 	},
			// 	Value:    nil,
			// 	Readonly: false,
			// 	Required: false,
			// 	Disabled: false,
			// 	Hidden:   false,
			// },
			// // device:   lookup(devices)
			// // devices:  list[lookup(devices)]
			// // hotdesks: list[lookup(devices)]
			// // ----------------------- //
			// &createdAt,
			// &createdBy,
			// &updatedAt,
			// &updatedBy,
			// // ----------------------- //
		},
		Primary: fieldSerialId.GetId(),
		Display: "name",
		// Indexes: map[string]*custompb.Index{
		// 	"email":     {Unique: true},
		// 	"username":  {Unique: true},
		// 	"extension": {Unique: true},
		// },
		Readonly:   true,
		Extendable: false,
		// Administered: true,
	}
	Roles = custompb.Dataset{
		Name: "Roles",
		Repo: "roles",
		Path: "roles",
		Fields: []*custompb.Field{
			// ----------------------- //
			&fieldSerialId,
			// ----------------------- //
			{
				Id:   "name",
				Name: "Name",
				Hint: "Group of users",
				Kind: customrel.STRING,
				Type: &custompb.Field_String_{
					String_: &datatypb.Text{
						MaxChars: 255, // Unicode
					},
				},
				Value:    nil,
				Readonly: false,
				Required: true,
				Disabled: false,
				Hidden:   false,
			},
			// {
			// 	Id:   "description",
			// 	Name: "Description",
			// 	Hint: "",
			// 	Kind: types.STRING,
			// 	// Type: &custompb.Field_String_{
			// 	// 	String_: &datatype.Text{
			// 	// 		MaxChars: 4092, // Unicode
			// 	// 	},
			// 	// },
			// 	Value:    nil,
			// 	Readonly: false,
			// 	Required: false,
			// 	Disabled: false,
			// 	Hidden:   false,
			// },
			// {
			// 	Id:   "user",
			// 	Name: "User",
			// 	Hint: "Is user ?",
			// 	Kind: types.BOOL,
			// 	// Type: &custompb.Field_Bool{
			// 	// 	Bool: &datatype.Bool{},
			// 	// },
			// 	Value: &custompb.Field_Default{
			// 		Default: structpb.NewBoolValue(false),
			// 	},
			// 	Readonly: true,
			// 	Required: true,
			// 	Disabled: false,
			// 	Hidden:   true,
			// },
			// // permissions
			// // members
			// // metadata
			// // ----------------------- //
			// &createdAt,
			// &createdBy,
			// &updatedAt,
			// &updatedBy,
			// // ----------------------- //
		},
		Primary: fieldSerialId.GetId(),
		Display: "name",
		// Indexes:      map[string]*custompb.Index{},
		Readonly:   true,
		Extendable: false,
		// Administered: true,
	}
	Contacts = custompb.Dataset{
		Name: "Contacts",
		Repo: "contacts",
		Path: "contacts",
		Fields: []*custompb.Field{
			// ----------------------- //
			&fieldSerialId,
			// ----------------------- //
			{
				Id:   "name.common_name",
				Name: "Contact name",
				Hint: "",
				Kind: customrel.STRING,
				Type: &custompb.Field_String_{
					String_: &datatypb.Text{
						MaxChars: 255, // Unicode
					},
				},
				// Value: &custompb.Field_Default{
				// 	Default: structpb.NewStringValue("$(.name.given_name)$(if .name.middle_name) $(.name.middle_name)$(end)$(if .name.family_name) $(.name.family_name)$(end)"),
				// },
				Readonly: false,
				Required: true,
				Disabled: false,
				Hidden:   false,
			},
			// {
			// 	Id:   "name.given_name",
			// 	Name: "Given name",
			// 	Hint: "First name",
			// 	Kind: types.STRING,
			// 	Type: &custompb.Field_String_{
			// 		String_: &datatype.Text{
			// 			MaxChars: 64, // Unicode
			// 		},
			// 	},
			// 	Value:    nil,
			// 	Readonly: false,
			// 	Required: false,
			// 	Disabled: false,
			// 	Hidden:   false,
			// },
			// {
			// 	Id:   "name.middle_name",
			// 	Name: "Middle name",
			// 	Hint: "",
			// 	Kind: types.STRING,
			// 	Type: &custompb.Field_String_{
			// 		String_: &datatype.Text{
			// 			MaxChars: 64, // Unicode
			// 		},
			// 	},
			// 	Value:    nil,
			// 	Readonly: false,
			// 	Required: false,
			// 	Disabled: false,
			// 	Hidden:   false,
			// },
			// {
			// 	Id:   "name.family_name",
			// 	Name: "Family name",
			// 	Hint: "Last name",
			// 	Kind: types.STRING,
			// 	Type: &custompb.Field_String_{
			// 		String_: &datatype.Text{
			// 			MaxChars: 64, // Unicode
			// 		},
			// 	},
			// 	Value:    nil,
			// 	Readonly: false,
			// 	Required: false,
			// 	Disabled: false,
			// 	Hidden:   false,
			// },
			// {
			// 	Id:   "about",
			// 	Name: "BIO",
			// 	Hint: "Short description",
			// 	Kind: types.STRING,
			// 	// Type: &custompb.Field_String_{
			// 	// 	String_: &datatype.Text{
			// 	// 		MaxBytes: 6, // ASCII: DIGITS
			// 	// 	},
			// 	// },
			// 	Value:    nil,
			// 	Readonly: false,
			// 	Required: false,
			// 	Disabled: false,
			// 	Hidden:   false,
			// },
			// // labels: list[lookup(contact.label)]
			// // emails: list[lookup(contact.email)]
			// // phones: list[lookup(contact.phone)]
			// // photos:
			// // managers:
			// // comments:
			// // languages:
			// // timezones:
			// // imclients:
			// // variables:
			// // ----------------------- //
			// &createdAt,
			// &createdBy,
			// &updatedAt,
			// &updatedBy,
			// // ----------------------- //
		},
		Primary: fieldSerialId.GetId(),
		Display: "name.common_name",
		// Indexes: map[string]*custompb.Index{},
		Readonly:   true, // system
		Extendable: true,
		// Administered: true,
	}
	ContactGroups = custompb.Dataset{
		Name: "Contact Groups",
		Repo: "groups", // "contact_groups",
		Path: "contacts/groups",
		Fields: []*custompb.Field{
			// ----------------------- //
			&fieldSerialId,
			// ----------------------- //
			{
				Id:   "name",
				Name: "Group",
				Hint: "",
				Kind: customrel.STRING,
				Type: &custompb.Field_String_{
					String_: &datatypb.Text{
						MaxChars: 255, // Unicode
					},
				},
				// Value: &custompb.Field_Default{
				// 	Default: structpb.NewStringValue("$(.name.given_name)$(if .name.middle_name) $(.name.middle_name)$(end)$(if .name.family_name) $(.name.family_name)$(end)"),
				// },
				Readonly: false,
				Required: true,
				Disabled: false,
				Hidden:   false,
			},
		},
		Primary: fieldSerialId.GetId(),
		Display: "name",
		// Indexes: map[string]*custompb.Index{},
		Readonly: true, // system
		// Extendable: false,
		// Administered: false,
	}
	Calendars = custompb.Dataset{
		Name: "Calendars",
		Repo: "calendars",
		Path: "calendars",
		Fields: []*custompb.Field{
			// ----------------------- //
			&fieldSerialId,
			// ----------------------- //
			{
				Id:   "name",
				Name: "Calendar",
				Hint: "",
				Kind: customrel.STRING,
				Type: &custompb.Field_String_{
					String_: &datatypb.Text{
						MaxChars: 255, // Unicode
					},
				},
				Value:    nil,
				Readonly: false,
				Required: true,
				Disabled: false,
				Hidden:   false,
			},
		},
		Primary: fieldSerialId.GetId(),
		Display: "name",
		// Indexes: map[string]*custompb.Index{},
		Readonly:   true, // system
		Extendable: false,
		// Administered: true,
	}
	BlockLists = custompb.Dataset{
		Name: "Lists", // "Block List",
		Repo: "list",
		Path: "call_center/list",
		Fields: []*custompb.Field{
			// ----------------------- //
			&fieldSerialId,
			// ----------------------- //
			{
				Id:   "name",
				Name: "List",
				Hint: "",
				Kind: customrel.STRING,
				Type: &custompb.Field_String_{
					String_: &datatypb.Text{
						MaxChars: 255, // Unicode
					},
				},
				Value:    nil,
				Readonly: false,
				Required: true,
				Disabled: false,
				Hidden:   false,
			},
		},
		Primary: fieldSerialId.GetId(),
		Display: "name",
		// Indexes: map[string]*custompb.Index{},
		Readonly:   true, // system
		Extendable: false,
		// Administered: true,
	}
	Operators = custompb.Dataset{
		Name: "Agents",
		Repo: "agents",
		Path: "call_center/agents",
		Fields: []*custompb.Field{
			// ----------------------- //
			&fieldSerialId,
			// ----------------------- //
			{
				Id:   "name", // "user.name",
				Name: "Operator",
				Hint: "",
				Kind: customrel.STRING,
				Type: &custompb.Field_String_{
					String_: &datatypb.Text{
						MaxChars: 255, // Unicode
					},
				},
				Value:    nil,
				Readonly: false,
				Required: true,
				Disabled: false,
				Hidden:   false,
			},
			// {
			// 	Id:   "user",
			// 	Name: "Email",
			// 	Hint: "Email address",
			// 	Kind: types.STRING,
			// 	Type: &custompb.Field_Lookup{
			// 		Lookup: &datatype.Lookup{
			// 			Path: "users",
			// 		},
			// 	},
			// 	Value:    nil,
			// 	Readonly: false,
			// 	Required: true,
			// 	Disabled: false,
			// 	Hidden:   false,
			// },
			// // ----------------------- //
			// &createdAt,
			// &createdBy,
			// &updatedAt,
			// &updatedBy,
			// // ----------------------- //
		},
		Primary: fieldSerialId.GetId(),
		Display: "name", // "user.name",
		// Indexes: map[string]*custompb.Index{
		// 	"user": {Unique: true},
		// },
		Readonly:   true,
		Extendable: false,
		// Administered: true,
	}
	Queues = custompb.Dataset{
		Name: "Queues",
		Repo: "queues",
		Path: "call_center/queues",
		Fields: []*custompb.Field{
			// ----------------------- //
			&fieldSerialId,
			// ----------------------- //
			{
				Id:   "name",
				Name: "Queue",
				Hint: "",
				Kind: customrel.STRING,
				Type: &custompb.Field_String_{
					String_: &datatypb.Text{
						MaxChars: 255, // Unicode
					},
				},
				Value:    nil,
				Readonly: false,
				Required: true,
				Disabled: false,
				Hidden:   false,
			},
		},
		Primary: fieldSerialId.GetId(),
		Display: "name",
		// Indexes: map[string]*custompb.Index{},
		Readonly:   true, // system
		Extendable: false,
		// Administered: true,
	}
	CommunicationTypes = custompb.Dataset{
		Name: "Communication Types",
		Repo: "communication_type",
		Path: "call_center/communication_type",
		Fields: []*custompb.Field{
			// ----------------------- //
			&fieldSerialId,
			// ----------------------- //
			{
				Id:   "name",
				Name: "Communication Type",
				Hint: "",
				Kind: customrel.STRING,
				Type: &custompb.Field_String_{
					String_: &datatypb.Text{
						MaxChars: 255, // Unicode
					},
				},
				Value:    nil,
				Readonly: false,
				Required: true,
				Disabled: false,
				Hidden:   false,
			},
		},
		Primary: fieldSerialId.GetId(),
		Display: "name",
		// Indexes: map[string]*custompb.Index{},
		Readonly:   true, // system
		Extendable: false,
		// Administered: false,
	}

	Issues = custompb.Dataset{
		Name: "Cases",
		Repo: "cases",
		Path: "cases",
		Fields: []*custompb.Field{
			// ----------------------- //
			&fieldSerialId,
			// ----------------------- //
			{
				Id:   "name",
				Name: "Case",
				Hint: "",
				Kind: customrel.STRING,
				Type: &custompb.Field_String_{
					String_: &datatypb.Text{
						MaxChars: 255, // Unicode
					},
				},
				Value:    nil,
				Readonly: false,
				Required: true,
				Disabled: false,
				Hidden:   false,
			},
		},
		Primary: fieldSerialId.GetId(),
		Display: "name",
		// Indexes: map[string]*custompb.Index{},
		Readonly:   true, // system
		Extendable: true,
		// Administered: true,
	}
	IssueSources = custompb.Dataset{
		Name: "Case Sources",
		Repo: "sources", // "case_sources",
		Path: "cases/sources",
		Fields: []*custompb.Field{
			// ----------------------- //
			&fieldSerialId,
			// ----------------------- //
			{
				Id:   "name",
				Name: "Source",
				Hint: "",
				Kind: customrel.STRING,
				Type: &custompb.Field_String_{
					String_: &datatypb.Text{
						MaxChars: 255, // Unicode
					},
				},
				Value:    nil,
				Readonly: false,
				Required: true,
				Disabled: false,
				Hidden:   false,
			},
		},
		Primary: fieldSerialId.GetId(),
		Display: "name",
		// Indexes: map[string]*custompb.Index{},
		Readonly: true, // system
		// Extendable: false,
		// Administered: false,
	}
	IssuePriorities = custompb.Dataset{
		Name: "Priorities",
		Repo: "priorities", // "case_priorities",
		Path: "cases/priorities",
		Fields: []*custompb.Field{
			// ----------------------- //
			&fieldSerialId,
			// ----------------------- //
			{
				Id:   "name",
				Name: "Priority",
				Hint: "",
				Kind: customrel.STRING,
				Type: &custompb.Field_String_{
					String_: &datatypb.Text{
						MaxChars: 255, // Unicode
					},
				},
				Value:    nil,
				Readonly: false,
				Required: true,
				Disabled: false,
				Hidden:   false,
			},
		},
		Primary: fieldSerialId.GetId(),
		Display: "name",
		// Indexes: map[string]*custompb.Index{},
		Readonly: true, // system
		// Extendable: false,
		// Administered: true,
	}
)

func init() {

	regedit := customreg.GlobalTypes
	for _, descriptor := range []*custompb.Dataset{
		// global type(s)
		&Users,
		&Roles,
		&Contacts,
		&ContactGroups,
		&Calendars,
		&BlockLists,
		&CommunicationTypes,
		&Operators,
		&Queues,
		&Issues,
		&IssueSources,
		&IssuePriorities,
		// more here ...
	} {

		// custom
		// System wide (global, dc:0) types ...
		regtype := DictionaryOf(0, descriptor)
		err := regtype.Err()
		if err != nil {
			panic(fmt.Errorf("custom: invalid dataset type structure; error: %w", err))
		}
		err = regedit.Register(regtype)
		if err != nil {
			panic(err)
		}
	}
}
