package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceOptionalHaveSelfDamage covers the controller
// "you may have this creature deal ..." causative: the trigger body lowers to a
// single Damage instruction marked Optional.
func TestGenerateExecutableCardSourceOptionalHaveSelfDamage(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Searing Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature enters, you may have this creature deal 1 damage to each creature.",
		Power:      new("2"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"TriggeredAbilities: []game.TriggeredAbility",
		"Primitive: game.Damage",
		"Optional: true",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceOptionalHaveItDamage covers the controller
// "you may have it deal ..." causative whose subject is the referenced object
// (the dying/blocking source). Both the structural "have" and the real Damage
// effect compile optional in this pronoun form; the lowerer still produces a
// single Optional Damage instruction sourced from the event permanent.
func TestGenerateExecutableCardSourceOptionalHaveItDamage(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Spiteful Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, you may have it deal 1 damage to any target.",
		Power:      new("2"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.Damage",
		"DamageSource: opt.Val(game.EventPermanentReference())",
		"Optional: true",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceRejectsUnsupportedOptionalHave keeps the
// fail-closed boundary: a non-controller "<player> may have" optional stays
// unsupported rather than silently dropping the player gate, since the runtime
// cannot model the non-controller decider as a controller-gated optional.
func TestGenerateExecutableCardSourceRejectsUnsupportedOptionalHave(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		oracleText string
		power      *string
	}{
		{
			name:       "non-controller controller-may",
			typeLine:   "Enchantment",
			oracleText: "Whenever a creature enters, that creature's controller may have it deal damage equal to its power to any target.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Unsupported Have",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
				Power:      test.power,
				Toughness:  test.power,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "u")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" {
				t.Fatalf("source = %q, want no partial card", source)
			}
			if len(diagnostics) == 0 || diagnostics[0].Summary != "unsupported optional effect" {
				t.Fatalf("diagnostics = %#v, want unsupported optional effect", diagnostics)
			}
		})
	}
}

// TestGenerateExecutableCardSourceOptionalHavePlayerSubject covers controller
// "you may have <player/subject> <action>" causatives whose action is a
// fully-determined player-or-creature effect (draw, mill, discard, life change,
// power/toughness change). The causative action's clause uses the base-verb form
// ("each player draw a card"), which the parser leaves non-exact, but the body
// still lowers to a single Optional instruction targeting the correct player or
// permanent because every field the runtime needs is parsed.
func TestGenerateExecutableCardSourceOptionalHavePlayerSubject(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		oracleText string
		power      *string
		wantParts  []string
	}{
		{
			name:       "each player draw",
			typeLine:   "Sorcery",
			oracleText: "You may have each player draw a card.",
			wantParts: []string{
				"Primitive: game.Draw",
				"PlayerGroup: game.AllPlayersReference()",
				"Optional: true",
			},
		},
		{
			name:       "each opponent discard",
			typeLine:   "Creature — Bear",
			oracleText: "When this creature enters, you may have each opponent discard a card.",
			power:      new("2"),
			wantParts: []string{
				"Primitive: game.Discard",
				"PlayerGroup: game.OpponentsReference()",
				"Optional: true",
			},
		},
		{
			name:       "target player mill",
			typeLine:   "Sorcery",
			oracleText: "You may have target player mill two cards.",
			wantParts: []string{
				"Primitive: game.Mill",
				"Player: game.TargetPlayerReference(0)",
				"Constraint: \"target player\"",
				"Optional: true",
			},
		},
		{
			name:       "target opponent discard",
			typeLine:   "Sorcery",
			oracleText: "You may have target opponent discard a card.",
			wantParts: []string{
				"Primitive: game.Discard",
				"Player: game.TargetPlayerReference(0)",
				"Constraint: \"target opponent\"",
				"Optional: true",
			},
		},
		{
			name:       "that player lose life",
			typeLine:   "Creature — Bear",
			oracleText: "Whenever this creature deals combat damage to a player, you may have that player lose 2 life.",
			power:      new("2"),
			wantParts: []string{
				"Primitive: game.LoseLife",
				"Player: game.EventPlayerReference()",
				"Optional: true",
			},
		},
		{
			name:       "target creature pump",
			typeLine:   "Sorcery",
			oracleText: "You may have target creature get +1/+1 until end of turn.",
			wantParts: []string{
				"Primitive: game.ModifyPT",
				"game.TargetPermanentReference(0)",
				"game.DurationUntilEndOfTurn",
				"Optional: true",
			},
		},
		{
			name:       "target creature shrink",
			typeLine:   "Creature — Bear",
			oracleText: "When this creature enters, you may have target creature get -2/-2 until end of turn.",
			power:      new("2"),
			wantParts: []string{
				"Primitive: game.ModifyPT",
				"PowerDelta:     game.Fixed(-2)",
				"ToughnessDelta: game.Fixed(-2)",
				"game.DurationUntilEndOfTurn",
				"Optional: true",
			},
		},
		{
			name:       "that player lose life dynamic",
			typeLine:   "Creature — Ally",
			oracleText: "When this creature enters, you may have target player lose life equal to the number of Allies you control.",
			power:      new("2"),
			wantParts: []string{
				"Primitive: game.LoseLife",
				"Amount: game.Dynamic(",
				"Optional: true",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Player Subject Have",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
				Power:      test.power,
				Toughness:  test.power,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "p")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, want := range test.wantParts {
				if !strings.Contains(source, want) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
		})
	}
}

// TestGenerateExecutableCardSourceOptionalHaveKeywordGrant covers controller
// "you may have target creature gain <keyword> until end of turn" causatives. The
// granted keyword rides the action clause rather than the structural "have", so
// the body lowers to a single Optional ApplyContinuous that adds the keyword to
// the chosen target until end of turn.
func TestGenerateExecutableCardSourceOptionalHaveKeywordGrant(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		wantParts  []string
	}{
		{
			name:       "gain flying",
			oracleText: "When this creature enters, you may have target creature gain flying until end of turn.",
			wantParts: []string{
				"Primitive: game.ApplyContinuous",
				"AddKeywords: []game.Keyword{",
				"game.Flying,",
				"Duration: game.DurationUntilEndOfTurn",
				"Optional: true",
			},
		},
		{
			name:       "gain double strike",
			oracleText: "When this creature enters, you may have target creature gain double strike until end of turn.",
			wantParts: []string{
				"Primitive: game.ApplyContinuous",
				"game.DoubleStrike,",
				"Optional: true",
			},
		},
		{
			name:       "gain two keywords",
			oracleText: "When this creature enters, you may have target creature gain flying and vigilance until end of turn.",
			wantParts: []string{
				"Primitive: game.ApplyContinuous",
				"game.Flying,",
				"game.Vigilance,",
				"Optional: true",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			power := new("2")
			card := &ScryfallCard{
				Name:       "Keyword Grant Have",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: test.oracleText,
				Power:      power,
				Toughness:  power,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "k")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, want := range test.wantParts {
				if !strings.Contains(source, want) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
		})
	}
}
