package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: official CR effective 2026-11-06 (verified copy retained in
// the unit's .cardwork/cr-20261106.txt), read 509.1h, 509.3a-g, 509.4, 509.4a,
// 509.4b, 702.22h-i and 702.25a. 509.4: a creature put onto the battlefield
// blocking is "blocking" but, for trigger events and effects, it never
// "blocked"; 509.3a-b: "blocks" / "blocks a creature" don't trigger for it.
// 509.3c: "becomes blocked" triggers when the attacker is blocked by a creature
// put onto the battlefield blocking, but only if it was unblocked; 509.3d:
// "becomes blocked by a creature" triggers per new blocking pair. 509.4a: if the
// named creature is no longer attacking, or the entering creature's controller
// isn't its defending player, the creature enters but never blocks. 509.4b:
// blocking requirements and restrictions don't apply. 509.1h: an attacker stays
// blocked after its blockers leave combat. 702.22h-i: blocking one member of a
// band blocks the whole band, including through an effect. Card faces (pinned
// corpus): Konda, Lord of Eiganjo ({5}{W} 3/3 Vigilance, indestructible,
// Bushido 5 — "Whenever this creature blocks or becomes blocked, it gets +5/+5
// until end of turn."), Cave Tiger ({2}{G} 2/2 "Whenever this creature becomes
// blocked by a creature, this creature gets +1/+1 until end of turn."), Ashmouth
// Hound ({1}{R} 2/1 "Whenever this creature blocks or becomes blocked by a
// creature, this creature deals 1 damage to that creature."), Zhalfirin Knight
// ({2}{W}{W} 2/2 Flanking), Benalish Hero ({W} 1/1 Banding). Test faces take the
// effect shapes of #3504's cards: Flash Foliage ("Create a 1/1 green Saproling
// creature token that's blocking target creature attacking you."), Aetherplasm
// ("you may put a creature card from your hand onto the battlefield blocking
// that creature"), Mirror Match ("create a token that's a copy of that creature
// and that's blocking that creature") and Brimaz, King of Oreskos ("Whenever
// Brimaz blocks a creature, create a 1/1 white Cat Soldier creature token with
// vigilance that's blocking that creature.").

func attackingCreatureTarget() game.TargetSpec {
	return game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: "target attacking creature",
		Allow:      game.TargetAllowPermanent,
		Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}),
	}
}

func effectBlockInstant(name string, mode game.Mode) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:         name,
		ManaCost:     opt.Val(cost.Mana{cost.O(0)}),
		Types:        []types.Card{types.Instant},
		SpellAbility: opt.Val(mode.Ability()),
	}}
}

// saprolingBlockSpell is Flash Foliage's shape: "Create a 1/1 green Saproling
// creature token that's blocking target attacking creature."
func saprolingBlockSpell() *game.CardDef {
	return effectBlockInstant("Saproling Block Test", game.Mode{
		Targets: []game.TargetSpec{attackingCreatureTarget()},
		Sequence: []game.Instruction{{Primitive: game.CreateToken{
			Amount:        game.Fixed(1),
			Source:        game.TokenDef(vanillaCreature("Saproling", 1, 1)),
			EntryBlocking: opt.Val(game.TargetPermanentReference(0)),
		}}},
	})
}

// putBlockingSpell is Aetherplasm's shape: "Put a creature card from your hand
// onto the battlefield blocking target attacking creature."
func putBlockingSpell() *game.CardDef {
	return effectBlockInstant("Put Blocking Test", game.Mode{
		Targets: []game.TargetSpec{attackingCreatureTarget()},
		Sequence: []game.Instruction{{Primitive: game.ChooseFromZone{
			Player:      game.ControllerReference(),
			SourceZone:  zone.Hand,
			Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
			Quantity:    game.Fixed(1),
			Destination: game.ChooseDestination{Zone: zone.Battlefield},
			Riders:      game.ChooseRiders{EntersBlocking: opt.Val(game.TargetPermanentReference(0))},
		}}},
	})
}

