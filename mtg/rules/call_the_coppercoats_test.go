package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func coppercoatsDef() *game.CardDef {
	token := &game.CardDef{CardFace: game.CardFace{
		Name:      "Human Soldier",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Human, types.Soldier},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Call the Coppercoats",
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.W}),
		Types:    []types.Card{types.Instant},
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectCostModifier,
			AffectedSource: true,
			CostModifier: game.CostModifier{
				Kind:                         game.CostModifierSpell,
				PerTargetBeyondFirstIncrease: cost.Mana{cost.O(1), cost.W},
			},
		}}}},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 0,
				MaxTargets: 99,
				Constraint: "target opponent",
				Allow:      game.TargetAllowPlayer,
				Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
			}},
			Sequence: []game.Instruction{{Primitive: game.CreateToken{
				Amount: game.Dynamic(game.DynamicAmount{
					Kind:       game.DynamicAmountCountSelector,
					Multiplier: 1,
					Group: game.PlayerGroupControlledGroup(
						game.TargetedPlayersReference(),
						game.Selection{RequiredTypes: []types.Card{types.Creature}},
					),
				}),
				Source: game.TokenDef(token),
			}}},
		}.Ability()),
	}}
}

func setupCoppercoatsCast(t *testing.T, targets []game.Target, colorless, white int) (*Engine, *game.Game) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Players[game.Player1].ManaPool.Add(mana.C, colorless)
	g.Players[game.Player1].ManaPool.Add(mana.W, white)
	spellID := addCardToHand(g, game.Player1, coppercoatsDef())
	if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(spellID, zone.Hand, game.FaceFront, targets, 0, nil)) {
		t.Fatal("cast failed")
	}
	return engine, g
}

func TestCoppercoatsStriveCostByChosenTargetCount(t *testing.T) {
	tests := []struct {
		name      string
		targets   []game.Target
		colorless int
		white     int
		want      bool
	}{
		{"zero", nil, 2, 1, true},
		{"one", []game.Target{game.PlayerTarget(game.Player2)}, 2, 1, true},
		{"two", []game.Target{game.PlayerTarget(game.Player2), game.PlayerTarget(game.Player3)}, 3, 2, true},
		{"two missing white", []game.Target{game.PlayerTarget(game.Player2), game.PlayerTarget(game.Player3)}, 4, 1, false},
		{"three", []game.Target{game.PlayerTarget(game.Player2), game.PlayerTarget(game.Player3), game.PlayerTarget(game.Player4)}, 4, 3, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			g.Turn.Phase = game.PhasePrecombatMain
			g.Players[game.Player1].ManaPool.Add(mana.C, tc.colorless)
			g.Players[game.Player1].ManaPool.Add(mana.W, tc.white)
			spellID := addCardToHand(g, game.Player1, coppercoatsDef())
			act := action.CastSpellFaceFromZone(spellID, zone.Hand, game.FaceFront, tc.targets, 0, nil)
			if got := engine.applyAction(g, game.Player1, act); got != tc.want {
				t.Fatalf("cast = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCoppercoatsTargetEnumerationHonorsTargetability(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Players[game.Player1].ManaPool.Add(mana.C, 10)
	g.Players[game.Player1].ManaPool.Add(mana.W, 10)
	spellID := addCardToHand(g, game.Player1, coppercoatsDef())
	applyPlayerRuleEffect(engine, g, game.Player2, game.RuleEffectPlayerHexproof)
	legal := engine.legalActions(g, game.Player1)
	for _, targets := range [][]game.Target{
		nil,
		{game.PlayerTarget(game.Player3)},
		{game.PlayerTarget(game.Player3), game.PlayerTarget(game.Player4)},
	} {
		act := action.CastSpellFaceFromZone(spellID, zone.Hand, game.FaceFront, targets, 0, nil)
		if !containsAction(legal, act) {
			t.Fatalf("legal actions missing targets %#v", targets)
		}
	}
	if containsAction(legal, action.CastSpellFaceFromZone(
		spellID, zone.Hand, game.FaceFront, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil,
	)) {
		t.Fatal("hexproof opponent remained a legal target")
	}
}

func TestCoppercoatsResolutionUsesStillLegalTargetsAndCurrentControl(t *testing.T) {
	engine, g := setupCoppercoatsCast(t,
		[]game.Target{game.PlayerTarget(game.Player2), game.PlayerTarget(game.Player3)}, 10, 10)
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Creature}}})
	moved := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Creature}}})
	addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Creature}}})
	addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Creature}}})
	moved.Controller = game.Player4
	applyPlayerRuleEffect(engine, g, game.Player2, game.RuleEffectPlayerHexproof)
	addReplacementPermanent(t, g, game.Player1, anyControllerTokenDoublingCardDef())

	engine.resolveTopOfStack(g, &TurnLog{})
	if got := countTokenPermanentsNamed(g, "Human Soldier"); got != 4 {
		t.Fatalf("tokens = %d, want 4 (two Player3 creatures, doubled)", got)
	}

}

func TestCoppercoatsCountsCreatureControllersAtResolution(t *testing.T) {
	engine, g := setupCoppercoatsCast(t, []game.Target{game.PlayerTarget(game.Player2)}, 10, 10)
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Creature}}})
	moved := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Creature}}})
	moved.Controller = game.Player4

	engine.resolveTopOfStack(g, &TurnLog{})
	if got := countTokenPermanentsNamed(g, "Human Soldier"); got != 1 {
		t.Fatalf("tokens = %d, want 1 after one creature changed controllers", got)
	}
}

func TestCoppercoatsFizzlesWhenAllTargetsBecomeIllegal(t *testing.T) {
	engine, g := setupCoppercoatsCast(t, []game.Target{game.PlayerTarget(game.Player2)}, 10, 10)
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Creature}}})
	applyPlayerRuleEffect(engine, g, game.Player2, game.RuleEffectPlayerHexproof)
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := countTokenPermanentsNamed(g, "Human Soldier"); got != 0 {
		t.Fatalf("tokens = %d, want 0 after fizzle", got)
	}
}

func TestCoppercoatsZeroTargetsResolvesLegally(t *testing.T) {
	specs := coppercoatsDef().SpellAbility.Val.Modes[0].Targets
	obj := &game.StackObject{Controller: game.Player1}
	if !stackObjectHasAnyLegalTargetsForSpecs(
		game.NewGame([game.NumPlayers]game.PlayerConfig{}), coppercoatsDef(), 0, specs, obj,
	) {
		t.Fatal("zero-target Coppercoats was treated as fizzled")
	}
}
