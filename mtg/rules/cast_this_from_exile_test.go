package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// misthollowGriffinDef builds a self-scoped exile-cast creature ("You may cast
// this card from exile.", Misthollow Griffin).
func misthollowGriffinDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Griffin",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			Text:           "You may cast this card from exile.",
			ZoneOfFunction: zone.Exile,
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCastFromZone,
				AffectedPlayer: game.PlayerYou,
				CastFromZone:   zone.Exile,
				AffectedSource: true,
			}},
		}},
	}}
}

func addCardToExile(g *game.Game, playerID game.PlayerID, def *game.CardDef) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: def, Owner: playerID}
	g.Players[playerID].Exile.Add(cardID)
	return cardID
}

func TestCastThisFromExile(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToExile(g, game.Player1, misthollowGriffinDef())

	if !canCastFromZoneByRuleEffect(g, game.Player1, cardID, zone.Exile, game.FaceFront) {
		t.Fatal("self-permission should make the card castable from exile")
	}
	if canCastFromZoneByRuleEffect(g, game.Player2, cardID, zone.Exile, game.FaceFront) {
		t.Fatal("only the exile owner may cast the card from their exile")
	}
}

func TestCastThisFromExileIsSelfScoped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	griffinID := addCardToExile(g, game.Player1, misthollowGriffinDef())
	otherID := addCardToExile(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Plain Spell",
		Types: []types.Card{types.Sorcery},
	}})

	if !canCastFromZoneByRuleEffect(g, game.Player1, griffinID, zone.Exile, game.FaceFront) {
		t.Fatal("the source card should be castable from exile")
	}
	if canCastFromZoneByRuleEffect(g, game.Player1, otherID, zone.Exile, game.FaceFront) {
		t.Fatal("the permission must not extend to other exiled cards")
	}
}
