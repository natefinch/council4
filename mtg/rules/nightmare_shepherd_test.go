package rules

import (
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const nightmareShepherdLink = game.LinkedKey("event-card-exile-copy")

func nightmareShepherdContent() game.AbilityContent {
	const result = game.ResultKey("if-you-do")
	return game.Mode{Sequence: []game.Instruction{
		{
			Optional: true,
			Primitive: game.MoveCard{
				Card:                            game.CardReference{Kind: game.CardReferenceEvent},
				FromZone:                        zone.Graveyard,
				Destination:                     zone.Exile,
				PublishLinked:                   nightmareShepherdLink,
				ReplacePublishedLinked:          true,
				IncludeEventPermanentComponents: true,
			},
			PublishResult: result,
		},
		{
			ResultGate: opt.Val(game.InstructionResultGate{Key: result, Succeeded: game.TriTrue}),
			Primitive: game.CreateToken{
				Amount: game.Fixed(1),
				Source: game.TokenCopyOf(game.TokenCopySpec{
					Source:       game.TokenCopySourceObject,
					Object:       game.LinkedObjectReference(string(nightmareShepherdLink)),
					SetPower:     opt.Val(game.PT{Value: 1}),
					SetToughness: opt.Val(game.PT{Value: 1}),
					AddSubtypes:  []types.Sub{types.Nightmare},
				}),
			},
		},
	}}.Ability()
}

func nightmareShepherdDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Nightmare Shepherd",
		Types:    []types.Card{types.Enchantment, types.Creature},
		Subtypes: []types.Sub{types.Demon},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:       game.EventPermanentDied,
					Controller:  game.TriggerControllerYou,
					ExcludeSelf: true,
					SubjectSelection: game.Selection{
						RequiredTypes: []types.Card{types.Creature},
						NonToken:      true,
					},
				},
			},
			Content: nightmareShepherdContent(),
		}},
	}}
}

func shepherdVictimDef(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		ManaCost:  opt.Val(cost.Mana{cost.G, cost.G}),
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Bear},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 5}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
	}}
}

func resolveShepherdTriggers(
	t *testing.T,
	g *game.Game,
	engine *Engine,
	agents [game.NumPlayers]PlayerAgent,
) {
	t.Helper()
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("Nightmare Shepherd trigger was not put on the stack")
	}
	for g.Stack.Size() > 0 {
		engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	}
}

func shepherdTokens(g *game.Game, name string) []*game.Permanent {
	var result []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.TokenDef != nil && permanent.TokenDef.Name == name {
			result = append(result, permanent)
		}
	}
	return result
}

func TestNightmareShepherdCopiesExactDiedObjectLKI(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	shepherd := addCombatPermanent(g, game.Player1, nightmareShepherdDef())
	victim := addCombatPermanent(g, game.Player1, shepherdVictimDef("Relic Bear"))
	cardID := victim.CardInstanceID

	if !movePermanentToZone(g, victim, zone.Graveyard) {
		t.Fatal("victim did not die")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Nightmare Shepherd did not trigger")
	}
	shepherd.Controller = game.Player2
	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("exact died card did not reach exile")
	}
	tokens := shepherdTokens(g, "Relic Bear")
	if len(tokens) != 1 {
		t.Fatalf("tokens = %d, want 1", len(tokens))
	}
	token := tokens[0]
	if token.Controller != game.Player1 {
		t.Fatalf("token controller = %v, want trigger controller Player1", token.Controller)
	}
	def := token.TokenDef
	if !def.Power.Exists || def.Power.Val.Value != 1 ||
		!def.Toughness.Exists || def.Toughness.Val.Value != 1 {
		t.Fatalf("token P/T = %+v/%+v, want 1/1", def.Power, def.Toughness)
	}
	if !slices.Contains(def.Types, types.Artifact) ||
		!slices.Contains(def.Types, types.Creature) ||
		!slices.Contains(def.Subtypes, types.Bear) ||
		!slices.Contains(def.Subtypes, types.Nightmare) ||
		!slices.Contains(def.Colors, color.Green) ||
		!reflect.DeepEqual(def.ManaCost, shepherdVictimDef("ignored").ManaCost) ||
		!def.HasKeyword(game.Flying) {
		t.Fatalf("token did not preserve copied characteristics: %#v", def.CardFace)
	}
}

