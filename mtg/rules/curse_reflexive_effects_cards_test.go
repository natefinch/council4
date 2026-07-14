package rules

import (
	"testing"

	cardsc "github.com/natefinch/council4/mtg/cards/c"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// seedLibraryCard places a fresh, uniquely named card owned by the given player
// on top of that player's library and returns its ID, so a draw test can prove a
// player drew from its own library rather than another player's.
func seedLibraryCard(g *game.Game, owner game.PlayerID, name string) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   &game.CardDef{CardFace: game.CardFace{Name: name}},
		Owner: owner,
	}
	g.Players[owner].Library.Add(cardID)
	return cardID
}

// addTappedPermanent puts a tapped permanent of the given card type onto the
// battlefield under the player's control, so an untap test can observe whether a
// mass untap reaches it. A land argument models the permanents Curse of Bounty
// must leave tapped; any other type models the nonland permanents it untaps.
func addTappedPermanent(g *game.Game, controller game.PlayerID, cardType types.Card) *game.Permanent {
	permanent := addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bounty Target",
		Types: []types.Card{cardType},
	}})
	permanent.Tapped = true
	return permanent
}

// TestGeneratedCurseOfVitalityGainsLifeForControllerAndAttackers proves the
// generated Curse of Vitality resolves "Whenever enchanted player is attacked, you
// gain 2 life. Each opponent attacking that player does the same." to 2 life for
// the controller and 2 life for each opponent attacking the enchanted player.
// Player3, which attacked with two creatures, gains 2 rather than 4 (the group
// deduplicates), and the enchanted Player2 gains nothing. It is the end-to-end
// proof that the folded gain-life rider lowers to the reusable
// opponents-attacking group.
func TestGeneratedCurseOfVitalityGainsLifeForControllerAndAttackers(t *testing.T) {
	g := resolveCurseAttackTrigger(t, cardsc.CurseOfVitality())

	want := map[game.PlayerID]int{
		game.Player1: 42, // controller
		game.Player2: 40, // enchanted, attacked player gains nothing
		game.Player3: 42, // attacked with two creatures, still one gain
		game.Player4: 42, // attacked directly (its planeswalker attack does not double it)
	}
	for _, pid := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		if got := g.Players[pid].Life; got != want[pid] {
			t.Fatalf("player %v life = %d, want %d", pid, got, want[pid])
		}
	}
}

// TestGeneratedCurseOfVerbosityDrawsFromEachRecipientOwnLibrary proves the
// generated Curse of Verbosity resolves "Whenever enchanted player is attacked,
// you draw a card. Each opponent attacking that player does the same." to one draw
// from the controller's own library and one from each attacking opponent's own
// library. Every player's library is seeded with two cards so a would-be double
// draw for Player3 (two attackers) would be visible; each recipient draws exactly
// its own top card, and the enchanted Player2 draws nothing.
func TestGeneratedCurseOfVerbosityDrawsFromEachRecipientOwnLibrary(t *testing.T) {
	type libState struct{ top, bottom id.ID }
	seeded := map[game.PlayerID]libState{}
	setup := func(g *game.Game) {
		for _, pid := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
			bottom := seedLibraryCard(g, pid, "Verbosity Bottom")
			top := seedLibraryCard(g, pid, "Verbosity Top")
			seeded[pid] = libState{top: top, bottom: bottom}
		}
	}
	g := resolveCurseAttackTriggerWithSetup(t, cardsc.CurseOfVerbosity(), setup)

	for _, pid := range []game.PlayerID{game.Player1, game.Player3, game.Player4} {
		hand := g.Players[pid].Hand
		if hand.Size() != 1 || !hand.Contains(seeded[pid].top) {
			t.Fatalf("player %v hand size = %d, want only its own top card drawn", pid, hand.Size())
		}
		lib := g.Players[pid].Library
		if lib.Size() != 1 || !lib.Contains(seeded[pid].bottom) {
			t.Fatalf("player %v library size = %d, want only its own bottom card retained", pid, lib.Size())
		}
	}
	if hand := g.Players[game.Player2].Hand; hand.Size() != 0 {
		t.Fatalf("enchanted Player2 hand size = %d, want no draw", hand.Size())
	}
	if lib := g.Players[game.Player2].Library; lib.Size() != 2 {
		t.Fatalf("enchanted Player2 library size = %d, want both cards retained", lib.Size())
	}
}

// TestGeneratedCurseOfBountyUntapsEachRecipientNonlandsButNotLands proves the
// generated Curse of Bounty resolves "Whenever enchanted player is attacked, untap
// all nonland permanents you control. Each opponent attacking that player untaps
// all nonland permanents they control." by untapping each recipient's own nonland
// permanents while leaving their lands tapped. The controller's per-attacker untap
// is attributed to each attacking opponent (not the controller), so the enchanted
// Player2 — not a recipient — keeps everything tapped.
func TestGeneratedCurseOfBountyUntapsEachRecipientNonlandsButNotLands(t *testing.T) {
	nonland := map[game.PlayerID]*game.Permanent{}
	land := map[game.PlayerID]*game.Permanent{}
	setup := func(g *game.Game) {
		for _, pid := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
			nonland[pid] = addTappedPermanent(g, pid, types.Artifact)
			land[pid] = addTappedPermanent(g, pid, types.Land)
		}
	}
	g := resolveCurseAttackTriggerWithSetup(t, cardsc.CurseOfBounty(), setup)
	_ = g

	for _, pid := range []game.PlayerID{game.Player1, game.Player3, game.Player4} {
		if nonland[pid].Tapped {
			t.Fatalf("player %v nonland stayed tapped, want untapped", pid)
		}
		if !land[pid].Tapped {
			t.Fatalf("player %v land untapped, want left tapped", pid)
		}
	}
	if !nonland[game.Player2].Tapped {
		t.Fatal("enchanted Player2 nonland untapped, want left tapped")
	}
	if !land[game.Player2].Tapped {
		t.Fatal("enchanted Player2 land untapped, want left tapped")
	}
}

