package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Rules review (CR 2026-09-25, https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt):
// 702.170d "A plotted card's owner may cast it from exile without paying its
// mana cost during their main phase while the stack is empty during any turn
// after the turn in which it became plotted." 702.143a lets the player who
// foretold a card from their hand cast it after that turn for its foretell
// cost, and 702.143d gives that right to the owner of a card that becomes
// foretold by an effect. No other player gains either alternative cost; another
// player's independent permission to cast the card follows its own cost.

// grantCastFromExile gives playerID a permanent per-card permission to cast
// cardID from exile for its normal costs, independent of plot or foretell.
func grantCastFromExile(g *game.Game, playerID game.PlayerID, cardID id.ID) {
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID: g.IDGen.Next(), Kind: game.RuleEffectCastFromZone, Controller: playerID,
		AffectedPlayer: game.PlayerYou, CastFromZone: zone.Exile, AffectedCardID: cardID,
		Duration: game.DurationPermanent,
	})
}

// requireCastRejected proves playerID cannot cast cardID from exile through
// enumeration, direct announcement validation, a cloned simulator branch, or
// direct application, and that the rejected attempt leaves the card exiled with
// its plotted/foretold record intact.
func requireCastRejected(t *testing.T, engine *Engine, g *game.Game, playerID game.PlayerID, cardID id.ID, records map[id.ID]int) {
	t.Helper()
	cast := action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)
	if actionsContain(engine.legalActions(g, playerID), cast) {
		t.Fatal("legal actions offer the cast from exile")
	}
	if engine.announcementLegal(g, playerID, cast) {
		t.Fatal("direct announcement of the cast from exile validated")
	}
	if _, accepted := engine.Simulator().Apply(g, playerID, cast, simPassPolicies()); accepted {
		t.Fatal("simulator branch accepted the cast from exile")
	}
	if engine.applyAction(g, playerID, cast) {
		t.Fatal("cast from exile applied")
	}
	if _, onStack := g.Stack.Peek(); onStack {
		t.Fatal("rejected cast put a spell on the stack")
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("rejected cast moved the card out of its owner's exile")
	}
	if _, ok := records[cardID]; !ok {
		t.Fatal("rejected cast cleared the plotted/foretold record")
	}
}

// requireCastAccepted proves playerID can cast cardID from exile through
// enumeration and direct validation, then applies it and checks the spell is on
// the stack under playerID's control.
func requireCastAccepted(t *testing.T, engine *Engine, g *game.Game, playerID game.PlayerID, cardID id.ID) {
	t.Helper()
	cast := action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)
	if !actionsContain(engine.legalActions(g, playerID), cast) {
		t.Fatal("legal actions do not offer the cast from exile")
	}
	if !engine.announcementLegal(g, playerID, cast) {
		t.Fatal("direct announcement of the cast from exile did not validate")
	}
	if !engine.applyAction(g, playerID, cast) {
		t.Fatal("cast from exile failed")
	}
	top, onStack := g.Stack.Peek()
	if !onStack || top.SourceID != cardID || top.Controller != playerID {
		t.Fatalf("stack top = %+v, want %v's spell from the cast card", top, playerID)
	}
}

func setPriority(g *game.Game, active, priority game.PlayerID) {
	g.Turn.ActivePlayer = active
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = priority
}

// TestPlottedCardCastOnlyByOwner verifies an opponent with priority in their
// own main phase cannot cast another player's plotted card through the plot
// permission by any route, while the owner still casts it free on a later turn.
func TestPlottedCardCastOnlyByOwner(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, plotSorcery(cost.Mana{cost.G}))
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	setSorceryTiming(g, game.Player1)
	g.Turn.TurnNumber = 3
	if !engine.applyAction(g, game.Player1, action.PlotCard(cardID)) {
		t.Fatal("plot action failed")
	}
	for range 9 {
		addBasicLandPermanent(g, game.Player2, types.Forest)
	}

	g.Turn.TurnNumber = 4
	setSorceryTiming(g, game.Player2)
	requireCastRejected(t, engine, g, game.Player2, cardID, g.PlottedCards)

	g.Turn.TurnNumber = 5
	setSorceryTiming(g, game.Player1)
	if !forest.Tapped {
		t.Fatal("precondition: the owner's only land should still be tapped")
	}
	requireCastAccepted(t, engine, g, game.Player1, cardID)
}

