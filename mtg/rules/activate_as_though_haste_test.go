package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
)

// activateAsThoughHasteStaticPermanent gives playerID a battlefield permanent
// whose static ability lets them activate abilities of creatures they control as
// though those creatures had haste (Thousand-Year Elixir).
func activateAsThoughHasteStaticPermanent(g *game.Game, playerID game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, playerID, &game.CardDef{CardFace: game.CardFace{
		Name:            "Test Haste Activation Permission",
		StaticAbilities: []game.StaticAbility{game.ActivateAbilitiesAsThoughHasteStaticBody},
	}})
}

// summoningSickTapAbilityCreature adds a summoning-sick creature controlled by
// playerID whose only activated ability costs {T}.
func summoningSickTapAbilityCreature(g *game.Game, playerID game.PlayerID) *game.Permanent {
	creature := addCombatPermanent(g, playerID, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: cost.Tap,
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
		}.Ability(),
	}))
	creature.SummoningSick = true
	return creature
}

// TestActivateAsThoughHasteAllowsTapAbilityWhileSummoningSick proves the static
// lets a summoning-sick creature the controller owns pay a {T} activation cost,
// and that the permission ends when the static's source leaves the battlefield.
func TestActivateAsThoughHasteAllowsTapAbilityWhileSummoningSick(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	creature := summoningSickTapAbilityCreature(g, game.Player1)
	act := action.ActivateAbility(creature.ObjectID, 0, nil, 0)

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("tap ability was legal for a summoning-sick creature without the static")
	}

	source := activateAsThoughHasteStaticPermanent(g, game.Player1)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("tap ability was not legal for a summoning-sick creature under the static")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(tap ability) = false, want true")
	}
	if !creature.Tapped {
		t.Fatal("summoning-sick creature was not tapped to pay its activation cost")
	}

	if _, ok := removePermanentFromBattlefield(g, source.ObjectID); !ok {
		t.Fatal("failed to remove the static source from the battlefield")
	}
	creature.Tapped = false
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("tap ability stayed legal after the static's source left the battlefield")
	}
}

// TestActivateAsThoughHasteDoesNotGrantAttack proves the static is an activation
// permission only: a summoning-sick creature still cannot attack while it applies
// unless the creature has real haste.
func TestActivateAsThoughHasteDoesNotGrantAttack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := summoningSickTapAbilityCreature(g, game.Player1)
	activateAsThoughHasteStaticPermanent(g, game.Player1)

	if canAttackWith(g, creature, game.Player1) {
		t.Fatal("summoning-sick creature could attack under an activation-only haste permission")
	}
	if !hasKeyword(g, creature, game.Haste) {
		creature.SummoningSick = false
		if !canAttackWith(g, creature, game.Player1) {
			t.Fatal("creature could not attack after summoning sickness ended")
		}
	}
}

// TestActivateAsThoughHasteIsControllerScoped proves an opponent's summoning-sick
// creature gains no benefit from the controller's static.
func TestActivateAsThoughHasteIsControllerScoped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.ActivePlayer = game.Player2

	activateAsThoughHasteStaticPermanent(g, game.Player1)
	creature := summoningSickTapAbilityCreature(g, game.Player2)
	act := action.ActivateAbility(creature.ObjectID, 0, nil, 0)

	if containsAction(engine.legalActions(g, game.Player2), act) {
		t.Fatal("opponent's summoning-sick creature gained tap activation from the controller's static")
	}
}
