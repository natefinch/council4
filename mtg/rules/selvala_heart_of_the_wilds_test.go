package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func selvalaEnterCreature(g *game.Game, controller game.PlayerID, name string, power int) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: power}),
	}})
}

func selvalaEnterTriggerInstruction() *game.Instruction {
	return &game.Instruction{
		Primitive: game.Draw{
			Amount: game.Fixed(1),
			Player: game.ObjectControllerReference(game.EventPermanentReference()),
		},
		Condition: opt.Val(game.EffectCondition{
			Condition: opt.Val(game.Condition{
				EventPermanentPowerGreaterThanEachOtherCreature: true,
			}),
		}),
		Optional:      true,
		OptionalActor: opt.Val(game.ObjectControllerReference(game.EventPermanentReference())),
	}
}

// TestSelvalaEnterTriggerDrawsForStrictlyGreatestPower exercises Selvala, Heart
// of the Wilds' enter trigger "its controller may draw a card if its power is
// greater than each other creature's power." The entering creature's controller
// draws only when that creature's power is strictly greater than every other
// creature on the battlefield; a tie or a larger creature leaves the gate closed.
func TestSelvalaEnterTriggerDrawsForStrictlyGreatestPower(t *testing.T) {
	resolve := func(t *testing.T, setup func(g *game.Game) *game.Permanent) (before, after int, ownerBefore, ownerAfter int) {
		t.Helper()
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		selvala := selvalaEnterCreature(g, game.Player1, "Selvala, Heart of the Wilds", 2)
		entering := setup(g)
		addCardToLibraryNamed(g, entering.Controller, "Draw Fodder")
		addCardToLibraryNamed(g, game.Player1, "Owner Fodder")
		obj := &game.StackObject{
			Kind:            game.StackTriggeredAbility,
			SourceID:        selvala.ObjectID,
			SourceCardID:    selvala.CardInstanceID,
			Controller:      game.Player1,
			HasTriggerEvent: true,
			TriggerEvent: game.Event{
				Kind:        game.EventPermanentEnteredBattlefield,
				PermanentID: entering.ObjectID,
			},
		}
		agents := [game.NumPlayers]PlayerAgent{
			game.Player1: defaultChoiceAgent{},
			game.Player2: defaultChoiceAgent{},
		}
		before = g.Players[entering.Controller].Hand.Size()
		ownerBefore = g.Players[game.Player1].Hand.Size()
		newEffectResolver(engine, g, obj, agents, &TurnLog{}).
			resolveInstruction(selvalaEnterTriggerInstruction())
		after = g.Players[entering.Controller].Hand.Size()
		ownerAfter = g.Players[game.Player1].Hand.Size()
		return before, after, ownerBefore, ownerAfter
	}

	t.Run("strictly greatest draws for entering controller", func(t *testing.T) {
		before, after, ownerBefore, ownerAfter := resolve(t, func(g *game.Game) *game.Permanent {
			selvalaEnterCreature(g, game.Player1, "Bystander", 4)
			selvalaEnterCreature(g, game.Player2, "Ally", 3)
			return selvalaEnterCreature(g, game.Player2, "Behemoth", 5)
		})
		if after != before+1 {
			t.Fatalf("entering controller hand = %d, want %d (its controller may draw)", after, before+1)
		}
		if ownerAfter != ownerBefore {
			t.Fatalf("Selvala's controller hand = %d, want %d (only the entering controller draws)", ownerAfter, ownerBefore)
		}
	})

	t.Run("tied for greatest does not draw", func(t *testing.T) {
		before, after, _, _ := resolve(t, func(g *game.Game) *game.Permanent {
			selvalaEnterCreature(g, game.Player1, "Rival", 4)
			return selvalaEnterCreature(g, game.Player2, "Contender", 4)
		})
		if after != before {
			t.Fatalf("entering controller hand = %d, want %d (a tie is not strictly greatest)", after, before)
		}
	})

	t.Run("larger creature present does not draw", func(t *testing.T) {
		before, after, _, _ := resolve(t, func(g *game.Game) *game.Permanent {
			selvalaEnterCreature(g, game.Player1, "Titan", 7)
			return selvalaEnterCreature(g, game.Player2, "Newcomer", 5)
		})
		if after != before {
			t.Fatalf("entering controller hand = %d, want %d (a larger creature exists)", after, before)
		}
	})
}

