package rules

import (
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/s"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// copyTargetHost is a distinctively named 2/2 creature so the landfall copy
// token (a copy of the enchanted creature) is unambiguous to identify and its
// copiable 2/2 differs from the 3/3 it becomes while Springheart's +1/+1 grant
// applies.
func copyTargetHost(g *game.Game) *game.Permanent {
	return addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Copy Target Beast",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
}

// bestowSpringheartOnto casts the real generated Springheart Nantuko for its
// bestow cost onto host and resolves it, returning the attached Aura permanent.
func bestowSpringheartOnto(t *testing.T, g *game.Game, engine *Engine, host *game.Permanent) *game.Permanent {
	t.Helper()
	spellID := addCardToHand(g, game.Player1, cards.SpringheartNantuko())
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	targets := []game.Target{game.PermanentTarget(host.ObjectID)}
	if !engine.applyAction(g, game.Player1, action.CastBestowSpell(spellID, targets, 0, nil)) {
		t.Fatal("bestowed cast of Springheart Nantuko failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	springheart, ok := findPermanentByCardID(g, spellID)
	if !ok {
		t.Fatal("Springheart Nantuko did not enter the battlefield")
	}
	if !springheart.Bestowed || !springheart.AttachedTo.Exists || springheart.AttachedTo.Val != host.ObjectID {
		t.Fatalf("Springheart did not bestow onto host: bestowed=%v attached=%v", springheart.Bestowed, springheart.AttachedTo)
	}
	return springheart
}

// pushSpringheartLandfall puts the real card's landfall triggered ability on the
// stack with a land-entered trigger event, as if a land the controller controls
// had just entered.
func pushSpringheartLandfall(g *game.Game, springheart *game.Permanent) {
	inst, ok := g.GetCardInstance(springheart.CardInstanceID)
	if !ok {
		panic("springheart card instance not found")
	}
	land := addBasicLandPermanent(g, game.Player1, types.Forest)
	ability := &inst.Def.TriggeredAbilities[0]
	g.Stack.Push(&game.StackObject{
		ID:              g.IDGen.Next(),
		Kind:            game.StackTriggeredAbility,
		SourceID:        springheart.ObjectID,
		SourceCardID:    springheart.CardInstanceID,
		Controller:      springheart.Controller,
		InlineTrigger:   ability,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:        game.EventPermanentEnteredBattlefield,
			PermanentID: land.ObjectID,
			Controller:  game.Player1,
		},
	})
}

// springheartTokenCounts tallies the tokens Player1 controls by name after a
// landfall resolution.
func springheartTokenCounts(g *game.Game) (copies, insects int, copyToken *game.Permanent) {
	for _, permanent := range g.Battlefield {
		if !permanent.Token || permanent.TokenDef == nil {
			continue
		}
		switch permanent.TokenDef.Name {
		case "Copy Target Beast":
			copies++
			copyToken = permanent
		case "Insect":
			insects++
		default:
			// Ignore any unrelated tokens on the battlefield.
		}
	}
	return copies, insects, copyToken
}

// TestSpringheartNantukoBestowGrantsAndFallsOff exercises the real generated
// card's Bestow keyword and +1/+1 grant end to end: cast bestowed it attaches as
// an Aura, stops being a creature, and grants the enchanted creature +1/+1; when
// the enchanted creature leaves, it becomes an unattached creature (CR 702.103f).
func TestSpringheartNantukoBestowGrantsAndFallsOff(t *testing.T) {
	g, engine := setupBestowMain(t)
	host := copyTargetHost(g)
	springheart := bestowSpringheartOnto(t, g, engine, host)

	if permanentHasType(g, springheart, types.Creature) {
		t.Fatal("bestowed Springheart is still a creature, want Aura only")
	}
	if got := effectivePower(g, host); got != 3 {
		t.Fatalf("enchanted host power = %d, want 3 (2/2 + 1/1 grant)", got)
	}
	if got, _ := effectiveToughness(g, host); got != 3 {
		t.Fatalf("enchanted host toughness = %d, want 3", got)
	}

	movePermanentToZone(g, host, zone.Graveyard)
	engine.applyStateBasedActionsWithDeaths(g)

	fallen, ok := findPermanentByCardID(g, springheart.CardInstanceID)
	if !ok {
		t.Fatal("Springheart left the battlefield after its host died, want it to survive as a creature")
	}
	if fallen.Bestowed || fallen.AttachedTo.Exists {
		t.Fatalf("Springheart still bestowed/attached after host left: %#v", fallen)
	}
	if !permanentHasType(g, fallen, types.Creature) {
		t.Fatal("fallen-off Springheart is not a creature")
	}
}

// TestSpringheartNantukoLandfallPaidCopiesEnchantedCreature proves the paid path:
// while attached, paying {1}{G} arms the reflexive trigger that creates a token
// copy of the enchanted creature. The copy uses copiable values (a 2/2, not the
// 3/3 the host is while Springheart's grant applies), is owned and controlled by
// the ability's controller, and no Insect is created.
func TestSpringheartNantukoLandfallPaidCopiesEnchantedCreature(t *testing.T) {
	g, engine := setupBestowMain(t)
	host := copyTargetHost(g)
	springheart := bestowSpringheartOnto(t, g, engine, host)

	// Fund the {1}{G} landfall payment with two Forests.
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	pushSpringheartLandfall(g, springheart)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}
	resolveStackWithTriggers(engine, g, agents)

	copies, insects, copyToken := springheartTokenCounts(g)
	if copies != 1 || insects != 0 {
		t.Fatalf("paid landfall tokens: copies=%d insects=%d, want copies=1 insects=0", copies, insects)
	}
	if copyToken.Owner != game.Player1 || copyToken.Controller != game.Player1 {
		t.Fatalf("copy token owner=%v controller=%v, want Player1/Player1", copyToken.Owner, copyToken.Controller)
	}
	if got := effectivePower(g, copyToken); got != 2 {
		t.Fatalf("copy token power = %d, want 2 (copiable value, not the +1/+1-buffed 3)", got)
	}
	if len(g.PendingReflexiveTriggers) != 0 {
		t.Fatalf("pending reflexive triggers = %d, want 0 (drained)", len(g.PendingReflexiveTriggers))
	}
}

// TestSpringheartNantukoLandfallDeclinedCreatesInsect proves the declined path:
// while attached, declining the optional payment creates exactly one 1/1 green
// Insect and no copy.
func TestSpringheartNantukoLandfallDeclinedCreatesInsect(t *testing.T) {
	g, engine := setupBestowMain(t)
	host := copyTargetHost(g)
	springheart := bestowSpringheartOnto(t, g, engine, host)

	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	pushSpringheartLandfall(g, springheart)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: false}}
	resolveStackWithTriggers(engine, g, agents)

	copies, insects, _ := springheartTokenCounts(g)
	if copies != 0 || insects != 1 {
		t.Fatalf("declined landfall tokens: copies=%d insects=%d, want copies=0 insects=1", copies, insects)
	}
}

