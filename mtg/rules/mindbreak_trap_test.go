package rules

import (
	"reflect"
	"slices"
	"testing"

	cardm "github.com/natefinch/council4/mtg/cards/m"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// mindbreakExile extracts the ExileTargetSpells primitive from the real
// generated Mindbreak Trap definition so the runtime tests drive the curated
// card, not a hand-written stand-in.
func mindbreakExile(t *testing.T) game.ExileTargetSpells {
	t.Helper()
	def := cardm.MindbreakTrap()
	if !def.SpellAbility.Exists {
		t.Fatal("Mindbreak Trap has no spell ability")
	}
	seq := def.SpellAbility.Val.Modes[0].Sequence
	if len(seq) != 1 {
		t.Fatalf("spell ability sequence length = %d, want 1", len(seq))
	}
	prim, ok := seq[0].Primitive.(game.ExileTargetSpells)
	if !ok {
		t.Fatalf("spell ability primitive is %T, want game.ExileTargetSpells", seq[0].Primitive)
	}
	return prim
}

// pushVictimSpell registers a card instance owned and controlled by controller
// with a life-gain body and pushes a spell stack object for it, returning the
// object. The life-gain body is a resolution witness: if the spell is exiled it
// never gains life, and if it wrongly resolves the controller's life rises.
func pushVictimSpell(g *game.Game, controller game.PlayerID, name string, gain int) *game.StackObject {
	sourceID := g.IDGen.Next()
	g.CardInstances[sourceID] = &game.CardInstance{
		ID: sourceID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:  name,
			Types: []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
				Primitive: game.GainLife{Amount: game.Fixed(gain), Player: game.ControllerReference()},
			}}}.Ability()),
		}},
		Owner: controller,
	}
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   sourceID,
		Controller: controller,
	}
	g.Stack.Push(obj)
	return obj
}

// pushMindbreak registers a real Mindbreak Trap spell controlled by controller
// that targets the given stack objects and pushes it on top of the stack,
// recording one target spec's worth of counts so the variable-count group
// reference resolves the whole chosen set.
func pushMindbreak(g *game.Game, controller game.PlayerID, targetIDs ...id.ID) *game.StackObject {
	sourceID := g.IDGen.Next()
	g.CardInstances[sourceID] = &game.CardInstance{ID: sourceID, Def: cardm.MindbreakTrap(), Owner: controller}
	targets := make([]game.Target, len(targetIDs))
	for i, targetID := range targetIDs {
		targets[i] = game.StackObjectTarget(targetID)
	}
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     sourceID,
		Controller:   controller,
		Targets:      targets,
		TargetCounts: []int{len(targetIDs)},
	}
	g.Stack.Push(obj)
	return obj
}

func stackContainsObject(g *game.Game, objectID id.ID) bool {
	for _, obj := range g.Stack.Objects() {
		if obj.ID == objectID {
			return true
		}
	}
	return false
}

