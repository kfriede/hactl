package cmd

import (
	"testing"
)

func TestParseJSONInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, result map[string]any)
	}{
		{
			name:  "valid simple object",
			input: `{"name":"test","count":42}`,
			check: func(t *testing.T, result map[string]any) {
				if result["name"] != "test" {
					t.Errorf("expected name=test, got %v", result["name"])
				}
			},
		},
		{
			name:  "nested object",
			input: `{"item":{"enabled":true}}`,
			check: func(t *testing.T, result map[string]any) {
				nested, ok := result["item"].(map[string]any)
				if !ok {
					t.Fatalf("expected nested map, got %T", result["item"])
				}
				if nested["enabled"] != true {
					t.Errorf("expected enabled=true, got %v", nested["enabled"])
				}
			},
		},
		{
			name:    "invalid JSON",
			input:   `{not valid json}`,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   ``,
			wantErr: true,
		},
		{
			name:    "json array instead of object",
			input:   `[1, 2, 3]`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseJSONInput(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestConfirmActionWithYes(t *testing.T) {
	saved := flagYes
	defer func() { flagYes = saved }()

	flagYes = true
	if !confirmAction("delete this resource") {
		t.Error("confirmAction should return true when flagYes is set")
	}
}
