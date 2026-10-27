package rules

import (
	"strconv"
	"testing"

	"github.com/natefinch/council4/internal/carddata"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func compileTargetCountSpell(t *testing.T, body string) *game.CardDef {
	t.Helper()
	defs := legacyCardDefs(t, carddata.ScryfallCard{
		Name: "Count Parameters", Layout: "normal", TypeLine: "Sorcery", ManaCost: "{X}{B}",
		OracleText: body,
	})

	return defs[0]
}

func TestCompiledExactXTargetChoicesAndResolution(t *testing.T) {
	for _, x := range []int{0, 1, 2, 3} {
		t.Run(strconv.Itoa(x), func(t *testing.T) {
			def := compileTargetCountSpell(t, "Destroy X target artifacts. Draw X cards.")
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			var chosen []*game.Permanent
			var targets []game.Target
			for range x {
				permanent := addArtifactPermanent(g, game.Player2)
				chosen = append(chosen, permanent)
				targets = append(targets, game.PermanentTarget(permanent.ObjectID))
			}
			untargeted := addArtifactPermanent(g, game.Player2)
			for range 4 {
				addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Library Card"}})
			}
			if !spellTargetCountsMatchX(g, game.Player1, def, nil, targets, x, game.CastBranch{}) {
				t.Fatal("exact X chosen set was rejected")
			}
			for _, other := range []int{0, 1, 2, 3} {
				if other != x && spellTargetCountsMatchX(g, game.Player1, def, nil, targets, other, game.CastBranch{}) {
					t.Fatalf("%d targets accepted with X=%d", x, other)
				}
			}
			spell := addCardToHand(g, game.Player1, def)
			g.Players[game.Player1].Hand.Remove(spell)
			g.Stack.Push(&game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: spell, Controller: game.Player1,
				XValue: x, Targets: targets, TargetCounts: []int{x},
			})
			NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
			for _, permanent := range chosen {
				if _, exists := permanentByObjectID(g, permanent.ObjectID); exists {
					t.Fatal("chosen artifact remained")
				}
			}
			if _, exists := permanentByObjectID(g, untargeted.ObjectID); !exists {
				t.Fatal("untargeted artifact was destroyed")
			}
			if got := len(g.Players[game.Player1].Hand.All()); got != x {
				t.Fatalf("drawn cards=%d, want %d", got, x)
			}
		})
	}
}

func TestCompiledLastVariableTargetUsesItsOwnSpec(t *testing.T) {
	def := compileTargetCountSpell(t, "Target player draws a card. Destroy X target artifacts.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addArtifactPermanent(g, game.Player2)
	second := addArtifactPermanent(g, game.Player3)
	addCardToLibrary(g, game.Player4, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Card"}})
	spell := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].Hand.Remove(spell)
	g.Stack.Push(&game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: spell, Controller: game.Player1, XValue: 2,
		Targets: []game.Target{
			game.PlayerTarget(game.Player4), game.PermanentTarget(first.ObjectID), game.PermanentTarget(second.ObjectID),
		},
		TargetCounts: []int{1, 2},
	})
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
	if len(g.Players[game.Player4].Hand.All()) != 1 {
		t.Fatal("the first fixed target did not draw")
	}

	for _, permanent := range []*game.Permanent{first, second} {
		if _, exists := permanentByObjectID(g, permanent.ObjectID); exists {
			t.Fatal("last variable target group addressed the wrong spec")
		}
	}
}

