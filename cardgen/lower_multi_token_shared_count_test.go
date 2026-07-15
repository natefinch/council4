package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

// multiTokenTriggerInstructions returns the CreateToken primitives lowered from
// an enters-the-battlefield triggered ability whose single mode is entirely
// CreateToken instructions. It mirrors multiTokenCreateInstructions but reads the
// first triggered ability instead of the spell ability, so it covers permanents
// (like Farmer Cotton) whose multi-token create fires from an ETB trigger.
func multiTokenTriggerInstructions(t *testing.T, face loweredFaceAbilities) []game.CreateToken {
	t.Helper()
	if len(face.TriggeredAbilities) == 0 {
		t.Fatal("no triggered ability lowered")
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	creates := make([]game.CreateToken, 0, len(mode.Sequence))
	for i, inst := range mode.Sequence {
		create, ok := inst.Primitive.(game.CreateToken)
		if !ok {
			t.Fatalf("sequence[%d] = %T, want game.CreateToken", i, inst.Primitive)
		}
		creates = append(creates, create)
	}
	return creates
}

func createTokenDef(t *testing.T, create game.CreateToken) *game.CardDef {
	t.Helper()
	def, ok := create.Source.TokenDefRef()
	if !ok {
		t.Fatal("token source is not a token definition")
	}
	return def
}

func assertDynamicX(t *testing.T, q game.Quantity, label string) {
	t.Helper()
	if !q.IsDynamic() {
		t.Fatalf("%s amount = %d (fixed), want dynamic X", label, q.Value())
	}
	da := q.DynamicAmount()
	if !da.Exists || da.Val.Kind != game.DynamicAmountX {
		t.Fatalf("%s dynamic amount = %+v, want DynamicAmountX", label, da)
	}
}

// TestLowerMultiTokenSharedVariableX proves Farmer Cotton's "create X 1/1 white
// Halfling creature tokens and X Food tokens." lowers to two CreateToken
// instructions that share the spell's variable X count: a synthesized 1/1 white
// Halfling creature token and the predefined Food artifact token. This is the
// reusable multi-token-with-shared-count path — a synthesized creature spec and a
// predefined artifact spec created together under one dynamic X.
func TestLowerMultiTokenSharedVariableX(t *testing.T) {
	t.Parallel()
	pt := "1"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Farmer",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Halfling Peasant",
		ManaCost:   "{X}{G}{W}",
		OracleText: `When this creature enters, create X 1/1 white Halfling creature tokens and X Food tokens. (They're artifacts with "{2}, {T}, Sacrifice this token: You gain 3 life.")`,
		Power:      &pt,
		Toughness:  &pt,
		Colors:     []string{"G", "W"},
	})
	creates := multiTokenTriggerInstructions(t, face)
	if len(creates) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(creates))
	}

	assertDynamicX(t, creates[0].Amount, "Halfling")
	halfling := createTokenDef(t, creates[0])
	if halfling.Name != "Halfling" {
		t.Errorf("token[0] name = %q, want Halfling", halfling.Name)
	}
	if !halfling.Power.Exists || halfling.Power.Val.Value != 1 ||
		!halfling.Toughness.Exists || halfling.Toughness.Val.Value != 1 {
		t.Errorf("token[0] PT = %+v/%+v, want 1/1", halfling.Power, halfling.Toughness)
	}
	if len(halfling.Colors) != 1 || halfling.Colors[0] != color.White {
		t.Errorf("token[0] colors = %v, want [White]", halfling.Colors)
	}
	if len(halfling.Types) != 1 || halfling.Types[0] != types.Creature {
		t.Errorf("token[0] types = %v, want [Creature]", halfling.Types)
	}

	assertDynamicX(t, creates[1].Amount, "Food")
	food := createTokenDef(t, creates[1])
	if food.Name != "Food" {
		t.Errorf("token[1] name = %q, want Food", food.Name)
	}
	if len(food.Types) != 1 || food.Types[0] != types.Artifact {
		t.Errorf("token[1] types = %v, want [Artifact]", food.Types)
	}
	if len(food.Subtypes) != 1 || food.Subtypes[0] != types.Food {
		t.Errorf("token[1] subtypes = %v, want [Food]", food.Subtypes)
	}
	if len(food.ActivatedAbilities) != 1 {
		t.Errorf("token[1] activated abilities = %d, want 1 (Food sac-for-life)", len(food.ActivatedAbilities))
	}
}

// TestLowerMultiTokenSharedFixedPredefined proves the shared-count multi-token
// path also creates two predefined artifact tokens under a single fixed count
// (Madame Vastra's "create a Clue token and a Food token."). Both reuse the
// predefined token definitions and each emits one token.
func TestLowerMultiTokenSharedFixedPredefined(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Vastra",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create a Clue token and a Food token.",
	})
	creates := multiTokenCreateInstructions(t, face)
	if len(creates) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(creates))
	}
	want := []struct {
		name string
		sub  types.Sub
	}{
		{"Clue", types.Clue},
		{"Food", types.Food},
	}
	for i, w := range want {
		if creates[i].Amount.IsDynamic() || creates[i].Amount.Value() != 1 {
			t.Errorf("token[%d] amount = %v, want fixed 1", i, creates[i].Amount)
		}
		def := createTokenDef(t, creates[i])
		if def.Name != w.name {
			t.Errorf("token[%d] name = %q, want %q", i, def.Name, w.name)
		}
		if len(def.Types) != 1 || def.Types[0] != types.Artifact {
			t.Errorf("token[%d] types = %v, want [Artifact]", i, def.Types)
		}
		if len(def.Subtypes) != 1 || def.Subtypes[0] != w.sub {
			t.Errorf("token[%d] subtypes = %v, want [%v]", i, def.Subtypes, w.sub)
		}
	}
}

// TestLowerMultiTokenMixedCountFailsClosed proves a multi-token clause whose
// specs do not share one representable count ("a 1/1 ... and X ... tokens") does
// not lower to the shared-count path — the specs cannot carry one Quantity, so
// the clause fails closed rather than mis-counting a token type.
func TestLowerMultiTokenMixedCountFailsClosed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Mixed",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{X}{G}",
		OracleText: "Create a 1/1 green Saproling creature token and X Food tokens.",
		Colors:     []string{"G"},
	})
	if face.SpellAbility.Exists {
		t.Fatal("mixed-count multi-token clause lowered instead of failing closed")
	}
}
