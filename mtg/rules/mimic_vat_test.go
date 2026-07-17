package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const mimicVatImprintLink = game.LinkedKey("imprint")

func mimicVatImprintContent() game.AbilityContent {
	return game.Mode{Sequence: []game.Instruction{{
		Optional: true,
		Primitive: game.ReplaceLinkedExiledCard{
			Card:     game.CardReference{Kind: game.CardReferenceEvent},
			FromZone: zone.Graveyard,
			LinkID:   mimicVatImprintLink,
		},
	}}}.Ability()
}

func mimicVatTokenContent() game.AbilityContent {
	const tokenLink = game.LinkedKey("imprint-created-token")
	const resultKey = game.ResultKey("imprint-token-created")
	return game.Mode{Sequence: []game.Instruction{
		{
			Primitive: game.CreateToken{
				Amount: game.Fixed(1),
				Source: game.TokenCopyOf(game.TokenCopySpec{
					Source:      game.TokenCopySourceLinkedExiledCard,
					LinkID:      mimicVatImprintLink,
					AddKeywords: []game.Keyword{game.Haste},
				}),
				PublishLinked: tokenLink,
			},
			PublishResult: resultKey,
		},
		{
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       resultKey,
				Succeeded: game.TriTrue,
			}),
			Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
				Timing:              game.DelayedAtBeginningOfNextEndStep,
				CapturedObjectGroup: opt.Val(game.LinkedObjectReference(string(tokenLink))),
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.Exile{Group: game.CapturedObjectsGroup()},
				}}}.Ability(),
			}},
		},
	}}.Ability()
}

func mimicVatDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Mimic Vat",
		Types: []types.Card{types.Artifact},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event: game.EventPermanentDied,
					SubjectSelection: game.Selection{
						RequiredTypes: []types.Card{types.Creature},
						NonToken:      true,
					},
				},
			},
			Content: mimicVatImprintContent(),
		}},
		ActivatedAbilities: []game.ActivatedAbility{{
			Content: mimicVatTokenContent(),
		}},
	}}
}

func mimicCreature(name string, power int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: power}),
	}}
}

func mimicDiedCard(g *game.Game, owner game.PlayerID, def *game.CardDef) (game.ObjectID, game.Event) {
	cardID := addCardToGraveyard(g, owner, def)
	card, _ := g.GetCardInstance(cardID)
	card.ZoneVersion = 1
	return cardID, game.Event{
		Kind:            game.EventPermanentDied,
		CardID:          cardID,
		CardZoneVersion: card.ZoneVersion,
		FromZone:        zone.Battlefield,
		ToZone:          zone.Graveyard,
	}
}

func resolveMimicImprint(
	g *game.Game,
	engine *Engine,
	vat *game.Permanent,
	event game.Event,
	agents [game.NumPlayers]PlayerAgent,
) {
	engine.resolveAbilityContentWithChoices(g, &game.StackObject{
		ID:              g.IDGen.Next(),
		Kind:            game.StackTriggeredAbility,
		SourceID:        vat.ObjectID,
		SourceCardID:    vat.CardInstanceID,
		Controller:      vat.Controller,
		HasTriggerEvent: true,
		TriggerEvent:    event,
	}, mimicVatImprintContent(), agents, &TurnLog{})
}

func resolveMimicActivation(g *game.Game, engine *Engine, vat *game.Permanent) {
	engine.resolveAbilityContentWithChoices(g, &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackActivatedAbility,
		SourceID:     vat.ObjectID,
		SourceCardID: vat.CardInstanceID,
		Controller:   vat.Controller,
	}, mimicVatTokenContent(), [game.NumPlayers]PlayerAgent{}, &TurnLog{})
}

func mimicTokens(g *game.Game, name string) []*game.Permanent {
	var tokens []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.TokenDef != nil && permanent.TokenDef.Name == name {
			tokens = append(tokens, permanent)
		}
	}
	return tokens
}

func TestMimicVatImprintReplacesOnlyAfterSuccessfulExile(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	vat := addCombatPermanent(g, game.Player1, mimicVatDef())
	oldID, oldEvent := mimicDiedCard(g, game.Player2, mimicCreature("Old Corpse", 2))
	resolveMimicImprint(g, engine, vat, oldEvent, [game.NumPlayers]PlayerAgent{})

	declinedID, declinedEvent := mimicDiedCard(g, game.Player2, mimicCreature("Declined Corpse", 3))
	resolveMimicImprint(g, engine, vat, declinedEvent, [game.NumPlayers]PlayerAgent{
		game.Player1: declineChoiceAgent{},
	})
	if !g.Players[game.Player2].Exile.Contains(oldID) ||
		!g.Players[game.Player2].Graveyard.Contains(declinedID) {
		t.Fatal("declining a new imprint changed the old imprint")
	}

	commanderID, commanderEvent := mimicDiedCard(g, game.Player2, mimicCreature("Commander Corpse", 4))
	g.CommanderIDs[commanderID] = true
	resolveMimicImprint(g, engine, vat, commanderEvent, [game.NumPlayers]PlayerAgent{})
	if !g.Players[game.Player2].CommandZone.Contains(commanderID) ||
		!g.Players[game.Player2].Exile.Contains(oldID) {
		t.Fatal("commander replacement did not preserve the prior imprint")
	}

	newID, newEvent := mimicDiedCard(g, game.Player3, mimicCreature("New Corpse", 5))
	resolveMimicImprint(g, engine, vat, newEvent, [game.NumPlayers]PlayerAgent{})
	if !g.Players[game.Player3].Exile.Contains(newID) ||
		!g.Players[game.Player2].Graveyard.Contains(oldID) {
		t.Fatal("successful replacement did not establish one current imprint")
	}
}