// TestMindbreakTrapRealCardDefShape locks the curated definition to both
// features issue #1779 added: the per-opponent spells-cast {0} alternative cost
// and the variable-count exile-target-spells effect. Sourcing the assertions
// from the registered definition proves the generated card — not a stand-in —
// carries the typed shapes the parser, compiler, lowering, and render produced.
func TestMindbreakTrapRealCardDefShape(t *testing.T) {
	def := cardm.MindbreakTrap()

	if got, want := def.ManaCost.Val, (cost.Mana{cost.O(2), cost.U, cost.U}); !def.ManaCost.Exists || !reflect.DeepEqual(got, want) {
		t.Fatalf("mana cost = %v (exists %t), want %v", got, def.ManaCost.Exists, want)
	}
	if !slices.Equal(def.Types, []types.Card{types.Instant}) {
		t.Fatalf("types = %v, want [Instant]", def.Types)
	}
	if !slices.Equal(def.Subtypes, []types.Sub{types.Trap}) {
		t.Fatalf("subtypes = %v, want [Trap]", def.Subtypes)
	}
	if !reflect.DeepEqual(def.ColorIdentity, color.NewIdentity(color.Blue)) {
		t.Fatalf("color identity = %v, want blue", def.ColorIdentity)
	}

	if got := len(def.AlternativeCosts); got != 1 {
		t.Fatalf("alternative costs length = %d, want 1", got)
	}
	wantAlt := cost.Alternative{
		Label:          "Pay {0}",
		ManaCost:       opt.Val(cost.Mana{cost.O(0)}),
		Condition:      cost.AlternativeConditionOpponentCastSpellsThisTurn,
		ConditionCount: 3,
	}
	if got := def.AlternativeCosts[0]; !reflect.DeepEqual(got, wantAlt) {
		t.Fatalf("alternative cost = %+v, want %+v", got, wantAlt)
	}

	wantExile := game.ExileTargetSpells{Object: game.AllTargetStackObjectsReference(0)}
	if got := mindbreakExile(t); !reflect.DeepEqual(got, wantExile) {
		t.Fatalf("exile primitive = %+v, want %+v", got, wantExile)
	}

	specs := def.SpellAbility.Val.Modes[0].Targets
	if len(specs) != 1 {
		t.Fatalf("target specs = %d, want 1", len(specs))
	}
	spec := specs[0]
	if spec.MinTargets != 0 || spec.MaxTargets != 99 {
		t.Fatalf("target cardinality = [%d,%d], want [0,99] (any number)", spec.MinTargets, spec.MaxTargets)
	}
	if spec.Allow&game.TargetAllowStackObject == 0 {
		t.Fatalf("target spec Allow = %v, want stack objects allowed", spec.Allow)
	}
	if !slices.Equal(spec.Predicate.StackObjectKinds, []game.StackObjectKind{game.StackSpell}) {
		t.Fatalf("stack object kinds = %v, want [StackSpell] (spells only, never abilities)", spec.Predicate.StackObjectKinds)
	}
}

// TestMindbreakTrapExilesChosenSpellCounts drives the real card over the whole
// "any number" cardinality band — zero, one, and many targets — proving each
// chosen spell is removed from the stack and its card put into its owner's
// exile, while a zero-target cast resolves as a legal no-op that touches
// nothing.
func TestMindbreakTrapExilesChosenSpellCounts(t *testing.T) {
	t.Run("zero targets resolves as a no-op", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		bystander := pushVictimSpell(g, game.Player2, "Bystander", 5)
		mindbreak := pushMindbreak(g, game.Player1)

		engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

		if stackContainsObject(g, mindbreak.ID) {
			t.Fatal("zero-target Mindbreak did not leave the stack (should resolve, not fizzle-hang)")
		}
		if !stackContainsObject(g, bystander.ID) {
			t.Fatal("zero-target Mindbreak removed an untargeted spell")
		}
		if got := g.Players[game.Player1].Exile.Size() + g.Players[game.Player2].Exile.Size(); got != 0 {
			t.Fatalf("exiled cards = %d, want 0 for a zero-target cast", got)
		}
	})

	t.Run("one target is exiled", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		victim := pushVictimSpell(g, game.Player2, "Victim", 5)
		victimCard := g.CardInstances[victim.SourceID].ID
		mindbreak := pushMindbreak(g, game.Player1, victim.ID)

		engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

		if stackContainsObject(g, victim.ID) {
			t.Fatal("targeted spell remained on the stack")
		}
		if !g.Players[game.Player2].Exile.Contains(victimCard) {
			t.Fatal("exiled spell's card is not in its owner's exile zone")
		}
		if stackContainsObject(g, mindbreak.ID) {
			t.Fatal("resolved Mindbreak remained on the stack")
		}
	})

	t.Run("many targets are all exiled", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		victims := []*game.StackObject{
			pushVictimSpell(g, game.Player2, "Victim A", 5),
			pushVictimSpell(g, game.Player3, "Victim B", 5),
			pushVictimSpell(g, game.Player2, "Victim C", 5),
		}
		targetIDs := make([]id.ID, len(victims))
		for i, victim := range victims {
			targetIDs[i] = victim.ID
		}
		pushMindbreak(g, game.Player1, targetIDs...)

		engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

		for i, victim := range victims {
			if stackContainsObject(g, victim.ID) {
				t.Fatalf("victim %d remained on the stack", i)
			}
			card := g.CardInstances[victim.SourceID].ID
			if !g.Players[victim.Controller].Exile.Contains(card) {
				t.Fatalf("victim %d card not exiled", i)
			}
		}
	})
}

