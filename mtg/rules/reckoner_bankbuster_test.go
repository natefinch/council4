package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func bankbusterTestDef() *game.CardDef {
	treasure := &game.CardDef{CardFace: game.CardFace{
		Name:     "Treasure",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Treasure},
	}}
	pilot := &game.CardDef{CardFace: game.CardFace{
		Name:      "Pilot",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Pilot},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{{
			CrewPowerBonus: 2,
		}},
	}}
	noChargeCounters := game.EffectCondition{Condition: opt.Val(game.Condition{
		Negate: true,
		Object: opt.Val(game.SourcePermanentReference()),
		ObjectMatches: opt.Val(game.Selection{
			RequiredCounter: counter.Charge,
			RequiredCounterCount: opt.Val(compare.Int{
				Op:    compare.GreaterOrEqual,
				Value: 1,
			}),
		}),
	})}
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Reckoner Bankbuster",
		Types:     []types.Card{types.Artifact},
		Subtypes:  []types.Sub{types.Vehicle},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
		ActivatedAbilities: []game.ActivatedAbility{{
			ManaCost: opt.Val(cost.Mana{cost.O(2)}),
			AdditionalCosts: []cost.Additional{
				{Kind: cost.AdditionalTap},
				{Kind: cost.AdditionalRemoveCounter, Amount: 1, CounterKind: counter.Charge},
			},
			Content: game.Mode{Sequence: []game.Instruction{
				{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}},
				{Primitive: game.CreateToken{Amount: game.Fixed(1), Source: game.TokenDef(treasure)}, Condition: opt.Val(noChargeCounters)},
				{Primitive: game.CreateToken{Amount: game.Fixed(1), Source: game.TokenDef(pilot)}, Condition: opt.Val(noChargeCounters)},
			}}.Ability(),
		}},
	}}
}

func TestReckonerBankbusterRepeatedActivationsAndThreshold(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	bankbuster := addCombatPermanent(g, game.Player1, bankbusterTestDef())
	bankbuster.Counters.Add(counter.Charge, 3)
	for range 3 {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Card"}})
	}

	for activation := 1; activation <= 3; activation++ {
		bankbuster.Tapped = false
		g.Players[game.Player1].ManaPool.Add(mana.C, 2)
		if !engine.applyAction(g, game.Player1, action.ActivateAbility(bankbuster.ObjectID, 0, nil, 0)) {
			t.Fatalf("activation %d rejected", activation)
		}
		if got := bankbuster.Counters.Get(counter.Charge); got != 3-activation {
			t.Fatalf("counters after paying activation %d = %d, want %d", activation, got, 3-activation)
		}
		engine.resolveTopOfStack(g, nil)
		if got := g.Players[game.Player1].Hand.Size(); got != activation {
			t.Fatalf("hand size after activation %d = %d, want %d", activation, got, activation)
		}
		wantTokens := 0
		if activation == 3 {
			wantTokens = 1
		}
		if got := countTokensNamed(g, "Treasure", game.Player1); got != wantTokens {
			t.Fatalf("Treasure tokens after activation %d = %d, want %d", activation, got, wantTokens)
		}
		if got := countTokensNamed(g, "Pilot", game.Player1); got != wantTokens {
			t.Fatalf("Pilot tokens after activation %d = %d, want %d", activation, got, wantTokens)
		}
	}

	bankbuster.Tapped = false
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	if engine.applyAction(g, game.Player1, action.ActivateAbility(bankbuster.ObjectID, 0, nil, 0)) {
		t.Fatal("activation without a charge counter succeeded")
	}
	if bankbuster.Tapped {
		t.Fatal("failed activation tapped Bankbuster")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.C); got != 2 {
		t.Fatalf("failed activation spent mana: %d remains, want 2", got)
	}
}
