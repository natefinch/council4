package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const beseechLinkID = "beseech-exiled"

// beseechPayoffBody is the lowered three-instruction spell body of the reusable
// search/exile/conditional-cast payoff (Beseech the Mirror): search the library
// for a card and exile it face down publishing a link, then optionally cast the
// linked card for free when the spell was bargained and the linked card's mana
// value is 4 or less, then move the linked card from exile to hand as an ungated
// fallback (a no-op when the card was cast and has left exile).
func beseechPayoffBody() []game.Instruction {
	linked := game.CardReference{Kind: game.CardReferenceLinked, LinkID: beseechLinkID}
	return []game.Instruction{
		{
			Primitive: game.Search{
				Player:        game.ControllerReference(),
				Amount:        game.Fixed(1),
				Spec:          game.SearchSpec{SourceZone: zone.Library, Destination: zone.Exile, ExileFaceDown: true},
				PublishLinked: game.LinkedKey(beseechLinkID),
			},
		},
		{
			Primitive: game.CastForFree{Player: game.ControllerReference(), Zone: zone.Exile, Card: linked},
			Optional:  true,
			Condition: opt.Val(game.EffectCondition{
				Condition: opt.Val(game.Condition{SpellWasBargained: true}),
			}),
			CardCondition: opt.Val(game.CardSelection{
				Card:      linked,
				Selection: game.Selection{ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 4})},
			}),
		},
		{Primitive: game.MoveCard{Card: linked, FromZone: zone.Exile, Destination: zone.Hand}},
	}
}

// pushBeseechPayoff seeds a Beseech-payoff sorcery source card and pushes a
// matching stack object under controller, recording the bargained/copy cast
// branch so the resolution sees the same state a real cast would produce.
func pushBeseechPayoff(g *game.Game, controller game.PlayerID, bargained, isCopy bool) *game.StackObject {
	sourceID := g.IDGen.Next()
	g.CardInstances[sourceID] = &game.CardInstance{
		ID: sourceID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:         "Beseech Payoff",
			Types:        []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{Sequence: beseechPayoffBody()}.Ability()),
		}},
		Owner: controller,
	}
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   sourceID,
		Controller: controller,
		Bargained:  bargained,
		Copy:       isCopy,
	}
	g.Stack.Push(obj)
	return obj
}

// beseechAgent accepts or declines the optional free cast and always takes the
// first offered option for any other resolution choice (the singular search
// pick), so a single-card library is found deterministically.
type beseechAgent struct {
	acceptCast bool
}

func (beseechAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a beseechAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceMay {
		if a.acceptCast {
			return []int{1}
		}
		return []int{0}
	}
	return []int{0}
}

func instantDef(name string, manaValue int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.O(manaValue)}),
	}}
}

// TestBeseechBargainedCastsFoundCardMV4 verifies the bargained MV<=4 branch: the
// found card is exiled face down, cast for free during resolution (so it leaves
// exile for the stack), and the move-to-hand fallback is a no-op.
func TestBeseechBargainedCastsFoundCardMV4(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToLibrary(g, game.Player1, instantDef("Cheap Bolt", 4))
	pushBeseechPayoff(g, game.Player1, true, false)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: beseechAgent{acceptCast: true}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("cast card must not be put into hand")
	}
	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("cast card must have left exile")
	}
	top, ok := g.Stack.Peek()
	if !ok || top.SourceID != cardID {
		t.Fatalf("free-cast spell not on the stack (top=%+v, want source %v)", top, cardID)
	}
	if top.Controller != game.Player1 {
		t.Fatalf("free-cast controller = %v, want the caster %v", top.Controller, game.Player1)
	}
}

// TestBeseechBargainedMV5ToHand verifies the bargained-but-too-expensive branch:
// a mana value 5 card fails the linked mana-value gate, so it is not cast and
// falls to hand from exile.
func TestBeseechBargainedMV5ToHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToLibrary(g, game.Player1, instantDef("Pricey Bolt", 5))
	pushBeseechPayoff(g, game.Player1, true, false)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: beseechAgent{acceptCast: true}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("MV5 card must be put into hand (mana-value gate fails)")
	}
	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("card must have left exile for hand")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0 (nothing cast)", g.Stack.Size())
	}
}

// TestBeseechNotBargainedToHand verifies the unbargained branch: the free cast is
// gated on the spell being bargained, so an unbargained resolution never casts
// and the found card goes to hand.
func TestBeseechNotBargainedToHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToLibrary(g, game.Player1, instantDef("Cheap Bolt", 2))
	pushBeseechPayoff(g, game.Player1, false, false)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: beseechAgent{acceptCast: true}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("unbargained found card must be put into hand")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0 (unbargained: no free cast)", g.Stack.Size())
	}
}

// TestBeseechCopyNeverCasts verifies a copied Beseech spell is never bargained
// (CR 707.10c / 702.166), so its found card falls to hand even when the caster
// would accept the free cast.
func TestBeseechCopyNeverCasts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToLibrary(g, game.Player1, instantDef("Cheap Bolt", 1))
	pushBeseechPayoff(g, game.Player1, false, true)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: beseechAgent{acceptCast: true}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("copy's found card must be put into hand (copies are never bargained)")
	}
}