// TestMindbreakTrapExilesOwnAndOpponentSpells proves the effect is owner-blind:
// it exiles the caster's own spell and an opponent's spell chosen together,
// each card landing in its own owner's exile zone.
func TestMindbreakTrapExilesOwnAndOpponentSpells(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	own := pushVictimSpell(g, game.Player1, "Own Spell", 5)
	opponent := pushVictimSpell(g, game.Player2, "Opponent Spell", 5)
	pushMindbreak(g, game.Player1, own.ID, opponent.ID)

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if stackContainsObject(g, own.ID) || stackContainsObject(g, opponent.ID) {
		t.Fatal("a targeted spell (own or opponent) remained on the stack")
	}
	if !g.Players[game.Player1].Exile.Contains(g.CardInstances[own.SourceID].ID) {
		t.Fatal("caster's own spell was not exiled to their exile zone")
	}
	if !g.Players[game.Player2].Exile.Contains(g.CardInstances[opponent.SourceID].ID) {
		t.Fatal("opponent's spell was not exiled to their exile zone")
	}
}

// TestMindbreakTrapExiledSpellsNeverResolve proves an exiled spell's effect is
// cancelled: two of three spells are exiled and never gain their controllers
// life as the stack drains, while the one untargeted spell resolves normally.
func TestMindbreakTrapExiledSpellsNeverResolve(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	p2Base := g.Players[game.Player2].Life
	p3Base := g.Players[game.Player3].Life

	survivor := pushVictimSpell(g, game.Player2, "Survivor", 4)
	exiledA := pushVictimSpell(g, game.Player2, "Exiled A", 7)
	exiledB := pushVictimSpell(g, game.Player3, "Exiled B", 9)
	pushMindbreak(g, game.Player1, exiledA.ID, exiledB.ID)

	// Drain the whole stack: Mindbreak first (exiling A and B), then the
	// survivor resolves.
	for !g.Stack.IsEmpty() {
		engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	}

	if got, want := g.Players[game.Player2].Life, p2Base+4; got != want {
		t.Fatalf("player2 life = %d, want %d (only the untargeted survivor's +4 applies)", got, want)
	}
	if got := g.Players[game.Player3].Life; got != p3Base {
		t.Fatalf("player3 life = %d, want %d (exiled spell must not gain life)", got, p3Base)
	}
	if !stackContainsObject(g, survivor.ID) && g.Players[game.Player2].Graveyard.Size() == 0 {
		t.Fatal("survivor spell disappeared without resolving")
	}
}

// TestMindbreakTrapPreservesStackOrder proves removing several spells at once
// keeps the surviving spells in their original relative stack order and never
// resolves a removed spell in place.
func TestMindbreakTrapPreservesStackOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	a := pushVictimSpell(g, game.Player2, "A", 5)
	b := pushVictimSpell(g, game.Player2, "B", 5)
	c := pushVictimSpell(g, game.Player2, "C", 5)
	d := pushVictimSpell(g, game.Player2, "D", 5)
	// Exile the two interior spells, leaving A (bottom) and C.
	pushMindbreak(g, game.Player1, b.ID, d.ID)

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	var remaining []id.ID
	for _, obj := range g.Stack.Objects() {
		if obj.Kind == game.StackSpell {
			remaining = append(remaining, obj.ID)
		}
	}
	if want := []id.ID{a.ID, c.ID}; !slices.Equal(remaining, want) {
		t.Fatalf("surviving stack spells = %v, want %v (bottom-to-top A then C)", remaining, want)
	}
}