// TestSelvalaEnterTriggerUsesLastKnownPowerWhenCreatureLeft confirms that when
// the entered creature has already left the battlefield by the time Selvala's
// trigger resolves, its controller may still draw using the creature's power as
// it last existed on the battlefield (Selvala ruling 2016-08-23), rather than
// the gate failing closed.
func TestSelvalaEnterTriggerUsesLastKnownPowerWhenCreatureLeft(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	selvala := selvalaEnterCreature(g, game.Player1, "Selvala, Heart of the Wilds", 2)
	selvalaEnterCreature(g, game.Player1, "Bystander", 3)
	entering := selvalaEnterCreature(g, game.Player2, "Departed Behemoth", 6)
	addCardToLibraryNamed(g, game.Player2, "Draw Fodder")

	// The entering creature leaves the battlefield in response to the trigger,
	// but its last-known power (6, strictly greatest) is remembered.
	snapshot := snapshotPermanent(g, entering, zone.Battlefield)
	rememberLastKnown(g, &snapshot)
	if _, ok := removePermanentFromBattlefield(g, entering.ObjectID); !ok {
		t.Fatal("failed to remove entering creature from the battlefield")
	}

	obj := &game.StackObject{
		Kind:            game.StackTriggeredAbility,
		SourceID:        selvala.ObjectID,
		SourceCardID:    selvala.CardInstanceID,
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:        game.EventPermanentEnteredBattlefield,
			PermanentID: entering.ObjectID,
		},
	}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: defaultChoiceAgent{},
		game.Player2: defaultChoiceAgent{},
	}
	before := g.Players[game.Player2].Hand.Size()
	newEffectResolver(engine, g, obj, agents, &TurnLog{}).
		resolveInstruction(selvalaEnterTriggerInstruction())
	if after := g.Players[game.Player2].Hand.Size(); after != before+1 {
		t.Fatalf("entering controller hand = %d, want %d (may draw using last-known power)", after, before+1)
	}
}

func selvalaManaSource() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Selvala, Heart of the Wilds",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 3}),
		ManaAbilities: []game.ManaAbility{{
			AdditionalCosts: cost.Tap,
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.AddMana{
					Amount: game.Dynamic(game.DynamicAmount{
						Kind:       game.DynamicAmountGreatestPowerInGroup,
						Multiplier: 1,
						Group: game.BattlefieldGroup(game.Selection{
							RequiredTypes: []types.Card{types.Creature},
							Controller:    game.ControllerYou,
						}),
					}),
					CombinationColors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
				},
			}}}.Ability(),
		}},
	}}
}

// TestSelvalaManaAbilityAddsGreatestPowerInChosenColors exercises Selvala, Heart
// of the Wilds' mana ability "Add X mana in any combination of colors, where X is
// the greatest power among creatures you control." The amount equals the greatest
// power among the activating player's creatures (ignoring opponents' creatures),
// and the produced mana lands in the colors the controller chooses.
func TestSelvalaManaAbilityAddsGreatestPowerInChosenColors(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, selvalaManaSource())
	selvalaEnterCreature(g, game.Player1, "Mightiest", 6)
	selvalaEnterCreature(g, game.Player1, "Lesser", 3)
	selvalaEnterCreature(g, game.Player2, "Opposing Colossus", 9)

	// Greatest power among Player1's creatures is 6; split it as U=4, G=2 across
	// the WUBRG offer (per-unit color indices: U=1, G=4).
	split := []int{1, 1, 1, 1, 4, 4}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{split}},
	}
	activate := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	if !engine.applyActionWithChoices(g, game.Player1, activate, agents, &TurnLog{}) {
		t.Fatal("Selvala mana ability activation failed")
	}

	pool := g.Players[game.Player1].ManaPool
	if got := pool.Total(); got != 6 {
		t.Fatalf("total mana = %d, want 6 (greatest power among controlled creatures)", got)
	}
	if got := pool.Amount(mana.U); got != 4 {
		t.Fatalf("blue mana = %d, want 4", got)
	}
	if got := pool.Amount(mana.G); got != 2 {
		t.Fatalf("green mana = %d, want 2", got)
	}
}
