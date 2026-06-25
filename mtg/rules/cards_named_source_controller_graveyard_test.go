package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// growthCyclePumpDef mirrors the dynamic clause of the generated Growth Cycle
// card: a target creature gets +2/+2 for each card named Growth Cycle in the
// controller's graveyard (DynamicAmountCardsNamedSourceInControllerGraveyard).
func growthCyclePumpDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Growth Cycle",
		Types: []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target creature",
				Allow:      game.TargetAllowPermanent,
			}},
			Sequence: []game.Instruction{
				{Primitive: game.ModifyPT{
					Object: game.TargetPermanentReference(0),
					PowerDelta: game.Dynamic(game.DynamicAmount{
						Kind:       game.DynamicAmountCardsNamedSourceInControllerGraveyard,
						Multiplier: 2,
					}),
					ToughnessDelta: game.Dynamic(game.DynamicAmount{
						Kind:       game.DynamicAmountCardsNamedSourceInControllerGraveyard,
						Multiplier: 2,
					}),
					Duration: game.DurationUntilEndOfTurn,
				}},
			},
		}.Ability()),
	}}
}

// resolveControllerGraveyardCount resolves a stack object whose source is the
// given def and returns the power delta a single ModifyPT instruction would
// apply, computed from the controller's graveyard alone.
func controllerGraveyardCount(g *game.Game, controller game.PlayerID) int {
	spellID := g.IDGen.Next()
	g.CardInstances[spellID] = &game.CardInstance{
		ID:    spellID,
		Def:   growthCyclePumpDef(),
		Owner: controller,
	}
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   spellID,
		Controller: controller,
	}
	dynamic := game.DynamicAmount{Kind: game.DynamicAmountCardsNamedSourceInControllerGraveyard}
	return dynamicAmountValue(g, obj, controller, dynamic)
}

// TestCardsNamedSourceInControllerGraveyardCountsControllerOnly verifies that
// the controller-scoped count includes only the resolving controller's
// graveyard, ignoring cards named the source in other players' graveyards.
func TestCardsNamedSourceInControllerGraveyardCountsControllerOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Two copies in the controller's graveyard, one in the opponent's, and a
	// differently named card in the controller's graveyard.
	addNamedCardToGraveyard(g, game.Player1, "Growth Cycle")
	addNamedCardToGraveyard(g, game.Player1, "Growth Cycle")
	addNamedCardToGraveyard(g, game.Player2, "Growth Cycle")
	addNamedCardToGraveyard(g, game.Player1, "Lightning Bolt")

	if got := controllerGraveyardCount(g, game.Player1); got != 2 {
		t.Fatalf("controller graveyard count = %d, want 2 (only Player1's graveyard)", got)
	}
}

// TestCardsNamedSourceInControllerGraveyardEmptyIsZero verifies the count is
// zero when the controller's graveyard holds no matching card, even if an
// opponent's graveyard does.
func TestCardsNamedSourceInControllerGraveyardEmptyIsZero(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addNamedCardToGraveyard(g, game.Player2, "Growth Cycle")

	if got := controllerGraveyardCount(g, game.Player1); got != 0 {
		t.Fatalf("controller graveyard count = %d, want 0", got)
	}
}
