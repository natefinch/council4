package parser

import "testing"

func TestExpandEntersOrLeavesBattlefieldTrigger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "self trigger splits into enters and leaves triggers",
			source: "When this creature enters or leaves the battlefield, create a Food token.",
			want: []string{
				"When this creature enters, create a Food token.",
				"Whenever this creature leaves the battlefield, create a Food token.",
			},
		},
		{
			name:   "whenever introduction preserved on enters half",
			source: "Whenever another artifact you control enters or leaves the battlefield, you may pay {1}.",
			want: []string{
				"Whenever another artifact you control enters, you may pay {1}.",
				"Whenever another artifact you control leaves the battlefield, you may pay {1}.",
			},
		},
		{
			name:   "ability word prefix carried to each",
			source: "Eerie — When this creature enters or leaves the battlefield, draw a card.",
			want: []string{
				"Eerie — When this creature enters, draw a card.",
				"Eerie — Whenever this creature leaves the battlefield, draw a card.",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := expandEntersOrLeavesBattlefieldTrigger(test.source)
			lines := splitSourceLines(got)
			if len(lines) != len(test.want) {
				t.Fatalf("expandEntersOrLeavesBattlefieldTrigger(%q) = %q, want %d lines", test.source, got, len(test.want))
			}
			for i := range lines {
				if lines[i] != test.want[i] {
					t.Fatalf("expandEntersOrLeavesBattlefieldTrigger(%q) line %d = %q, want %q", test.source, i, lines[i], test.want[i])
				}
			}
		})
	}
}

func TestExpandEntersOrLeavesBattlefieldTriggerLeavesOtherTextAlone(t *testing.T) {
	t.Parallel()
	unchanged := []string{
		"When this creature enters, draw a card.",
		"Whenever this creature leaves the battlefield, draw a card.",
		"When this creature enters or leaves the battlefield",
		"When this creature enters or leaves the battlefield,",
		"Whenever Wernog, Rider's Chaplain enters or leaves the battlefield, draw a card.",
		"Enchanted creature enters or leaves the battlefield, draw a card.",
	}
	for _, source := range unchanged {
		if got := expandEntersOrLeavesBattlefieldTrigger(source); got != source {
			t.Fatalf("expandEntersOrLeavesBattlefieldTrigger(%q) = %q, want unchanged", source, got)
		}
	}
}
