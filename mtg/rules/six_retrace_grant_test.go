package rules

import (
	"reflect"
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/s"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// These tests exercise the real, compiler-generated Six's second ability —
// "During your turn, nonland permanent cards in your graveyard have retrace." —
// through the runtime that #3052 activated. They prove the grant is scoped to
// Six's controller during that controller's turn, applies only to nonland
// permanent cards, synthesizes retrace's intrinsic cost (the card's mana cost
// plus discarding a land card), and casts end to end. The during-your-turn gate
// was previously unverified for a retrace grant; the opponent-turn test closes
// that gap. All checks flow through the same text-blind APIs Underworld Breach
// uses, so no card-name or Oracle-text inspection is involved.

func sixGraveCreature() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Grave Bear",
		Types:    []types.Card{types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(1), cost.G}),
	}}
}

func sixGraveLand() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Grave Land", Types: []types.Card{types.Land}}}
}

func sixGraveSorcery() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Grave Sorcery",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.R}),
	}}
}

// TestSixRealCardGrantsRetraceWithIntrinsicCost proves the registered card grants
// a retrace alternative to a nonland permanent card in its controller's graveyard
// whose cost is the card's own mana cost plus discarding a land card, reusing the
// repeatable escape mechanic.
func TestSixRealCardGrantsRetraceWithIntrinsicCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	addPermanentForSBA(g, game.Player1, cards.Six())
	creature := sixGraveCreature()

	alternatives := grantedGraveyardCastAlternatives(g, game.Player1, creature)
	if len(alternatives) != 1 {
		t.Fatalf("granted alternatives = %d, want 1 retrace grant", len(alternatives))
	}
	alt := alternatives[0]
	if alt.Mechanic != cost.AlternativeMechanicEscape {
		t.Fatalf("mechanic = %v, want the repeatable escape mechanic retrace reuses", alt.Mechanic)
	}
	if !alt.ManaCost.Exists || !reflect.DeepEqual(alt.ManaCost.Val, creature.ManaCost.Val) {
		t.Fatalf("retrace mana cost = %+v, want the card's own %+v", alt.ManaCost, creature.ManaCost.Val)
	}
	want := []cost.Additional{{
		Kind:          cost.AdditionalDiscard,
		Amount:        1,
		MatchCardType: true,
		CardType:      types.Land,
		Text:          "Discard a land card",
	}}
	if !reflect.DeepEqual(alt.AdditionalCosts, want) {
		t.Fatalf("retrace additional costs = %+v, want discard a land %+v", alt.AdditionalCosts, want)
	}
}

// TestSixRealCardRetraceExcludesLandAndNonpermanent proves the nonland permanent
// selection: a land card and a sorcery card in the controller's graveyard are not
// granted retrace.
func TestSixRealCardRetraceExcludesLandAndNonpermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	addPermanentForSBA(g, game.Player1, cards.Six())

	if got := len(grantedGraveyardCastAlternatives(g, game.Player1, sixGraveLand())); got != 0 {
		t.Fatalf("retrace alternatives for a land = %d, want 0", got)
	}
	if got := len(grantedGraveyardCastAlternatives(g, game.Player1, sixGraveSorcery())); got != 0 {
		t.Fatalf("retrace alternatives for a sorcery = %d, want 0 (permanent cards only)", got)
	}
}

// TestSixRealCardRetraceOnlyDuringControllersTurn closes the previously untested
// during-your-turn gate: the retrace grant is offered on the controller's turn
// and withdrawn on an opponent's turn even though the source stays on the
// battlefield.
func TestSixRealCardRetraceOnlyDuringControllersTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addPermanentForSBA(g, game.Player1, cards.Six())
	creature := sixGraveCreature()

	setMainPhasePriority(g, game.Player1)
	if got := len(grantedGraveyardCastAlternatives(g, game.Player1, creature)); got != 1 {
		t.Fatalf("retrace alternatives during controller's turn = %d, want 1", got)
	}

	setMainPhasePriority(g, game.Player2)
	if got := len(grantedGraveyardCastAlternatives(g, game.Player1, creature)); got != 0 {
		t.Fatalf("retrace alternatives during opponent's turn = %d, want 0 (during your turn only)", got)
	}
}

// TestSixRealCardRetraceAppliesOnlyToController proves the "your graveyard"
// scoping: only Six's controller may retrace their graveyard cards, not an
// opponent from theirs, even during that opponent's own turn.
func TestSixRealCardRetraceAppliesOnlyToController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addPermanentForSBA(g, game.Player1, cards.Six())
	creature := sixGraveCreature()

	setMainPhasePriority(g, game.Player1)
	if got := len(grantedGraveyardCastAlternatives(g, game.Player1, creature)); got != 1 {
		t.Fatalf("controller retrace alternatives = %d, want 1", got)
	}
	// On the opponent's own turn the opponent still gets nothing: the grant only
	// ever reaches Six's controller.
	setMainPhasePriority(g, game.Player2)
	if got := len(grantedGraveyardCastAlternatives(g, game.Player2, creature)); got != 0 {
		t.Fatalf("opponent retrace alternatives = %d, want 0 (grant is controller-only)", got)
	}
}

// TestSixRealCardRetraceStopsWhenSourceLeaves proves the grant is continuous:
// once Six leaves the battlefield the retrace alternative is no longer offered.
func TestSixRealCardRetraceStopsWhenSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	six := addPermanentForSBA(g, game.Player1, cards.Six())
	creature := sixGraveCreature()

	if got := len(grantedGraveyardCastAlternatives(g, game.Player1, creature)); got != 1 {
		t.Fatalf("retrace alternatives while Six is present = %d, want 1", got)
	}
	if !movePermanentToZone(g, six, zone.Graveyard) {
		t.Fatal("moving Six off the battlefield failed")
	}
	if got := len(grantedGraveyardCastAlternatives(g, game.Player1, creature)); got != 0 {
		t.Fatalf("retrace alternatives after Six left = %d, want 0", got)
	}
}

// TestSixRealCardRetraceCastsPermanentFromGraveyard is the end-to-end path: with
// the real Six on the battlefield, a nonland permanent card is cast from the
// graveyard by paying its mana cost and discarding a land card, and enters the
// battlefield on resolution.
func TestSixRealCardRetraceCastsPermanentFromGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addPermanentForSBA(g, game.Player1, cards.Six())
	cardID := addCardToGraveyard(g, game.Player1, sixGraveCreature())
	addCardToHand(g, game.Player1, sixGraveLand())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil)) {
		t.Fatal("real Six did not grant a payable retrace cast from the graveyard")
	}
	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("retraced card was not removed from the graveyard when cast")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceZone != zone.Graveyard {
		t.Fatalf("stack object = %+v, want granted retrace graveyard cast", obj)
	}
	if obj.Flashback {
		t.Fatal("granted retrace must not be marked flashback")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if !permanentForCardOnBattlefield(g, cardID) {
		t.Fatal("retraced permanent did not enter the battlefield on resolution")
	}
}

// TestSixRealCardRetraceRequiresLandToDiscard proves the retrace cast is
// unpayable without a land card to discard, even with mana available.
func TestSixRealCardRetraceRequiresLandToDiscard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addPermanentForSBA(g, game.Player1, cards.Six())
	cardID := addCardToGraveyard(g, game.Player1, sixGraveCreature())
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Not A Land", Types: []types.Card{types.Instant}}})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	setMainPhasePriority(g, game.Player1)

	if engine.applyAction(g, game.Player1, action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil)) {
		t.Fatal("retrace cast succeeded without a land card to discard")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("retrace card left the graveyard despite an unpayable discard cost")
	}
}
