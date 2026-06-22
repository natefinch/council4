package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func escapeSpell() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Escape Spell",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.O(5)}),
		SpellAbility: opt.Val(game.Mode{
			Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
		}.Ability()),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.SimpleKeyword{Kind: game.Escape}},
		}},
		AlternativeCosts: []cost.Alternative{{
			Label:    "Escape",
			ManaCost: opt.Val(cost.Mana{cost.G}),
			AdditionalCosts: []cost.Additional{
				{Kind: cost.AdditionalExile, Text: "Exile two other cards from your graveyard", Source: zone.Graveyard, Amount: 2, ExcludeSource: true},
			},
		}}},
	}
}

func TestEscapeCastsFromGraveyardPayingExileCostAndResolvesNormally(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToGraveyard(g, game.Player1, escapeSpell())
	first := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fuel One"}})
	second := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fuel Two"}})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	setMainPhasePriority(g, game.Player1)

	act := action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("escape cast from graveyard failed")
	}

	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("escape spell was not removed from the graveyard when cast")
	}
	if !g.Players[game.Player1].Exile.Contains(first) || !g.Players[game.Player1].Exile.Contains(second) {
		t.Fatal("escape did not exile the two other graveyard cards as its cost")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceZone != zone.Graveyard {
		t.Fatalf("stack object = %+v, want escape graveyard cast", obj)
	}
	if obj.Flashback {
		t.Fatal("escape spell must not be marked as flashback (it is not exiled on resolution)")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("escape spell was exiled on resolution; escape spells return to the graveyard")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("escape spell did not return to the graveyard after resolving")
	}
}

func TestEscapePaymentExcludesSourceCardFromGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToGraveyard(g, game.Player1, escapeSpell())
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Only Fuel"}})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	setMainPhasePriority(g, game.Player1)

	act := action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil)
	if engine.applyAction(g, game.Player1, act) {
		t.Fatal("escape cast succeeded with only one other graveyard card; the escaping card must not pay its own exile cost")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("escape spell left the graveyard despite an unpayable cost")
	}
}
