package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// sacrificeChoiceFromOptionalTrigger lowers a single-face card whose only
// triggered ability is an optional "you may sacrifice ..." effect and returns
// the SacrificePermanents primitive that heads its reflexive sequence.
func sacrificeChoiceFromOptionalTrigger(t *testing.T, oracleText string) game.SacrificePermanents {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Sac Choice",
		Layout:     "normal",
		TypeLine:   "Creature — Test",
		OracleText: oracleText,
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Sequence) == 0 {
		t.Fatal("empty triggered sequence")
	}
	sacrifice, ok := mode.Sequence[0].Primitive.(game.SacrificePermanents)
	if !ok {
		t.Fatalf("first primitive = %T, want game.SacrificePermanents", mode.Sequence[0].Primitive)
	}
	if !mode.Sequence[0].Optional {
		t.Fatal("sacrifice instruction = mandatory, want optional")
	}
	return sacrifice
}

// TestLowerSacrificeChoiceHeterogeneousDisjunction proves the optional "you may
// sacrifice <A> or <B>" cost broadens to a Selection.AnyOf union when the two
// sides occupy different permanent dimensions and at least one side names a
// distinct artifact/token class: a card type with an artifact subtype (Gastal
// Blockbuster's "creature or Vehicle"), a card type with a bare token (Old Man
// Willow's "another creature or a token"), or a bare token with a card type
// (West Wind Avatar's "token or a land"). Each side lowers independently and the
// runtime sacrifice selection accepts any of them.
func TestLowerSacrificeChoiceHeterogeneousDisjunction(t *testing.T) {
	t.Parallel()
	creature := game.Selection{RequiredTypes: []types.Card{types.Creature}}
	land := game.Selection{RequiredTypes: []types.Card{types.Land}}
	vehicle := game.Selection{SubtypesAny: []types.Sub{types.Sub("Vehicle")}}
	token := game.Selection{TokenOnly: true}

	cases := []struct {
		name       string
		oracleText string
		want       game.Selection
	}{
		{
			name:       "card type or artifact subtype",
			oracleText: "When this creature enters, you may sacrifice a creature or Vehicle. When you do, draw a card.",
			want:       game.Selection{AnyOf: []game.Selection{creature, vehicle}},
		},
		{
			name:       "another card type or token",
			oracleText: "Whenever this creature attacks, you may sacrifice another creature or a token. When you do, draw a card.",
			want:       game.Selection{AnyOf: []game.Selection{creature, token}, ExcludeSource: true},
		},
		{
			name:       "token or card type",
			oracleText: "When this creature enters, you may sacrifice a token or a land. When you do, draw a card.",
			want:       game.Selection{AnyOf: []game.Selection{token, land}},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			sacrifice := sacrificeChoiceFromOptionalTrigger(t, test.oracleText)
			if sacrifice.Amount != game.Fixed(1) {
				t.Fatalf("sacrifice amount = %+v, want Fixed(1)", sacrifice.Amount)
			}
			if !selectionsEqual(sacrifice.Selection, test.want) {
				t.Fatalf("sacrifice selection = %#v, want %#v", sacrifice.Selection, test.want)
			}
		})
	}
}

// TestSacrificeChoiceThreeWayDisjunctionStaysUnsupported guards the fail-closed
// boundary: the broadened reconstruction only spans a two-sided union, so a
// three-way "creature, Vehicle, or token" sacrifice has no canonical wording the
// round-trip reproduces and must stay unsupported rather than mislower.
func TestSacrificeChoiceThreeWayDisjunctionStaysUnsupported(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Triple Sac",
		Layout:     "normal",
		TypeLine:   "Creature — Test",
		OracleText: "When this creature enters, you may sacrifice a creature, Vehicle, or token. When you do, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
}

// selectionsEqual compares the fields the disjunctive sacrifice family exercises:
// the AnyOf union, its ordered sides' required types and subtypes, the token-only
// flag, and the source exclusion carried from a leading "another".
func selectionsEqual(got, want game.Selection) bool {
	if got.ExcludeSource != want.ExcludeSource || got.TokenOnly != want.TokenOnly {
		return false
	}
	if len(got.RequiredTypes) != len(want.RequiredTypes) {
		return false
	}
	for i := range want.RequiredTypes {
		if got.RequiredTypes[i] != want.RequiredTypes[i] {
			return false
		}
	}
	if len(got.SubtypesAny) != len(want.SubtypesAny) {
		return false
	}
	for i := range want.SubtypesAny {
		if got.SubtypesAny[i] != want.SubtypesAny[i] {
			return false
		}
	}
	if len(got.AnyOf) != len(want.AnyOf) {
		return false
	}
	for i := range want.AnyOf {
		if !selectionsEqual(got.AnyOf[i], want.AnyOf[i]) {
			return false
		}
	}
	return true
}
