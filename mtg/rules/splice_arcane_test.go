package rules

import (
	"strings"
	"testing"

	cardd "github.com/natefinch/council4/mtg/cards/d"
	cardg "github.com/natefinch/council4/mtg/cards/g"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// spliceScriptAgent scripts the Splice onto Arcane choices during a cast: which
// eligible splice card (by name) to select at each ChoiceZoneSelection prompt in
// order (an empty name, or running past the end of the list, declines), which
// target to announce at each ChoiceTarget prompt in order, and it always takes
// the first offered payment option. It leaves every other choice at the engine's
// presented order so the surrounding cast is otherwise unscripted.
type spliceScriptAgent struct {
	splices   []string
	targets   []game.Target
	spliceIdx int
	targetIdx int
}

func (*spliceScriptAgent) ChooseAction(_ PlayerObservation, _ []action.Action) action.Action {
	return action.Pass()
}

func (a *spliceScriptAgent) ChooseChoice(_ PlayerObservation, req game.ChoiceRequest) []int {
	switch req.Kind {
	case game.ChoiceZoneSelection:
		if a.spliceIdx >= len(a.splices) {
			return nil
		}
		want := a.splices[a.spliceIdx]
		a.spliceIdx++
		if want == "" {
			return nil
		}
		for _, o := range req.Options {
			if strings.Contains(o.Label, want) {
				return []int{o.Index}
			}
		}
		return nil
	case game.ChoiceTarget:
		if a.targetIdx < len(a.targets) {
			want := a.targets[a.targetIdx]
			a.targetIdx++
			for _, o := range req.Options {
				if len(o.Targets) == 1 && spliceTargetsEqual(o.Targets[0], want) {
					return []int{o.Index}
				}
			}
		}
		if len(req.Options) > 0 {
			return []int{req.Options[0].Index}
		}
		return nil
	case game.ChoicePayment:
		if len(req.Options) > 0 {
			return []int{req.Options[0].Index}
		}
		return nil
	default:
		idx := make([]int, len(req.Options))
		for i, o := range req.Options {
			idx[i] = o.Index
		}
		return idx
	}
}

func spliceTargetsEqual(a, b game.Target) bool {
	return a.Kind == b.Kind && a.PermanentID == b.PermanentID && a.PlayerID == b.PlayerID
}

// addToughCreature puts a creature with the given toughness onto controller's
// battlefield so splice damage can be marked without a state-based sacrifice.
func addToughCreature(g *game.Game, controller game.PlayerID, toughness int) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:      "Splice Target Creature",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: toughness}),
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

// grantShroud gives an existing permanent shroud after a spell has been put on
// the stack so it becomes an illegal target at resolution (CR 702.18), exercising
// the resolution-time target recheck (CR 608.2b) for host and spliced targets.
func grantShroud(g *game.Game, permanent *game.Permanent) {
	card := g.CardInstances[permanent.CardInstanceID]
	card.Def.StaticAbilities = append(card.Def.StaticAbilities, game.ShroudStaticBody)
}

func newSpliceGame() (*game.Game, *Engine) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	return g, NewEngine(nil)
}