// TestForetoldCardCastOnlyByOwner verifies another player with priority cannot
// cast a foretold card for its foretell cost by any route, while its owner can.
func TestForetoldCardCastOnlyByOwner(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, foretellInstant(cost.Mana{cost.U}))
	addBasicLandPermanent(g, game.Player1, types.Island)
	addBasicLandPermanent(g, game.Player1, types.Island)
	ownerIsland := addBasicLandPermanent(g, game.Player1, types.Island)
	setSorceryTiming(g, game.Player1)
	g.Turn.TurnNumber = 3
	if !engine.applyAction(g, game.Player1, action.ForetellCard(cardID)) {
		t.Fatal("foretell action failed")
	}
	addBasicLandPermanent(g, game.Player2, types.Island)

	g.Turn.TurnNumber = 4
	setPriority(g, game.Player1, game.Player2)
	requireCastRejected(t, engine, g, game.Player2, cardID, g.ForetoldCards)

	setPriority(g, game.Player1, game.Player1)
	requireCastAccepted(t, engine, g, game.Player1, cardID)
	if !ownerIsland.Tapped {
		t.Fatal("owner's foretold cast did not pay the {U} foretell cost")
	}
}

// TestNonOwnerPermissionCastsPlottedCardForItsOwnCost verifies a non-owner's
// independent cast-from-exile permission over a plotted card follows that
// permission's cost route: the non-owner pays the card's mana cost rather than
// casting it free through the owner's plot permission.
func TestNonOwnerPermissionCastsPlottedCardForItsOwnCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:         "Plotted Growth",
		Types:        []types.Card{types.Sorcery},
		ManaCost:     opt.Val(cost.Mana{cost.G, cost.G}),
		SpellAbility: opt.Val(game.AbilityContent{}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.PlotKeyword{Cost: cost.Mana{cost.G}}},
		}},
	}})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	setSorceryTiming(g, game.Player1)
	g.Turn.TurnNumber = 3
	if !engine.applyAction(g, game.Player1, action.PlotCard(cardID)) {
		t.Fatal("plot action failed")
	}
	grantCastFromExile(g, game.Player2, cardID)
	first := addBasicLandPermanent(g, game.Player2, types.Forest)

	g.Turn.TurnNumber = 4
	setSorceryTiming(g, game.Player2)
	requireCastRejected(t, engine, g, game.Player2, cardID, g.PlottedCards)

	second := addBasicLandPermanent(g, game.Player2, types.Forest)
	requireCastAccepted(t, engine, g, game.Player2, cardID)
	if !first.Tapped || !second.Tapped {
		t.Fatal("non-owner's permission cast did not pay the {G}{G} mana cost")
	}
}

// TestNonOwnerPermissionCastsForetoldCardForItsOwnCost verifies a non-owner's
// independent cast-from-exile permission over a foretold card pays the card's
// mana cost, not its owner's foretell cost.
func TestNonOwnerPermissionCastsForetoldCardForItsOwnCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:         "Foretold Insight",
		Types:        []types.Card{types.Instant},
		ManaCost:     opt.Val(cost.Mana{cost.U, cost.U}),
		SpellAbility: opt.Val(game.AbilityContent{}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.ForetellKeyword{Cost: cost.Mana{cost.U}}},
		}},
	}})
	addBasicLandPermanent(g, game.Player1, types.Island)
	addBasicLandPermanent(g, game.Player1, types.Island)
	setSorceryTiming(g, game.Player1)
	g.Turn.TurnNumber = 3
	if !engine.applyAction(g, game.Player1, action.ForetellCard(cardID)) {
		t.Fatal("foretell action failed")
	}
	grantCastFromExile(g, game.Player2, cardID)
	first := addBasicLandPermanent(g, game.Player2, types.Island)

	g.Turn.TurnNumber = 4
	setSorceryTiming(g, game.Player2)
	requireCastRejected(t, engine, g, game.Player2, cardID, g.ForetoldCards)

	second := addBasicLandPermanent(g, game.Player2, types.Island)
	requireCastAccepted(t, engine, g, game.Player2, cardID)
	if !first.Tapped || !second.Tapped {
		t.Fatal("non-owner's permission cast did not pay the {U}{U} mana cost")
	}
}
