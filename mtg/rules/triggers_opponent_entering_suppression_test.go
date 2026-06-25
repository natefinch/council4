package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func opponentEnteringSuppressorDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Opponent Entering Suppressor",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind: game.RuleEffectSuppressOpponentEnteringTriggers,
			}},
		}},
	}}
}

// TestSuppressOpponentEnteringTriggersDropsOpponentTrigger verifies that an
// opponent's entering-caused triggered ability is suppressed while a controller
// controls the Elesh Norn-style suppressor, but the suppressor controller's own
// entering trigger still fires.
func TestSuppressOpponentEnteringTriggersDropsOpponentTrigger(t *testing.T) {
	for name, tc := range map[string]struct {
		triggerController game.PlayerID
		wantStack         int
	}{
		"opponent entering trigger suppressed": {triggerController: game.Player2, wantStack: 0},
		"controller entering trigger fires":    {triggerController: game.Player1, wantStack: 1},
	} {
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			addCombatPermanent(g, game.Player1, opponentEnteringSuppressorDef())
			source := addCombatPermanent(g, tc.triggerController, selfEntersTriggerSourceDef([]types.Card{types.Creature}))

			emitEvent(g, game.Event{
				Kind:        game.EventZoneChanged,
				Controller:  tc.triggerController,
				Player:      tc.triggerController,
				PermanentID: source.ObjectID,
				CardID:      source.CardInstanceID,
				FromZone:    zone.Stack,
				ToZone:      zone.Battlefield,
			})
			engine.putTriggeredAbilitiesOnStack(g)
			if got := g.Stack.Size(); got != tc.wantStack {
				t.Fatalf("stack size = %d, want %d", got, tc.wantStack)
			}
		})
	}
}

// TestSuppressOpponentEnteringTriggersLeavesNonEnteringTriggers verifies the
// suppressor only drops entering-caused triggers, not other triggered abilities
// of an opponent's permanents.
func TestSuppressOpponentEnteringTriggersLeavesNonEnteringTriggers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, opponentEnteringSuppressorDef())
	source := addCombatPermanent(g, game.Player2, attacksTriggerSourceDef())

	emitEvent(g, game.Event{
		Kind:           game.EventAttackerDeclared,
		Controller:     game.Player2,
		Player:         game.Player2,
		SourceObjectID: source.ObjectID,
		PermanentID:    source.ObjectID,
		CardID:         source.CardInstanceID,
	})
	engine.putTriggeredAbilitiesOnStack(g)
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want 1 (non-entering trigger is not suppressed)", got)
	}
}

func attacksTriggerSourceDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Attacks Trigger Source",
		Types: []types.Card{types.Creature},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
				Event:  game.EventAttackerDeclared,
				Source: game.TriggerSourceSelf,
			}},
			Content: game.Mode{Sequence: []game.Instruction{
				{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}},
			}}.Ability(),
		}},
	}}
}
