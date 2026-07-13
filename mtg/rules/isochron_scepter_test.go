package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// declineChoiceAgent answers "No" to every may/optional choice and defers to the
// request default otherwise, letting a test observe the decline path.
type declineChoiceAgent struct{}

func (declineChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Action{}
}

func (declineChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceMay {
		return []int{0}
	}
	return request.DefaultSelection
}

// imprintCopyCastAbility builds the AbilityContent produced by lowering the
// imprint copy/cast idiom "{2}, {T}: You may copy the exiled card. If you do,
// you may cast the copy without paying its mana cost." (Isochron Scepter): a
// two-instruction sequence composing the generic CopyCard consent (publishing a
// success result) with the gated PlayLinkedExiledCard free copy cast.
func imprintCopyCastAbility() game.AbilityContent {
	return game.Mode{Sequence: []game.Instruction{
		{
			Optional:      true,
			Primitive:     game.CopyCard{Player: game.ControllerReference(), LinkID: testImprintLink},
			PublishResult: game.ResultKey("imprint-copy-made"),
		},
		{
			Optional: true,
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       game.ResultKey("imprint-copy-made"),
				Succeeded: game.TriTrue,
			}),
			Primitive: game.PlayLinkedExiledCard{
				Player:                game.ControllerReference(),
				LinkID:                testImprintLink,
				Copy:                  true,
				WithoutPayingManaCost: true,
			},
		},
	}}.Ability()
}

// isochronScepterDef builds a reusable imprint artifact with only the copy/cast
// activated ability. No card-name-specific behavior backs it.
func isochronScepterDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Isochron Scepter",
		Types: []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{
			{Content: imprintCopyCastAbility()},
		},
	}}
}

// activatedObjFor builds an activated-ability stack object whose source is the
// given permanent, matching the object identity the imprint link is keyed by.
func activatedObjFor(permanent *game.Permanent) *game.StackObject {
	return &game.StackObject{
		Kind:         game.StackActivatedAbility,
		SourceID:     permanent.ObjectID,
		SourceCardID: permanent.CardInstanceID,
		Controller:   permanent.Controller,
	}
}

// addExiledInstant creates an instant card owned by player and places it in
// exile, returning its instance ID.
func addExiledInstant(g *game.Game, owner game.PlayerID, def *game.CardDef) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Owner: owner, Def: def}
	g.Players[owner].Exile.Add(cardID)
	return cardID
}

// plainInstantDef is a targetless instant used to observe the copy without
// requiring target legality.
func plainInstantDef(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:         name,
		Types:        []types.Card{types.Instant},
		ManaCost:     opt.Val(cost.Mana{cost.O(1)}),
		SpellAbility: opt.Val(game.AbilityContent{}),
	}}
}

// damageInstantDef is an instant that deals 2 damage to a target opponent, used
// to observe target selection and effect resolution of the copy.
func damageInstantDef(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		SpellAbility: opt.Val(game.Mode{
			Targets:  []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "opponent"}},
			Sequence: []game.Instruction{{Primitive: game.Damage{Recipient: game.AnyTargetDamageRecipient(0), Amount: game.Fixed(2)}}},
		}.Ability()),
	}}
}

// resolveImprintCopyCast resolves the copy/cast ability content once for the
// given source object, imprinting cardID under the source's object-scoped
// imprint link, and returns the game so callers can inspect the stack.
func imprintLink(g *game.Game, obj *game.StackObject, cardID id.ID) {
	rememberLinkedObject(g, linkedObjectByObjectKey(g, obj, testImprintLink), game.LinkedObjectRef{CardID: cardID})
}

// TestImprintCopyCastPutsFreeCopyOnStackLeavingOriginalExiled verifies the
// activated ability puts a copy spell (a non-card copy carrying the imprinted
// card's copiable values) onto the stack while the original card stays in exile.
func TestImprintCopyCastPutsFreeCopyOnStackLeavingOriginalExiled(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	scepter := addCombatPermanent(g, game.Player1, isochronScepterDef())
	obj := activatedObjFor(scepter)
	cardID := addExiledInstant(g, game.Player1, plainInstantDef("Silence"))
	imprintLink(g, obj, cardID)

	engine.resolveAbilityContentWithChoices(g, obj, imprintCopyCastAbility(), [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("imprinted card left exile after copy")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1 copy spell", g.Stack.Size())
	}
	top := g.Stack.Objects()[0]
	if !top.Copy {
		t.Fatal("stacked spell is not a copy")
	}
	if top.SourceTokenDef == nil {
		t.Fatal("copy carries no embedded token definition of the imprinted card")
	}
	if top.SourceTokenDef.Name != "Silence" {
		t.Fatalf("copy name = %q, want Silence", top.SourceTokenDef.Name)
	}
	if top.Controller != game.Player1 {
		t.Fatalf("copy controller = %v, want Player1", top.Controller)
	}
}

// TestImprintCopyCastCopyCeasesToExistLeavingOriginalExiled verifies the copy is
// not a card token: once it resolves it ceases to exist, and the original
// imprinted card remains in exile (never joins a graveyard).
func TestImprintCopyCastCopyCeasesToExistLeavingOriginalExiled(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	scepter := addCombatPermanent(g, game.Player1, isochronScepterDef())
	obj := activatedObjFor(scepter)
	cardID := addExiledInstant(g, game.Player1, plainInstantDef("Silence"))
	imprintLink(g, obj, cardID)

	engine.resolveAbilityContentWithChoices(g, obj, imprintCopyCastAbility(), [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if g.Stack.Size() != 0 {
		t.Fatalf("stack size after resolution = %d, want 0 (copy ceased to exist)", g.Stack.Size())
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("imprinted card left exile after the copy resolved")
	}
	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("imprinted card entered the graveyard; the copy must not move the card")
	}
}

// TestImprintCopyCastDeclineLeavesStateUnchanged verifies declining the copy
// consent leaves the imprinted card in exile and puts nothing on the stack.
func TestImprintCopyCastDeclineLeavesStateUnchanged(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	scepter := addCombatPermanent(g, game.Player1, isochronScepterDef())
	obj := activatedObjFor(scepter)
	cardID := addExiledInstant(g, game.Player1, plainInstantDef("Silence"))
	imprintLink(g, obj, cardID)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: declineChoiceAgent{}}
	engine.resolveAbilityContentWithChoices(g, obj, imprintCopyCastAbility(), agents, &TurnLog{})

	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0 after declining the copy", g.Stack.Size())
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("imprinted card left exile after declining the copy")
	}
}

