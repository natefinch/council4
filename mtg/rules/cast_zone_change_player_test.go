package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3549. CR 2026-09-25 (verified local copy
// .cardwork/cr.txt, "effective as of September 25, 2026"): 601.2a moves the
// cast card from where it is to the stack and makes the casting player the
// spell's controller; 400.3 keeps a card only in its owner's library, graveyard
// or hand, so the graveyard a cast card leaves is its owner's; 109.5 makes
// "your" on Syr Konrad its controller; 603.10a makes "leaves a graveyard"
// triggers look back in time; 702.37c casts a morph card face down from any
// zone it could normally be cast from. Default-registry cards from the pinned
// corpus: Syr Konrad, the Grim oracle 14c3ff84-1e82-4606-a433-869fc52cc382
// ("... or a creature card leaves your graveyard, Syr Konrad deals 1 damage to
// each opponent"), Muldrotha, the Gravetide oracle
// e4625704-1d52-44e4-804f-2f45644d76ac, Birchlore Rangers oracle
// 667c5676-7a18-4fa9-b965-b26797538e23 (Creature, {G}, Morph {G}) and Zoetic
// Cavern oracle 3763de30-28e1-4689-a71c-07d2fea3a466 (Land, Morph {2}).

func castSpellEvent(t *testing.T, g *game.Game, cardID id.ID) game.Event {
	t.Helper()
	for _, event := range g.Events {
		if event.Kind == game.EventSpellCast && event.CardID == cardID {
			return event
		}
	}
	t.Fatalf("no spell-cast event for card %v", cardID)
	return game.Event{}
}

func otherPlayer(player game.PlayerID) game.PlayerID {
	if player == game.Player1 {
		return game.Player2
	}
	return game.Player1
}

// assertOnlyKonradTriggered puts pending triggers on the stack and requires
// exactly want's "leaves your graveyard" trigger, then resolves it.
func assertOnlyKonradTriggered(t *testing.T, g *game.Game, e *Engine, want *game.Permanent) {
	t.Helper()
	if !e.putTriggeredAbilitiesOnStack(g) {
		t.Fatalf("Player%d's Syr Konrad did not trigger", want.Controller+1)
	}
	var triggers []*game.StackObject
	for _, obj := range g.Stack.Objects() {
		if obj.Kind == game.StackTriggeredAbility {
			triggers = append(triggers, obj)
		}
	}
	if len(triggers) != 1 || triggers[0].SourceID != want.ObjectID {
		t.Fatalf("triggers = %+v, want only Player%d's Syr Konrad", triggers, want.Controller+1)
	}
	opponent := otherPlayer(want.Controller)
	before := [game.NumPlayers]int{}
	for i := range g.Players {
		before[i] = g.Players[i].Life
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[opponent].Life; got != before[opponent]-1 {
		t.Fatalf("Player%d life = %d, want %d", opponent+1, got, before[opponent]-1)
	}
	if got := g.Players[want.Controller].Life; got != before[want.Controller] {
		t.Fatalf("Konrad's controller life = %d, want unchanged %d", got, before[want.Controller])
	}
}

// Each player casting a creature card from their own graveyard through
// Muldrotha triggers only their own Syr Konrad, face up and face down; the
// opponent's Konrad is the negative control.
func TestGraveyardCastZoneChangeNamesGraveyardOwner(t *testing.T) {
	t.Parallel()
	for _, player := range []game.PlayerID{game.Player1, game.Player2} {
		for _, faceDown := range []bool{false, true} {
			name := fmt.Sprintf("Player%d face up", player+1)
			if faceDown {
				name = fmt.Sprintf("Player%d face down", player+1)
			}
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				g, e, registry := lifetimeGame(t)
				g.Turn.ActivePlayer = player
				g.Turn.PriorityPlayer = player
				muldrotha := addCombatPermanent(g, player, lifetimeDefinition(t, registry, muldrothaID))
				konrad := addCombatPermanent(g, player, productionChoiceCard(t, "Syr Konrad, the Grim"))
				addCombatPermanent(g, otherPlayer(player), productionChoiceCard(t, "Syr Konrad, the Grim"))
				var cardID id.ID
				var cast action.Action
				if faceDown {
					cardID = addCardToGraveyard(g, player, productionChoiceCard(t, "Zoetic Cavern"))
					g.Players[player].ManaPool.Add(mana.C, 3)
					cast = action.CastFaceDownFromZone(cardID, zone.Graveyard, game.FaceFront, game.FaceDownMorph).
						WithPlayPermission(muldrothaPermission(muldrotha, types.Creature))
				} else {
					cardID = addCardToGraveyard(g, player, productionChoiceCard(t, "Birchlore Rangers"))
					g.Players[player].ManaPool.Add(mana.G, 1)
					cast = graveyardCast(cardID, muldrotha, types.Creature)
				}
				if !e.applyAction(g, player, cast) {
					t.Fatalf("%s graveyard cast failed", name)
				}
				move := castZoneChangeEvent(t, g, cardID, zone.Graveyard)
				if move.Player != player || move.Controller != player {
					t.Fatalf("zone change Player=%v Controller=%v, want both %v", move.Player, move.Controller, player)
				}
				if spell := castSpellEvent(t, g, cardID); spell.Player != player || spell.Controller != player {
					t.Fatalf("spell cast Player=%v Controller=%v, want both %v", spell.Player, spell.Controller, player)
				}
				assertOnlyKonradTriggered(t, g, e, konrad)
			})
		}
	}
}

