package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func enteringTriggerDoublerDef(filter []types.Card) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Entering Trigger Doubler",
		Types: []types.Card{types.Artifact},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectAdditionalTriggerForEnteringPermanent,
				PermanentTypes: filter,
			}},
		}},
	}}
}

func selfEntersTriggerSourceDef(cardTypes []types.Card) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Entering Trigger Source",
		Types: cardTypes,
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
				Event:       game.EventZoneChanged,
				Source:      game.TriggerSourceSelf,
				MatchToZone: true,
				ToZone:      zone.Battlefield,
			}},
			Content: game.Mode{Sequence: []game.Instruction{
				{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}},
			}}.Ability(),
		}},
	}}
}

func TestEnteringPermanentRuleEffectMultipliesMatchingTrigger(t *testing.T) {
	for name, tc := range map[string]struct {
		filter     []types.Card
		enterTypes []types.Card
		wantStack  int
	}{
		"artifact-or-creature filter doubles entering creature": {
			filter:     []types.Card{types.Artifact, types.Creature},
			enterTypes: []types.Card{types.Creature},
			wantStack:  2,
		},
		"artifact-or-creature filter ignores entering land": {
			filter:     []types.Card{types.Artifact, types.Creature},
			enterTypes: []types.Card{types.Land},
			wantStack:  1,
		},
		"empty filter doubles any entering permanent": {
			filter:     nil,
			enterTypes: []types.Card{types.Land},
			wantStack:  2,
		},
	} {
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			addCombatPermanent(g, game.Player1, enteringTriggerDoublerDef(tc.filter))
			source := addCombatPermanent(g, game.Player1, selfEntersTriggerSourceDef(tc.enterTypes))

			emitEvent(g, game.Event{
				Kind:        game.EventZoneChanged,
				Controller:  game.Player1,
				Player:      game.Player1,
				PermanentID: source.ObjectID,
				CardID:      source.CardInstanceID,
				FromZone:    zone.Stack,
				ToZone:      zone.Battlefield,
			})
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("entering trigger was not put on the stack")
			}
			if got := g.Stack.Size(); got != tc.wantStack {
				t.Fatalf("stack size = %d, want %d", got, tc.wantStack)
			}
		})
	}
}

func TestEnteringPermanentRuleEffectIgnoresOpponentControlledTriggerSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, enteringTriggerDoublerDef(nil))
	source := addCombatPermanent(g, game.Player2, selfEntersTriggerSourceDef([]types.Card{types.Creature}))

	emitEvent(g, game.Event{
		Kind:        game.EventZoneChanged,
		Controller:  game.Player2,
		Player:      game.Player2,
		PermanentID: source.ObjectID,
		CardID:      source.CardInstanceID,
		FromZone:    zone.Stack,
		ToZone:      zone.Battlefield,
	})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("entering trigger was not put on the stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want 1 (opponent's trigger is not doubled)", got)
	}
}
