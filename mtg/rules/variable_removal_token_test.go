package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// variableRemovalSourceDef is a minimal sorcery stand-in that hosts the
// variable-target removal-token instructions under test.
func variableRemovalSourceDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Variable Removal",
		Types: []types.Card{types.Sorcery},
	}}
}

// namedTokenDef builds a vanilla creature token definition for payoff assertions.
func namedTokenDef(name string, c color.Color) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Colors:    []color.Color{c},
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}}
}

func namedTokenCount(g *game.Game, controller game.PlayerID, name string) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent != nil && permanent.Token && permanent.Controller == controller &&
			permanent.TokenDef != nil && permanent.TokenDef.Name == name {
			count++
		}
	}
	return count
}

func removalStackObjectWithTargets(source *game.Permanent, permanents ...*game.Permanent) *game.StackObject {
	obj := linkedSourceObject(source)
	for _, permanent := range permanents {
		obj.Targets = append(obj.Targets, game.PermanentTarget(permanent.ObjectID))
	}
	return obj
}

// TestRemoveTargetsForTokenDestroysAllTargetsUnderLink verifies the destroy form
// (Descent of the Dragons): every chosen target is destroyed under the removal
// link, including targets controlled by different players.
func TestRemoveTargetsForTokenDestroysAllTargetsUnderLink(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mine := addCombatCreaturePermanent(g, game.Player1)
	theirsA := addCombatCreaturePermanent(g, game.Player2)
	theirsB := addCombatCreaturePermanent(g, game.Player2)
	source := addCombatPermanent(g, game.Player1, variableRemovalSourceDef())
	obj := removalStackObjectWithTargets(source, mine, theirsA, theirsB)

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.RemoveTargetsForToken{
		LinkedKey: game.LinkedKey("removed-targets-for-token"),
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	for _, removed := range []*game.Permanent{mine, theirsA, theirsB} {
		if permanentByCardID(g, removed.CardInstanceID) != nil {
			t.Fatal("a chosen target remained on the battlefield after destroy")
		}
		if !g.Players[removed.Controller].Graveyard.Contains(removed.CardInstanceID) {
			t.Fatalf("destroyed creature %d did not reach its owner's graveyard", removed.CardInstanceID)
		}
	}
	key := linkedObjectSourceKey(g, obj, "removed-targets-for-token")
	if got := len(linkedObjects(g, key)); got != 3 {
		t.Fatalf("linked removed objects = %d, want 3 (one per target)", got)
	}
}

// TestRemoveTargetsForTokenExilesAllTargetsUnderLink verifies the exile form
// (Curse of the Swine): every chosen target is exiled under the removal link.
func TestRemoveTargetsForTokenExilesAllTargetsUnderLink(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mine := addCombatCreaturePermanent(g, game.Player1)
	theirs := addCombatCreaturePermanent(g, game.Player2)
	source := addCombatPermanent(g, game.Player1, variableRemovalSourceDef())
	obj := removalStackObjectWithTargets(source, mine, theirs)

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.RemoveTargetsForToken{
		Exile:     true,
		LinkedKey: game.LinkedKey("removed-targets-for-token"),
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	for _, removed := range []*game.Permanent{mine, theirs} {
		if permanentByCardID(g, removed.CardInstanceID) != nil {
			t.Fatal("a chosen target remained on the battlefield after exile")
		}
		if !g.Players[removed.Controller].Exile.Contains(removed.CardInstanceID) {
			t.Fatalf("exiled creature %d did not reach its owner's exile zone", removed.CardInstanceID)
		}
	}
	key := linkedObjectSourceKey(g, obj, "removed-targets-for-token")
	if got := len(linkedObjects(g, key)); got != 2 {
		t.Fatalf("linked removed objects = %d, want 2 (one per target)", got)
	}
}

// TestRemoveTargetsForTokenChainMintsPerController verifies the full chain: each
// removed creature's last-known controller creates exactly one token, and the
// link is cleared after the payoff.
func TestRemoveTargetsForTokenChainMintsPerController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mine := addCombatCreaturePermanent(g, game.Player1)
	theirsA := addCombatCreaturePermanent(g, game.Player2)
	theirsB := addCombatCreaturePermanent(g, game.Player2)
	source := addCombatPermanent(g, game.Player1, variableRemovalSourceDef())
	obj := removalStackObjectWithTargets(source, mine, theirsA, theirsB)

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.RemoveTargetsForToken{
		LinkedKey: game.LinkedKey("removed-targets-for-token"),
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.CreateTokenForEachDestroyed{
		Source:    game.TokenDef(namedTokenDef("Dragon", color.Red)),
		LinkedKey: game.LinkedKey("removed-targets-for-token"),
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := namedTokenCount(g, game.Player1, "Dragon"); got != 1 {
		t.Fatalf("Player1 Dragon tokens = %d, want 1", got)
	}
	if got := namedTokenCount(g, game.Player2, "Dragon"); got != 2 {
		t.Fatalf("Player2 Dragon tokens = %d, want 2", got)
	}
	key := linkedObjectSourceKey(g, obj, "removed-targets-for-token")
	if got := len(linkedObjects(g, key)); got != 0 {
		t.Fatalf("linked removed objects after payoff = %d, want 0 (link cleared)", got)
	}
}

// TestRemoveTargetsForTokenChainMintsForRemovedToken proves the payoff is
// text-faithful for tokens: removing a token target still mints the payoff token
// for that token's controller. A token has CardInstanceID == 0, so the link must
// preserve its ObjectID (permanentObjectBindingRef) or the payoff would skip it.
func TestRemoveTargetsForTokenChainMintsForRemovedToken(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mine := addCombatCreaturePermanent(g, game.Player1)
	theirToken := addTokenCreaturePermanent(g, game.Player2, "Boar")
	source := addCombatPermanent(g, game.Player1, variableRemovalSourceDef())
	obj := removalStackObjectWithTargets(source, mine, theirToken)

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.RemoveTargetsForToken{
		Exile:     true,
		LinkedKey: game.LinkedKey("removed-targets-for-token"),
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.CreateTokenForEachDestroyed{
		Source:    game.TokenDef(namedTokenDef("Pig", color.Green)),
		LinkedKey: game.LinkedKey("removed-targets-for-token"),
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := namedTokenCount(g, game.Player1, "Pig"); got != 1 {
		t.Fatalf("Player1 Pig tokens = %d, want 1 (its removed nontoken creature)", got)
	}
	if got := namedTokenCount(g, game.Player2, "Pig"); got != 1 {
		t.Fatalf("Player2 Pig tokens = %d, want 1 (its removed TOKEN creature must still get the payoff)", got)
	}
}

// TestRemoveTargetsForTokenDestroyRespectsIndestructible verifies the destroy
// form honors indestructibility: an indestructible target is neither destroyed
// nor linked, while its destructible companion is.
func TestRemoveTargetsForTokenDestroyRespectsIndestructible(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mortal := addCombatCreaturePermanent(g, game.Player1)
	tough := addCombatCreaturePermanent(g, game.Player2, game.Indestructible)
	source := addCombatPermanent(g, game.Player1, variableRemovalSourceDef())
	obj := removalStackObjectWithTargets(source, mortal, tough)

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.RemoveTargetsForToken{
		LinkedKey: game.LinkedKey("removed-targets-for-token"),
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if permanentByCardID(g, mortal.CardInstanceID) != nil {
		t.Fatal("destructible target survived destroy")
	}
	if permanentByCardID(g, tough.CardInstanceID) == nil {
		t.Fatal("indestructible target was destroyed")
	}
	key := linkedObjectSourceKey(g, obj, "removed-targets-for-token")
	if got := len(linkedObjects(g, key)); got != 1 {
		t.Fatalf("linked removed objects = %d, want 1 (only the destroyed creature)", got)
	}
}

// countEqualsXSpellDef builds a Curse-of-the-Swine-like sorcery whose lone
// creature target spec binds its target count to X via CountEqualsX.
func countEqualsXSpellDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Exile X Creatures",
		Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets:   0,
				MaxTargets:   20,
				Constraint:   "target creatures",
				Allow:        game.TargetAllowPermanent,
				Selection:    opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				CountEqualsX: true,
			}},
			Sequence: []game.Instruction{{Primitive: game.RemoveTargetsForToken{
				Exile:     true,
				LinkedKey: game.LinkedKey("removed-targets-for-token"),
			}}},
		}.Ability()),
	}}
}

