package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// playAndCastFromExileWithCroakPermanent gives playerID a battlefield permanent
// whose static ability lets that player play lands and cast spells from among
// cards they own in exile with croak counters on them (Grolnok, the Omnivore).
func playAndCastFromExileWithCroakPermanent(g *game.Game, playerID game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, playerID, &game.CardDef{CardFace: game.CardFace{
		Name: "Test Omnivore",
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{
				{
					Kind:               game.RuleEffectPlayLandsFromZone,
					AffectedPlayer:     game.PlayerYou,
					CastFromZone:       zone.Exile,
					PermanentTypes:     []types.Card{types.Land},
					ExileCounterFilter: opt.Val(counter.Croak),
				},
				{
					Kind:               game.RuleEffectCastSpellsFromZone,
					AffectedPlayer:     game.PlayerYou,
					CastFromZone:       zone.Exile,
					ExileCounterFilter: opt.Val(counter.Croak),
				},
			},
		}},
	}})
}

// TestExileWithCounterMoveCardPlacesCounter verifies a MoveCard whose Counter
// rider names a croak counter places that counter on the moved card once it
// lands in exile (Grolnok's "exile it with a croak counter on it" trigger).
func TestExileWithCounterMoveCardPlacesCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Milled Frog",
		Types: []types.Card{types.Creature},
	}})

	addEffectSpellToStack(g, game.Player1, game.MoveCard{
		Card:        game.CardReference{Kind: game.CardReferenceTarget},
		FromZone:    zone.Graveyard,
		Destination: zone.Exile,
		Counter:     opt.Val(counter.Croak),
	}, []game.Target{currentCardTarget(t, g, cardID)})
	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("card did not move to its owner's exile")
	}
	if !g.HasExileCounter(cardID, counter.Croak) {
		t.Fatal("exiled card is missing the croak counter placed by the move")
	}
}

// TestExileWithoutCounterMoveCardPlacesNoCounter verifies an ordinary exiling
// MoveCard (no Counter rider) leaves no exile counter behind.
func TestExileWithoutCounterMoveCardPlacesNoCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Milled Frog",
		Types: []types.Card{types.Creature},
	}})

	addEffectSpellToStack(g, game.Player1, game.MoveCard{
		Card:        game.CardReference{Kind: game.CardReferenceTarget},
		FromZone:    zone.Graveyard,
		Destination: zone.Exile,
	}, []game.Target{currentCardTarget(t, g, cardID)})
	engine.resolveTopOfStack(g, &TurnLog{})

	if g.HasExileCounter(cardID, counter.Croak) {
		t.Fatal("plain exile placed a croak counter it should not have")
	}
}

// TestPlayLandFromExileRequiresCroakCounter verifies the play-from-exile
// permission only reaches exiled lands that carry a croak counter, and only for
// the player who controls the permission.
func TestPlayLandFromExileRequiresCroakCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	playAndCastFromExileWithCroakPermanent(g, game.Player1)

	withCounter := addCardToExile(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land}}})
	withoutCounter := addCardToExile(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Island", Types: []types.Card{types.Land}}})
	g.AddExileCounter(withCounter, counter.Croak, 1)

	if !canPlayLandFromZoneByRuleEffect(g, game.Player1, withCounter, zone.Exile) {
		t.Fatal("croak-countered exiled land is not playable despite the permission")
	}
	if canPlayLandFromZoneByRuleEffect(g, game.Player1, withoutCounter, zone.Exile) {
		t.Fatal("exiled land without a croak counter is playable")
	}
	if canPlayLandFromZoneByRuleEffect(g, game.Player2, withCounter, zone.Exile) {
		t.Fatal("opponent may play from an exile permission they do not control")
	}
}

// TestCastSpellFromExileRequiresCroakCounter verifies the cast-from-exile
// permission only reaches exiled spells that carry a croak counter.
func TestCastSpellFromExileRequiresCroakCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	playAndCastFromExileWithCroakPermanent(g, game.Player1)

	withCounter := addCardToExile(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Croaked Bolt", Types: []types.Card{types.Instant}}})
	withoutCounter := addCardToExile(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Plain Bolt", Types: []types.Card{types.Instant}}})
	g.AddExileCounter(withCounter, counter.Croak, 1)

	if !canCastSpellsFromZoneByRuleEffect(g, game.Player1, withCounter, zone.Exile, game.FaceFront) {
		t.Fatal("croak-countered exiled spell is not castable despite the permission")
	}
	if canCastSpellsFromZoneByRuleEffect(g, game.Player1, withoutCounter, zone.Exile, game.FaceFront) {
		t.Fatal("exiled spell without a croak counter is castable")
	}
	if canCastSpellsFromZoneByRuleEffect(g, game.Player2, withCounter, zone.Exile, game.FaceFront) {
		t.Fatal("opponent may cast from an exile permission they do not control")
	}
}

// TestApplyPlayLandFromExileWithCounter drives the full action path: playing a
// croak-countered land from exile succeeds and consumes the land drop.
func TestApplyPlayLandFromExileWithCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	playAndCastFromExileWithCroakPermanent(g, game.Player1)
	landID := addCardToExile(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land}}})
	g.AddExileCounter(landID, counter.Croak, 1)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.PlayLandFaceFromZone(landID, zone.Exile, game.FaceFront)) {
		t.Fatal("playing a croak-countered land from exile was rejected despite the permission")
	}
	if g.Turn.LandsPlayedThisTurn != 1 {
		t.Fatalf("lands played = %d, want 1", g.Turn.LandsPlayedThisTurn)
	}
	if g.Players[game.Player1].Exile.Contains(landID) {
		t.Fatal("land remained in exile after being played")
	}
}

