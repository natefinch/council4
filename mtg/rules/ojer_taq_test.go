package rules

import (
	"slices"
	"testing"

	cardo "github.com/natefinch/council4/mtg/cards/o"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// soldierTokenDef is a bare creature token used to exercise Ojer Taq's
// creature-token tripling replacement.
func soldierTokenDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Soldier", Types: []types.Card{types.Creature}}}
}

// TestOjerTaqTriplesCreatureTokensUnderYourControl drives the real curated card
// through the token-creation replacement machinery: one creature token created
// under its controller becomes three (front face: "three times that many").
func TestOjerTaqTriplesCreatureTokensUnderYourControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, cardo.OjerTaqDeepestFoundation())

	if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player1, soldierTokenDef(), 1, false, [game.NumPlayers]PlayerAgent{}, nil) {
		t.Fatal("createTokenPermanentsWithChoices() = false, want true")
	}
	if got := countTokenPermanentsNamed(g, "Soldier"); got != 3 {
		t.Fatalf("creature tokens under Ojer Taq = %d, want 3", got)
	}
}

// TestOjerTaqTriplingCompoundsWithAnotherMultiplier proves the generic
// multiplier composes with a second replacement multiplier rather than being a
// hardcoded doubling: Ojer Taq (x3) plus an any-controller doubler (x2) turns
// one creature token into six. Multiplication order is irrelevant (1*3*2 ==
// 1*2*3 == 6), and the empty agents resolve the ordering choice by default.
func TestOjerTaqTriplingCompoundsWithAnotherMultiplier(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, cardo.OjerTaqDeepestFoundation())
	addReplacementPermanent(t, g, game.Player1, anyControllerTokenDoublingCardDef())

	if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player1, soldierTokenDef(), 1, false, [game.NumPlayers]PlayerAgent{}, nil) {
		t.Fatal("createTokenPermanentsWithChoices() = false, want true")
	}
	if got := countTokenPermanentsNamed(g, "Soldier"); got != 6 {
		t.Fatalf("creature tokens under Ojer Taq x3 and an x2 doubler = %d, want 6", got)
	}
}

// TestOjerTaqDoesNotTripleNoncreatureTokens proves the replacement is gated on
// the creature type: a Treasure (artifact) token is created singly.
func TestOjerTaqDoesNotTripleNoncreatureTokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, cardo.OjerTaqDeepestFoundation())
	treasure := &game.CardDef{CardFace: game.CardFace{
		Name:     "Treasure",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Treasure},
	}}

	if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player1, treasure, 1, false, [game.NumPlayers]PlayerAgent{}, nil) {
		t.Fatal("createTokenPermanentsWithChoices() = false, want true")
	}
	if got := countTokenPermanentsNamed(g, "Treasure"); got != 1 {
		t.Fatalf("non-creature tokens under Ojer Taq = %d, want 1 (only creature batches are tripled)", got)
	}
}

// TestOjerTaqDoesNotTripleTokensUnderOpponentControl proves the "under your
// control" scope: creature tokens an opponent creates are not tripled.
func TestOjerTaqDoesNotTripleTokensUnderOpponentControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, cardo.OjerTaqDeepestFoundation())

	if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player2, soldierTokenDef(), 1, false, [game.NumPlayers]PlayerAgent{}, nil) {
		t.Fatal("createTokenPermanentsWithChoices(Player2) = false, want true")
	}
	if got := countTokenPermanentsNamed(g, "Soldier"); got != 1 {
		t.Fatalf("opponent creature tokens under Ojer Taq = %d, want 1 (your-control scope)", got)
	}
}

