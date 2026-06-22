package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// astralCornucopiaPermanent adds a permanent whose tap mana ability lets the
// controller choose a color and adds one mana of that color for each charge
// counter on itself (Astral Cornucopia).
func astralCornucopiaPermanent(g *game.Game, controller game.PlayerID, charges int) *game.Permanent {
	def := &game.CardDef{CardFace: game.CardFace{
		Name:  "Astral Cornucopia",
		Types: []types.Card{types.Artifact},
	}}
	ability := game.TapManaChosenColorDynamicAbility("", game.DynamicAmount{
		Kind:        game.DynamicAmountObjectCounters,
		Multiplier:  1,
		CounterKind: counter.Charge,
		Object:      game.SourcePermanentReference(),
	})
	def.ManaAbilities = append(def.ManaAbilities, ability)
	cardID := g.IDGen.Next()
	card := &game.CardInstance{ID: cardID, Def: def, Owner: controller}
	g.CardInstances[cardID] = card
	permanent, ok := createCardPermanent(g, card, controller, zone.Stack)
	if !ok {
		panic("astral cornucopia permanent was not created")
	}
	permanent.Counters.Add(counter.Charge, charges)
	return permanent
}

func TestChosenColorSourceCounterManaScalesWithCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cornucopia := astralCornucopiaPermanent(g, game.Player1, 4)

	act := action.ActivateAbility(cornucopia.ObjectID, 0, nil, 0)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(astral cornucopia mana ability) = false, want true")
	}

	pool := g.Players[game.Player1].ManaPool
	colored := pool.Amount(mana.W) + pool.Amount(mana.U) + pool.Amount(mana.B) +
		pool.Amount(mana.R) + pool.Amount(mana.G)
	if colored != 4 {
		t.Fatalf("colored mana = %d, want 4 (one of the chosen color per charge counter)", colored)
	}
	if got := pool.Amount(mana.C); got != 0 {
		t.Fatalf("colorless mana = %d, want 0 (the chosen color is one of the five colors)", got)
	}
}

// TestChosenColorSourceCounterManaWithoutCountersAddsNothing verifies the ability
// produces no mana when the source has no counters of the kind, so the chosen
// color is offered but nothing is added.
func TestChosenColorSourceCounterManaWithoutCountersAddsNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cornucopia := astralCornucopiaPermanent(g, game.Player1, 0)

	act := action.ActivateAbility(cornucopia.ObjectID, 0, nil, 0)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(astral cornucopia mana ability) = false, want true")
	}

	pool := g.Players[game.Player1].ManaPool
	total := pool.Amount(mana.W) + pool.Amount(mana.U) + pool.Amount(mana.B) +
		pool.Amount(mana.R) + pool.Amount(mana.G) + pool.Amount(mana.C)
	if total != 0 {
		t.Fatalf("total mana = %d, want 0 (no charge counters)", total)
	}
}
