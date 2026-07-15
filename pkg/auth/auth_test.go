package auth

import "testing"

func TestCheckValues(t *testing.T) {
	tests := []struct {
		name     string
		check    RoleCheckType
		actual   []string
		required []string
		want     bool
	}{
		{name: "any match", check: RoleCheckTypeAny, actual: []string{"user.read"}, required: []string{"user.read", "user.delete"}, want: true},
		{name: "any miss", check: RoleCheckTypeAny, actual: []string{"user.read"}, required: []string{"user.delete"}, want: false},
		{name: "all match", check: RoleCheckTypeAll, actual: []string{"user.read", "user.delete"}, required: []string{"user.read", "user.delete"}, want: true},
		{name: "all miss", check: RoleCheckTypeAll, actual: []string{"user.read"}, required: []string{"user.read", "user.delete"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkValues(tt.check, tt.actual, tt.required); got != tt.want {
				t.Fatalf("checkValues() = %v, want %v", got, tt.want)
			}
		})
	}
}