// mirrorBlockSpell is Mirror Match's shape without its defender filter: "For
// each attacking creature, create a token that's a copy of that creature and
// that's blocking that creature."
func mirrorBlockSpell() *game.CardDef {
	return effectBlockInstant("Mirror Block Test", game.Mode{
		Sequence: []game.Instruction{{Primitive: game.CreateToken{
			Amount: game.Fixed(1),
			Source: game.TokenCopyOf(game.TokenCopySpec{
				Source: game.TokenCopySourceEachInGroup,
				Group: game.GroupRef(game.BattlefieldGroup(game.Selection{
					RequiredTypes: []types.Card{types.Creature},
					CombatState:   game.CombatStateAttacking,
				})),
			}),
			EntryBlockingCopySource: true,
		}}},
	})
}

// brimazShapeCreature is Brimaz's block trigger on a 1/10: "Whenever this
// creature blocks a creature, create a 1/1 white Cat Soldier creature token
// that's blocking that creature."
func brimazShapeCreature() *game.CardDef {
	def := vanillaCreature("Brimaz Shape Test", 1, 10)
	def.TriggeredAbilities = []game.TriggeredAbility{{
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhenever,
			Pattern: game.TriggerPattern{
				Event:        game.EventBlockerDeclared,
				Source:       game.TriggerSourceSelf,
				BlockerScope: game.BlockerTriggerEachBlockedCreature,
			},
		},
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.CreateToken{
			Amount:        game.Fixed(1),
			Source:        game.TokenDef(vanillaCreature("Cat Soldier", 1, 1)),
			EntryBlocking: opt.Val(game.EventRelatedPermanentReference()),
		}}}}.Ability(),
	}}
	return def
}

// becameBlockedBy returns the attacker's becomes-blocked events whose blocker
// is not one of the given declared blockers, keyed by blocker.
func becameBlockedByOthers(events []game.Event, attacker id.ID, declared ...id.ID) map[id.ID]game.Event {
	found := make(map[id.ID]game.Event)
	for _, event := range events {
		if event.Kind != game.EventAttackerBecameBlocked || event.PermanentID != attacker {
			continue
		}
		if slices.Contains(declared, event.RelatedPermanentID) {
			continue
		}
		found[event.RelatedPermanentID] = event
	}
	return found
}

func attackPlayer2(attacker *game.Permanent) action.DeclareAttackersAction {
	attacker.SummoningSick = false
	return action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{attackAt(attacker, game.AttackTarget{Player: game.Player2})}}
}

func tokensControlledBy(g *game.Game, player game.PlayerID) []*game.Permanent {
	var tokens []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.Controller == player {
			tokens = append(tokens, permanent)
		}
	}
	return tokens
}

func onlyEvent(t *testing.T, events map[id.ID]game.Event) (id.ID, game.Event) {
	t.Helper()
	if len(events) != 1 {
		t.Fatalf("effect-added blocking pairs = %d (%+v), want 1", len(events), events)
	}
	for blocker, event := range events {
		return blocker, event
	}
	return 0, game.Event{}
}

func noBlockerDeclared(t *testing.T, events []game.Event, blockers ...id.ID) {
	t.Helper()
	for _, event := range events {
		if event.Kind == game.EventBlockerDeclared && slices.Contains(blockers, event.PermanentID) {
			t.Fatalf("creature put onto the battlefield blocking emitted %+v; it never blocked (CR 509.4)", event)
		}
	}
}

