package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// grantImproviseStatic builds the reusable static that grants Improvise to the
// controller's spells matching selection (empty selection grants to every
// spell), modeling Inspiring Statuary's "Nonartifact spells you cast have
// improvise." and Ironheart's "Noncreature spells you cast have improvise."
func grantImproviseStatic(text string, selection game.Selection) game.StaticAbility {
	return game.StaticAbility{
		Text: text,
		RuleEffects: []game.RuleEffect{{
			Kind:               game.RuleEffectGrantSpellKeyword,
			GrantedKeyword:     game.Improvise,
			AffectedController: game.ControllerYou,
			CardSelection:      selection,
		}},
	}
}

func inspiringStatuaryPermanent() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Inspiring Statuary",
		Types: []types.Card{types.Artifact},
		StaticAbilities: []game.StaticAbility{
			game.ImproviseStaticBody,
			grantImproviseStatic(
				"Nonartifact spells you cast have improvise.",
				game.Selection{ExcludedTypes: []types.Card{types.Artifact}},
			),
		}},
	}
}

func plainSpell(name string, cardType types.Card, manaCost cost.Mana) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: name,
		Types:        []types.Card{cardType},
		ManaCost:     opt.Val(manaCost),
		SpellAbility: opt.Val(game.AbilityContent{})},
	}
}

// TestGrantedImproviseTapsArtifactsForNonartifactSpell proves the static group
// grant lets a nonartifact spell without native Improvise pay a generic cost by
// tapping artifacts.
func TestGrantedImproviseTapsArtifactsForNonartifactSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, inspiringStatuaryPermanent())
	spellID := addCardToHand(g, game.Player1, plainSpell("Grizzly Bears", types.Creature, cost.Mana{cost.O(2)}))
	first := addCombatPermanent(g, game.Player1, improviseArtifact())
	second := addCombatPermanent(g, game.Player1, improviseArtifact())
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("granted-improvise spell cast failed")
	}
	if !first.Tapped && !second.Tapped {
		t.Fatal("granted improvise did not tap any artifact for generic mana")
	}
	tapped := 0
	for _, permanent := range g.Battlefield {
		if permanent.Tapped {
			tapped++
		}
	}
	if tapped != 2 {
		t.Fatalf("tapped %d artifacts for a {2} spell, want 2", tapped)
	}
}

// TestGrantedImproviseExcludesArtifactSpell proves the nonartifact filter keeps
// an artifact spell from receiving the grant, so it cannot tap artifacts.
func TestGrantedImproviseExcludesArtifactSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, inspiringStatuaryPermanent())
	spellID := addCardToHand(g, game.Player1, plainSpell("Ornithopter", types.Artifact, cost.Mana{cost.O(2)}))
	addCombatPermanent(g, game.Player1, improviseArtifact())
	addCombatPermanent(g, game.Player1, improviseArtifact())
	setMainPhasePriority(g, game.Player1)

	if engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("artifact spell received the nonartifact-only improvise grant, want failure")
	}
}

// TestGrantedImproviseRemovedWhenSourceLeaves proves the static grant is derived
// from the source permanent, so removing the source removes the grant.
func TestGrantedImproviseRemovedWhenSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	statuary := addCombatPermanent(g, game.Player1, inspiringStatuaryPermanent())
	spellID := addCardToHand(g, game.Player1, plainSpell("Grizzly Bears", types.Creature, cost.Mana{cost.O(2)}))
	addCombatPermanent(g, game.Player1, improviseArtifact())
	addCombatPermanent(g, game.Player1, improviseArtifact())
	setMainPhasePriority(g, game.Player1)

	g.Battlefield = removePermanent(g.Battlefield, statuary)

	if engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("grant persisted after its source left the battlefield, want failure")
	}
}

// TestGrantedImproviseDoesNotDoubleCountNativeImprovise proves a spell that has
// Improvise natively and also matches a grant still taps one artifact per
// generic symbol, not two.
func TestGrantedImproviseDoesNotDoubleCountNativeImprovise(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Grant Improvise to every spell the controller casts.
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Universal Grant",
		Types:           []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{grantImproviseStatic("Spells you cast have improvise.", game.Selection{})}}})
	spellID := addCardToHand(g, game.Player1, improviseSpell(cost.Mana{cost.O(1)}))
	first := addCombatPermanent(g, game.Player1, improviseArtifact())
	second := addCombatPermanent(g, game.Player1, improviseArtifact())
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("natively-improvise spell with a redundant grant failed to cast")
	}
	tapped := 0
	for _, permanent := range []*game.Permanent{first, second} {
		if permanent.Tapped {
			tapped++
		}
	}
	if tapped != 1 {
		t.Fatalf("tapped %d artifacts for a {1} spell, want 1", tapped)
	}
}

func removePermanent(battlefield []*game.Permanent, target *game.Permanent) []*game.Permanent {
	kept := battlefield[:0]
	for _, permanent := range battlefield {
		if permanent != target {
			kept = append(kept, permanent)
		}
	}
	return kept
}