// TestSpliceDesperateRitualOntoRealArcaneHost is the core end-to-end case: a real
// Arcane host (Desperate Ritual) is cast from hand and a second Desperate Ritual
// is spliced onto it. The splice source stays in hand (only revealed), its mana
// splice cost is paid as an additional cost, and both the host's and the spliced
// card's Add {R}{R}{R} effects resolve — six red mana in the caster's pool.
func TestSpliceDesperateRitualOntoRealArcaneHost(t *testing.T) {
	g, engine := newSpliceGame()
	hostID := addCardToHand(g, game.Player1, cardd.DesperateRitual())
	spliceID := addCardToHand(g, game.Player1, cardd.DesperateRitual())
	// Host {1}{R} + splice {1}{R} = two generic + two red, paid exactly.
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &spliceScriptAgent{splices: []string{"Desperate Ritual"}}}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(hostID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("casting Desperate Ritual with a spliced Desperate Ritual failed")
	}

	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("no spell on the stack after cast")
	}
	if len(obj.SplicedContent) != 1 {
		t.Fatalf("SplicedContent length = %d, want 1", len(obj.SplicedContent))
	}
	if !g.Players[game.Player1].Hand.Contains(spliceID) {
		t.Fatal("spliced card left the hand; splice reveals but does not move the card")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("mana pool total after paying host + splice = %d, want 0 (both costs paid)", g.Players[game.Player1].ManaPool.Total())
	}

	resolveStackWithTriggers(engine, g, agents)

	if got := g.Players[game.Player1].ManaPool.Amount(mana.R); got != 6 {
		t.Fatalf("red mana after resolving host + spliced Add {R}{R}{R} = %d, want 6", got)
	}
	if !g.Players[game.Player1].Graveyard.Contains(hostID) {
		t.Fatal("host spell did not resolve to the graveyard")
	}
	if !g.Players[game.Player1].Hand.Contains(spliceID) {
		t.Fatal("spliced card must remain in hand after resolution")
	}
}

// TestSpliceDeclineLeavesHostAlone confirms that declining the splice offer (an
// empty ChoiceZoneSelection answer, which the generic simulation strategy also
// gives) casts the host unchanged: no splice content, only the host effect.
func TestSpliceDeclineLeavesHostAlone(t *testing.T) {
	g, engine := newSpliceGame()
	hostID := addCardToHand(g, game.Player1, cardd.DesperateRitual())
	spliceID := addCardToHand(g, game.Player1, cardd.DesperateRitual())
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &spliceScriptAgent{splices: []string{""}}}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(hostID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("casting Desperate Ritual failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || len(obj.SplicedContent) != 0 {
		t.Fatalf("SplicedContent = %v, want none after declining", obj.SplicedContent)
	}
	resolveStackWithTriggers(engine, g, agents)
	if got := g.Players[game.Player1].ManaPool.Amount(mana.R); got != 3 {
		t.Fatalf("red mana after resolving host only = %d, want 3", got)
	}
	if !g.Players[game.Player1].Hand.Contains(spliceID) {
		t.Fatal("declined splice card must remain in hand")
	}
}

// TestSpliceUnaffordableNotOffered checks the fail-closed affordability gate: when
// the caster cannot pay the host cost plus the splice cost, the splice card is not
// an eligible candidate, so no splice choice is even presented.
func TestSpliceUnaffordableNotOffered(t *testing.T) {
	g, engine := newSpliceGame()
	hostID := addCardToHand(g, game.Player1, cardd.DesperateRitual())
	addCardToHand(g, game.Player1, cardd.DesperateRitual())
	// Only enough for the host {1}{R}; the splice {1}{R} is unaffordable.
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &spliceScriptAgent{splices: []string{"Desperate Ritual"}}}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(hostID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("casting Desperate Ritual failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || len(obj.SplicedContent) != 0 {
		t.Fatalf("SplicedContent = %v, want none (splice unaffordable)", obj.SplicedContent)
	}
}

// nonArcaneInstant is a red instant with no Arcane subtype and a simple Add {G}
// effect, used as a host onto which splice cards may never be attached.
func nonArcaneInstant() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Plain Bolt",
		ManaCost: opt.Val(cost.Mana{cost.R}),
		Types:    []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Sequence: []game.Instruction{
				{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.G}},
			},
		}.Ability()),
	}}
}

// TestSpliceNonArcaneHostNotOffered confirms the Arcane gate: even with an
// affordable splice card in hand, a non-Arcane host is never spliceable.
func TestSpliceNonArcaneHostNotOffered(t *testing.T) {
	g, engine := newSpliceGame()
	hostID := addCardToHand(g, game.Player1, nonArcaneInstant())
	addCardToHand(g, game.Player1, cardd.DesperateRitual())
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &spliceScriptAgent{splices: []string{"Desperate Ritual"}}}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(hostID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("casting the non-Arcane host failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || len(obj.SplicedContent) != 0 {
		t.Fatalf("SplicedContent = %v, want none (host is not Arcane)", obj.SplicedContent)
	}
}