// TestTokenEntersBlockingUnblockedAttacker resolves Flash Foliage's shape
// against an unblocked attacker during the declare blockers step. The token is
// blocking but never blocked (CR 509.4), so the attacker becomes blocked
// (CR 509.3c), its per-pair trigger sees the token (CR 509.3d), and it deals no
// damage to the defending player.
func TestTokenEntersBlockingUnblockedAttacker(t *testing.T) {
	t.Run("bushido becomes blocked", func(t *testing.T) {
		g := bandCombatGame()
		konda := registryCombatPermanent(t, g, game.Player1, "Konda, Lord of Eiganjo")
		spell := addCardToHand(g, game.Player2, saprolingBlockSpell())
		life := g.Players[game.Player2].Life
		defender := &bandBlockingAgent{cast: spell, castStep: game.StepDeclareBlockers, target: konda.ObjectID}
		policies := simPassPolicies()
		policies[game.Player2] = defender

		after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, attackPlayer2(konda), policies)
		if !ok || !defender.casted {
			t.Fatalf("combat accepted = %v, spell cast = %v", ok, defender.casted)
		}
		token, event := onlyEvent(t, becameBlockedByOthers(after.Events, konda.ObjectID))
		if event.AttackerAlreadyBlocked {
			t.Fatalf("Konda's block by the token = %+v, want a newly blocked attacker", event)
		}
		noBlockerDeclared(t, after.Events, token)
		kondaAfter, _ := permanentByObjectID(after, konda.ObjectID)
		if got := effectivePower(after, kondaAfter); got != 8 {
			t.Fatalf("Konda power = %d, want 8 after bushido from becoming blocked", got)
		}
		if _, alive := permanentByObjectID(after, token); alive {
			t.Fatal("Saproling survived blocking an 8/8")
		}
		if after.Players[game.Player2].Life != life {
			t.Fatalf("defender life = %d, want %d: Konda was blocked", after.Players[game.Player2].Life, life)
		}
	})
	t.Run("flanking per pair keeps attacker blocked", func(t *testing.T) {
		g := bandCombatGame()
		knight := registryCombatPermanent(t, g, game.Player1, "Zhalfirin Knight")
		spell := addCardToHand(g, game.Player2, saprolingBlockSpell())
		life := g.Players[game.Player2].Life
		defender := &bandBlockingAgent{cast: spell, castStep: game.StepDeclareBlockers, target: knight.ObjectID}
		policies := simPassPolicies()
		policies[game.Player2] = defender

		after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, attackPlayer2(knight), policies)
		if !ok || !defender.casted {
			t.Fatalf("combat accepted = %v, spell cast = %v", ok, defender.casted)
		}
		onlyEvent(t, becameBlockedByOthers(after.Events, knight.ObjectID))
		if tokens := tokensControlledBy(after, game.Player2); len(tokens) != 0 {
			t.Fatalf("Saproling survived flanking: %+v", tokens)
		}
		if after.Players[game.Player2].Life != life {
			t.Fatalf("defender life = %d, want %d: the knight stays blocked after flanking kills its blocker (CR 509.1h)", after.Players[game.Player2].Life, life)
		}
		if got := markedDamage(t, after, knight); got != 0 {
			t.Fatalf("knight damage = %d, want 0", got)
		}
	})
}

// TestBlockTriggerTokenJoinsExistingBlock plays Brimaz's shape: its declared
// block triggers a token entering blocking the same attacker. The attacker is
// already blocked, so a once-only "becomes blocked" trigger doesn't fire again
// (CR 509.3c), while a per-pair "becomes blocked by a creature" trigger fires
// for the token (CR 509.3d).
func TestBlockTriggerTokenJoinsExistingBlock(t *testing.T) {
	t.Run("per pair trigger sees the token", func(t *testing.T) {
		g := bandCombatGame()
		hound := registryCombatPermanent(t, g, game.Player1, "Ashmouth Hound")
		brimaz := addCombatPermanent(g, game.Player2, brimazShapeCreature())
		defender := &bandBlockingAgent{blocks: []game.BlockDeclaration{{Blocker: brimaz.ObjectID, Blocking: hound.ObjectID}}}
		policies := simPassPolicies()
		policies[game.Player2] = defender

		after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, attackPlayer2(hound), policies)
		if !ok {
			t.Fatal("combat not accepted")
		}
		token, event := onlyEvent(t, becameBlockedByOthers(after.Events, hound.ObjectID, brimaz.ObjectID))
		if !event.AttackerAlreadyBlocked {
			t.Fatalf("token block = %+v, want the attacker already blocked", event)
		}
		noBlockerDeclared(t, after.Events, token)
		if _, alive := permanentByObjectID(after, token); alive {
			t.Fatal("Cat Soldier survived Ashmouth Hound's becomes-blocked-by trigger")
		}
		if got := markedDamage(t, after, brimaz); got != 3 {
			t.Fatalf("Brimaz damage = %d, want 3 (1 from the hound's trigger, 2 combat)", got)
		}
	})
	t.Run("once only becomes blocked trigger", func(t *testing.T) {
		g := bandCombatGame()
		konda := registryCombatPermanent(t, g, game.Player1, "Konda, Lord of Eiganjo")
		brimaz := addCombatPermanent(g, game.Player2, brimazShapeCreature())
		defender := &bandBlockingAgent{blocks: []game.BlockDeclaration{{Blocker: brimaz.ObjectID, Blocking: konda.ObjectID}}}
		policies := simPassPolicies()
		policies[game.Player2] = defender

		after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, attackPlayer2(konda), policies)
		if !ok {
			t.Fatal("combat not accepted")
		}
		onlyEvent(t, becameBlockedByOthers(after.Events, konda.ObjectID, brimaz.ObjectID))
		kondaAfter, _ := permanentByObjectID(after, konda.ObjectID)
		if got := effectivePower(after, kondaAfter); got != 8 {
			t.Fatalf("Konda power = %d, want 8: bushido triggers once when it becomes blocked", got)
		}
	})
}