func TestCompiledEachTargetDamageUsesLastSpecAndKeepsIllegalSlot(t *testing.T) {
	def := compileTargetCountSpell(t, "Target player draws a card. Count Parameters deals 4 damage to each of X targets.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creatureDef := &game.CardDef{CardFace: game.CardFace{
		Name: "Durable Creature", Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 8}), Toughness: opt.Val(game.PT{Value: 8}),
	}}
	chosen := addCombatPermanent(g, game.Player2, creatureDef)
	departed := addCombatPermanent(g, game.Player2, creatureDef)
	addCardToLibrary(g, game.Player4, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Card"}})
	spell := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].Hand.Remove(spell)
	g.Stack.Push(&game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: spell, Controller: game.Player1, XValue: 3,
		Targets: []game.Target{
			game.PlayerTarget(game.Player4), game.PermanentTarget(chosen.ObjectID),
			game.PermanentTarget(departed.ObjectID), game.PlayerTarget(game.Player3),
		},
		TargetCounts: []int{1, 3},
	})
	if _, destroyed := destroyPermanentInBatch(g, departed.ObjectID, 0, false); !destroyed {
		t.Fatal("could not make one selected target illegal before resolution")
	}
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
	if chosen.MarkedDamage != 4 || g.Players[game.Player3].Life != 36 {
		t.Fatalf("mixed damage: creature=%d player3=%d player4=%d",
			chosen.MarkedDamage, g.Players[game.Player3].Life, g.Players[game.Player4].Life)
	}
	if g.Players[game.Player4].Life != 40 || len(g.Players[game.Player4].Hand.All()) != 1 {
		t.Fatal("the preceding fixed target was damaged or failed to draw")
	}
}

func TestCompiledTargetCountSelectionAndOwner(t *testing.T) {
	def := compileTargetCountSpell(t, "Return X target creatures you don't control to their owners' hands.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	mine := addCreaturePermanent(g, game.Player1)
	borrowed := addCreaturePermanent(g, game.Player1)
	borrowed.Controller = game.Player2
	opponent := addCreaturePermanent(g, game.Player3)
	spec := def.SpellAbility.Val.Modes[0].Targets[0]
	choices := targetChoicesForSpecs(g, game.Player1, def, 0, game.Event{}, []game.TargetSpec{spec})
	for _, choice := range choices.choices {
		for _, target := range choice {
			if target.PermanentID == mine.ObjectID {
				t.Fatal("you-don't-control selector offered a controlled creature")
			}
		}
	}
	spell := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].Hand.Remove(spell)
	g.Stack.Push(&game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: spell, Controller: game.Player1, XValue: 2,
		Targets:      []game.Target{game.PermanentTarget(borrowed.ObjectID), game.PermanentTarget(opponent.ObjectID)},
		TargetCounts: []int{2},
	})
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Hand.Contains(borrowed.CardInstanceID) ||
		!g.Players[game.Player3].Hand.Contains(opponent.CardInstanceID) {
		t.Fatal("returned objects did not reach their owners' hands")
	}
	if _, exists := permanentByObjectID(g, mine.ObjectID); !exists {
		t.Fatal("controlled untargeted creature was moved")
	}
}

func TestCompiledExactXSameGraveyard(t *testing.T) {
	def := compileTargetCountSpell(t, "Exile X target cards from a single graveyard.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "First", Types: []types.Card{types.Instant}}})
	second := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Second", Types: []types.Card{types.Sorcery}}})
	other := addCardToGraveyard(g, game.Player3, &game.CardDef{CardFace: game.CardFace{Name: "Other", Types: []types.Card{types.Instant}}})
	same := []game.Target{currentCardTarget(t, g, first), currentCardTarget(t, g, second)}
	cross := []game.Target{currentCardTarget(t, g, first), currentCardTarget(t, g, other)}
	if !spellTargetCountsMatchX(g, game.Player1, def, nil, same, 2, game.CastBranch{}) ||
		spellTargetCountsMatchX(g, game.Player1, def, nil, cross, 2, game.CastBranch{}) {
		t.Fatal("exact X did not preserve the same-graveyard choice constraint")
	}
	spell := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].Hand.Remove(spell)
	g.Stack.Push(&game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: spell, Controller: game.Player1,
		XValue: 2, Targets: same, TargetCounts: []int{2},
	})
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player2].Exile.Contains(first) || !g.Players[game.Player2].Exile.Contains(second) ||
		!g.Players[game.Player3].Graveyard.Contains(other) {
		t.Fatal("graveyard X exile did not affect exactly the chosen cards")
	}
}
