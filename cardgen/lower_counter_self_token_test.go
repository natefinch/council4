package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerCounterThenSelfCreatesCreatureToken proves the ordered
// counter-then-create family whose token is owned by the caster lowers to a
// CounterObject followed by a controller-recipient CreateToken (Geist Snatch,
// Summoner's Bane, Launch Mishap).
func TestLowerCounterThenSelfCreatesCreatureToken(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		oracleText string
		subtype    types.Sub
		power      int
	}{
		{
			name:       "Geist Snatch",
			oracleText: "Counter target creature spell. Create a 1/1 blue Spirit creature token with flying.",
			subtype:    types.Spirit,
			power:      1,
		},
		{
			name:       "Summoner's Bane",
			oracleText: "Counter target creature spell. Create a 2/2 blue Illusion creature token.",
			subtype:    types.Illusion,
			power:      2,
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       test.name,
				Layout:     "normal",
				TypeLine:   "Instant",
				ManaCost:   "{2}{U}{U}",
				Colors:     []string{"U"},
				OracleText: test.oracleText,
			})
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Sequence) != 2 {
				t.Fatalf("sequence = %#v, want counter then create token", mode.Sequence)
			}
			if _, ok := mode.Sequence[0].Primitive.(game.CounterObject); !ok {
				t.Fatalf("first primitive = %T, want game.CounterObject", mode.Sequence[0].Primitive)
			}
			create, ok := mode.Sequence[1].Primitive.(game.CreateToken)
			if !ok {
				t.Fatalf("second primitive = %T, want game.CreateToken", mode.Sequence[1].Primitive)
			}
			if create.Recipient.Exists {
				t.Fatalf("recipient = %#v, want unset controller recipient", create.Recipient)
			}
			if create.Amount.Value() != 1 {
				t.Fatalf("amount = %d, want 1", create.Amount.Value())
			}
			token, ok := create.Source.TokenDefRef()
			if !ok ||
				!slices.Equal(token.Subtypes, []types.Sub{test.subtype}) ||
				!token.Power.Exists || token.Power.Val.Value != test.power {
				t.Fatalf("token = %#v, want %s", token, test.subtype)
			}
		})
	}
}

// TestLowerCounterThenSelfCreatesTreasureToken proves the caster-token family
// also accepts predefined artifact tokens (Hornswoggle).
func TestLowerCounterThenSelfCreatesTreasureToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Hornswoggle",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{2}{U}",
		Colors:     []string{"U"},
		OracleText: "Counter target creature spell. You create a Treasure token. (It's an artifact with \"{T}, Sacrifice this token: Add one mana of any color.\")",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want counter then create token", mode.Sequence)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.CounterObject); !ok {
		t.Fatalf("first primitive = %T, want game.CounterObject", mode.Sequence[0].Primitive)
	}
	create, ok := mode.Sequence[1].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("second primitive = %T, want game.CreateToken", mode.Sequence[1].Primitive)
	}
	if create.Recipient.Exists {
		t.Fatalf("recipient = %#v, want unset controller recipient", create.Recipient)
	}
	token, ok := create.Source.TokenDefRef()
	if !ok ||
		token.Name != string(types.Treasure) ||
		!slices.Equal(token.Subtypes, []types.Sub{types.Treasure}) {
		t.Fatalf("token = %#v, want Treasure", token)
	}
}

// TestLowerCounterThenSelfTokenRejectsColorlessGate keeps the caster-token
// lowerer narrow: an X-sized token count is dynamic and not yet lowered here, so
// it stays unsupported rather than lowering partially.
func TestLowerCounterThenSelfTokenRejectsDynamicCount(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Spell Swindle",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{3}{U}{U}",
		Colors:     []string{"U"},
		OracleText: "Counter target spell. Create X Treasure tokens, where X is that spell's mana value.",
	})
}
