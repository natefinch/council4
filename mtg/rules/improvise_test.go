package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestImproviseMakesGenericSpellPayableAndTapsArtifacts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, improviseSpell(cost.Mana{cost.O(2)}))
	first := addCombatPermanent(g, game.Player1, improviseArtifact())
	second := addCombatPermanent(g, game.Player1, improviseArtifact())
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("improvise spell cast failed")
	}
	if !first.Tapped || !second.Tapped {
		t.Fatal("improvise did not tap artifacts for generic mana")
	}
}

func TestImproviseDoesNotPayColoredSymbols(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, improviseSpell(cost.Mana{cost.G}))
	artifact := addCombatPermanent(g, game.Player1, improviseArtifact())
	setMainPhasePriority(g, game.Player1)

	if engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("improvise paid a colored pip with an artifact, want failure")
	}
	if artifact.Tapped {
		t.Fatal("artifact was tapped even though improvise cannot pay colored mana")
	}
}

func TestImproviseIgnoresTappedArtifacts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, improviseSpell(cost.Mana{cost.O(1)}))
	artifact := addCombatPermanent(g, game.Player1, improviseArtifact())
	artifact.Tapped = true
	setMainPhasePriority(g, game.Player1)

	if engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("improvise spell cast with only a tapped artifact, want failure")
	}
}

func improviseSpell(manaCost cost.Mana) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Improvise Spell",
		Types:           []types.Card{types.Sorcery},
		ManaCost:        opt.Val(manaCost),
		SpellAbility:    opt.Val(game.AbilityContent{}),
		StaticAbilities: []game.StaticAbility{game.ImproviseStaticBody}},
	}
}

func improviseArtifact() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Improvise Artifact",
		Types: []types.Card{types.Artifact}},
	}
}