// TestSpliceTargetedGlacialRayAnnouncesTarget covers a targeted splice: Glacial
// Ray spliced onto an Arcane host announces its own target while casting, and its
// two damage resolves against that target.
func TestSpliceTargetedGlacialRayAnnouncesTarget(t *testing.T) {
	g, engine := newSpliceGame()
	hostID := addCardToHand(g, game.Player1, cardd.DesperateRitual())
	spliceID := addCardToHand(g, game.Player1, cardg.GlacialRay())
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)

	target := game.PlayerTarget(game.Player2)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &spliceScriptAgent{
		splices: []string{"Glacial Ray"},
		targets: []game.Target{target},
	}}
	lifeBefore := g.Players[game.Player2].Life
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(hostID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("casting host with a spliced Glacial Ray failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || len(obj.SplicedContent) != 1 {
		t.Fatalf("SplicedContent length = %d, want 1", len(obj.SplicedContent))
	}
	if len(obj.SplicedTargets) != 1 || len(obj.SplicedTargets[0]) != 1 ||
		!spliceTargetsEqual(obj.SplicedTargets[0][0], target) {
		t.Fatalf("SplicedTargets = %v, want the announced Player2 target", obj.SplicedTargets)
	}
	if !g.Players[game.Player1].Hand.Contains(spliceID) {
		t.Fatal("Glacial Ray left the hand; splice reveals but does not move the card")
	}

	resolveStackWithTriggers(engine, g, agents)

	if got := lifeBefore - g.Players[game.Player2].Life; got != 2 {
		t.Fatalf("Player2 lost %d life, want 2 from the spliced Glacial Ray", got)
	}
}

// TestSpliceMultipleAppliedInChoiceOrder splices two Glacial Rays onto one host
// and confirms both are recorded, each with its own announced target, in the
// order the caster chose them.
func TestSpliceMultipleAppliedInChoiceOrder(t *testing.T) {
	g, engine := newSpliceGame()
	hostID := addCardToHand(g, game.Player1, cardd.DesperateRitual())
	addCardToHand(g, game.Player1, cardg.GlacialRay())
	addCardToHand(g, game.Player1, cardg.GlacialRay())
	victim := addToughCreature(g, game.Player2, 4)
	// Host {1}{R} + two splices {1}{R} each = three generic + three red.
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Players[game.Player1].ManaPool.Add(mana.R, 3)

	firstTarget := game.PlayerTarget(game.Player2)
	secondTarget := game.PermanentTarget(victim.ObjectID)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &spliceScriptAgent{
		splices: []string{"Glacial Ray", "Glacial Ray"},
		targets: []game.Target{firstTarget, secondTarget},
	}}
	lifeBefore := g.Players[game.Player2].Life
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(hostID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("casting host with two spliced Glacial Rays failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || len(obj.SplicedContent) != 2 {
		t.Fatalf("SplicedContent length = %d, want 2", len(obj.SplicedContent))
	}
	if len(obj.SplicedTargets) != 2 ||
		!spliceTargetsEqual(obj.SplicedTargets[0][0], firstTarget) ||
		!spliceTargetsEqual(obj.SplicedTargets[1][0], secondTarget) {
		t.Fatalf("SplicedTargets = %v, want [Player2, victim] in choice order", obj.SplicedTargets)
	}

	resolveStackWithTriggers(engine, g, agents)

	if got := lifeBefore - g.Players[game.Player2].Life; got != 2 {
		t.Fatalf("Player2 lost %d life, want 2 from the first spliced Glacial Ray", got)
	}
	if victim.MarkedDamage != 2 {
		t.Fatalf("victim marked damage = %d, want 2 from the second spliced Glacial Ray", victim.MarkedDamage)
	}
}

