package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// cagedSunSource builds a Caged Sun permanent whose entry-time color choice is
// fixed to chosen and whose chosen-color land-mana trigger adds an additional
// mana of that same chosen color.
func cagedSunSource(g *game.Game, controller game.PlayerID, chosen mana.Color) *game.Permanent {
	source := addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Caged Sun",
		Types: []types.Card{types.Artifact},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:                                   game.EventPermanentTapped,
					Controller:                              game.TriggerControllerYou,
					RequireTappedForMana:                    true,
					RequireProducedManaColorFromEntryChoice: true,
					SubjectSelection:                        game.Selection{RequiredTypes: []types.Card{types.Land}},
				},
			},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.AddMana{Amount: game.Fixed(1), EntryChoiceFrom: game.EntryColorChoiceKey},
			}}}.Ability(),
		}},
	}})
	source.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryColorChoiceKey: {Kind: game.ResolutionChoiceMana, Color: chosen},
	}
	return source
}

func basicColorLand(g *game.Game, controller game.PlayerID, name string, subtype types.Sub, m mana.Color) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:          name,
		Types:         []types.Card{types.Land},
		Subtypes:      []types.Sub{subtype},
		ManaAbilities: []game.ManaAbility{game.TapManaAbility(m)},
	}})
}

// TestCagedSunTriggerDoublesChosenColorLandMana proves Caged Sun's trigger:
// tapping a land for the chosen color (red) yields the land's {R} plus the
// trigger's additional {R} routed through the source's entry color choice.
func TestCagedSunTriggerDoublesChosenColorLandMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cagedSunSource(g, game.Player1, mana.R)
	mountain := basicColorLand(g, game.Player1, "Mountain", types.Mountain, mana.R)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(mountain.ObjectID, 0, nil, 0)) {
		t.Fatal("activating Mountain mana ability = false, want true")
	}
	engine.putTriggeredAbilitiesOnStack(g)
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].ManaPool.Amount(mana.R); got != 2 {
		t.Fatalf("red mana = %d, want 2 (1 from Mountain + 1 from Caged Sun trigger)", got)
	}
}

// TestCagedSunTriggerIgnoresNonChosenColorLandMana proves the chosen-color
// filter: tapping a land for a color other than the chosen one (green when red
// was chosen) does not fire the trigger, so no additional mana is added.
func TestCagedSunTriggerIgnoresNonChosenColorLandMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cagedSunSource(g, game.Player1, mana.R)
	forest := basicColorLand(g, game.Player1, "Forest", types.Forest, mana.G)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(forest.ObjectID, 0, nil, 0)) {
		t.Fatal("activating Forest mana ability = false, want true")
	}
	engine.putTriggeredAbilitiesOnStack(g)

	if got := g.Players[game.Player1].ManaPool.Amount(mana.G); got != 1 {
		t.Fatalf("green mana = %d, want 1 (Forest only; trigger must not fire)", got)
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.R); got != 0 {
		t.Fatalf("red mana = %d, want 0 (chosen color is red but no chosen-color mana was produced)", got)
	}
}

// TestCagedSunAnthemAffectsMulticoloredButNotColorless proves the "+1/+1 to
// creatures you control of the chosen color" static: a multicolored creature
// that includes the chosen color is buffed, while a colorless creature is not.
func TestCagedSunAnthemAffectsMulticoloredButNotColorless(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	anchor := cagedSunSource(g, game.Player1, mana.R)
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:             1,
		Controller:     game.Player1,
		SourceObjectID: anchor.ObjectID,
		Layer:          game.LayerPowerToughnessModify,
		Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			ColorChoice:   game.ColorChoiceSourceEntry,
		}),
		PowerDelta:     1,
		ToughnessDelta: 1,
	})

	multicolored := addColoredCreaturePermanent(g, game.Player1, 2, color.Red, color.Green)
	colorless := addColoredCreaturePermanent(g, game.Player1, 2)

	if got := effectivePower(g, multicolored); got != 3 {
		t.Fatalf("multicolored chosen-color creature power = %d, want 3 (buffed)", got)
	}
	if got := effectivePower(g, colorless); got != 2 {
		t.Fatalf("colorless creature power = %d, want 2 (unbuffed)", got)
	}
}
