package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// groupPerObjectReducerCard models the dynamic group cast-cost discount the
// cardgen backend emits for "[<filter>] spells you cast cost {N} less to cast for
// each <permanent> you control" (Temur Battlecrier, Hamza, Guardian of Arashin):
// a controller-scoped (non-AffectedSource) spell modifier carrying a per-object
// reduction and battlefield count selection, optionally gated to the controller's
// turn.
func groupPerObjectReducerCard(perObject int, count game.Selection, duringTurn bool) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Group Reducer",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:                           game.RuleEffectCostModifier,
				AffectedPlayer:                 game.PlayerYou,
				RestrictedDuringControllerTurn: duringTurn,
				CostModifier: game.CostModifier{
					Kind:               game.CostModifierSpell,
					PerObjectReduction: perObject,
					CountSelection:     &count,
				},
			}},
		}},
	}}
}

func TestGroupSpellCostReductionPerCreatureYouControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCreaturePermanent(g, game.Player1)
	addCreaturePermanent(g, game.Player1)
	addCreaturePermanent(g, game.Player2)
	addCombatPermanent(g, game.Player1, groupPerObjectReducerCard(1,
		game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}, false))

	spell := &game.CardDef{CardFace: game.CardFace{
		Name:     "Big Spell",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.O(8)}),
	}}
	// Three creatures the reducer's controller controls (two vanilla plus the
	// reducer itself); opponent creatures are not counted.
	if got := sourceSpellGenericReduction(g, game.Player1, spell); got != 3 {
		t.Fatalf("group reduction = %d, want 3", got)
	}
}

func TestGroupSpellCostReductionDuringTurnGate(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCreaturePermanent(g, game.Player1)
	addCombatPermanent(g, game.Player1, groupPerObjectReducerCard(1,
		game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}, true))
	spell := &game.CardDef{CardFace: game.CardFace{
		Name:     "Big Spell",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.O(8)}),
	}}

	g.Turn.ActivePlayer = game.Player1
	if got := sourceSpellGenericReduction(g, game.Player1, spell); got != 2 {
		t.Fatalf("group reduction during your turn = %d, want 2", got)
	}
	g.Turn.ActivePlayer = game.Player2
	if got := sourceSpellGenericReduction(g, game.Player1, spell); got != 0 {
		t.Fatalf("group reduction outside your turn = %d, want 0", got)
	}
}