// TestSpringheartNantukoLandfallUnattachedCreatesInsect proves the unattached
// path: cast as an ordinary creature (not bestowed), landfall offers no payment
// because Springheart is not attached to a creature its controller controls, so
// it creates exactly one Insect.
func TestSpringheartNantukoLandfallUnattachedCreatesInsect(t *testing.T) {
	g, engine := setupBestowMain(t)
	springheart := addCombatPermanent(g, game.Player1, cards.SpringheartNantuko())

	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	pushSpringheartLandfall(g, springheart)

	// Even an agent willing to pay is never offered the payment when unattached.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}
	resolveStackWithTriggers(engine, g, agents)

	copies, insects, _ := springheartTokenCounts(g)
	if copies != 0 || insects != 1 {
		t.Fatalf("unattached landfall tokens: copies=%d insects=%d, want copies=0 insects=1", copies, insects)
	}
}

// TestSpringheartNantukoLandfallCopyOfLegendaryObeysLegendRule proves legendary
// handling: the copy faithfully copies the enchanted creature's Legendary
// supertype (a copiable value), so when the controller then owns both the
// original legendary creature and its freshly created copy, the legend-rule
// state-based action (CR 704.5j) culls the duplicate down to one.
func TestSpringheartNantukoLandfallCopyOfLegendaryObeysLegendRule(t *testing.T) {
	g, engine := setupBestowMain(t)
	host := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:       "Legend Beast",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Power:      opt.Val(game.PT{Value: 2}),
		Toughness:  opt.Val(game.PT{Value: 2}),
	}})
	springheart := bestowSpringheartOnto(t, g, engine, host)

	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	pushSpringheartLandfall(g, springheart)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}
	resolveStackWithTriggers(engine, g, agents)

	// The copy is a legendary token bearing the copied name.
	var copyToken *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.TokenDef != nil && permanent.TokenDef.Name == "Legend Beast" {
			copyToken = permanent
		}
	}
	if copyToken == nil {
		t.Fatal("landfall copy of the legendary host was not created")
	}
	if !permanentHasSupertype(g, copyToken, types.Legendary) {
		t.Fatal("copy of a legendary creature is not itself legendary")
	}

	// Two "Legend Beast" legendaries under one controller trip the legend rule.
	engine.applyStateBasedActionsWithDeaths(g)
	remaining := 0
	for _, permanent := range g.Battlefield {
		if permanentEffectiveName(g, permanent) == "Legend Beast" && permanent.Controller == game.Player1 {
			remaining++
		}
	}
	if remaining != 1 {
		t.Fatalf("Legend Beast permanents after legend rule = %d, want 1", remaining)
	}
}

// payoff is evaluated at reflexive resolution: paying while attached arms the
// reflexive trigger, but if Springheart detaches before that trigger resolves,
// the copy of the (now absent) enchanted creature cannot be made, so the
// fixed-Insect fallback fires instead. Exactly one Insect and no copy result.
func TestSpringheartNantukoLandfallPaidButDetachedCreatesInsect(t *testing.T) {
	g, engine := setupBestowMain(t)
	host := copyTargetHost(g)
	springheart := bestowSpringheartOnto(t, g, engine, host)

	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	pushSpringheartLandfall(g, springheart)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}
	log := &TurnLog{}
	// Resolve the outer ability (pays, arms the reflexive trigger) and put the
	// reflexive trigger on the stack, then detach Springheart before it resolves.
	engine.resolveTopOfStackWithChoices(g, agents, log)
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log)
	detachPermanent(g, springheart)
	// Drain the reflexive trigger (and any follow-ups).
	for {
		if _, ok := g.Stack.Peek(); !ok {
			break
		}
		engine.resolveTopOfStackWithChoices(g, agents, log)
		engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log)
	}

	copies, insects, _ := springheartTokenCounts(g)
	if copies != 0 || insects != 1 {
		t.Fatalf("paid-but-detached landfall tokens: copies=%d insects=%d, want copies=0 insects=1", copies, insects)
	}
}
