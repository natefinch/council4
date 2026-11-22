package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: pinned Oracle corpus 6da9ced0…; Protector of the Crown CN2
// #21 (24447332-4bc5-420d-972c-6d472669cfa3), Oracle f1052b21-ba96-499f-b9e6-9dba4ed82e1e;
// {5}{W} 2/5, becomes monarch on entry and redirects all damage to you to itself.
// Earthquake CM2 #95 (7cf2f3c7-66b2-48f6-97ef-09150903df7e), Oracle 9a40614b-50a3-422c-849e-53c8b7d3d204.
// {X}{R}, deals X to each creature without flying and each player.
// Read https://api.scryfall.com/cards/{printing-id}/rulings on 2026-11-21:
// Protector's 2016-08-23 rulings confirm redirection keeps the damage's combat
// status and does not require the monarch; Earthquake's 2004-10-04 ruling
// checks flying on resolution. CR effective 2026-11-06, from
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 608.2f (one simultaneous action); 614.1a, 614.6, 614.9 and 616.1f
// (redirection changes the recipient before damage is dealt); 120.3e,
// 120.4b–d and 603.2c (one resulting occurrence for the same source and
// recipient, before marking damage).
func TestReviewStaticRedirectEarthquakeProtectorOfTheCrown(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := cards.NewDefaultRegistry().Lookup("Protector of the Crown")
	if def == nil {
		t.Fatal("Protector of the Crown is not registered")
	}
	card := addCardToHand(g, game.Player2, def)
	protector, ok := createCardPermanent(g, g.CardInstances[card], game.Player2, zone.Hand)
	if !ok {
		t.Fatal("Protector of the Crown did not enter")
	}
	spell := castEarthquake(t, engine, g, 2)
	stackSpell, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("Earthquake missing from the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if protector.MarkedDamage != 4 {
		t.Fatalf("Protector marked damage = %d, want 4", protector.MarkedDamage)
	}
	for player := range g.Players {
		want := 38
		if game.PlayerID(player) == game.Player2 {
			want = 40
		}
		if got := g.Players[player].Life; got != want {
			t.Fatalf("player %d life = %d, want %d", player, got, want)
		}
	}
	var protectorHits []int
	for _, event := range g.Events {
		if event.Kind != game.EventDamageDealt {
			continue
		}
		if event.SourceID != spell || event.SourceObjectID != stackSpell.ID || event.CombatDamage {
			t.Fatalf("damage has wrong source or combat status: %+v", event)
		}
		if event.PermanentID == protector.ObjectID {
			protectorHits = append(protectorHits, event.Amount)
		}
	}
	if !slices.Equal(protectorHits, []int{4}) {
		t.Fatalf("Earthquake damage occurrences to Protector = %v, want [4]", protectorHits)
	}
	batches := damageEventBatchIDs(g)
	if len(batches) != game.NumPlayers {
		t.Fatalf("damage events = %d, want one per unprotected player plus one to Protector", len(batches))
	}
	for _, batch := range batches {
		if batch == 0 || batch != batches[0] {
			t.Fatalf("damage batch IDs = %v, want one shared nonzero ID", batches)
		}
	}
}

func TestStaticDamageRedirectCoalescesWithoutLosingAssignmentAttribution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCreaturePermanent(g, game.Player1)
	other := addCreaturePermanent(g, game.Player1)
	protector := addCreaturePermanent(g, game.Player2)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind:           game.RuleEffectRedirectDamageToSource,
		Controller:     game.Player2,
		AffectedPlayer: game.PlayerYou,
		SourceObjectID: protector.ObjectID,
	})
	batchID, nextBatchID := g.IDGen.Next(), g.IDGen.Next()
	results := resolveDamageBatch(g, []damageEvent{
		{sourceID: source.CardInstanceID, sourceObjectID: source.ObjectID, controller: game.Player1, player: game.Player2, amount: 2},
		{sourceID: source.CardInstanceID, sourceObjectID: source.ObjectID, controller: game.Player1, permanent: protector, amount: 1},
		{sourceID: other.CardInstanceID, sourceObjectID: other.ObjectID, controller: game.Player1, player: game.Player2, amount: 4},
	}, batchID, false, nil)
	for index, want := range []struct {
		amount     int
		redirected bool
	}{{2, true}, {1, false}, {4, true}} {
		got := results[index]
		if got.amount != want.amount || got.redirected != want.redirected || got.event.permanent != protector {
			t.Errorf("assignment %d: dealt %d, redirected %v, recipient %p; want %d, %v, %p",
				index, got.amount, got.redirected, got.event.permanent, want.amount, want.redirected, protector)
		}
	}
	resolveDamageBatch(g, []damageEvent{
		{sourceID: source.CardInstanceID, sourceObjectID: source.ObjectID, controller: game.Player1, player: game.Player2, amount: 1},
	}, nextBatchID, false, nil)

	type occurrence struct{ source, batch id.ID }
	want := map[occurrence]int{{source.ObjectID, batchID}: 3, {other.ObjectID, batchID}: 4, {source.ObjectID, nextBatchID}: 1}
	for _, event := range g.Events {
		if event.Kind != game.EventDamageDealt {
			continue
		}
		key := occurrence{event.SourceObjectID, event.SimultaneousID}
		if event.PermanentID != protector.ObjectID || event.Amount != want[key] {
			t.Errorf("unexpected damage occurrence: source %d, batch %d, recipient %d, amount %d",
				key.source, key.batch, event.PermanentID, event.Amount)
		}
		delete(want, key)
	}
	if len(want) != 0 {
		t.Fatalf("missing damage occurrences: %v", want)
	}
}

