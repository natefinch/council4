package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// addHideawayLandPermanent puts a bare land permanent onto the battlefield and
// returns it. The Hideaway abilities are exercised directly through resolved
// instructions, so the permanent only needs a stable identity to key the
// source-scoped hideaway link.
func addHideawayLandPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:  "Hidden Vault",
			Types: []types.Card{types.Land},
		}},
		Owner: controller,
	}
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          controller,
		Controller:     controller,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

// hideawaySourceObject builds a triggered/activated-ability stack object whose
// source is the given land permanent, so handleHideawayExile and
// handlePlayHideawayCard resolve from the same source-scoped hideaway key.
func hideawaySourceObject(land *game.Permanent) *game.StackObject {
	return &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		SourceID:     land.ObjectID,
		SourceCardID: land.CardInstanceID,
		Controller:   land.Controller,
	}
}

func TestHideawayExileLinksChosenCardAndBottomsRest(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	land := addHideawayLandPermanent(g, game.Player1)
	// Library bottom-to-top: bottomFiller, c1, c2, c3, c4 (c4 is on top).
	bottomFiller := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bottom Filler"}})
	c1 := addCardToLibrary(g, game.Player1, simpleGainLifeInstant("Look One"))
	c2 := addCardToLibrary(g, game.Player1, simpleGainLifeInstant("Look Two"))
	c3 := addCardToLibrary(g, game.Player1, simpleGainLifeInstant("Look Three"))
	c4 := addCardToLibrary(g, game.Player1, simpleGainLifeInstant("Look Four"))

	obj := hideawaySourceObject(land)
	// The top four cards are c4, c3, c2, c1; choosing index 2 exiles c2.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{2}}}}
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.HideawayExile{Amount: game.Fixed(4)}}, agents, &TurnLog{})

	player := g.Players[game.Player1]
	if !player.Exile.Contains(c2) {
		t.Fatalf("chosen card %v not exiled", c2)
	}
	if !player.Exile.IsFaceDown(c2) {
		t.Fatal("chosen card was not exiled face down")
	}
	key := linkedObjectSourceKey(g, obj, hideawayLinkID)
	refs := linkedObjects(g, key)
	if len(refs) != 1 || refs[0].CardID != c2 {
		t.Fatalf("linked objects = %+v, want single ref to %v", refs, c2)
	}
	if got := player.Library.All(); !sameCardIDs(got, []id.ID{bottomFiller, c1, c3, c4}) {
		t.Fatalf("library = %+v, want bottomed look cards plus filler", got)
	}
	for _, cardID := range []id.ID{c1, c3, c4} {
		if !player.Library.Contains(cardID) {
			t.Fatalf("looked card %v not returned to library", cardID)
		}
	}
}

func TestPlayHideawayCardCastsExiledSpellForFree(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	land := addHideawayLandPermanent(g, game.Player1)
	spell := addCardToLibrary(g, game.Player1, simpleGainLifeInstant("Hidden Spell"))
	obj := hideawaySourceObject(land)
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.HideawayExile{Amount: game.Fixed(1)}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !g.Players[game.Player1].Exile.Contains(spell) {
		t.Fatalf("spell %v was not exiled by hideaway", spell)
	}

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.PlayHideawayCard{}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	top, ok := g.Stack.Peek()
	if !ok || top.SourceID != spell {
		t.Fatalf("stack top = %+v, want free-cast hidden spell %v", top, spell)
	}
	if g.Players[game.Player1].Exile.Contains(spell) {
		t.Fatal("hidden spell remained in exile after being played")
	}
	if refs := linkedObjects(g, linkedObjectSourceKey(g, obj, hideawayLinkID)); len(refs) != 0 {
		t.Fatalf("hideaway link not cleared after play: %+v", refs)
	}
}

func TestPlayHideawayCardPutsExiledLandOntoBattlefield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	land := addHideawayLandPermanent(g, game.Player1)
	exiledLand := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Hidden Forest",
		Types: []types.Card{types.Land},
	}})
	obj := hideawaySourceObject(land)
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.HideawayExile{Amount: game.Fixed(1)}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.PlayHideawayCard{}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if g.Players[game.Player1].Exile.Contains(exiledLand) {
		t.Fatal("hidden land remained in exile after being played")
	}
	found := false
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == exiledLand {
			found = true
		}
	}
	if !found {
		t.Fatalf("hidden land %v was not put onto the battlefield", exiledLand)
	}
}

func TestPlayHideawayCardDeclinedLeavesCardExiled(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	land := addHideawayLandPermanent(g, game.Player1)
	spell := addCardToLibrary(g, game.Player1, simpleGainLifeInstant("Hidden Spell"))
	obj := hideawaySourceObject(land)
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.HideawayExile{Amount: game.Fixed(1)}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	// Optional "may" declined: the envelope skips the play entirely.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{
		Primitive: game.PlayHideawayCard{},
		Optional:  true,
	}, agents, &TurnLog{})

	if !g.Players[game.Player1].Exile.Contains(spell) {
		t.Fatal("declined hideaway play should leave the card exiled")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want nothing cast after decline", g.Stack.Size())
	}
}
