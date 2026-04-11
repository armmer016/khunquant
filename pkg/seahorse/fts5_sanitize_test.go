package seahorse

import (
	"testing"
)

func TestSanitizeFTS5Query(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// Basic tokens
		{"hello world", `"hello" "world"`},
		{"database", `"database"`},

		// FTS5 operators neutralized
		{"sub-agent", `"sub-agent"`},
		{"agent:main", `"agent:main"`},
		{"+required", `"+required"`},
		{"prefix*", `"prefix*"`},
		{"^initial", `"^initial"`},
		{"crash OR restart", `"crash" "OR" "restart"`},
		{"NOT excluded", `"NOT" "excluded"`},
		{"(grouped)", `"(grouped)"`},

		// User-quoted phrases preserved
		{`"exact phrase" other`, `"exact phrase" "other"`},
		{`before "middle phrase" after`, `"before" "middle phrase" "after"`},

		// Unmatched quotes stripped
		{`"unmatched`, `"unmatched"`},
		{`hello"world`, `"helloworld"`},

		// NEAR operator neutralized
		{"NEAR/2 agent", `"NEAR/2" "agent"`},

		// Empty input
		{"", ""},
		{"   ", ""},

		// CJK unaffected
		{"数据库连接", `"数据库连接"`},
		{"数据库 连接", `"数据库" "连接"`},
		{"sub-agent重启", `"sub-agent重启"`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := SanitizeFTS5Query(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeFTS5Query(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
