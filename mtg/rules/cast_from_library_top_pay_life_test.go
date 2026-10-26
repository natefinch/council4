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

// castSpellsFromLibraryTopPayLifePermanent grants playerID, until it leaves,
// permission to cast spells from the top of their library while paying life
// equal to each spell's mana value rather than its mana cost ("If you cast a
// spell this way, pay life equal to its mana value rather than pay its mana
// cost.", Bolas's Citadel, Gwenom, Remorseless).
func castSpellsFromLibraryTopPayLifePermanent(g *game.Game, playerID game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, playerID, &game.CardDef{CardFace: game.CardFace{
		Name: "Test Citadel",
		StaticAbilities: []game.StaticAbility{{
			Text: "You may play lands and cast spells from the top of your library. If you cast a spell this way, pay life equal to its mana value rather than pay its mana cost.",
			RuleEffects: []game.RuleEffect{{
				Kind:                    game.RuleEffectCastSpellsFromZone,
				AffectedPlayer:          game.PlayerYou,
				CastFromZone:            zone.Library,
				TopCardOnly:             true,
				PayLifeEqualToManaValue: true,
			}},
		}},
	}})
}

func TestLibraryCastRouteRequiresPayLifeOnlyForFlaggedPermission(t *testing.T) {
	plainGame := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	plainSpellID := addCardToLibrary(plainGame, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Top Bear", Types: []types.Card{types.Creature}}})
	castSpellsFromLibraryTopPermanent(plainGame, game.Player1, nil)
	routes := spellCastRoutes(plainGame, game.Player1, plainSpellID, zone.Library, game.FaceFront, castCostMode{})
	if len(routes) != 1 || routes[0].payLife {
		t.Fatalf("plain cast-from-top routes = %+v, want one route that pays the mana cost", routes)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	spellID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Top Bear", Types: []types.Card{types.Creature}}})
	castSpellsFromLibraryTopPayLifePermanent(g, game.Player1)
	routes = spellCastRoutes(g, game.Player1, spellID, zone.Library, game.FaceFront, castCostMode{})
	if len(routes) != 1 || !routes[0].payLife {
		t.Fatalf("pay-life cast-from-top routes = %+v, want one pay-life route", routes)
	}
	if routes := spellCastRoutes(g, game.Player2, spellID, zone.Library, game.FaceFront, castCostMode{}); len(routes) != 0 {
		t.Fatalf("opponent routes = %+v, want none from the controller's pay-life permission", routes)
	}
}

func TestCastSpellFromLibraryTopPaysLifeInsteadOfMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	castSpellsFromLibraryTopPayLifePermanent(g, game.Player1)
	spellID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Top Ogre",
		Types:    []types.Card{types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.R}),
	}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Players[game.Player1].Life = 20

	act := action.CastSpellFaceFromZone(spellID, zone.Library, game.FaceFront, nil, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("casting the top library spell for life is not a legal action")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("casting the top library spell for life failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != spellID {
		t.Fatalf("stack top = %#v, want the cast spell", obj)
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana pool = %d, want 0 (paid with life, not mana)", got)
	}
	if got := g.Players[game.Player1].Life; got != 17 {
		t.Fatalf("life = %d, want 17 (paid 3 life for a mana value 3 spell)", got)
	}
}

func TestCastSpellFromLibraryTopForLifeRequiresEnoughLife(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	castSpellsFromLibraryTopPayLifePermanent(g, game.Player1)
	spellID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Top Ogre",
		Types:    []types.Card{types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.R}),
	}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Players[game.Player1].Life = 2

	act := action.CastSpellFaceFromZone(spellID, zone.Library, game.FaceFront, nil, 0, nil)
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("casting for life must be illegal when life is below the spell's mana value")
	}
}
