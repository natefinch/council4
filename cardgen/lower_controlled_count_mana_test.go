package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerCabalCoffersManaAbility verifies that "Add {B} for each Swamp you
// control" lowers to a mana ability whose AddMana amount is a dynamic count of
// the Swamps the controller has on the battlefield.
func TestLowerCabalCoffersManaAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Cabal Coffers",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{2}, {T}: Add {B} for each Swamp you control.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("mana abilities = %d, want 1", len(face.ManaAbilities))
	}
	sequence := face.ManaAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence = %#v, want single AddMana", sequence)
	}
	add, ok := sequence[0].Primitive.(game.AddMana)
	if !ok || add.ManaColor != mana.B || !add.Amount.IsDynamic() {
		t.Fatalf("mana primitive = %#v", sequence[0].Primitive)
	}
	dynamic := add.Amount.DynamicAmount().Val
	if dynamic.Kind != game.DynamicAmountCountSelector || dynamic.Multiplier != 1 {
		t.Fatalf("dynamic amount = %#v", dynamic)
	}
	selection := dynamic.Group.Selection()
	if selection.Controller != game.ControllerYou ||
		len(selection.SubtypesAny) != 1 || selection.SubtypesAny[0] != types.Swamp {
		t.Fatalf("count selection = %#v", selection)
	}
	if err := game.ValidateInstructionSequence(sequence); err != nil {
		t.Fatalf("instruction sequence invalid: %v", err)
	}
}

// TestLowerControlledCountManaFamily verifies the generic category lowers across
// produced colors and counted permanent filters.
func TestLowerControlledCountManaFamily(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		oracle  string
		color   mana.Color
		require func(game.Selection) bool
	}{
		{
			name:   "Gaea's Cradle",
			oracle: "{T}: Add {G} for each creature you control.",
			color:  mana.G,
			require: func(s game.Selection) bool {
				return s.Controller == game.ControllerYou &&
					len(s.RequiredTypes) == 1 && s.RequiredTypes[0] == types.Creature
			},
		},
		{
			name:   "Serra's Sanctum",
			oracle: "{T}: Add {W} for each enchantment you control.",
			color:  mana.W,
			require: func(s game.Selection) bool {
				return s.Controller == game.ControllerYou &&
					len(s.RequiredTypes) == 1 && s.RequiredTypes[0] == types.Enchantment
			},
		},
		{
			name:   "Tolarian Academy",
			oracle: "{T}: Add {U} for each artifact you control.",
			color:  mana.U,
			require: func(s game.Selection) bool {
				return s.Controller == game.ControllerYou &&
					len(s.RequiredTypes) == 1 && s.RequiredTypes[0] == types.Artifact
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       tc.name,
				Layout:     "normal",
				TypeLine:   "Land",
				OracleText: tc.oracle,
			})
			if len(face.ManaAbilities) != 1 {
				t.Fatalf("mana abilities = %d, want 1", len(face.ManaAbilities))
			}
			sequence := face.ManaAbilities[0].Content.Modes[0].Sequence
			if len(sequence) != 1 {
				t.Fatalf("sequence = %#v", sequence)
			}
			add, ok := sequence[0].Primitive.(game.AddMana)
			if !ok || add.ManaColor != tc.color || !add.Amount.IsDynamic() {
				t.Fatalf("mana primitive = %#v", sequence[0].Primitive)
			}
			dynamic := add.Amount.DynamicAmount().Val
			if dynamic.Kind != game.DynamicAmountCountSelector || !tc.require(dynamic.Group.Selection()) {
				t.Fatalf("dynamic amount = %#v", dynamic)
			}
			if err := game.ValidateInstructionSequence(sequence); err != nil {
				t.Fatalf("instruction sequence invalid: %v", err)
			}
		})
	}
}

// TestLowerCountManaTappedAndOpponentFilters verifies the generic category
// extends to tapped/untapped state and opponent-controlled count selectors
// (Mana Geyser's "Add {R} for each tapped land your opponents control").
func TestLowerCountManaTappedAndOpponentFilters(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		oracle  string
		color   mana.Color
		require func(game.Selection) bool
	}{
		{
			name:   "Mana Geyser",
			oracle: "{T}: Add {R} for each tapped land your opponents control.",
			color:  mana.R,
			require: func(s game.Selection) bool {
				return s.Controller == game.ControllerOpponent &&
					s.Tapped == game.TriTrue &&
					len(s.RequiredTypes) == 1 && s.RequiredTypes[0] == types.Land
			},
		},
		{
			name:   "Tapped Land Source",
			oracle: "{T}: Add {C} for each tapped land you control.",
			color:  mana.C,
			require: func(s game.Selection) bool {
				return s.Controller == game.ControllerYou &&
					s.Tapped == game.TriTrue &&
					len(s.RequiredTypes) == 1 && s.RequiredTypes[0] == types.Land
			},
		},
		{
			name:   "Untapped Creature Source",
			oracle: "{T}: Add {G} for each untapped creature you control.",
			color:  mana.G,
			require: func(s game.Selection) bool {
				return s.Controller == game.ControllerYou &&
					s.Tapped == game.TriFalse &&
					len(s.RequiredTypes) == 1 && s.RequiredTypes[0] == types.Creature
			},
		},
		{
			name:   "Opponent Land Source",
			oracle: "{T}: Add {R} for each land your opponents control.",
			color:  mana.R,
			require: func(s game.Selection) bool {
				return s.Controller == game.ControllerOpponent &&
					s.Tapped == game.TriAny &&
					len(s.RequiredTypes) == 1 && s.RequiredTypes[0] == types.Land
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       tc.name,
				Layout:     "normal",
				TypeLine:   "Land",
				OracleText: tc.oracle,
			})
			if len(face.ManaAbilities) != 1 {
				t.Fatalf("mana abilities = %d, want 1", len(face.ManaAbilities))
			}
			sequence := face.ManaAbilities[0].Content.Modes[0].Sequence
			add, ok := sequence[0].Primitive.(game.AddMana)
			if !ok || add.ManaColor != tc.color || !add.Amount.IsDynamic() {
				t.Fatalf("mana primitive = %#v", sequence[0].Primitive)
			}
			dynamic := add.Amount.DynamicAmount().Val
			if dynamic.Kind != game.DynamicAmountCountSelector || !tc.require(dynamic.Group.Selection()) {
				t.Fatalf("dynamic amount = %#v", dynamic.Group.Selection())
			}
			if err := game.ValidateInstructionSequence(sequence); err != nil {
				t.Fatalf("instruction sequence invalid: %v", err)
			}
		})
	}
}

// TestLowerControlledCountManaFailsClosed verifies that near-miss bodies do not
// lower a mana ability and fail closed with a diagnostic.
func TestLowerControlledCountManaFailsClosed(t *testing.T) {
	t.Parallel()
	for name, oracle := range map[string]string{
		"multi symbol": "{T}: Add {G}{G} for each creature you control.",
		"any color":    "{T}: Add one mana of any color for each creature you control.",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Near Miss " + name,
				Layout:     "normal",
				TypeLine:   "Land",
				OracleText: oracle,
			}, "k")
			if err == nil && len(diagnostics) == 0 {
				t.Fatalf("expected fail-closed diagnostics for %q", oracle)
			}
		})
	}
}
