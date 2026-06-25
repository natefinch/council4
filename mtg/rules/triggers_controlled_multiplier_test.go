package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func controlledTriggerDoublerDef(filter game.Selection) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Controlled Trigger Doubler",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:              game.RuleEffectAdditionalTriggerForControlledPermanent,
				AffectedSelection: filter,
			}},
		}},
	}}
}

func selfEntersTypedTriggerSourceDef(name string, supertypes []types.Super, cardTypes []types.Card, subtypes []types.Sub) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       name,
		Supertypes: supertypes,
		Types:      cardTypes,
		Subtypes:   subtypes,
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

func emitSelfEnter(g *game.Game, source *game.Permanent) {
	emitEvent(g, game.Event{
		Kind:        game.EventZoneChanged,
		Controller:  source.Controller,
		Player:      source.Controller,
		PermanentID: source.ObjectID,
		CardID:      source.CardInstanceID,
		FromZone:    zone.Stack,
		ToZone:      zone.Battlefield,
	})
}

func TestControlledPermanentRuleEffectMultipliesMatchingTrigger(t *testing.T) {
	legendaryCreature := game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		Supertypes:    []types.Super{types.Legendary},
	}
	allySubtype := game.Selection{SubtypesAny: []types.Sub{types.Sub("Ally")}}

	for name, tc := range map[string]struct {
		filter     game.Selection
		supertypes []types.Super
		cardTypes  []types.Card
		subtypes   []types.Sub
		wantStack  int
	}{
		"legendary-creature filter doubles legendary creature trigger": {
			filter:     legendaryCreature,
			supertypes: []types.Super{types.Legendary},
			cardTypes:  []types.Card{types.Creature},
			wantStack:  2,
		},
		"legendary-creature filter ignores nonlegendary creature": {
			filter:    legendaryCreature,
			cardTypes: []types.Card{types.Creature},
			wantStack: 1,
		},
		"legendary-creature filter ignores legendary noncreature": {
			filter:     legendaryCreature,
			supertypes: []types.Super{types.Legendary},
			cardTypes:  []types.Card{types.Artifact},
			wantStack:  1,
		},
		"subtype filter doubles matching subtype trigger": {
			filter:    allySubtype,
			cardTypes: []types.Card{types.Creature},
			subtypes:  []types.Sub{types.Sub("Ally")},
			wantStack: 2,
		},
		"subtype filter ignores other subtype": {
			filter:    allySubtype,
			cardTypes: []types.Card{types.Creature},
			subtypes:  []types.Sub{types.Sub("Soldier")},
			wantStack: 1,
		},
	} {
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			addCombatPermanent(g, game.Player1, controlledTriggerDoublerDef(tc.filter))
			source := addCombatPermanent(g, game.Player1, selfEntersTypedTriggerSourceDef("Trigger Source", tc.supertypes, tc.cardTypes, tc.subtypes))

			emitSelfEnter(g, source)
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("controlled trigger was not put on the stack")
			}
			if got := g.Stack.Size(); got != tc.wantStack {
				t.Fatalf("stack size = %d, want %d", got, tc.wantStack)
			}
		})
	}
}

func TestControlledPermanentRuleEffectIgnoresOpponentControlledTriggerSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, controlledTriggerDoublerDef(game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		Supertypes:    []types.Super{types.Legendary},
	}))
	source := addCombatPermanent(g, game.Player2, selfEntersTypedTriggerSourceDef(
		"Opponent Source", []types.Super{types.Legendary}, []types.Card{types.Creature}, nil))

	emitSelfEnter(g, source)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("controlled trigger was not put on the stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want 1 (opponent's trigger is not doubled)", got)
	}
}

func TestControlledPermanentRuleEffectDoublesOwnTrigger(t *testing.T) {
	// "a ... you control" includes the doubler itself, unlike the "another"
	// chosen-type and entering-permanent families.
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	doubler := &game.CardDef{CardFace: game.CardFace{
		Name:       "Self-Doubling Legend",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind: game.RuleEffectAdditionalTriggerForControlledPermanent,
				AffectedSelection: game.Selection{
					RequiredTypes: []types.Card{types.Creature},
					Supertypes:    []types.Super{types.Legendary},
				},
			}},
		}},
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
	source := addCombatPermanent(g, game.Player1, doubler)

	emitSelfEnter(g, source)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("controlled trigger was not put on the stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want 2 (the doubler doubles its own trigger)", got)
	}
}
