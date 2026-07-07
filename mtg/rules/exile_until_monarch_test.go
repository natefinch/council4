package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// exileUntilMonarchLinkKey mirrors the constant the cardgen lowering emits for
// the Palace Jailer exile link.
const exileUntilMonarchLinkKey = game.LinkedKey("exile-until-opponent-monarch")

// resolveExileUntilMonarch resolves the two instructions the cardgen lowering
// produces for "exile <target> until an opponent becomes the monarch": a linked
// exile of victim followed by the persistent become-monarch return delayed
// trigger, both scoped to source.
func resolveExileUntilMonarch(t *testing.T, engine *Engine, g *game.Game, source, victim *game.Permanent) {
	t.Helper()
	obj := linkedSourceObject(source)
	obj.Targets = []game.Target{game.PermanentTarget(victim.ObjectID)}
	resolveInstruction(engine, g, obj, game.Exile{
		Object:         game.TargetPermanentReference(0),
		ExileLinkedKey: exileUntilMonarchLinkKey,
	}, nil)
	resolveInstruction(engine, g, obj, game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
		EventPattern: opt.Val(game.TriggerPattern{
			Event:  game.EventBecameMonarch,
			Player: game.TriggerPlayerOpponent,
		}),
		OneShot: true,
		Window:  game.DelayedWindowUntilFires,
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.PutOnBattlefield{Source: game.LinkedBattlefieldSource(exileUntilMonarchLinkKey)},
		}}}.Ability(),
	}}, nil)
}

func exileUntilMonarchSourceDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Warden of the Crown", Types: []types.Card{types.Creature}}}
}

// TestExileUntilOpponentBecomesMonarchReturnsWhenOpponentTakesCrown models Palace
// Jailer while its source stays in play: the exiled creature stays exiled while
// its controller holds the crown and returns to its owner's control once an
// opponent becomes the monarch.
func TestExileUntilOpponentBecomesMonarchReturnsWhenOpponentTakesCrown(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := addCombatCreaturePermanent(g, game.Player2)
	source := addCombatPermanent(g, game.Player1, exileUntilMonarchSourceDef())
	setMonarch(g, game.Player1)

	resolveExileUntilMonarch(t, engine, g, source, victim)

	if permanentByCardID(g, victim.CardInstanceID) != nil {
		t.Fatal("victim remained on the battlefield after exile-until-monarch")
	}
	if !g.Players[game.Player2].Exile.Contains(victim.CardInstanceID) {
		t.Fatal("victim did not reach its owner's exile zone")
	}

	// The controller re-taking the crown must not return the card.
	setMonarch(g, game.Player1)
	if engine.putTriggeredAbilitiesOnStack(g) {
		engine.resolveTopOfStack(g, &TurnLog{})
	}
	if permanentByCardID(g, victim.CardInstanceID) != nil {
		t.Fatal("victim returned before an opponent became the monarch")
	}

	// An opponent taking the crown returns the exiled card to its owner.
	setMonarch(g, game.Player2)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("become-monarch return trigger did not fire when an opponent took the crown")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	returned := permanentByCardID(g, victim.CardInstanceID)
	if returned == nil || returned.Controller != game.Player2 {
		t.Fatalf("returned permanent = %+v, want victim back under owner Player2 control", returned)
	}
	if g.Players[game.Player2].Exile.Contains(victim.CardInstanceID) {
		t.Fatal("victim remained in exile after an opponent became the monarch")
	}
}

// TestExileUntilOpponentBecomesMonarchReturnsAfterSourceLeaves is the key
// ruling: Palace Jailer leaving the battlefield does NOT return the exiled
// creature; the game keeps watching, and the creature returns the next time an
// opponent becomes the monarch. A persistent event delayed trigger (not a
// battlefield-scoped ability) makes this hold after the source is gone.
func TestExileUntilOpponentBecomesMonarchReturnsAfterSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := addCombatCreaturePermanent(g, game.Player2)
	source := addCombatPermanent(g, game.Player1, exileUntilMonarchSourceDef())
	setMonarch(g, game.Player1)

	resolveExileUntilMonarch(t, engine, g, source, victim)
	if !g.Players[game.Player2].Exile.Contains(victim.CardInstanceID) {
		t.Fatal("victim did not reach exile")
	}

	// Source leaves the battlefield; the creature must stay exiled.
	movePermanentToZone(g, source, zone.Graveyard)
	if engine.putTriggeredAbilitiesOnStack(g) {
		engine.resolveTopOfStack(g, &TurnLog{})
	}
	if permanentByCardID(g, victim.CardInstanceID) != nil {
		t.Fatal("victim returned when the source left the battlefield; should stay exiled")
	}

	// Later, an opponent becomes the monarch; the persistent delayed trigger fires.
	setMonarch(g, game.Player2)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("delayed return trigger did not fire after the source had left")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	returned := permanentByCardID(g, victim.CardInstanceID)
	if returned == nil || returned.Controller != game.Player2 {
		t.Fatalf("returned permanent = %+v, want victim back under owner Player2 control", returned)
	}
}

// TestExileUntilReturnStaysExileOnlyWhenCardLeftExile is the Banishing Light /
// Oblivion Ring ruling: an exile-until return must return the card only while it
// is still the same object in exile. If the exiled card leaves exile before the
// return fires — here another effect moves it from exile to its owner's
// graveyard — the return does nothing, even though a same-CardID card now sits
// in the graveyard. The default (nil LinkedReturnZones) PutOnBattlefield is
// exile-only, so it must never reanimate the card from the graveyard.
func TestExileUntilReturnStaysExileOnlyWhenCardLeftExile(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := addCombatCreaturePermanent(g, game.Player2)
	source := addCombatPermanent(g, game.Player1, exileUntilMonarchSourceDef())
	setMonarch(g, game.Player1)

	resolveExileUntilMonarch(t, engine, g, source, victim)
	if !g.Players[game.Player2].Exile.Contains(victim.CardInstanceID) {
		t.Fatal("victim did not reach exile")
	}

	// Another effect moves the exiled card from exile to its owner's graveyard
	// while the return link is still live.
	if !g.Players[game.Player2].Exile.Remove(victim.CardInstanceID) {
		t.Fatal("victim was not in exile to move")
	}
	g.Players[game.Player2].Graveyard.Add(victim.CardInstanceID)

	// An opponent becomes the monarch: the return trigger fires but must do
	// nothing because the card is no longer the exiled object.
	setMonarch(g, game.Player2)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("become-monarch return trigger did not fire when an opponent took the crown")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if permanentByCardID(g, victim.CardInstanceID) != nil {
		t.Fatal("card was wrongly returned to the battlefield from the graveyard; exile-until return must be exile-only")
	}
	if !g.Players[game.Player2].Graveyard.Contains(victim.CardInstanceID) {
		t.Fatal("card left the graveyard; the exile-until return must leave a card that already left exile untouched")
	}
}