func TestNightmareShepherdRequiresExactCardToReachExile(t *testing.T) {
	tests := []struct {
		name   string
		before func(*game.Game, game.ObjectID)
		agents [game.NumPlayers]PlayerAgent
	}{
		{
			name: "declined",
			agents: [game.NumPlayers]PlayerAgent{
				game.Player1: declineChoiceAgent{},
			},
		},
		{
			name: "left graveyard",
			before: func(g *game.Game, cardID game.ObjectID) {
				moveCardBetweenZones(g, game.Player1, cardID, zone.Graveyard, zone.Hand)
			},
		},
		{
			name: "commander redirect",
			before: func(g *game.Game, cardID game.ObjectID) {
				g.CommanderIDs[cardID] = true
			},
		},
		{
			name: "zone change replacement",
			before: func(g *game.Game, _ game.ObjectID) {
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					ID:            g.IDGen.Next(),
					Description:   "put graveyard cards into hand instead of exile",
					MatchEvent:    game.EventZoneChanged,
					MatchFromZone: true,
					FromZone:      zone.Graveyard,
					MatchToZone:   true,
					ToZone:        zone.Exile,
					ReplaceToZone: zone.Hand,
				})
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			addCombatPermanent(g, game.Player1, nightmareShepherdDef())
			victim := addCombatPermanent(g, game.Player1, shepherdVictimDef("No Copy"))
			cardID := victim.CardInstanceID
			if !movePermanentToZone(g, victim, zone.Graveyard) {
				t.Fatal("victim did not die")
			}
			if !engine.putTriggeredAbilitiesOnStackWithChoices(g, test.agents, &TurnLog{}) {
				t.Fatal("Nightmare Shepherd did not trigger")
			}
			if test.before != nil {
				test.before(g, cardID)
			}
			engine.resolveTopOfStackWithChoices(g, test.agents, &TurnLog{})
			if got := len(shepherdTokens(g, "No Copy")); got != 0 {
				t.Fatalf("tokens = %d, want 0", got)
			}
		})
	}
}

func TestNightmareShepherdSimultaneousDeathsSourceLeavesAndTokenDoubler(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	shepherd := addCombatPermanent(g, game.Player1, nightmareShepherdDef())
	first := addCombatPermanent(g, game.Player1, shepherdVictimDef("First Corpse"))
	second := addCombatPermanent(g, game.Player1, shepherdVictimDef("Second Corpse"))
	addReplacementPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Anointed Procession",
		Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{
			game.TokenCreationReplacement("double tokens", 2, game.TriggerControllerYou),
		},
	}})

	if !movePermanentsToZoneSimultaneously(g, []*game.Permanent{shepherd, first, second}, zone.Graveyard) {
		t.Fatal("simultaneous deaths failed")
	}
	resolveShepherdTriggers(t, g, engine, [game.NumPlayers]PlayerAgent{})

	if got := len(shepherdTokens(g, "First Corpse")); got != 2 {
		t.Fatalf("First Corpse tokens = %d, want 2", got)
	}
	if got := len(shepherdTokens(g, "Second Corpse")); got != 2 {
		t.Fatalf("Second Corpse tokens = %d, want 2", got)
	}
	if got := len(shepherdTokens(g, "Nightmare Shepherd")); got != 0 {
		t.Fatalf("self-copy tokens = %d, want 0", got)
	}
}

func TestNightmareShepherdCopiesMergedCreatureCopiableValues(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, nightmareShepherdDef())
	victim := addCombatPermanent(g, game.Player1, shepherdVictimDef("Merged Top"))
	lowerDef := &game.CardDef{CardFace: game.CardFace{
		Name:  "Merged Lower",
		Types: []types.Card{types.Creature},
		ActivatedAbilities: []game.ActivatedAbility{{
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	}}
	lowerID := addCardToHand(g, game.Player1, lowerDef)
	g.Players[game.Player1].Hand.Remove(lowerID)
	victim.MergedCards = []game.MergedCard{{CardInstanceID: lowerID, Owner: game.Player1}}

	if !movePermanentToZone(g, victim, zone.Graveyard) {
		t.Fatal("merged creature did not die")
	}
	resolveShepherdTriggers(t, g, engine, [game.NumPlayers]PlayerAgent{})

	tokens := shepherdTokens(g, "Merged Top")
	if len(tokens) != 1 {
		t.Fatalf("tokens = %d, want 1", len(tokens))
	}
	if got := len(tokens[0].TokenDef.ActivatedAbilities); got != 1 {
		t.Fatalf("copied activated abilities = %d, want merged lower ability", got)
	}
	if !g.Players[game.Player1].Exile.Contains(lowerID) {
		t.Fatal("merged lower component was not exiled with the tracked permanent")
	}
}

func TestNightmareShepherdExcludesTokenDeaths(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, nightmareShepherdDef())
	token, ok := createTokenPermanent(g, game.Player1, shepherdVictimDef("Token Victim"))
	if !ok || !movePermanentToZone(g, token, zone.Graveyard) {
		t.Fatal("token did not die")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Nightmare Shepherd triggered for a token death")
	}
}