// TestMindbreakTrapExilesSpellCopyWithoutCard proves a targeted spell copy
// (CR 707.10) simply ceases to exist when exiled: it leaves the stack and no
// card is moved to exile, because a copy has no physical card.
func TestMindbreakTrapExilesSpellCopyWithoutCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	copySpell := pushVictimSpell(g, game.Player2, "Copied Spell", 5)
	copySpell.Copy = true
	pushMindbreak(g, game.Player1, copySpell.ID)

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if stackContainsObject(g, copySpell.ID) {
		t.Fatal("exiled spell copy remained on the stack")
	}
	if got := g.Players[game.Player2].Exile.Size(); got != 0 {
		t.Fatalf("exile zone size = %d, want 0 (a copy has no card to exile)", got)
	}
}

// TestMindbreakTrapExilesCommanderToCommandZone proves exile is an ordinary
// zone change that honors replacements: a commander spell exiled by Mindbreak
// is redirected to its owner's command zone (CR 903.9a), not the exile zone.
func TestMindbreakTrapExilesCommanderToCommandZone(t *testing.T) {
	configs := [game.NumPlayers]game.PlayerConfig{
		game.Player2: {Commander: commanderDef("Exiled General", color.Green)},
	}
	g := game.NewGame(configs)
	engine := NewEngine(nil)
	commanderID := g.Players[game.Player2].CommanderInstanceID
	if commanderID == 0 {
		t.Fatal("player2 has no commander instance")
	}
	commanderSpell := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   commanderID,
		Controller: game.Player2,
	}
	g.Stack.Push(commanderSpell)
	pushMindbreak(g, game.Player1, commanderSpell.ID)

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if stackContainsObject(g, commanderSpell.ID) {
		t.Fatal("commander spell remained on the stack")
	}
	if g.Players[game.Player2].Exile.Contains(commanderID) {
		t.Fatal("commander went to exile instead of the command zone")
	}
	if !g.Players[game.Player2].CommandZone.Contains(commanderID) {
		t.Fatal("exiled commander was not redirected to the command zone")
	}
}

// TestMindbreakTrapResolutionRechecksTargetLegality proves the resolution-time
// recheck (CR 608.2b) governs the group exile: an interior target that has left
// the stack is skipped while the still-legal targets are exiled (partial
// legality), and a cast whose every target has become illegal is countered with
// nothing exiled (full fizzle).
func TestMindbreakTrapResolutionRechecksTargetLegality(t *testing.T) {
	t.Run("partial legality exiles only still-legal spells", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		gone := pushVictimSpell(g, game.Player2, "Gone", 5)
		legal := pushVictimSpell(g, game.Player2, "Legal", 5)
		mindbreak := pushMindbreak(g, game.Player1, gone.ID, legal.ID)
		// The first target leaves the stack before Mindbreak resolves.
		g.Stack.RemoveByID(gone.ID)

		engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

		if !g.Players[game.Player2].Exile.Contains(g.CardInstances[legal.SourceID].ID) {
			t.Fatal("still-legal target was not exiled")
		}
		if stackContainsObject(g, legal.ID) {
			t.Fatal("still-legal target remained on the stack")
		}
		if stackContainsObject(g, mindbreak.ID) {
			t.Fatal("Mindbreak did not resolve with a partially legal target set")
		}
	})

	t.Run("all targets illegal fizzles with nothing exiled", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		goneA := pushVictimSpell(g, game.Player2, "Gone A", 5)
		goneB := pushVictimSpell(g, game.Player2, "Gone B", 5)
		mindbreak := pushMindbreak(g, game.Player1, goneA.ID, goneB.ID)
		g.Stack.RemoveByID(goneA.ID)
		g.Stack.RemoveByID(goneB.ID)

		engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

		if stackContainsObject(g, mindbreak.ID) {
			t.Fatal("fizzled Mindbreak remained on the stack")
		}
		if got := g.Players[game.Player2].Exile.Size(); got != 0 {
			t.Fatalf("exiled cards = %d, want 0 when every target is illegal", got)
		}
		// A fizzled instant goes to its owner's graveyard.
		if !g.Players[game.Player1].Graveyard.Contains(mindbreak.SourceID) {
			t.Fatal("fizzled Mindbreak was not put into its owner's graveyard")
		}
	})
}