// TestCreatureFromHandEntersBlocking plays Aetherplasm's shape: Ashmouth Hound
// is put from hand onto the battlefield blocking Cave Tiger. The Tiger's
// becomes-blocked-by trigger fires (CR 509.3d) but the Hound's "blocks" half
// doesn't, because the Hound never blocked (CR 509.3b, 509.4). The pumped 3/3
// Tiger survives the Hound's 2 combat damage.
func TestCreatureFromHandEntersBlocking(t *testing.T) {
	g := bandCombatGame()
	tiger := registryCombatPermanent(t, g, game.Player1, "Cave Tiger")
	houndCard := addCardToHand(g, game.Player2, registryCardDef(t, "Ashmouth Hound"))
	spell := addCardToHand(g, game.Player2, putBlockingSpell())
	life := g.Players[game.Player2].Life
	defender := &bandBlockingAgent{cast: spell, castStep: game.StepDeclareBlockers, target: tiger.ObjectID}
	policies := simPassPolicies()
	policies[game.Player2] = defender

	after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, attackPlayer2(tiger), policies)
	if !ok || !defender.casted {
		t.Fatalf("combat accepted = %v, spell cast = %v", ok, defender.casted)
	}
	hound, event := onlyEvent(t, becameBlockedByOthers(after.Events, tiger.ObjectID))
	if event.AttackerAlreadyBlocked {
		t.Fatalf("Tiger's block = %+v, want a newly blocked attacker", event)
	}
	if !after.Players[game.Player2].Graveyard.Contains(houndCard) {
		t.Fatal("Ashmouth Hound was not put from hand onto the battlefield blocking and killed")
	}
	noBlockerDeclared(t, after.Events, hound)
	tigerAfter, ok := permanentByObjectID(after, tiger.ObjectID)
	if !ok {
		t.Fatal("Cave Tiger died: the entered Hound's blocks trigger fired")
	}
	if got, _ := effectiveToughness(after, tigerAfter); got != 3 {
		t.Fatalf("Cave Tiger toughness = %d, want 3 after its becomes-blocked-by trigger", got)
	}
	if got := tigerAfter.MarkedDamage; got != 2 {
		t.Fatalf("Cave Tiger damage = %d, want 2 combat damage only", got)
	}
	if _, alive := permanentByObjectID(after, hound); alive {
		t.Fatal("Ashmouth Hound survived the 3/3 Tiger")
	}
	if after.Players[game.Player2].Life != life {
		t.Fatalf("defender life = %d, want %d", after.Players[game.Player2].Life, life)
	}
}