// TestSpliceCopiedSpellCopiesSplicedText confirms that a copy of a spell with
// spliced-on content carries that content (CR 707.2/706): copies copy the spliced
// text so the copy resolves the same combined effects.
func TestSpliceCopiedSpellCopiesSplicedText(t *testing.T) {
	g, engine := newSpliceGame()
	hostID := addCardToHand(g, game.Player1, cardd.DesperateRitual())
	addCardToHand(g, game.Player1, cardd.DesperateRitual())
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &spliceScriptAgent{splices: []string{"Desperate Ritual"}}}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(hostID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("casting host with a spliced Desperate Ritual failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || len(obj.SplicedContent) != 1 {
		t.Fatalf("host SplicedContent length = %d, want 1", len(obj.SplicedContent))
	}

	depthBefore := g.Stack.Size()
	addEffectSpellToStack(g, game.Player1,
		game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(obj.ID)})
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Stack.Size(); got != depthBefore+1 {
		t.Fatalf("stack size after copy = %d, want %d", got, depthBefore+1)
	}
	top, ok := g.Stack.Peek()
	if !ok || !top.Copy {
		t.Fatal("copy of the spliced spell not on top of the stack")
	}
	if top.ID == obj.ID {
		t.Fatal("copy shares the original's ID")
	}
	if len(top.SplicedContent) != 1 {
		t.Fatalf("copy SplicedContent length = %d, want 1 (copy carries spliced text)", len(top.SplicedContent))
	}
}

// TestSpliceResolvesSplicedOnlyWhenHostTargetIllegal covers the combined
// resolution recheck (CR 608.2b, CR 702.47): a targeted Arcane host (Glacial Ray)
// whose creature target becomes illegal still resolves for its legal spliced
// Glacial Ray, which deals its damage while the host's own damage is skipped.
func TestSpliceResolvesSplicedOnlyWhenHostTargetIllegal(t *testing.T) {
	g, engine := newSpliceGame()
	hostID := addCardToHand(g, game.Player1, cardg.GlacialRay())
	addCardToHand(g, game.Player1, cardg.GlacialRay())
	hostVictim := addToughCreature(g, game.Player2, 4)
	// Host {1}{R} + splice {1}{R} = two generic + two red, paid exactly.
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)

	hostTarget := game.PermanentTarget(hostVictim.ObjectID)
	spliceTarget := game.PlayerTarget(game.Player2)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &spliceScriptAgent{
		splices: []string{"Glacial Ray"},
		targets: []game.Target{spliceTarget},
	}}
	lifeBefore := g.Players[game.Player2].Life
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(hostID, []game.Target{hostTarget}, 0, nil), agents, &TurnLog{}) {
		t.Fatal("casting Glacial Ray host with a spliced Glacial Ray failed")
	}
	// The host's creature target gains shroud after the spell is on the stack, so
	// only the host target is illegal at resolution; the spliced player target
	// stays legal and keeps the whole spell resolving.
	grantShroud(g, hostVictim)

	resolveStackWithTriggers(engine, g, agents)

	if got := lifeBefore - g.Players[game.Player2].Life; got != 2 {
		t.Fatalf("Player2 lost %d life, want 2 from the still-legal spliced Glacial Ray", got)
	}
	if hostVictim.MarkedDamage != 0 {
		t.Fatalf("host creature marked damage = %d, want 0 (illegal host target deferred)", hostVictim.MarkedDamage)
	}
}

