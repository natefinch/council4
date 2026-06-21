package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// gravecrawlerDef builds a self-scoped graveyard-cast creature. When condition
// is present the permission is gated on controlling a Zombie.
func gravecrawlerDef(condition opt.V[game.Condition]) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Crawler",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Zombie},
		StaticAbilities: []game.StaticAbility{{
			Text:           "You may cast this card from your graveyard.",
			ZoneOfFunction: zone.Graveyard,
			Condition:      condition,
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCastFromZone,
				AffectedPlayer: game.PlayerYou,
				CastFromZone:   zone.Graveyard,
				AffectedSource: true,
			}},
		}},
	}}
}

func zombieControlCondition() opt.V[game.Condition] {
	return opt.Val(game.Condition{
		ControlsMatching: opt.Val(game.SelectionCount{
			Selection: game.Selection{SubtypesAny: []types.Sub{types.Zombie}},
		}),
	})
}

func TestCastThisFromGraveyardUnconditional(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToGraveyard(g, game.Player1, gravecrawlerDef(opt.V[game.Condition]{}))

	if !canCastFromZoneByRuleEffect(g, game.Player1, cardID, zone.Graveyard, game.FaceFront) {
		t.Fatal("unconditional self-permission should make the card castable from the graveyard")
	}
	if canCastFromZoneByRuleEffect(g, game.Player2, cardID, zone.Graveyard, game.FaceFront) {
		t.Fatal("only the graveyard owner may cast the card from their graveyard")
	}
}

func TestCastThisFromGraveyardConditionGate(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToGraveyard(g, game.Player1, gravecrawlerDef(zombieControlCondition()))

	if canCastFromZoneByRuleEffect(g, game.Player1, cardID, zone.Graveyard, game.FaceFront) {
		t.Fatal("card must not be castable while the controller controls no Zombie")
	}

	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Walking Corpse",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Zombie},
	}})

	if !canCastFromZoneByRuleEffect(g, game.Player1, cardID, zone.Graveyard, game.FaceFront) {
		t.Fatal("card should be castable once the controller controls a Zombie")
	}
}

func TestCastThisFromGraveyardIsSelfScoped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	crawlerID := addCardToGraveyard(g, game.Player1, gravecrawlerDef(opt.V[game.Condition]{}))
	otherID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Plain Spell",
		Types: []types.Card{types.Sorcery},
	}})

	if !canCastFromZoneByRuleEffect(g, game.Player1, crawlerID, zone.Graveyard, game.FaceFront) {
		t.Fatal("the source card should be castable from the graveyard")
	}
	if canCastFromZoneByRuleEffect(g, game.Player1, otherID, zone.Graveyard, game.FaceFront) {
		t.Fatal("the permission must not extend to other graveyard cards")
	}
}
