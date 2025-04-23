package postgres

import "testing"

func Test_customSqlIdentifier(t *testing.T) {
	tests := []struct {
		name      string
		wantIdent string
	}{
		// TODO: Add test cases.
		{
			name:      "a",
			wantIdent: `a`,
		},
		{
			name:      "A",
			wantIdent: `"A"`,
		},
		{
			name:      "id",
			wantIdent: `id`,
		},
		{
			name:      "Oid",
			wantIdent: `"Oid"`,
		},
		{
			name:      `"`,
			wantIdent: `""""`,
		},
		{
			name:      `""`,
			wantIdent: `""""""`,
		},
		{
			name:      "my_field",
			wantIdent: "my_field",
		},
		{
			name:      "camelCaseFieldName",
			wantIdent: `"camelCaseFieldName"`,
		},
		{
			name:      "user",
			wantIdent: `"user"`,
		},
		{
			name:      "user_defined_type_code",
			wantIdent: `user_defined_type_code`,
		},
		{
			name:      "my\"field",
			wantIdent: `"my""field"`,
		},
		{
			name:      "my\"\"field",
			wantIdent: `"my""""field"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotIdent := CustomSqlIdentifier(tt.name); gotIdent != tt.wantIdent {
				t.Errorf("customSqlIdentifier() = %v, want %v", gotIdent, tt.wantIdent)
			}
		})
	}
}
