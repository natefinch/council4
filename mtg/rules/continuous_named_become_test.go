package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestNamedBecomeRenamesAndAddsSupertype verifies that a permanent named-become
// continuous effect ("becomes a 6/6 legendary Horror creature named Fenric")
// changes the affected permanent's effective name at the text layer and adds the
// legendary supertype at the type layer, so both characteristics are observable
// through the continuous layer system.
func TestNamedBecomeRenamesAndAddsSupertype(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addPermanentForSBA(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Grizzly Bears",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})

	g.ContinuousEffects = append(g.ContinuousEffects,
		game.ContinuousEffect{
			ID:               1,
			AffectedObjectID: creature.ObjectID,
			Layer:            game.LayerText,
			Duration:         game.DurationPermanent,
			SetName:          "Fenric",
		},
		game.ContinuousEffect{
			ID:               2,
			AffectedObjectID: creature.ObjectID,
			Layer:            game.LayerType,
			Duration:         game.DurationPermanent,
			AddSupertypes:    []types.Super{types.Legendary},
		},
	)

	if got := permanentEffectiveName(g, creature); got != "Fenric" {
		t.Fatalf("effective name = %q, want renamed Fenric", got)
	}
	if !permanentHasSupertype(g, creature, types.Legendary) {
		t.Fatal("named-become effect did not add the legendary supertype")
	}
}

// TestNamedBecomeTriggersLegendRule verifies that two distinct creatures both
// renamed Fenric and made legendary by named-become effects collide under the
// legend rule, proving the new text-layer name and added supertype are visible
// to state-based actions just like printed names and supertypes.
func TestNamedBecomeTriggersLegendRule(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addPermanentForSBA(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Grizzly Bears",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	second := addPermanentForSBA(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Runeclaw Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})

	for i, target := range []*game.Permanent{first, second} {
		base := id.ID(10 * (i + 1))
		g.ContinuousEffects = append(g.ContinuousEffects,
			game.ContinuousEffect{
				ID:               base,
				AffectedObjectID: target.ObjectID,
				Layer:            game.LayerText,
				Duration:         game.DurationPermanent,
				SetName:          "Fenric",
			},
			game.ContinuousEffect{
				ID:               base + 1,
				AffectedObjectID: target.ObjectID,
				Layer:            game.LayerType,
				Duration:         game.DurationPermanent,
				AddSupertypes:    []types.Super{types.Legendary},
			},
		)
	}

	changed, deaths := checkLegendaryRuleStateBasedActions(g, newPassBatchID(g))

	if !changed {
		t.Fatal("checkLegendaryRuleStateBasedActions() = false, want true for two Fenrics")
	}
	if _, ok := permanentByObjectID(g, first.ObjectID); !ok {
		t.Fatal("oldest Fenric should remain on battlefield")
	}
	if _, ok := permanentByObjectID(g, second.ObjectID); ok {
		t.Fatal("newer duplicate Fenric remained on battlefield")
	}
	if len(deaths) != 1 || deaths[0].Permanent != second.ObjectID || deaths[0].Reason != PermanentDeathReasonLegendaryRule {
		t.Fatalf("death logs = %+v, want newer duplicate legend-rule death", deaths)
	}
}
