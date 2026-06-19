package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// sourceSpellReductionCard models a spell that costs PerObjectReduction generic
// less to cast for each battlefield permanent matching selection, encoded as the
// AffectedSource spell cost modifier the cardgen backend emits for the
// "This spell costs {N} less to cast for each <object>" ability.
func sourceSpellReductionCard(name string, manaCost cost.Mana, selection game.Selection, perObject int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(manaCost),
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCostModifier,
				AffectedSource: true,
				CostModifier: game.CostModifier{
					Kind:               game.CostModifierSpell,
					PerObjectReduction: perObject,
					CountSelection:     selection,
				},
			}},
		}},
	}}
}

// sourceSpellGenericReduction sums the generic reductions the rules engine
// resolves for casting card from the player's hand, which for a clean game is
// exactly the source-scoped per-object reduction.
func sourceSpellGenericReduction(g *game.Game, playerID game.PlayerID, card *game.CardDef) int {
	state := &rulesPaymentState{g: g}
	total := 0
	for _, modifier := range state.CostModifiersForSpell(playerID, card, 0, zone.Hand) {
		total += modifier.GenericReduction
	}
	return total
}

func anyCreatureSelection() game.Selection {
	return game.Selection{RequiredTypes: []types.Card{types.Creature}}
}

func TestSourceSpellCostReductionZeroCreaturesNoReduction(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	card := sourceSpellReductionCard("Blasphemous Act", cost.Mana{cost.O(8), cost.R}, anyCreatureSelection(), 1)

	if got := sourceSpellGenericReduction(g, game.Player1, card); got != 0 {
		t.Fatalf("reduction with no creatures = %d, want 0", got)
	}
}

func TestSourceSpellCostReductionPerCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCreaturePermanent(g, game.Player1)
	addCreaturePermanent(g, game.Player1)
	addCreaturePermanent(g, game.Player2)
	card := sourceSpellReductionCard("Blasphemous Act", cost.Mana{cost.O(8), cost.R}, anyCreatureSelection(), 1)

	if got := sourceSpellGenericReduction(g, game.Player1, card); got != 3 {
		t.Fatalf("reduction with three battlefield creatures = %d, want 3", got)
	}
}

func TestSourceSpellCostReductionCountsControllerScopedSelection(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCreaturePermanent(g, game.Player1)
	addCreaturePermanent(g, game.Player1)
	addCreaturePermanent(g, game.Player2)
	selection := anyCreatureSelection()
	selection.Controller = game.ControllerOpponent
	card := sourceSpellReductionCard("Primeval Protector", cost.Mana{cost.O(6), cost.G}, selection, 1)

	if got := sourceSpellGenericReduction(g, game.Player1, card); got != 1 {
		t.Fatalf("reduction for each creature opponents control = %d, want 1", got)
	}
}

func TestSourceSpellCostReductionGenericFloorsAtZeroKeepsColored(t *testing.T) {
	makeGame := func(creatures int) (*game.Game, *game.CardDef) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		for range creatures {
			addCreaturePermanent(g, game.Player1)
		}
		addBasicLandPermanent(g, game.Player1, types.Mountain)
		card := sourceSpellReductionCard("Blasphemous Act", cost.Mana{cost.O(2), cost.R}, anyCreatureSelection(), 1)
		return g, card
	}

	t.Run("over-reduction floors generic at zero", func(t *testing.T) {
		g, card := makeGame(5)
		if !canPayTestSpellCosts(g, testSpellPaymentRequest{playerID: game.Player1, card: card, sourceZone: zone.Hand}) {
			t.Fatal("canPaySpellCosts() = false; five creatures should floor {2} to zero leaving {R} payable by one Mountain")
		}
	})

	t.Run("colored requirement preserved", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		for range 5 {
			addCreaturePermanent(g, game.Player1)
		}
		addBasicLandPermanent(g, game.Player1, types.Forest)
		card := sourceSpellReductionCard("Blasphemous Act", cost.Mana{cost.O(2), cost.R}, anyCreatureSelection(), 1)
		if canPayTestSpellCosts(g, testSpellPaymentRequest{playerID: game.Player1, card: card, sourceZone: zone.Hand}) {
			t.Fatal("canPaySpellCosts() = true; the {R} requirement must survive the generic reduction and a Forest cannot pay it")
		}
	})

	t.Run("no reduction below the printed cost without creatures", func(t *testing.T) {
		g, card := makeGame(0)
		if canPayTestSpellCosts(g, testSpellPaymentRequest{playerID: game.Player1, card: card, sourceZone: zone.Hand}) {
			t.Fatal("canPaySpellCosts() = true; {2}{R} needs three mana when no creatures reduce it")
		}
	})
}

func TestSourceSpellCostReductionAppliesOnlyToSourceSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCreaturePermanent(g, game.Player1)
	addCreaturePermanent(g, game.Player1)
	// A permanent on the battlefield that carries the self-scoped reduction static
	// ability must not reduce the cost of other spells the controller casts.
	addCombatPermanent(g, game.Player1, sourceSpellReductionCard(
		"Primeval Protector", cost.Mana{cost.O(6), cost.G}, anyCreatureSelection(), 1))

	other := &game.CardDef{CardFace: game.CardFace{
		Name:     "Unrelated Spell",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.O(3), cost.R}),
	}}

	if got := sourceSpellGenericReduction(g, game.Player1, other); got != 0 {
		t.Fatalf("a battlefield self-reduction leaked %d generic onto an unrelated spell, want 0", got)
	}
}
