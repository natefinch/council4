package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerCounterSpellThatTargets(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		want       []game.SpellTargetRequirement
	}{
		{
			name:       "a creature",
			oracleText: "Counter target spell that targets a creature.",
			want: []game.SpellTargetRequirement{{
				Kind:          game.SpellTargetRequirementPermanent,
				RequiredTypes: []types.Card{types.Creature},
			}},
		},
		{
			name:       "a permanent you control",
			oracleText: "Counter target spell that targets a permanent you control.",
			want: []game.SpellTargetRequirement{{
				Kind:       game.SpellTargetRequirementPermanent,
				Controller: game.ControllerYou,
			}},
		},
		{
			name:       "a creature you control",
			oracleText: "Counter target spell that targets a creature you control.",
			want: []game.SpellTargetRequirement{{
				Kind:          game.SpellTargetRequirementPermanent,
				RequiredTypes: []types.Card{types.Creature},
				Controller:    game.ControllerYou,
			}},
		},
		{
			name:       "an enchantment",
			oracleText: "Counter target spell that targets an enchantment.",
			want: []game.SpellTargetRequirement{{
				Kind:          game.SpellTargetRequirementPermanent,
				RequiredTypes: []types.Card{types.Enchantment},
			}},
		},
		{
			name:       "a player",
			oracleText: "Counter target spell that targets a player.",
			want: []game.SpellTargetRequirement{{
				Kind: game.SpellTargetRequirementPlayer,
			}},
		},
		{
			name:       "you",
			oracleText: "Counter target spell that targets you.",
			want: []game.SpellTargetRequirement{{
				Kind:   game.SpellTargetRequirementPlayer,
				Player: game.PlayerYou,
			}},
		},
		{
			name:       "you or a permanent you control",
			oracleText: "Counter target spell that targets you or a permanent you control.",
			want: []game.SpellTargetRequirement{
				{
					Kind:   game.SpellTargetRequirementPlayer,
					Player: game.PlayerYou,
				},
				{
					Kind:       game.SpellTargetRequirementPermanent,
					Controller: game.ControllerYou,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Counter Targets",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: test.oracleText,
			})
			if !face.SpellAbility.Exists {
				t.Fatal("spell ability missing")
			}
			ability := face.SpellAbility.Val
			if len(ability.Modes) != 1 {
				t.Fatalf("modes = %d, want 1", len(ability.Modes))
			}
			mode := ability.Modes[0]
			if len(mode.Targets) != 1 {
				t.Fatalf("targets = %d, want 1", len(mode.Targets))
			}
			got := mode.Targets[0].Predicate.SpellTargets
			if len(got) != len(test.want) {
				t.Fatalf("spell targets = %+v, want %+v", got, test.want)
			}
			for i := range got {
				if got[i].Kind != test.want[i].Kind ||
					!slices.Equal(got[i].RequiredTypes, test.want[i].RequiredTypes) ||
					got[i].Controller != test.want[i].Controller ||
					got[i].Player != test.want[i].Player {
					t.Fatalf("spell target[%d] = %+v, want %+v", i, got[i], test.want[i])
				}
			}
			if _, ok := mode.Sequence[0].Primitive.(game.CounterObject); !ok {
				t.Fatalf("primitive = %T, want game.CounterObject", mode.Sequence[0].Primitive)
			}
		})
	}
}

func TestLowerCounterSpellThatTargetsRejectsUnsupported(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Counter target spell that targets an opponent.",
		"Counter target spell that targets a tapped creature.",
		"Counter target spell that targets a creature or player.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Unsupported Counter Targets",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: text,
			})
			if len(diagnostics) == 0 {
				t.Fatal("unsupported spell-target restriction lowered")
			}
		})
	}
}
