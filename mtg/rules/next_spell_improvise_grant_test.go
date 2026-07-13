package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

// nextSpellImproviseGrant builds the global, turn-scoped, controller-scoped
// one-shot "the next spell you cast this turn has improvise" rule effect
// produced by Archway of Innovation. selection filters the granted spells; an
// empty selection grants to the next spell of any kind.
func nextSpellImproviseGrant(g *game.Game, selection game.Selection) game.RuleEffect {
	return game.RuleEffect{
		ID:                     g.IDGen.Next(),
		Kind:                   game.RuleEffectGrantSpellKeyword,
		GrantedKeyword:         game.Improvise,
		Controller:             game.Player1,
		AffectedController:     game.ControllerYou,
		CardSelection:          selection,
		AppliesToNextSpellOnly: true,
		Duration:               game.DurationUntilEndOfTurn,
		CreatedTurn:            g.Turn.TurnNumber,
	}
}

// TestNextSpellImproviseGrantConsumedByFirstSpell proves the one-shot grant is
// consumed by exactly the first matching spell its controller casts, leaving
// later spells without the grant.
func TestNextSpellImproviseGrantConsumedByFirstSpell(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.RuleEffects = append(g.RuleEffects, nextSpellImproviseGrant(g, game.Selection{}))
	def := elfCreatureDef()

	first := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1, SourceTokenDef: def}
	if !spellGrantedKeyword(g, game.Player1, elfFace(def), 0, 0, game.Improvise) {
		t.Fatal("first spell should see the one-shot improvise grant")
	}
	consumeNextSpellKeywordGrantEffects(g, first)
	if len(g.RuleEffects) != 0 {
		t.Fatalf("global rule effects = %#v, want the one-shot grant consumed", g.RuleEffects)
	}
	if spellGrantedKeyword(g, game.Player1, elfFace(def), 0, 0, game.Improvise) {
		t.Fatal("a later spell should not see the grant after it is consumed")
	}
}

// TestNextSpellImproviseGrantIgnoresOpponentSpell proves the controller-scoped
// one-shot grant is not consumed by an opponent's spell.
func TestNextSpellImproviseGrantIgnoresOpponentSpell(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.RuleEffects = append(g.RuleEffects, nextSpellImproviseGrant(g, game.Selection{}))
	def := elfCreatureDef()

	opponentSpell := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player2, SourceTokenDef: def}
	consumeNextSpellKeywordGrantEffects(g, opponentSpell)
	if len(g.RuleEffects) != 1 {
		t.Fatalf("global rule effects = %#v, want the one-shot grant untouched by an opponent's spell", g.RuleEffects)
	}
}

// TestNextSpellImproviseGrantNotConsumedByNonmatchingSpell proves a filtered
// one-shot grant is not consumed by a spell outside its selection.
func TestNextSpellImproviseGrantNotConsumedByNonmatchingSpell(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	effect := nextSpellImproviseGrant(g, game.Selection{ExcludedTypes: []types.Card{types.Artifact}})
	g.RuleEffects = append(g.RuleEffects, effect)
	artifactDef := &game.CardDef{CardFace: game.CardFace{Name: "Ornithopter", Types: []types.Card{types.Artifact}}}

	artifactSpell := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1, SourceTokenDef: artifactDef}
	consumeNextSpellKeywordGrantEffects(g, artifactSpell)
	if len(g.RuleEffects) != 1 {
		t.Fatalf("global rule effects = %#v, want the filtered grant untouched by a nonmatching spell", g.RuleEffects)
	}
}

// TestNextSpellImproviseGrantEndToEnd proves the one-shot grant lets the first
// spell pay with improvise and is then consumed so the second spell cannot.
func TestNextSpellImproviseGrantEndToEnd(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.RuleEffects = append(g.RuleEffects, nextSpellImproviseGrant(g, game.Selection{}))
	firstSpell := addCardToHand(g, game.Player1, plainSpell("Grizzly Bears", types.Creature, cost.Mana{cost.O(2)}))
	secondSpell := addCardToHand(g, game.Player1, plainSpell("Hill Giant", types.Creature, cost.Mana{cost.O(2)}))
	// Four artifacts: the first spell taps two, leaving two untapped so the
	// second spell fails only because its grant was consumed, not for lack of
	// artifacts.
	for range 4 {
		addCombatPermanent(g, game.Player1, improviseArtifact())
	}
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(firstSpell, nil, 0, nil)) {
		t.Fatal("first spell should be castable with the one-shot improvise grant")
	}
	tapped := 0
	for _, permanent := range g.Battlefield {
		if permanent.Tapped {
			tapped++
		}
	}
	if tapped != 2 {
		t.Fatalf("tapped %d artifacts for the first spell, want 2", tapped)
	}

	setMainPhasePriority(g, game.Player1)
	if engine.applyAction(g, game.Player1, action.CastSpell(secondSpell, nil, 0, nil)) {
		t.Fatal("second spell should not receive the consumed one-shot grant")
	}
}

func elfFace(def *game.CardDef) *game.CardDef {
	face, _ := def.FaceDef(game.FaceFront)
	return face
}
