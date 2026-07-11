package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

// targetsSourceLifeTaxPermanent gives playerID a battlefield permanent whose
// static ability charges opponents extra life to cast spells that target it
// ("Spells your opponents cast that target this creature cost an additional N
// life to cast.", Terror of the Peaks).
func targetsSourceLifeTaxPermanent(g *game.Game, playerID game.PlayerID, life int) *game.Permanent {
	return addCombatPermanent(g, playerID, &game.CardDef{CardFace: game.CardFace{
		Name: "Test Terror of the Peaks",
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCostModifier,
				AffectedPlayer: game.PlayerOpponent,
				CostModifier: game.CostModifier{
					Kind:          game.CostModifierSpell,
					TargetsSource: true,
					LifeIncrease:  life,
				},
			}},
		}},
	}})
}

func spellLifeIncreaseForCaster(g *game.Game, caster game.PlayerID, card *game.CardDef, targets []game.Target) int {
	total := 0
	for _, modifier := range staticCostModifiersForContext(g, caster, card, zone.Hand, targets) {
		total += modifier.LifeIncrease
	}
	return total
}

func TestSpellCostModifierTargetsSourceLifeTaxesOpponentSpellsTargetingSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := targetsSourceLifeTaxPermanent(g, game.Player1, 3)
	card := &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Bolt",
		Types:    []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
	}}

	targetingSource := []game.Target{game.PermanentTarget(source.ObjectID)}
	if got := spellLifeIncreaseForCaster(g, game.Player2, card, targetingSource); got != 3 {
		t.Fatalf("opponent spell targeting source life increase = %d, want 3", got)
	}

	other := []game.Target{game.PermanentTarget(g.IDGen.Next())}
	if got := spellLifeIncreaseForCaster(g, game.Player2, card, other); got != 0 {
		t.Fatalf("opponent spell targeting another permanent life increase = %d, want 0", got)
	}

	if got := spellLifeIncreaseForCaster(g, game.Player2, card, nil); got != 0 {
		t.Fatalf("opponent spell with no targets life increase = %d, want 0", got)
	}

	// The controller is not an opponent, so the tax never applies to their own
	// spells even when they target the source.
	if got := spellLifeIncreaseForCaster(g, game.Player1, card, targetingSource); got != 0 {
		t.Fatalf("controller spell targeting source life increase = %d, want 0", got)
	}
}

// TestSpellCostModifierTargetsSourceLifeGatesOpponentSpellByLifeTotal proves the
// life tax is a real additional cost: an opponent whose spell targets the source
// can only pay when their life total covers the extra life, while a spell that
// does not target the source is unaffected.
func TestSpellCostModifierTargetsSourceLifeGatesOpponentSpellByLifeTotal(t *testing.T) {
	freeInstant := func() *game.CardDef {
		return &game.CardDef{CardFace: game.CardFace{
			Name:     "Test Free Instant",
			Types:    []types.Card{types.Instant},
			ManaCost: opt.Val(cost.Mana{}),
		}}
	}

	targetOf := func(g *game.Game, source *game.Permanent, targetSource bool) []game.Target {
		if targetSource {
			return []game.Target{game.PermanentTarget(source.ObjectID)}
		}
		return []game.Target{game.PermanentTarget(g.IDGen.Next())}
	}

	tests := []struct {
		name         string
		life         int
		targetSource bool
		wantPayable  bool
	}{
		{name: "ample life pays the tax", life: 20, targetSource: true, wantPayable: true},
		{name: "life below the tax cannot pay", life: 2, targetSource: true, wantPayable: false},
		{name: "low life untaxed when not targeting source", life: 2, targetSource: false, wantPayable: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := targetsSourceLifeTaxPermanent(g, game.Player1, 3)
			g.Players[game.Player2].Life = test.life
			cardID := addCardToHand(g, game.Player2, freeInstant())

			req := payment.SpellRequest{
				PlayerID:   game.Player2,
				CardID:     cardID,
				SourceZone: zone.Hand,
				Card:       g.CardInstances[cardID].Def,
				Targets:    targetOf(g, source, test.targetSource),
			}
			options := paymentOrch.planner(g).PayableSpellOptions(req)
			payable := len(options) > 0
			if payable != test.wantPayable {
				t.Fatalf("PayableSpellOptions payable = %v, want %v", payable, test.wantPayable)
			}
		})
	}
}