// TestSpellTargetCountsMatchXBindsTargetCountToX verifies the CountEqualsX cast
// legality: a spell is legal only when the number of chosen targets equals the
// chosen X, so the variable cost binds the number of exiled creatures.
func TestSpellTargetCountsMatchXBindsTargetCountToX(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	a := addCombatCreaturePermanent(g, game.Player1)
	b := addCombatCreaturePermanent(g, game.Player2)
	card := countEqualsXSpellDef()
	all := []game.Target{game.PermanentTarget(a.ObjectID), game.PermanentTarget(b.ObjectID)}

	cases := []struct {
		name    string
		targets []game.Target
		xValue  int
		want    bool
	}{
		{"two targets X=2", all, 2, true},
		{"two targets X=1", all, 1, false},
		{"two targets X=0", all, 0, false},
		{"one target X=1", all[:1], 1, true},
		{"one target X=2", all[:1], 2, false},
		{"zero targets X=0", nil, 0, true},
		{"zero targets X=1", nil, 1, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := spellTargetCountsMatchX(g, game.Player1, card, nil, tc.targets, tc.xValue)
			if got != tc.want {
				t.Fatalf("spellTargetCountsMatchX(%d targets, X=%d) = %v, want %v",
					len(tc.targets), tc.xValue, got, tc.want)
			}
		})
	}
}

// TestSpellTargetCountsMatchXIgnoresNonXSpecs verifies a spell without a
// CountEqualsX spec is unaffected by the X-binding check.
func TestSpellTargetCountsMatchXIgnoresNonXSpecs(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	a := addCombatCreaturePermanent(g, game.Player1)
	card := &game.CardDef{CardFace: game.CardFace{
		Name:  "Plain Removal",
		Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target creature",
				Allow:      game.TargetAllowPermanent,
				Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
			}},
			Sequence: []game.Instruction{{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}}},
		}.Ability()),
	}}
	if !spellTargetCountsMatchX(g, game.Player1, card, nil, []game.Target{game.PermanentTarget(a.ObjectID)}, 5) {
		t.Fatal("non-CountEqualsX spell should pass the X-binding check regardless of X")
	}
}