// TestSpliceHostResolvesWhenSpliceTargetIllegal is the mirror case: a legal host
// resolves while a spliced Glacial Ray whose target became illegal is skipped, so
// the host damage lands and the spliced damage does not (CR 608.2b, CR 702.47).
func TestSpliceHostResolvesWhenSpliceTargetIllegal(t *testing.T) {
	g, engine := newSpliceGame()
	hostID := addCardToHand(g, game.Player1, cardg.GlacialRay())
	addCardToHand(g, game.Player1, cardg.GlacialRay())
	spliceVictim := addToughCreature(g, game.Player2, 4)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)

	hostTarget := game.PlayerTarget(game.Player2)
	spliceTarget := game.PermanentTarget(spliceVictim.ObjectID)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &spliceScriptAgent{
		splices: []string{"Glacial Ray"},
		targets: []game.Target{spliceTarget},
	}}
	lifeBefore := g.Players[game.Player2].Life
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(hostID, []game.Target{hostTarget}, 0, nil), agents, &TurnLog{}) {
		t.Fatal("casting Glacial Ray host with a spliced Glacial Ray failed")
	}
	grantShroud(g, spliceVictim)

	resolveStackWithTriggers(engine, g, agents)

	if got := lifeBefore - g.Players[game.Player2].Life; got != 2 {
		t.Fatalf("Player2 lost %d life, want 2 from the legal host Glacial Ray", got)
	}
	if spliceVictim.MarkedDamage != 0 {
		t.Fatalf("splice creature marked damage = %d, want 0 (illegal spliced target deferred)", spliceVictim.MarkedDamage)
	}
}

// TestSpliceWholeSpellCounteredWhenOnlySpliceTargetIllegal covers the countering
// case: an untargeted host (Desperate Ritual) whose only target instance is the
// spliced Glacial Ray is countered when that spliced target becomes illegal — the
// host's own Add {R}{R}{R} does not happen and the host goes to the graveyard
// (CR 608.2b: all of the spell's targets are illegal).
func TestSpliceWholeSpellCounteredWhenOnlySpliceTargetIllegal(t *testing.T) {
	g, engine := newSpliceGame()
	hostID := addCardToHand(g, game.Player1, cardd.DesperateRitual())
	addCardToHand(g, game.Player1, cardg.GlacialRay())
	spliceVictim := addToughCreature(g, game.Player2, 4)
	// Host {1}{R} + splice {1}{R} = two generic + two red, paid exactly.
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)

	spliceTarget := game.PermanentTarget(spliceVictim.ObjectID)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &spliceScriptAgent{
		splices: []string{"Glacial Ray"},
		targets: []game.Target{spliceTarget},
	}}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(hostID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("casting untargeted Desperate Ritual with a spliced Glacial Ray failed")
	}
	grantShroud(g, spliceVictim)

	resolveStackWithTriggers(engine, g, agents)

	if got := g.Players[game.Player1].ManaPool.Amount(mana.R); got != 0 {
		t.Fatalf("red mana after resolution = %d, want 0 (whole spell countered, Add {R}{R}{R} skipped)", got)
	}
	if spliceVictim.MarkedDamage != 0 {
		t.Fatalf("splice creature marked damage = %d, want 0 (spell countered)", spliceVictim.MarkedDamage)
	}
	if !g.Players[game.Player1].Graveyard.Contains(hostID) {
		t.Fatal("countered host spell should be put into its owner's graveyard")
	}
}