// TestApplyPlayLandFromExileWithoutCounterRejected verifies an exiled land that
// lacks the croak counter cannot be played through the action path.
func TestApplyPlayLandFromExileWithoutCounterRejected(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	playAndCastFromExileWithCroakPermanent(g, game.Player1)
	landID := addCardToExile(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Island", Types: []types.Card{types.Land}}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if engine.applyAction(g, game.Player1, action.PlayLandFaceFromZone(landID, zone.Exile, game.FaceFront)) {
		t.Fatal("exiled land without a croak counter was playable through the action path")
	}
}

// grolnokOmnivorePermanent gives playerID a battlefield permanent carrying
// Grolnok, the Omnivore's paired abilities: the play/cast-from-exile static and
// the "whenever a permanent card is put into your graveyard from your library,
// exile it with a croak counter on it" trigger, whose subject is filtered to the
// permanent card types so instants and sorceries are excluded.
func grolnokOmnivorePermanent(g *game.Game, playerID game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, playerID, &game.CardDef{CardFace: game.CardFace{
		Name: "Grolnok, the Omnivore",
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{
				{
					Kind:               game.RuleEffectPlayLandsFromZone,
					AffectedPlayer:     game.PlayerYou,
					CastFromZone:       zone.Exile,
					PermanentTypes:     []types.Card{types.Land},
					ExileCounterFilter: opt.Val(counter.Croak),
				},
				{
					Kind:               game.RuleEffectCastSpellsFromZone,
					AffectedPlayer:     game.PlayerYou,
					CastFromZone:       zone.Exile,
					ExileCounterFilter: opt.Val(counter.Croak),
				},
			},
		}},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:         game.EventZoneChanged,
					Player:        game.TriggerPlayerYou,
					MatchFromZone: true,
					FromZone:      zone.Library,
					MatchToZone:   true,
					ToZone:        zone.Graveyard,
					SubjectSelection: game.Selection{RequiredTypesAny: []types.Card{
						types.Artifact, types.Battle, types.Creature,
						types.Enchantment, types.Land, types.Planeswalker,
					}},
				},
			},
			Content: game.Mode{
				Sequence: []game.Instruction{{
					Primitive: game.MoveCard{
						Card:        game.CardReference{Kind: game.CardReferenceEvent},
						FromZone:    zone.Graveyard,
						Destination: zone.Exile,
						Counter:     opt.Val(counter.Croak),
					},
				}},
			}.Ability(),
		}},
	}})
}

// TestGrolnokTriggerExilesMilledPermanentWithCroak drives Grolnok's trigger
// end-to-end: a permanent card put into its owner's graveyard from their library
// fires the trigger, which exiles it with a croak counter, and the play/cast
// static then reaches it as a spell castable from exile.
func TestGrolnokTriggerExilesMilledPermanentWithCroak(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	grolnokOmnivorePermanent(g, game.Player1)
	creatureID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Milled Frog", Types: []types.Card{types.Creature}}})

	g.TriggerEventCursor = len(g.Events)
	if !moveCardBetweenZones(g, game.Player1, creatureID, zone.Library, zone.Graveyard) {
		t.Fatal("could not move the creature from library to graveyard")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("milling a permanent card did not fire Grolnok's exile trigger")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Exile.Contains(creatureID) {
		t.Fatal("milled permanent was not exiled by Grolnok's trigger")
	}
	if !g.HasExileCounter(creatureID, counter.Croak) {
		t.Fatal("exiled permanent is missing the croak counter placed by the trigger")
	}
	if !canCastSpellsFromZoneByRuleEffect(g, game.Player1, creatureID, zone.Exile, game.FaceFront) {
		t.Fatal("croak-countered permanent in exile is not castable despite Grolnok's static")
	}
}

// TestGrolnokTriggerIgnoresMilledInstantAndSorcery verifies the permanent-typed
// subject filter: instants and sorceries put into the graveyard from the library
// do not fire Grolnok's trigger and are never exiled.
func TestGrolnokTriggerIgnoresMilledInstantAndSorcery(t *testing.T) {
	for _, tc := range []struct {
		name     string
		cardType types.Card
	}{
		{"instant", types.Instant},
		{"sorcery", types.Sorcery},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			grolnokOmnivorePermanent(g, game.Player1)
			cardID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Milled Spell", Types: []types.Card{tc.cardType}}})

			g.TriggerEventCursor = len(g.Events)
			if !moveCardBetweenZones(g, game.Player1, cardID, zone.Library, zone.Graveyard) {
				t.Fatal("could not move the card from library to graveyard")
			}
			if engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatalf("milling a %s wrongly fired Grolnok's permanent-only exile trigger", tc.name)
			}
			if !g.Players[game.Player1].Graveyard.Contains(cardID) {
				t.Fatalf("milled %s left the graveyard despite the trigger not firing", tc.name)
			}
			if g.Players[game.Player1].Exile.Contains(cardID) {
				t.Fatalf("milled %s was exiled even though it is not a permanent card", tc.name)
			}
		})
	}
}