// Casting a card out of an opponent's graveyard leaves the owner's graveyard
// while the caster controls the spell (CR 400.3, 601.2a): only the owner's Syr
// Konrad sees its creature card leave, on the resolution-time paid and free
// cast routes, and the cast event still names the caster.
func TestOpponentGraveyardCastZoneChangeNamesOwnerNotCaster(t *testing.T) {
	t.Parallel()
	routes := []struct {
		name string
		cast func(e *Engine, g *game.Game, cardID id.ID) bool
	}{
		{"paid cast during resolution", func(e *Engine, g *game.Game, cardID id.ID) bool {
			g.Players[game.Player1].ManaPool.Add(mana.G, 1)
			return e.castPaidSpellFromSource(g, g.Players[game.Player2], game.Player1, cardID, zone.Graveyard,
				[game.NumPlayers]PlayerAgent{}, &TurnLog{})
		}},
		{"free cast during resolution", func(e *Engine, g *game.Game, cardID id.ID) bool {
			return e.chooseAndCastFreeSpellFromSource(g, g.Players[game.Player2], game.Player1, cardID, zone.Graveyard,
				game.Selection{}, opt.V[int]{}, false, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		}},
	}
	for _, route := range routes {
		t.Run(route.name, func(t *testing.T) {
			t.Parallel()
			g, e, _ := lifetimeGame(t)
			addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Syr Konrad, the Grim"))
			ownerKonrad := addCombatPermanent(g, game.Player2, productionChoiceCard(t, "Syr Konrad, the Grim"))
			cardID := addCardToGraveyard(g, game.Player2, productionChoiceCard(t, "Birchlore Rangers"))
			if !route.cast(e, g, cardID) {
				t.Fatal("Player1 cast from Player2's graveyard failed")
			}
			move := castZoneChangeEvent(t, g, cardID, zone.Graveyard)
			if move.Player != game.Player2 || move.Controller != game.Player1 {
				t.Fatalf("zone change Player=%v Controller=%v, want owner Player2 and caster Player1", move.Player, move.Controller)
			}
			if spell := castSpellEvent(t, g, cardID); spell.Player != game.Player1 || spell.Controller != game.Player1 {
				t.Fatalf("spell cast Player=%v Controller=%v, want caster Player1", spell.Player, spell.Controller)
			}
			assertOnlyKonradTriggered(t, g, e, ownerKonrad)
		})
	}
}

// A creature card cast from hand leaves no graveyard, so neither Konrad's
// "leaves your graveyard" trigger fires; the zone change still names the owner.
func TestHandCastZoneChangeNamesOwnerWithoutGraveyardTrigger(t *testing.T) {
	t.Parallel()
	g, e, _ := lifetimeGame(t)
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player2
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Syr Konrad, the Grim"))
	addCombatPermanent(g, game.Player2, productionChoiceCard(t, "Syr Konrad, the Grim"))
	cardID := addCardToHand(g, game.Player2, productionChoiceCard(t, "Birchlore Rangers"))
	g.Players[game.Player2].ManaPool.Add(mana.G, 1)
	if !e.applyAction(g, game.Player2, actionBuild.castSpell(cardID, zone.Hand, game.FaceFront, nil, 0, nil)) {
		t.Fatal("Player2 hand cast failed")
	}
	if move := castZoneChangeEvent(t, g, cardID, zone.Hand); move.Player != game.Player2 || move.Controller != game.Player2 {
		t.Fatalf("zone change Player=%v Controller=%v, want both Player2", move.Player, move.Controller)
	}
	if e.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("a hand cast triggered a graveyard-leaves ability")
	}
}
