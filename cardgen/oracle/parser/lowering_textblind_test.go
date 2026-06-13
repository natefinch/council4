package parser

import "testing"

func TestRecognizeModalChoiceCounts(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		source  string
		context Context
		min     int
		max     int
		known   bool
	}{
		{
			name:    "choose one",
			source:  "Choose one —\n• Draw a card.\n• You gain 3 life.",
			context: Context{InstantOrSorcery: true},
			min:     1, max: 1, known: true,
		},
		{
			name:    "choose one or both",
			source:  "Choose one or both —\n• Draw a card.\n• You gain 3 life.",
			context: Context{InstantOrSorcery: true},
			min:     1, max: 2, known: true,
		},
		{
			name:    "choose two",
			source:  "Choose two —\n• Draw a card.\n• You gain 3 life.\n• You lose 1 life.",
			context: Context{InstantOrSorcery: true},
			min:     2, max: 2, known: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(tc.source, tc.context)
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if len(document.Abilities) != 1 || document.Abilities[0].Modal == nil {
				t.Fatalf("abilities = %#v", document.Abilities)
			}
			modal := document.Abilities[0].Modal
			if !modal.ChoiceKnown || modal.MinModes != tc.min || modal.MaxModes != tc.max {
				t.Fatalf("min/max/known = %d/%d/%v, want %d/%d/%v",
					modal.MinModes, modal.MaxModes, modal.ChoiceKnown, tc.min, tc.max, tc.known)
			}
		})
	}
}

func TestRecognizeModalChoiceUnknownFailsClosed(t *testing.T) {
	t.Parallel()
	// "Choose any number" is not a fixed cardinal choice; min/max must be
	// reported as unknown so lowering fails closed instead of guessing a count.
	source := "Choose any number —\n• Draw a card.\n• You gain 3 life."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 || document.Abilities[0].Modal == nil {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	if document.Abilities[0].Modal.ChoiceKnown {
		t.Fatalf("expected unknown modal choice, got min/max = %d/%d",
			document.Abilities[0].Modal.MinModes, document.Abilities[0].Modal.MaxModes)
	}
}

func TestRecognizeSagaLoreReminder(t *testing.T) {
	t.Parallel()
	source := "(As this Saga enters and after your draw step, add a lore counter.)"
	document, _ := Parse(source, Context{Saga: true})
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	if !document.Abilities[0].SagaReminder {
		t.Fatalf("expected SagaReminder for %q", source)
	}

	// Outside Saga context the same reminder is not a saga lore reminder.
	outside, _ := Parse(source, Context{})
	if len(outside.Abilities) == 1 && outside.Abilities[0].SagaReminder {
		t.Fatal("SagaReminder must be false outside Saga context")
	}

	// Near miss: altered reminder vocabulary fails closed.
	nearMiss, _ := Parse("(As this Saga enters, add two lore counters.)", Context{Saga: true})
	if len(nearMiss.Abilities) == 1 && nearMiss.Abilities[0].SagaReminder {
		t.Fatal("SagaReminder must be false for altered reminder text")
	}
}

func TestRecognizeReadAheadReminder(t *testing.T) {
	t.Parallel()
	base := "Read ahead (Choose a chapter and start with that many lore counters. Add one after your draw step. Skipped chapters don't trigger."
	cases := []struct {
		name    string
		source  string
		chapter int
		ok      bool
	}{
		{name: "no sacrifice", source: base + ")", chapter: 0, ok: true},
		{name: "sacrifice after III", source: base + " Sacrifice after III.)", chapter: 3, ok: true},
		{name: "near miss vocabulary", source: "Read ahead (Choose a chapter.)", chapter: 0, ok: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(tc.source, Context{Saga: true})
			if len(document.Abilities) != 1 {
				t.Fatalf("abilities = %#v", document.Abilities)
			}
			ability := document.Abilities[0]
			if ability.ReadAheadRecognized != tc.ok {
				t.Fatalf("ReadAheadRecognized = %v, want %v", ability.ReadAheadRecognized, tc.ok)
			}
			if tc.ok && ability.ReadAheadSacrificeChapter != tc.chapter {
				t.Fatalf("ReadAheadSacrificeChapter = %d, want %d",
					ability.ReadAheadSacrificeChapter, tc.chapter)
			}
		})
	}
}

func TestRecognizeDevoid(t *testing.T) {
	t.Parallel()
	document, _ := Parse("Devoid (This card has no color.)", Context{})
	if len(document.Abilities) != 1 || !document.Abilities[0].DevoidRecognized {
		t.Fatalf("expected DevoidRecognized, got %#v", document.Abilities)
	}
	nearMiss, _ := Parse("Devoid (This permanent has no color.)", Context{})
	if len(nearMiss.Abilities) == 1 && nearMiss.Abilities[0].DevoidRecognized {
		t.Fatal("DevoidRecognized must be false for altered reminder text")
	}
}

func TestEntersTappedSelfSyntax(t *testing.T) {
	t.Parallel()
	cases := []struct {
		source string
		want   bool
	}{
		{source: "This land enters tapped.", want: true},
		{source: "This permanent enters tapped.", want: true},
		{source: "Nyx Lotus enters tapped.", want: true},
		{source: "As this enchantment enters, choose Khans or Dragons.", want: false},
		{source: "This creature enters with three +1/+1 counters on it.", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(tc.source, Context{})
			if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) == 0 {
				t.Fatalf("abilities = %#v", document.Abilities)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) == 0 {
				t.Fatalf("no effects for %q", tc.source)
			}
			if got := effects[0].EntersTappedSelf; got != tc.want {
				t.Fatalf("EntersTappedSelf = %v, want %v", got, tc.want)
			}
		})
	}
}
