package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// unionEntersPattern is the "another creature or Vehicle you control enters"
// trigger subject: a controller-scoped enters trigger whose subject unions a
// card type (Creature) with a subtype (Vehicle) via Selection.AnyOf, the sole
// disjunction the runtime honors.
func unionEntersPattern() *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:       game.EventPermanentEnteredBattlefield,
		Controller:  game.TriggerControllerYou,
		ExcludeSelf: true,
		SubjectSelection: game.Selection{
			AnyOf: []game.Selection{
				{RequiredTypesAny: []types.Card{types.Creature}},
				{SubtypesAny: []types.Sub{types.Vehicle}},
			},
		},
	}
}

// TestUnionEntersTriggerMatchesCreatureAndVehicle proves the type-or-subtype
// union enters trigger fires for a controlled creature and for a controlled
// noncreature Vehicle, but not for an unrelated artifact, an opponent's
// creature, or the source permanent itself. This exercises the real matching
// path (triggerMatchesEvent -> matchSelection) that the Prowl back face relies
// on, confirming Selection.AnyOf honors both a RequiredTypesAny type alternative
// and a SubtypesAny subtype alternative.
func TestUnionEntersTriggerMatchesCreatureAndVehicle(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	pattern := unionEntersPattern()

	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Creature",
		Types: []types.Card{types.Creature},
	}})
	vehicle := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Vehicle",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Vehicle},
	}})
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Artifact",
		Types: []types.Card{types.Artifact},
	}})
	opponentCreature := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Opponent Creature",
		Types: []types.Card{types.Creature},
	}})

	enters := func(p *game.Permanent, controller game.PlayerID) game.Event {
		return game.Event{
			Kind:        game.EventPermanentEnteredBattlefield,
			PermanentID: p.ObjectID,
			Controller:  controller,
		}
	}

	if !triggerMatchesEvent(g, source, pattern, enters(creature, game.Player1)) {
		t.Fatal("union enters trigger did not match a controlled creature")
	}
	if !triggerMatchesEvent(g, source, pattern, enters(vehicle, game.Player1)) {
		t.Fatal("union enters trigger did not match a controlled noncreature Vehicle")
	}
	if triggerMatchesEvent(g, source, pattern, enters(artifact, game.Player1)) {
		t.Fatal("union enters trigger matched an artifact that is neither a creature nor a Vehicle")
	}
	if triggerMatchesEvent(g, source, pattern, enters(opponentCreature, game.Player2)) {
		t.Fatal("union enters trigger matched an opponent's creature despite you-control scope")
	}

	selfEnters := game.Event{
		Kind:           game.EventPermanentEnteredBattlefield,
		PermanentID:    source.ObjectID,
		SourceObjectID: source.ObjectID,
		CardID:         source.CardInstanceID,
		Controller:     game.Player1,
	}
	if triggerMatchesEvent(g, source, pattern, selfEnters) {
		t.Fatal("union enters trigger matched its own entry despite ExcludeSelf")
	}
}