// TestImprintCopyCastWithoutImprintDoesNothing verifies the ability is a legal
// no-op when the source imprinted nothing: no copy is offered or made.
func TestImprintCopyCastWithoutImprintDoesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	scepter := addCombatPermanent(g, game.Player1, isochronScepterDef())
	obj := activatedObjFor(scepter)

	engine.resolveAbilityContentWithChoices(g, obj, imprintCopyCastAbility(), [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0 with no imprinted card", g.Stack.Size())
	}
}

// TestImprintCopyCastCardMovedOutOfExileDoesNothing verifies the ability makes
// no copy once the imprinted card has left exile (for example it changed zones).
func TestImprintCopyCastCardMovedOutOfExileDoesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	scepter := addCombatPermanent(g, game.Player1, isochronScepterDef())
	obj := activatedObjFor(scepter)
	cardID := addExiledInstant(g, game.Player1, plainInstantDef("Silence"))
	imprintLink(g, obj, cardID)

	g.Players[game.Player1].Exile.Remove(cardID)
	g.Players[game.Player1].Graveyard.Add(cardID)

	engine.resolveAbilityContentWithChoices(g, obj, imprintCopyCastAbility(), [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0 once the imprinted card left exile", g.Stack.Size())
	}
}

// TestImprintCopyCastCopiedScepterHasNoLink verifies a copied Isochron Scepter
// (a fresh object identity that never imprinted) shares no imprint link with the
// original, so activating its copy/cast ability does nothing.
func TestImprintCopyCastCopiedScepterHasNoLink(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	original := addCombatPermanent(g, game.Player1, isochronScepterDef())
	originalObj := activatedObjFor(original)
	cardID := addExiledInstant(g, game.Player1, plainInstantDef("Silence"))
	imprintLink(g, originalObj, cardID)

	copyScepter := addCombatPermanent(g, game.Player1, isochronScepterDef())
	copyObj := activatedObjFor(copyScepter)

	engine.resolveAbilityContentWithChoices(g, copyObj, imprintCopyCastAbility(), [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0 (a copied Scepter shares no imprint link)", g.Stack.Size())
	}
}

// TestImprintCopyCastMultipleSceptersAreIndependent verifies each Scepter's
// imprint is object-scoped: activating one copies its own imprinted card, not
// another Scepter's.
func TestImprintCopyCastMultipleSceptersAreIndependent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	scepterA := addCombatPermanent(g, game.Player1, isochronScepterDef())
	scepterB := addCombatPermanent(g, game.Player1, isochronScepterDef())
	objA := activatedObjFor(scepterA)
	objB := activatedObjFor(scepterB)
	cardA := addExiledInstant(g, game.Player1, plainInstantDef("Silence"))
	cardB := addExiledInstant(g, game.Player1, plainInstantDef("Counterspell"))
	imprintLink(g, objA, cardA)
	imprintLink(g, objB, cardB)

	engine.resolveAbilityContentWithChoices(g, objA, imprintCopyCastAbility(), [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1", g.Stack.Size())
	}
	if got := g.Stack.Objects()[0].SourceTokenDef.Name; got != "Silence" {
		t.Fatalf("copied card = %q, want Silence (Scepter A's imprint)", got)
	}
}

// TestImprintCopyCastChoosesTargetAndResolves verifies a targeted imprinted
// spell's copy chooses a legal target and its effect resolves against that
// target.
func TestImprintCopyCastChoosesTargetAndResolves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	scepter := addCombatPermanent(g, game.Player1, isochronScepterDef())
	obj := activatedObjFor(scepter)
	cardID := addExiledInstant(g, game.Player1, damageInstantDef("Shock"))
	imprintLink(g, obj, cardID)

	engine.resolveAbilityContentWithChoices(g, obj, imprintCopyCastAbility(), [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1 targeted copy", g.Stack.Size())
	}
	top := g.Stack.Objects()[0]
	if len(top.Targets) != 1 {
		t.Fatalf("copy targets = %d, want 1", len(top.Targets))
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if got := g.Players[game.Player2].Life; got != 38 {
		t.Fatalf("opponent life after copy resolved = %d, want 38", got)
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("imprinted card left exile after the targeted copy resolved")
	}
}
