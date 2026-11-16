package mqtttopic

import "testing"

func TestMatch(t *testing.T) {
	tests := []struct {
		name  string
		route string
		topic string
		want  bool
	}{
		{"multi player1 parent", "sport/tennis/player1/#", "sport/tennis/player1", true},
		{"multi ranking", "sport/tennis/player1/#", "sport/tennis/player1/ranking", true},
		{"multi deep", "sport/tennis/player1/#", "sport/tennis/player1/score/wimbledon", true},
		{"multi matches parent", "sport/#", "sport", true},
		{"hash matches all", "#", "sport/tennis", true},
		{"hash matches single level", "#", "sport", true},
		{"hash matches leading slash", "#", "/finance", true},
		{"sport/tennis/# valid", "sport/tennis/#", "sport/tennis/x", true},
		{"plus last level p1", "sport/tennis/+", "sport/tennis/player1", true},
		{"plus last level p2", "sport/tennis/+", "sport/tennis/player2", true},
		{"plus not multi level", "sport/tennis/+", "sport/tennis/player1/ranking", false},
		{"plus not parent", "sport/+", "sport", false},
		{"plus matches empty level", "sport/+", "sport/", true},
		{"plus alone", "+", "sport", true},
		{"plus alone not two levels", "+", "sport/tennis", false},
		{"plus and hash", "+/tennis/#", "sport/tennis/player1/score", true},
		{"plus and hash parent", "+/tennis/#", "sport/tennis", true},
		{"plus middle", "sport/+/player1", "sport/tennis/player1", true},
		{"plus middle wrong", "sport/+/player1", "sport/tennis/player2", false},
		{"leading slash plus plus", "+/+", "/finance", true},
		{"leading slash slash plus", "/+", "/finance", true},
		{"leading slash plus alone", "+", "/finance", false},
		{"sport+ is not wildcard", "sport+", "sport1", false},
		{"sport/tennis# not wildcard", "sport/tennis#", "sport/tennis/player1", false},
		{"hash no dollar", "#", "$SYS/monitor/Clients", false},
		{"plus no dollar", "+/monitor/Clients", "$SYS/monitor/Clients", false},
		{"plus alone no dollar", "+", "$SYS", false},
		{"dollar hash", "$SYS/#", "$SYS/monitor/Clients", true},
		{"dollar hash parent", "$SYS/#", "$SYS", true},
		{"dollar plus", "$SYS/monitor/+", "$SYS/monitor/Clients", true},
		{"dollar exact", "$SYS/monitor/Clients", "$SYS/monitor/Clients", true},
		{"hash matches non dollar", "#", "sport", true},
		{"case sensitive", "ACCOUNTS", "Accounts", false},
		{"case sensitive same", "Accounts", "Accounts", true},
		{"space", "Accounts payable", "Accounts payable", true},
		{"space differs", "Accounts payable", "Accountspayable", false},
		{"leading slash distinct", "/finance", "finance", false},
		{"leading slash distinct reverse", "finance", "/finance", false},
		{"trailing slash distinct", "finance/", "finance", false},
		{"trailing slash same", "finance/", "finance/", true},
		{"only slash", "/", "/", true},
		{"only slash vs empty level", "/", "a/", false},
		{"zero length level", "a//b", "a//b", true},
		{"zero length level plus", "a/+/b", "a//b", true},
		{"zero length level differs", "a//b", "a/b", false},
		{"no normalization", "café", "café", false},
		{"unicode exact", "café", "café", true},
		{"unrecognized char no substitution", "a�b", "a?b", false},
		{"many levels", "a/b/c/d/e/f/g/h", "a/b/c/d/e/f/g/h", true},
		{"many levels plus", "a/+/c/+/e/+/g/+", "a/b/c/d/e/f/g/h", true},
		{"level prefix not match", "sport/ten", "sport/tennis", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Match(tt.route, tt.topic); got != tt.want {
				t.Errorf("match(%v, %v) = %v, want %v", tt.route, tt.topic, got, tt.want)
			}
		})
	}
}
