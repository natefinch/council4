package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestGenerateCuriousHerdControlledTokenCount proves the full pipeline emits
// Curious Herd's executable source with no diagnostics: a targeted opponent, a
// controller-recipient CreateToken, and a count group anchored to the target
// player. It guards the end-to-end "You create X ... tokens, where X is the
// number of artifacts that player controls." generation.
func TestGenerateCuriousHerdControlledTokenCount(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Curious Herd",
		Layout:     "normal",
		ManaCost:   "{3}{G}",
		TypeLine:   "Instant",
		OracleText: "Choose target opponent. You create X 3/3 green Beast creature tokens, where X is the number of artifacts that player controls.",
		Colors:     []string{"G"},
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.CreateToken{",
		"game.DynamicAmountCountSelector",
		"game.PlayerControlledGroup(game.TargetPlayerReference(0), game.Selection{RequiredTypes: []types.Card{types.Artifact}})",
		`Constraint: "target opponent"`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

// TestLowerControllerTokenCountThatPlayerControls proves that a controller-
// recipient token count scoped to a chosen player's permanents ("You create X
// ... tokens, where X is the number of artifacts that player controls.", Curious
// Herd) lowers to a CreateToken whose amount counts the target player's matching
// permanents. It backs the token-count mirror of the Anathemancer damage amount:
// the lone target scopes the count, the tokens still enter under the controller,
// and the "that player" count reference is folded into the count group's
// target-player anchor.
func TestLowerControllerTokenCountThatPlayerControls(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Controlled Herd",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Choose target opponent. You create X 3/3 green Beast creature tokens, where X is the number of artifacts that player controls.",
		Colors:     []string{"G"},
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(mode.Targets))
	}
	if mode.Targets[0].Allow != game.TargetAllowPlayer {
		t.Fatalf("target allow = %v, want TargetAllowPlayer", mode.Targets[0].Allow)
	}
	if !mode.Targets[0].Selection.Exists || mode.Targets[0].Selection.Val.Player != game.PlayerOpponent {
		t.Fatalf("target selection = %#v, want opponent", mode.Targets[0].Selection)
	}
	create := createTokenPrimitive(t, face)
	if create.Recipient.Exists {
		t.Fatalf("recipient = %#v, want unset (controller)", create.Recipient)
	}
	if !create.Amount.IsDynamic() {
		t.Fatalf("amount = %#v, want dynamic count", create.Amount)
	}
	dyn := create.Amount.DynamicAmount().Val
	if dyn.Kind != game.DynamicAmountCountSelector {
		t.Fatalf("amount kind = %v, want DynamicAmountCountSelector", dyn.Kind)
	}
	anchor, ok := dyn.Group.PlayerAnchor()
	if !ok || anchor.Kind() != game.PlayerReferenceTargetPlayer || anchor.TargetIndex() != 0 {
		t.Fatalf("count anchor = %#v, want TargetPlayerReference(0)", anchor)
	}
	if dyn.Group.Domain() != game.GroupDomainPlayerControlled {
		t.Fatalf("count domain = %v, want GroupDomainPlayerControlled", dyn.Group.Domain())
	}
	selection := dyn.Group.Selection()
	if len(selection.RequiredTypes) != 1 || selection.RequiredTypes[0] != types.Artifact {
		t.Fatalf("count selection required types = %v, want [Artifact]", selection.RequiredTypes)
	}
}

// TestLowerControllerTokenCountThatPlayerControlsFailsClosed proves the
// generalization stays narrow: a controller-recipient token count with no target
// player ("... where X is the number of artifacts you control.") is not treated
// as a scoped count and still lowers through the ordinary "you control"
// battlefield count, while a plain fixed count carries no count group at all.
func TestLowerControllerTokenCountThatPlayerControlsFailsClosed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test You Control Herd",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "You create X 1/1 green Beast creature tokens, where X is the number of artifacts you control.",
		Colors:     []string{"G"},
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 {
		t.Fatalf("targets = %d, want 0 (no target for a you-control count)", len(mode.Targets))
	}
	create := createTokenPrimitive(t, face)
	dyn := create.Amount.DynamicAmount().Val
	if dyn.Group.Domain() != game.GroupDomainBattlefield {
		t.Fatalf("you-control count domain = %v, want GroupDomainBattlefield", dyn.Group.Domain())
	}
	if _, ok := dyn.Group.PlayerAnchor(); ok {
		t.Fatal("you-control count must carry no player anchor")
	}
}