// TestGeneratedCurseOfVitalityNoAttackingOpponentsGainsControllerOnly proves that
// when the enchanted player is attacked only by the curse's controller, the
// opponents-attacking group is empty and only the controller gains life. It
// exercises the gain-life group path with no recipients.
func TestGeneratedCurseOfVitalityNoAttackingOpponentsGainsControllerOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	p1a := addCombatCreaturePermanent(g, game.Player1)
	p3a := addCombatCreaturePermanent(g, game.Player3)
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: p1a.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: p3a.ObjectID, Target: game.AttackTarget{Player: game.Player4}},
	}}
	resolveCurseAttackedAbility(t, g, cardsc.CurseOfVitality())

	want := map[game.PlayerID]int{game.Player1: 42, game.Player2: 40, game.Player3: 40, game.Player4: 40}
	for _, pid := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		if got := g.Players[pid].Life; got != want[pid] {
			t.Fatalf("player %v life = %d, want %d", pid, got, want[pid])
		}
	}
}

// TestGeneratedCurseOfBountyNoAttackingOpponentsUntapsControllerOnly proves that
// when the enchanted player is attacked only by the curse's controller, the
// controller still untaps its own nonland permanents while the empty
// opponents-attacking group untaps nothing. It exercises the per-player
// ForEachPlayerGroup untap with no members.
func TestGeneratedCurseOfBountyNoAttackingOpponentsUntapsControllerOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	p1a := addCombatCreaturePermanent(g, game.Player1)
	p3a := addCombatCreaturePermanent(g, game.Player3)
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: p1a.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: p3a.ObjectID, Target: game.AttackTarget{Player: game.Player4}},
	}}
	controllerNonland := addTappedPermanent(g, game.Player1, types.Artifact)
	opponentNonland := addTappedPermanent(g, game.Player3, types.Artifact)
	resolveCurseAttackedAbility(t, g, cardsc.CurseOfBounty())

	if controllerNonland.Tapped {
		t.Fatal("controller nonland stayed tapped with no attacking opponents, want untapped")
	}
	if !opponentNonland.Tapped {
		t.Fatal("non-attacking opponent nonland untapped, want left tapped")
	}
}

// TestGeneratedCurseOfVitalityAttributesLifeToAttackerEffectiveController proves
// the gain follows an attacker's effective controller rather than its owner: a
// creature owned by Player3 but controlled by Player4 attacking the enchanted
// player makes Player4 (not Player3) gain the life, alongside the controller.
func TestGeneratedCurseOfVitalityAttributesLifeToAttackerEffectiveController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	stolen := addCombatCreaturePermanent(g, game.Player3)
	stolen.Controller = game.Player4
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: stolen.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}}
	resolveCurseAttackedAbility(t, g, cardsc.CurseOfVitality())

	want := map[game.PlayerID]int{game.Player1: 42, game.Player2: 40, game.Player3: 40, game.Player4: 42}
	for _, pid := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		if got := g.Players[pid].Life; got != want[pid] {
			t.Fatalf("player %v life = %d, want %d", pid, got, want[pid])
		}
	}
}

// TestGeneratedCurseOfVitalityExcludesPlaneswalkerAndBattleAttackers proves the
// reflexive group counts only creatures attacking the enchanted player directly: a
// player attacking that player solely through its planeswalker or battle is not a
// recipient (CR 508.1), preserving the #3060 invariant. Player3 attacks Player2
// directly and gains 2, while Player4 attacks Player2 only via a planeswalker and a
// battle and gains nothing. Player4's exclusion is isolated from the shared
// harness's multiple-attacker dedup, so it would fail if the group counted
// non-player attacks.
func TestGeneratedCurseOfVitalityExcludesPlaneswalkerAndBattleAttackers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	direct := addCombatCreaturePermanent(g, game.Player3)
	viaPlaneswalker := addCombatCreaturePermanent(g, game.Player4)
	viaBattle := addCombatCreaturePermanent(g, game.Player4)
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: direct.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: viaPlaneswalker.ObjectID, Target: game.AttackTarget{Player: game.Player2, PlaneswalkerID: g.IDGen.Next()}},
		{Attacker: viaBattle.ObjectID, Target: game.AttackTarget{Player: game.Player2, BattleID: g.IDGen.Next()}},
	}}
	resolveCurseAttackedAbility(t, g, cardsc.CurseOfVitality())

	want := map[game.PlayerID]int{
		game.Player1: 42, // controller
		game.Player2: 40, // enchanted, attacked player
		game.Player3: 42, // attacked the enchanted player directly
		game.Player4: 40, // attacked only via a planeswalker and a battle, excluded
	}
	for _, pid := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		if got := g.Players[pid].Life; got != want[pid] {
			t.Fatalf("player %v life = %d, want %d", pid, got, want[pid])
		}
	}
}