func TestMimicVatLinksFollowSourceObjectAndCurrentController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	vatA := addCombatPermanent(g, game.Player1, mimicVatDef())
	vatB := addCombatPermanent(g, game.Player2, mimicVatDef())
	_, eventA := mimicDiedCard(g, game.Player3, mimicCreature("Alpha", 2))
	_, eventB := mimicDiedCard(g, game.Player4, mimicCreature("Beta", 3))
	resolveMimicImprint(g, engine, vatA, eventA, [game.NumPlayers]PlayerAgent{})
	resolveMimicImprint(g, engine, vatB, eventB, [game.NumPlayers]PlayerAgent{})

	vatA.Controller = game.Player3
	resolveMimicActivation(g, engine, vatA)
	alpha := mimicTokens(g, "Alpha")
	if len(alpha) != 1 || alpha[0].Controller != game.Player3 || !hasKeyword(g, alpha[0], game.Haste) {
		t.Fatalf("controlled Vat created Alpha tokens = %#v, want one hasty Player3 token", alpha)
	}
	if len(mimicTokens(g, "Beta")) != 0 {
		t.Fatal("one Vat used another Vat's imprint")
	}

	movePermanentToZone(g, vatA, zone.Graveyard)
	card, _ := g.GetCardInstance(vatA.CardInstanceID)
	reentered, ok := createCardPermanent(g, card, game.Player1, zone.Graveyard)
	if !ok {
		t.Fatal("Mimic Vat did not reenter")
	}
	resolveMimicActivation(g, engine, reentered)
	if len(mimicTokens(g, "Alpha")) != 1 {
		t.Fatal("reentered Mimic Vat inherited its prior object's imprint")
	}
}

func TestMimicVatLinkedCardMustStillBeSameExileObject(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	vat := addCombatPermanent(g, game.Player1, mimicVatDef())
	cardID, event := mimicDiedCard(g, game.Player2, mimicCreature("Traveler", 2))
	resolveMimicImprint(g, engine, vat, event, [game.NumPlayers]PlayerAgent{})
	if !moveCardBetweenZonesWithPlacement(g, game.Player2, cardID, zone.Exile, zone.Graveyard, false) ||
		!moveCardBetweenZonesWithPlacement(g, game.Player2, cardID, zone.Graveyard, zone.Exile, false) {
		t.Fatal("failed to move imprinted card out of and back into exile")
	}

	resolveMimicActivation(g, engine, vat)
	if len(mimicTokens(g, "Traveler")) != 0 || len(g.DelayedTriggers) != 0 {
		t.Fatal("stale imprint followed a card that left and reentered exile")
	}
}

func TestMimicVatCopiesCardValuesAndExilesAllDoubledTokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	vat := addCombatPermanent(g, game.Player1, mimicVatDef())
	addReplacementPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Anointed Procession",
		Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{
			game.TokenCreationReplacement("double tokens", 2, game.TriggerControllerYou),
		},
	}})
	_, event := mimicDiedCard(g, game.Player2, mimicCreature("Printed Corpse", 2))
	event.TokenDef = mimicCreature("Last-Known Copy", 9)
	resolveMimicImprint(g, engine, vat, event, [game.NumPlayers]PlayerAgent{})

	resolveMimicActivation(g, engine, vat)
	tokens := mimicTokens(g, "Printed Corpse")
	if len(tokens) != 2 || len(mimicTokens(g, "Last-Known Copy")) != 0 {
		t.Fatalf("copy tokens = %d printed/%d LKI, want 2/0", len(tokens), len(mimicTokens(g, "Last-Known Copy")))
	}
	for _, token := range tokens {
		if !hasKeyword(g, token, game.Haste) {
			t.Fatal("doubled Mimic Vat token did not gain haste")
		}
	}
	resolveMimicActivation(g, engine, vat)
	if got := len(mimicTokens(g, "Printed Corpse")); got != 4 {
		t.Fatalf("tokens after repeated activation = %d, want 4", got)
	}
	if len(g.DelayedTriggers) != 2 {
		t.Fatalf("delayed triggers = %d, want 2", len(g.DelayedTriggers))
	}
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	if len(mimicTokens(g, "Printed Corpse")) != 0 {
		t.Fatal("next end step did not exile every doubled token")
	}
}

func TestMimicVatTokenDeathsAndActivationWithoutImprintDoNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	vat := addCombatPermanent(g, game.Player1, mimicVatDef())
	pattern := mimicVatDef().TriggeredAbilities[0].Trigger.Pattern
	if triggerMatchesEvent(g, vat, &pattern, game.Event{
		Kind:        game.EventPermanentDied,
		PermanentID: g.IDGen.Next(),
		TokenName:   "Goblin",
		TokenDef:    mimicCreature("Goblin", 1),
		FromZone:    zone.Battlefield,
		ToZone:      zone.Graveyard,
	}) {
		t.Fatal("Mimic Vat triggered for a token death")
	}

	resolveMimicActivation(g, engine, vat)
	tokenCount := 0
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			tokenCount++
		}
	}
	if tokenCount != 0 || len(g.DelayedTriggers) != 0 {
		t.Fatal("activation without an imprint created a token or delayed trigger")
	}
}