// TestCopyTokensEnterBlockingTheirSources plays Mirror Match's shape in a
// multiplayer combat. Each copy blocks the creature it copies only when its
// controller is that attacker's defending player (CR 509.4a): the copy of the
// attacker of Player2 blocks it, and the copy of the attacker of Player3 enters
// without blocking. The blocking copy of Ashmouth Hound never blocked, so only
// the attacker's per-pair trigger fires; the copy dies and the attacker stays
// blocked (CR 509.1h).
func TestCopyTokensEnterBlockingTheirSources(t *testing.T) {
	g := bandCombatGame()
	hound := registryCombatPermanent(t, g, game.Player1, "Ashmouth Hound")
	tiger := registryCombatPermanent(t, g, game.Player1, "Cave Tiger")
	spell := addCardToHand(g, game.Player2, mirrorBlockSpell())
	life2, life3 := g.Players[game.Player2].Life, g.Players[game.Player3].Life
	defender := &bandBlockingAgent{cast: spell, castStep: game.StepDeclareBlockers}
	policies := simPassPolicies()
	policies[game.Player2] = defender
	hound.SummoningSick, tiger.SummoningSick = false, false
	attack := action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{
		attackAt(hound, game.AttackTarget{Player: game.Player2}),
		attackAt(tiger, game.AttackTarget{Player: game.Player3}),
	}}

	after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, attack, policies)
	if !ok || !defender.casted {
		t.Fatalf("combat accepted = %v, spell cast = %v", ok, defender.casted)
	}
	copyHound, event := onlyEvent(t, becameBlockedByOthers(after.Events, hound.ObjectID))
	if event.AttackerAlreadyBlocked {
		t.Fatalf("hound's block = %+v, want a newly blocked attacker", event)
	}
	if blocks := becameBlockedByOthers(after.Events, tiger.ObjectID); len(blocks) != 0 {
		t.Fatalf("Player2's Tiger copy blocked an attacker of Player3: %+v", blocks)
	}
	noBlockerDeclared(t, after.Events, copyHound)
	if _, alive := permanentByObjectID(after, copyHound); alive {
		t.Fatal("Ashmouth Hound copy survived the attacker's becomes-blocked-by trigger")
	}
	if got := markedDamage(t, after, hound); got != 0 {
		t.Fatalf("attacking hound damage = %d, want 0: its copy never blocked", got)
	}
	tokens := tokensControlledBy(after, game.Player2)
	if len(tokens) != 1 || isBlockingCreature(after, tokens[0].ObjectID) {
		t.Fatalf("Player2 tokens = %+v, want only the non-blocking Tiger copy", tokens)
	}
	if after.Players[game.Player2].Life != life2 {
		t.Fatalf("Player2 life = %d, want %d: the hound stays blocked", after.Players[game.Player2].Life, life2)
	}
	if got, want := after.Players[game.Player3].Life, life3-2; got != want {
		t.Fatalf("Player3 life = %d, want %d from the unblocked Tiger", got, want)
	}
}

// TestTokenEntersBlockingBand targets one member of an attacking band; the
// token blocks every member (CR 702.22h-i), and Ashmouth Hound's per-pair
// trigger fires for its own pair (CR 509.3d).
func TestTokenEntersBlockingBand(t *testing.T) {
	g := bandCombatGame()
	hero := registryCombatPermanent(t, g, game.Player1, "Benalish Hero")
	hound := registryCombatPermanent(t, g, game.Player1, "Ashmouth Hound")
	spell := addCardToHand(g, game.Player2, saprolingBlockSpell())
	life := g.Players[game.Player2].Life
	defender := &bandBlockingAgent{cast: spell, castStep: game.StepDeclareBlockers, target: hero.ObjectID}
	policies := simPassPolicies()
	policies[game.Player1] = &bandingDivisionAgent{}
	policies[game.Player2] = defender

	after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, bandAttack(hero, hound), policies)
	if !ok || !defender.casted {
		t.Fatalf("combat accepted = %v, spell cast = %v", ok, defender.casted)
	}
	token, _ := onlyEvent(t, becameBlockedByOthers(after.Events, hero.ObjectID))
	houndToken, event := onlyEvent(t, becameBlockedByOthers(after.Events, hound.ObjectID))
	if houndToken != token || event.AttackerAlreadyBlocked {
		t.Fatalf("hound's block = %+v, want the Saproling newly blocking it", event)
	}
	if _, alive := permanentByObjectID(after, token); alive {
		t.Fatal("Saproling survived Ashmouth Hound's becomes-blocked-by trigger")
	}
	if after.Players[game.Player2].Life != life {
		t.Fatalf("defender life = %d, want %d: the whole band was blocked", after.Players[game.Player2].Life, life)
	}
}