// TestBeseechDeclineToHand verifies declining the optional free cast leaves the
// found card in exile at the cast step, so the fallback moves it to hand.
func TestBeseechDeclineToHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToLibrary(g, game.Player1, instantDef("Cheap Bolt", 3))
	pushBeseechPayoff(g, game.Player1, true, false)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: beseechAgent{acceptCast: false}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("declined card must be put into hand")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0 (declined: no free cast)", g.Stack.Size())
	}
}

// TestBeseechUncastableLandToHand verifies an uncastable found card (a land has
// no spell to cast) falls to hand even when bargained, its mana value is 0, and
// the caster accepts.
func TestBeseechUncastableLandToHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Swamp",
		Types: []types.Card{types.Land},
	}})
	pushBeseechPayoff(g, game.Player1, true, false)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: beseechAgent{acceptCast: true}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("uncastable land must fall to hand")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0 (land is not cast)", g.Stack.Size())
	}
}

// TestBeseechUnpayableTargetsToHand verifies a found spell that cannot be legally
// cast (a targeted spell with no legal target available) is not cast and falls to
// hand, exercising the additional-requirement/uncastable fallback path.
func TestBeseechUnpayableTargetsToHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := instantDef("Doom", 2)
	def.SpellAbility = opt.Val(game.Mode{
		Targets: []game.TargetSpec{{
			MinTargets: 1,
			MaxTargets: 1,
			Allow:      game.TargetAllowPermanent,
			Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
		}},
	}.Ability())
	cardID := addCardToLibrary(g, game.Player1, def)
	pushBeseechPayoff(g, game.Player1, true, false)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: beseechAgent{acceptCast: true}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("uncastable targeted spell must fall to hand")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0 (no legal target: not cast)", g.Stack.Size())
	}
}

// TestBeseechFailToFindNothingHappens verifies an empty library finds nothing:
// no card is exiled, no link is published, and the fallback has nothing to move.
func TestBeseechFailToFindNothingHappens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	pushBeseechPayoff(g, game.Player1, true, false)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: beseechAgent{acceptCast: true}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if g.Players[game.Player1].Hand.Size() != 0 {
		t.Fatalf("hand size = %d, want 0 (nothing found)", g.Players[game.Player1].Hand.Size())
	}
	if g.Players[game.Player1].Exile.Size() != 0 {
		t.Fatalf("exile size = %d, want 0 (nothing exiled)", g.Players[game.Player1].Exile.Size())
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0 (nothing cast)", g.Stack.Size())
	}
}

// TestBeseechFreeCastRetainsTargets verifies the free cast picks a legal target
// for a targeted found spell, so the spell reaches the stack with a target and
// leaves exile.
func TestBeseechFreeCastRetainsTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := instantDef("Shock", 1)
	def.SpellAbility = opt.Val(game.Mode{
		Targets: []game.TargetSpec{{
			MinTargets: 1,
			MaxTargets: 1,
			Allow:      game.TargetAllowPlayer,
		}},
	}.Ability())
	cardID := addCardToLibrary(g, game.Player1, def)
	pushBeseechPayoff(g, game.Player1, true, false)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: beseechAgent{acceptCast: true}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("cast targeted card must have left exile")
	}
	top, ok := g.Stack.Peek()
	if !ok || top.SourceID != cardID {
		t.Fatalf("free-cast targeted spell not on the stack (top=%+v, want %v)", top, cardID)
	}
	if len(top.Targets) == 0 {
		t.Fatal("free-cast spell must carry a chosen target")
	}
}

// TestBeseechExileFaceDownSecrecy verifies the search half exiles the found card
// face down and publishes it under the linked key before any later reveal.
func TestBeseechExileFaceDownSecrecy(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToLibrary(g, game.Player1, instantDef("Secret", 2))
	obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: g.IDGen.Next(), Controller: game.Player1}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: beseechAgent{}}

	got, found := engine.searchLibraryExileFaceDown(g, obj, agents, &TurnLog{}, game.Player1, game.SearchSpec{
		SourceZone: zone.Library, Destination: zone.Exile, ExileFaceDown: true,
	}, 1)

	if !found || got != cardID {
		t.Fatalf("searchLibraryExileFaceDown = (%v,%v), want (%v,true)", got, found, cardID)
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("found card must be in exile")
	}
	if !g.Players[game.Player1].Exile.IsFaceDown(cardID) {
		t.Fatal("exiled card must be face down (secrecy)")
	}
	if g.Players[game.Player1].Library.Contains(cardID) {
		t.Fatal("found card must have left the library")
	}
}

// TestBeseechMultipleResolutionsIndependent verifies two independent Beseech
// resolutions keep separate source-scoped links: each finds and banks its own
// card without disturbing the other's.
func TestBeseechMultipleResolutionsIndependent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCardToLibrary(g, game.Player1, instantDef("First", 2))
	agents := [game.NumPlayers]PlayerAgent{game.Player1: beseechAgent{}}

	pushBeseechPayoff(g, game.Player1, false, false)
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	second := addCardToLibrary(g, game.Player1, instantDef("Second", 3))
	pushBeseechPayoff(g, game.Player1, false, false)
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	hand := g.Players[game.Player1].Hand
	if !hand.Contains(first) || !hand.Contains(second) {
		t.Fatalf("both found cards must be in hand independently (first=%v second=%v)",
			hand.Contains(first), hand.Contains(second))
	}
	if got := hand.Size(); got != 2 {
		t.Fatalf("hand size = %d, want exactly 2", got)
	}
}