// TestMindbreakTrapTargetEnumerationExcludesSelfAndAbilities proves the target
// selection composes correctly from existing enumeration: the "any number of
// target spells" spec offers every other spell on the stack — the caster's own
// and opponents' alike — but never the Mindbreak spell itself (a spell cannot
// target itself) and never a non-spell stack object such as an activated
// ability.
func TestMindbreakTrapTargetEnumerationExcludesSelfAndAbilities(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ownSpell := pushVictimSpell(g, game.Player1, "Own Spell", 5)
	opponentSpell := pushVictimSpell(g, game.Player2, "Opponent Spell", 5)
	ability := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackActivatedAbility,
		SourceID:   g.IDGen.Next(),
		Controller: game.Player2,
	}
	g.Stack.Push(ability)
	mindbreak := pushMindbreak(g, game.Player1)

	def := cardm.MindbreakTrap()
	spec := def.SpellAbility.Val.Modes[0].Targets[0]
	candidates := targetCandidatesForSpec(g, game.Player1, def, mindbreak.ID, game.Event{}, &spec)

	got := make(map[id.ID]bool)
	for _, candidate := range candidates {
		if candidate.Kind != game.TargetStackObject {
			t.Fatalf("candidate kind = %v, want stack object", candidate.Kind)
		}
		got[candidate.StackObjectID] = true
	}
	if !got[ownSpell.ID] {
		t.Error("own spell was not offered as a target")
	}
	if !got[opponentSpell.ID] {
		t.Error("opponent spell was not offered as a target")
	}
	if got[mindbreak.ID] {
		t.Error("Mindbreak offered itself as a target (a spell cannot target itself)")
	}
	if got[ability.ID] {
		t.Error("an activated ability was offered as a target (spells only)")
	}
}

// TestMindbreakTrapCopyExilesThenCeasesToExist proves a copy of Mindbreak Trap
// (CR 707.10) resolves through the same path: it exiles the spells chosen for
// the copy and then ceases to exist without going to a graveyard, since a copy
// of a spell is not a card. This also shows a Mindbreak copy is an ordinary
// spell others could target.
func TestMindbreakTrapCopyExilesThenCeasesToExist(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := pushVictimSpell(g, game.Player2, "Victim", 5)
	victimCard := g.CardInstances[victim.SourceID].ID
	copySpell := pushMindbreak(g, game.Player1, victim.ID)
	copySpell.Copy = true

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if stackContainsObject(g, victim.ID) {
		t.Fatal("copy of Mindbreak did not exile its target")
	}
	if !g.Players[game.Player2].Exile.Contains(victimCard) {
		t.Fatal("target exiled by the copy is not in its owner's exile zone")
	}
	if g.Players[game.Player1].Graveyard.Contains(copySpell.SourceID) {
		t.Fatal("the Mindbreak copy went to a graveyard (a copy is not a card)")
	}
}

// TestMindbreakTrapDuplicateTargetIsExiledOnce proves the ID-first removal is
// robust to a repeated target: if the same spell is listed twice, it is exiled
// exactly once and the second removal is a harmless no-op that neither errors
// nor exiles a second card. Normal target selection forbids duplicates; this
// guards the resolution primitive directly.
func TestMindbreakTrapDuplicateTargetIsExiledOnce(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := pushVictimSpell(g, game.Player2, "Victim", 5)
	pushMindbreak(g, game.Player1, victim.ID, victim.ID)

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if stackContainsObject(g, victim.ID) {
		t.Fatal("duplicated target spell remained on the stack")
	}
	if got := g.Players[game.Player2].Exile.Size(); got != 1 {
		t.Fatalf("exile zone size = %d, want exactly 1 (no double exile)", got)
	}
}