// TestSpliceCopyRetargetsHostAndSpliceIndependently covers the copy retarget path
// (CR 707.10c, CR 702.47): a copy made with "you may choose new targets" re-chooses
// the host target and each spliced target independently, and the original spell
// keeps its announced targets.
func TestSpliceCopyRetargetsHostAndSpliceIndependently(t *testing.T) {
	g, engine := newSpliceGame()
	hostID := addCardToHand(g, game.Player1, cardg.GlacialRay())
	addCardToHand(g, game.Player1, cardg.GlacialRay())
	spliceVictim := addToughCreature(g, game.Player2, 4)
	newHostVictim := addToughCreature(g, game.Player2, 4)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)

	hostTarget := game.PlayerTarget(game.Player2)
	spliceTarget := game.PermanentTarget(spliceVictim.ObjectID)
	newHostTarget := game.PermanentTarget(newHostVictim.ObjectID)
	newSpliceTarget := game.PlayerTarget(game.Player1)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &spliceScriptAgent{
		splices: []string{"Glacial Ray"},
		// The first ChoiceTarget answers the cast-time splice announcement; the
		// copy's may-choose-new-targets rider then re-chooses the host target and
		// the spliced target in that order.
		targets: []game.Target{spliceTarget, newHostTarget, newSpliceTarget},
	}}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(hostID, []game.Target{hostTarget}, 0, nil), agents, &TurnLog{}) {
		t.Fatal("casting Glacial Ray host with a spliced Glacial Ray failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || len(obj.SplicedContent) != 1 {
		t.Fatalf("host SplicedContent length = %d, want 1", len(obj.SplicedContent))
	}

	addEffectSpellToStack(g, game.Player1,
		game.CopyStackObject{Object: game.TargetStackObjectReference(0), MayChooseNewTargets: true},
		[]game.Target{game.StackObjectTarget(obj.ID)})
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	top, ok := g.Stack.Peek()
	if !ok || !top.Copy {
		t.Fatal("copy of the spliced spell not on top of the stack")
	}
	if len(top.Targets) != 1 || !spliceTargetsEqual(top.Targets[0], newHostTarget) {
		t.Fatalf("copy host target = %v, want retargeted to the new creature", top.Targets)
	}
	if len(top.SplicedTargets) != 1 || len(top.SplicedTargets[0]) != 1 ||
		!spliceTargetsEqual(top.SplicedTargets[0][0], newSpliceTarget) {
		t.Fatalf("copy spliced target = %v, want retargeted to Player1", top.SplicedTargets)
	}
	// Retargeting the copy is independent: the original keeps its announced targets.
	if len(obj.Targets) != 1 || !spliceTargetsEqual(obj.Targets[0], hostTarget) {
		t.Fatalf("original host target = %v, want unchanged Player2", obj.Targets)
	}
	if len(obj.SplicedTargets) != 1 || len(obj.SplicedTargets[0]) != 1 ||
		!spliceTargetsEqual(obj.SplicedTargets[0][0], spliceTarget) {
		t.Fatalf("original spliced target = %v, want unchanged creature", obj.SplicedTargets)
	}
}

// TestSpliceCopyWithoutRetargetPreservesSplicedTargets confirms a copy made
// without the choose-new-targets rider keeps every original host and spliced
// target (CR 707.10a).
func TestSpliceCopyWithoutRetargetPreservesSplicedTargets(t *testing.T) {
	g, engine := newSpliceGame()
	hostID := addCardToHand(g, game.Player1, cardg.GlacialRay())
	addCardToHand(g, game.Player1, cardg.GlacialRay())
	spliceVictim := addToughCreature(g, game.Player2, 4)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)

	hostTarget := game.PlayerTarget(game.Player2)
	spliceTarget := game.PermanentTarget(spliceVictim.ObjectID)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &spliceScriptAgent{
		splices: []string{"Glacial Ray"},
		targets: []game.Target{spliceTarget},
	}}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(hostID, []game.Target{hostTarget}, 0, nil), agents, &TurnLog{}) {
		t.Fatal("casting Glacial Ray host with a spliced Glacial Ray failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("no spell on the stack after cast")
	}

	addEffectSpellToStack(g, game.Player1,
		game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(obj.ID)})
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	top, ok := g.Stack.Peek()
	if !ok || !top.Copy {
		t.Fatal("copy not on top of the stack")
	}
	if len(top.Targets) != 1 || !spliceTargetsEqual(top.Targets[0], hostTarget) {
		t.Fatalf("copy host target = %v, want preserved Player2", top.Targets)
	}
	if len(top.SplicedTargets) != 1 || len(top.SplicedTargets[0]) != 1 ||
		!spliceTargetsEqual(top.SplicedTargets[0][0], spliceTarget) {
		t.Fatalf("copy spliced target = %v, want preserved creature", top.SplicedTargets)
	}
}
