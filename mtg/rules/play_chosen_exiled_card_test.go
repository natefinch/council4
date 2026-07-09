package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// resolvePlayChosenExiledCard resolves a PlayChosenExiledCard ability controlled
// by Player1, letting agents supply the resolution-time card choice.
func resolvePlayChosenExiledCard(t *testing.T, g *game.Game, prim game.PlayChosenExiledCard, agents [game.NumPlayers]PlayerAgent) *TurnLog {
	t.Helper()
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, prim, nil)
	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agents, &log)
	return &log
}

func playFromZoneRuleEffect(g *game.Game, cardID id.ID) (game.RuleEffect, bool) {
	for i := range g.RuleEffects {
		effect := g.RuleEffects[i]
		if effect.Kind == game.RuleEffectPlayFromZone && effect.AffectedCardID == cardID {
			return effect, true
		}
	}
	return game.RuleEffect{}, false
}

// TestPlayChosenExiledCardGrantsFreePlayForChosenOpponentCard verifies that the
// activated ability offers only cards an opponent owns bearing the named exile
// counter, and grants the controller a free-play permission bound to the chosen
// card (Dauthi Voidwalker: "Choose an exiled card an opponent owns with a void
// counter on it. You may play it this turn without paying its mana cost.").
func TestPlayChosenExiledCardGrantsFreePlayForChosenOpponentCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	oppVoid := addCardToExile(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Opp Void"}})
	g.AddExileCounter(oppVoid, counter.Void, 1)
	// An opponent card with no void counter is excluded by the counter filter.
	addCardToExile(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Opp No Counter"}})
	// A card the controller owns is excluded by the opponent owner scope even
	// though it carries a void counter.
	ownVoid := addCardToExile(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Own Void"}})
	g.AddExileCounter(ownVoid, counter.Void, 1)

	log := resolvePlayChosenExiledCard(t, g, game.PlayChosenExiledCard{
		Player:                game.ControllerReference(),
		Zone:                  zone.Exile,
		OwnerScope:            game.PlayerOpponent,
		Counter:               opt.Val(counter.Void),
		Duration:              game.DurationThisTurn,
		WithoutPayingManaCost: true,
	}, [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
	})

	if len(log.Choices) != 1 {
		t.Fatalf("choices = %+v, want exactly one resolution choice", log.Choices)
	}
	if got := len(log.Choices[0].Request.Options); got != 1 {
		t.Fatalf("choice options = %d, want 1 (only the opponent-owned void-countered card)", got)
	}

	effect, ok := playFromZoneRuleEffect(g, oppVoid)
	if !ok {
		t.Fatal("no RuleEffectPlayFromZone was granted for the chosen card")
	}
	if !effect.WithoutPayingManaCost {
		t.Fatal("granted permission is not flagged without paying its mana cost")
	}
	if effect.Duration != game.DurationThisTurn {
		t.Fatalf("granted permission Duration = %v, want DurationThisTurn", effect.Duration)
	}
	if effect.CastFromZone != zone.Exile {
		t.Fatalf("granted permission CastFromZone = %v, want Exile", effect.CastFromZone)
	}
	if effect.Controller != game.Player1 {
		t.Fatalf("granted permission Controller = %v, want Player1", effect.Controller)
	}
	if _, ok := playFromZoneRuleEffect(g, ownVoid); ok {
		t.Fatal("controller-owned card must not receive a play permission (opponent scope)")
	}
	if !castFromZoneWithoutPayingManaCost(g, game.Player1, oppVoid, zone.Exile, game.FaceFront) {
		t.Fatal("controller should be able to cast the chosen card without paying its mana cost")
	}
}

// TestPlayChosenExiledCardNoEligibleCardIsLegalNoOp verifies that the ability
// resolves as a legal no-op, granting no permission, when no exiled card matches
// the opponent-owner and counter filter.
func TestPlayChosenExiledCardNoEligibleCardIsLegalNoOp(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Only a counterless opponent card exists, so nothing qualifies.
	addCardToExile(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Opp No Counter"}})

	log := resolvePlayChosenExiledCard(t, g, game.PlayChosenExiledCard{
		Player:                game.ControllerReference(),
		Zone:                  zone.Exile,
		OwnerScope:            game.PlayerOpponent,
		Counter:               opt.Val(counter.Void),
		Duration:              game.DurationThisTurn,
		WithoutPayingManaCost: true,
	}, [game.NumPlayers]PlayerAgent{})

	if len(log.Choices) != 0 {
		t.Fatalf("choices = %+v, want none when no card qualifies", log.Choices)
	}
	for i := range g.RuleEffects {
		if g.RuleEffects[i].Kind == game.RuleEffectPlayFromZone {
			t.Fatal("a play permission was granted with no eligible card")
		}
	}
}