// TestRedirectDamageToSourcePreventsMonarchTheft proves that redirected combat
// damage does not reach the defending player, so an attacker cannot steal the
// monarchy (Protector of the Crown's purpose) and the player takes no commander
// damage — the combat damage is dealt to the redirect creature instead.
func TestRedirectDamageToSourcePreventsMonarchTheft(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMonarch(g, game.Player1)
	protector := addCreaturePermanent(g, game.Player1)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind:           game.RuleEffectRedirectDamageToSource,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerYou,
		SourceObjectID: protector.ObjectID,
		Duration:       game.DurationPermanent,
		CreatedTurn:    g.Turn.TurnNumber,
	})
	attacker := addCreaturePermanent(g, game.Player2)

	lifeBefore := g.Players[game.Player1].Life
	markPlayerCombatDamage(g, attacker, game.Player1, 3, &TurnLog{})

	if !g.Players[game.Player1].IsMonarch {
		t.Fatal("monarchy was stolen by redirected combat damage, want retained")
	}
	if g.Players[game.Player1].Life != lifeBefore {
		t.Fatalf("player life changed by %d, want 0 (damage redirected)", lifeBefore-g.Players[game.Player1].Life)
	}
	if protector.MarkedDamage != 3 {
		t.Fatalf("redirect creature marked damage = %d, want 3", protector.MarkedDamage)
	}
}

// damage that would be dealt to you is dealt to this creature instead": damage
// aimed at a player with an active RuleEffectRedirectDamageToSource is dealt to
// the effect's source permanent instead, so the player's life is unchanged and
// the creature is marked with the damage. A player without the effect loses life
// normally.
func TestRedirectDamageToSourceDealsToCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	protector := addCreaturePermanent(g, game.Player1)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind:           game.RuleEffectRedirectDamageToSource,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerYou,
		SourceObjectID: protector.ObjectID,
		Duration:       game.DurationPermanent,
		CreatedTurn:    g.Turn.TurnNumber,
	})

	lifeBefore := g.Players[game.Player1].Life
	dealt := dealPlayerDamage(g, 0, 0, game.Player2, game.Player1, 5, true)
	if dealt != 5 {
		t.Fatalf("damage dealt = %d, want 5 (redirected damage is still dealt)", dealt)
	}
	if g.Players[game.Player1].Life != lifeBefore {
		t.Fatalf("protected player life changed by %d, want 0 (damage redirected)", lifeBefore-g.Players[game.Player1].Life)
	}
	if protector.MarkedDamage != 5 {
		t.Fatalf("redirect creature marked damage = %d, want 5", protector.MarkedDamage)
	}

	opponentBefore := g.Players[game.Player2].Life
	if got := dealPlayerDamage(g, 0, 0, game.Player1, game.Player2, 3, true); got != 3 {
		t.Fatalf("unredirected damage dealt = %d, want 3", got)
	}
	if lost := opponentBefore - g.Players[game.Player2].Life; lost != 3 {
		t.Fatalf("unredirected player life lost = %d, want 3", lost)
	}
}

func TestPlayerChoosesProtectionOrDamageRedirectionFirst(t *testing.T) {
	for _, test := range []struct {
		name       string
		prefer     string
		wantDamage int
	}{
		{name: "protection first", prefer: "protection", wantDamage: 0},
		{name: "redirection first", prefer: "redirection", wantDamage: 4},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			redirect := addCreaturePermanent(g, game.Player1)
			g.RuleEffects = append(g.RuleEffects,
				game.RuleEffect{
					Kind:           game.RuleEffectPlayerProtection,
					Controller:     game.Player1,
					AffectedPlayer: game.PlayerYou,
					Protection:     game.ProtectionKeyword{Everything: true},
				},
				game.RuleEffect{
					Kind:           game.RuleEffectRedirectDamageToSource,
					Controller:     game.Player1,
					AffectedPlayer: game.PlayerYou,
					SourceObjectID: redirect.ObjectID,
					Duration:       game.DurationPermanent,
				},
			)
			source := addCreaturePermanent(g, game.Player2)
			engine := NewEngine(nil)
			engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{
				game.Player1: replacementChoosingAgent{prefer: test.prefer},
			}, &TurnLog{})
			defer g.ClearChoiceContext()

			if dealt := dealPlayerDamage(g, source.CardInstanceID, source.ObjectID, game.Player2, game.Player1, 4, true); dealt != test.wantDamage {
				t.Fatalf("damage dealt = %d, want %d", dealt, test.wantDamage)
			}
			if redirect.MarkedDamage != test.wantDamage {
				t.Fatalf("redirect permanent marked damage = %d, want %d", redirect.MarkedDamage, test.wantDamage)
			}
		})
	}
}
