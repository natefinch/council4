package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// grafdiggersCagePermanent gives controller a battlefield permanent whose static
// ability stops creature cards from entering the battlefield out of any
// graveyard or library, mirroring Grafdigger's Cage.
func grafdiggersCagePermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Grafdigger's Cage",
		Types: []types.Card{types.Artifact},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCantEnterFromZones,
				EnterFromZones: []zone.Type{zone.Graveyard, zone.Library},
				PermanentTypes: []types.Card{types.Creature},
			}},
		}},
	}})
}

func TestEntryFromZoneProhibitedByGrafdiggersCage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	grafdiggersCagePermanent(g, game.Player1)

	creature := &game.CardDef{CardFace: game.CardFace{
		Name: "Reanimated Beast", Types: []types.Card{types.Creature}}}
	artifact := &game.CardDef{CardFace: game.CardFace{
		Name: "Recurred Relic", Types: []types.Card{types.Artifact}}}

	if !entryFromZoneProhibited(g, creature, zone.Graveyard) {
		t.Fatal("creature cards must not enter the battlefield from a graveyard")
	}
	if !entryFromZoneProhibited(g, creature, zone.Library) {
		t.Fatal("creature cards must not enter the battlefield from a library")
	}
	if entryFromZoneProhibited(g, creature, zone.Hand) {
		t.Fatal("creature cards may still enter the battlefield from hand")
	}
	if entryFromZoneProhibited(g, artifact, zone.Graveyard) {
		t.Fatal("the creature-only filter must not restrict noncreature cards")
	}
}

// TestCreatePermanentBlockedByEnterRestriction proves the entry gate actually
// stops the permanent from being created when it would enter from a restricted
// zone, while still allowing entry from hand.
func TestCreatePermanentBlockedByEnterRestriction(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	grafdiggersCagePermanent(g, game.Player1)

	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Reanimated Beast", Types: []types.Card{types.Creature}}}
	cardID := addCardToHand(g, game.Player2, def)
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("card instance not found")
	}
	g.Players[game.Player2].Hand.Remove(cardID)

	if _, ok := createCardPermanentFaceWithOptions(NewEngine(nil), g, card, game.Player2, zone.Graveyard, game.FaceFront, nil, permanentCreationOptions{}, [game.NumPlayers]PlayerAgent{}, nil); ok {
		t.Fatal("creature must not be created when entering from a graveyard")
	}
	if _, ok := createCardPermanentFaceWithOptions(NewEngine(nil), g, card, game.Player2, zone.Hand, game.FaceFront, nil, permanentCreationOptions{}, [game.NumPlayers]PlayerAgent{}, nil); !ok {
		t.Fatal("creature must still be created when entering from hand")
	}
}