// TestOjerTaqDiesReturnsTappedAndTransformed drives the real dies trigger end to
// end: destroying the front-face God returns it to the battlefield on its back
// face (Temple of Civilization), tapped and transformed, under its owner's
// control.
func TestOjerTaqDiesReturnsTappedAndTransformed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardInstance(g, game.Player1, cardo.OjerTaqDeepestFoundation())
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          game.Player1,
		Controller:     game.Player1,
		Face:           game.FaceFront,
	}
	g.Battlefield = append(g.Battlefield, permanent)

	if _, ok := destroyPermanent(g, permanent.ObjectID); !ok {
		t.Fatal("destroyPermanent() = false, want the front-face God to die")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("putTriggeredAbilitiesOnStack() = false, want the dies trigger")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	returned := permanentForCard(g, cardID)
	if returned == nil {
		t.Fatal("Ojer Taq was not returned to the battlefield")
	}
	if returned.Face != game.FaceBack || !returned.Transformed {
		t.Fatalf("returned face/transformed = %v/%v, want back/true (Temple of Civilization)", returned.Face, returned.Transformed)
	}
	if !returned.Tapped {
		t.Fatal("returned permanent Tapped = false, want true")
	}
	if returned.Owner != game.Player1 || returned.Controller != game.Player1 {
		t.Fatalf("returned owner/controller = %v/%v, want owner's control (Player1)", returned.Owner, returned.Controller)
	}
}

// TestTempleOfCivilizationTransformGatedByThreeAttackers evaluates the real
// back-face activation condition. It must require three attackers you control
// this turn (two is not enough), must not count an opponent's attacker, and the
// ability is sorcery-timed.
func TestTempleOfCivilizationTransformGatedByThreeAttackers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Temple Source",
		Types: []types.Card{types.Land},
	}})

	back := cardo.OjerTaqDeepestFoundation().Back
	if !back.Exists {
		t.Fatal("Ojer Taq has no back face")
	}
	if len(back.Val.ActivatedAbilities) != 1 {
		t.Fatalf("back-face activated abilities = %d, want 1", len(back.Val.ActivatedAbilities))
	}
	ability := back.Val.ActivatedAbilities[0]
	if ability.Timing != game.SorceryOnly {
		t.Fatalf("transform activation timing = %v, want SorceryOnly", ability.Timing)
	}
	cond := ability.ActivationCondition
	ctx := conditionContext{controller: game.Player1, source: source}

	// Two attackers you control is not enough.
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player1})
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player1})
	if conditionSatisfied(g, ctx, cond) {
		t.Fatal("transform condition satisfied after only two attackers")
	}

	// An opponent's attacker does not count toward your total.
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player2})
	if conditionSatisfied(g, ctx, cond) {
		t.Fatal("transform condition satisfied counting an opponent's attacker")
	}

	// A third attacker you control reaches the required count.
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player1})
	if !conditionSatisfied(g, ctx, cond) {
		t.Fatal("transform condition not satisfied after attacking with three creatures")
	}
}

// TestOjerTaqFacesAndAbilities locks the shape of the real curated card across
// both faces so a regeneration regression is caught: the front God carries
// vigilance, the creature-token tripling replacement, and the dies trigger; the
// back land produces white mana and can transform at sorcery speed.
func TestOjerTaqFacesAndAbilities(t *testing.T) {
	def := cardo.OjerTaqDeepestFoundation()
	if def.Layout != game.LayoutTransform {
		t.Fatalf("layout = %v, want LayoutTransform", def.Layout)
	}

	front := def.CardFace
	if !slices.Contains(front.Types, types.Creature) {
		t.Fatal("front face is not a creature")
	}
	if len(front.StaticAbilities) != 1 {
		t.Fatalf("front static abilities = %d, want 1 (vigilance)", len(front.StaticAbilities))
	}
	if len(front.ReplacementAbilities) != 1 {
		t.Fatalf("front replacement abilities = %d, want 1 (token tripling)", len(front.ReplacementAbilities))
	}
	if len(front.TriggeredAbilities) != 1 {
		t.Fatalf("front triggered abilities = %d, want 1 (dies return)", len(front.TriggeredAbilities))
	}

	if !def.Back.Exists {
		t.Fatal("Ojer Taq has no back face")
	}
	back := def.Back.Val
	if !slices.Contains(back.Types, types.Land) {
		t.Fatal("back face is not a land")
	}
	if len(back.ManaAbilities) != 1 {
		t.Fatalf("back mana abilities = %d, want 1 ({T}: add {W})", len(back.ManaAbilities))
	}
	if len(back.ActivatedAbilities) != 1 {
		t.Fatalf("back activated abilities = %d, want 1 (transform)", len(back.ActivatedAbilities))
	}
}
