package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerCallTheCoppercoats(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Call the Coppercoats",
		Layout:     "normal",
		ManaCost:   "{2}{W}",
		TypeLine:   "Instant",
		OracleText: "Strive — This spell costs {1}{W} more to cast for each target beyond the first.\nChoose any number of target opponents. Create X 1/1 white Human Soldier creature tokens, where X is the number of creatures those opponents control.",
		Colors:     []string{"W"},
	}
	face := lowerSingleFace(t, card)
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	modifier := face.StaticAbilities[0].Body.RuleEffects[0].CostModifier
	if !slices.Equal(modifier.PerTargetBeyondFirstIncrease, cost.Mana{cost.O(1), cost.W}) {
		t.Fatalf("strive cost = %#v", modifier.PerTargetBeyondFirstIncrease)
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || mode.Targets[0].MinTargets != 0 || mode.Targets[0].MaxTargets != 99 {
		t.Fatalf("targets = %#v", mode.Targets)
	}
	create := createTokenPrimitive(t, face)
	if create.Amount.DynamicAmount().Val.Group.Domain() != game.GroupDomainPlayerGroupControlled {
		t.Fatalf("count group = %#v", create.Amount.DynamicAmount().Val.Group)
	}
	group, ok := create.Amount.DynamicAmount().Val.Group.PlayerGroup()
	if !ok || group.Kind != game.PlayerGroupReferenceTargetedPlayers {
		t.Fatalf("player group = %#v", group)
	}
	token, ok := create.Source.TokenDefRef()
	if !ok || token.Power.Val.Value != 1 || token.Toughness.Val.Value != 1 ||
		!slices.Equal(token.Subtypes, []types.Sub{types.Human, types.Soldier}) {
		t.Fatalf("token = %#v", token)
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "c")
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("generate: err=%v diagnostics=%#v", err, diagnostics)
	}
	for _, want := range []string{
		"PerTargetBeyondFirstIncrease: cost.Mana{cost.O(1), cost.W}",
		"game.PlayerGroupControlledGroup(game.TargetedPlayersReference()",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
