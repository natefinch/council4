package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// The static-source frame cache must be transparent: a value computed inside a
// caller-owned frame must equal the value computed by standalone queries. These
// tests build a board that exercises the cached paths (static-ability sources,
// counters, control change, face-down) and assert framed and unframed results
// agree for every permanent, which catches any cache divergence.

func frameDiffCreature(name string, power, toughness int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: toughness}),
	}}
}

// frameDiffBoard builds a four-player board with several overlapping continuous
// effects, a static-ability source, counters, a control change, and a face-down
// permanent.
func frameDiffBoard(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	reach := frameDiffCreature("Reachy", 1, 3)
	reach.StaticAbilities = append(reach.StaticAbilities, game.ReachStaticBody)
	addCombatPermanent(g, game.Player1, reach)

	bear := addCombatPermanent(g, game.Player1, frameDiffCreature("Bear", 2, 2))
	bear.Counters.Add(counter.PlusOnePlusOne, 2)

	stolen := addCombatPermanent(g, game.Player2, frameDiffCreature("Stolen", 4, 4))

	addCombatPermanent(g, game.Player3, frameDiffCreature("Plain", 3, 3))

	faceDown := addCombatPermanent(g, game.Player4, frameDiffCreature("Hidden", 5, 5))
	faceDown.FaceDown = true

	g.ContinuousEffects = append(g.ContinuousEffects,
		game.ContinuousEffect{
			ID:               1,
			AffectedObjectID: bear.ObjectID,
			Timestamp:        10,
			Layer:            game.LayerPowerToughnessModify,
			PowerDelta:       1,
			ToughnessDelta:   1,
		},
		game.ContinuousEffect{
			ID:               2,
			AffectedObjectID: stolen.ObjectID,
			Timestamp:        20,
			Layer:            game.LayerControl,
			NewController:    opt.Val(game.Player1),
		},
	)
	return g
}

type frameDiffValues struct {
	power       int
	toughness   int
	toughnessOK bool
	controller  game.PlayerID
	abilities   int
	reach       bool
	// characteristics serializes the slice-valued fields shared by reference out
	// of the cache (name/types/subtypes/supertypes/colors), which are exactly
	// the fields a shared-slice mutation bug would corrupt.
	characteristics string
}

func frameDiffSnapshot(g *game.Game, permanent *game.Permanent) frameDiffValues {
	toughness, toughnessOK := effectiveToughness(g, permanent)
	values := effectivePermanentValues(g, permanent)
	return frameDiffValues{
		power:       effectivePower(g, permanent),
		toughness:   toughness,
		toughnessOK: toughnessOK,
		controller:  effectiveController(g, permanent),
		abilities:   len(permanentEffectiveAbilities(g, permanent)),
		reach:       hasKeyword(g, permanent, game.Reach),
		characteristics: fmt.Sprintf("name=%q types=%v subtypes=%v supertypes=%v colors=%v",
			values.name, values.types, values.subtypes, values.supertypes, values.colors),
	}
}

func TestStaticSourceFrameIsTransparent(t *testing.T) {
	g := frameDiffBoard(t)

	for _, permanent := range g.Battlefield {
		unframed := frameDiffSnapshot(g, permanent)

		g.BeginStaticSourceFrame()
		framed := frameDiffSnapshot(g, permanent)
		// A second read in the same frame must also match (exercises the hit).
		framedAgain := frameDiffSnapshot(g, permanent)
		g.EndStaticSourceFrame()

		if unframed != framed || unframed != framedAgain {
			t.Errorf("permanent %d: unframed %+v, framed %+v, framed-again %+v",
				permanent.CardInstanceID, unframed, framed, framedAgain)
		}
	}
}

// TestStaticSourceFrameMatchesAfterReopen confirms a fresh frame reflects state
// changes made between frames (the cache does not persist across frames).
func TestStaticSourceFrameMatchesAfterReopen(t *testing.T) {
	g := frameDiffBoard(t)
	bear := g.Battlefield[1]

	g.BeginStaticSourceFrame()
	before := effectivePower(g, bear)
	g.EndStaticSourceFrame()

	bear.Counters.Add(counter.PlusOnePlusOne, 3)

	g.BeginStaticSourceFrame()
	after := effectivePower(g, bear)
	g.EndStaticSourceFrame()

	unframed := effectivePower(g, bear)
	if after != unframed {
		t.Errorf("after adding counters, framed power %d != unframed %d", after, unframed)
	}
	if after <= before {
		t.Errorf("adding +1/+1 counters should raise power: before %d, after %d", before, after)
	}
}

// TestStaticSourceFrameReflectsTypeAndControlChangesBetweenFrames confirms a
// fresh frame picks up changes to the slice-valued and control characteristics
// made between frames, so the cache never serves them stale across a mutation.
func TestStaticSourceFrameReflectsTypeAndControlChangesBetweenFrames(t *testing.T) {
	g := frameDiffBoard(t)
	stolen := g.Battlefield[2]

	g.BeginStaticSourceFrame()
	_ = frameDiffSnapshot(g, stolen)
	g.EndStaticSourceFrame()

	// Drop the control-change effect and add a subtype to the underlying card.
	g.ContinuousEffects = g.ContinuousEffects[:1]
	if card, ok := g.GetCardInstance(stolen.CardInstanceID); ok {
		card.Def.Subtypes = append(card.Def.Subtypes, types.Elf)
	}

	g.BeginStaticSourceFrame()
	framed := frameDiffSnapshot(g, stolen)
	g.EndStaticSourceFrame()
	unframed := frameDiffSnapshot(g, stolen)

	if framed != unframed {
		t.Errorf("after between-frame changes, framed %+v != unframed %+v", framed, unframed)
	}
	if framed.controller != stolen.Owner {
		t.Errorf("controller should revert to owner %v after the control effect is removed, got %v",
			stolen.Owner, framed.controller)
	}
}

func TestStandaloneEvaluatorReadFrames(t *testing.T) {
	for _, test := range []struct {
		name string
		read func(*game.Game, *game.Permanent)
	}{
		{"characteristics", func(g *game.Game, p *game.Permanent) { _ = effectivePermanentValues(g, p) }},
		{"controller", func(g *game.Game, p *game.Permanent) { _ = effectiveController(g, p) }},
		{"before layer", func(g *game.Game, p *game.Permanent) {
			_ = permanentValuesBeforeLayer(g, p, game.LayerPowerToughnessModify)
		}},
		{"rule effects", func(g *game.Game, _ *game.Permanent) { _ = activeRuleEffects(g) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := frameDiffBoard(t)
			p := g.Battlefield[1]
			test.read(g, p)
			if g.InStaticSourceFrame() {
				t.Fatal("standalone read retained its frame")
			}
			func() {
				g.BeginStaticSourceFrame()
				defer g.EndStaticSourceFrame()
				outer := frameCacheFor(g)
				for range 2 {
					test.read(g, p)
					if !g.InStaticSourceFrame() || frameCacheFor(g) != outer || !outer.sourcesBuilt {
						t.Error("nested read discarded or replaced its caller's cache")
					}
				}
			}()
			if g.InStaticSourceFrame() {
				t.Fatal("nested read left an unbalanced frame")
			}
		})
	}
}

func TestStandaloneEvaluatorSeesMutationsAndClones(t *testing.T) {
	g := frameDiffBoard(t)
	permanent := g.Battlefield[2]
	if effectivePower(g, permanent) != 4 || effectiveController(g, permanent) != game.Player1 {
		t.Fatal("initial control effect or power was not applied")
	}

	g.BeginStaticSourceFrame()
	_ = effectivePermanentValues(g, permanent)
	branch := g.Clone()
	g.EndStaticSourceFrame()
	if branch.InStaticSourceFrame() {
		t.Fatal("clone inherited its parent's read frame")
	}
	copied, ok := permanentByObjectID(branch, permanent.ObjectID)
	if !ok {
		t.Fatal("clone lost the fixture permanent")
	}
	branch.ContinuousEffects = branch.ContinuousEffects[:1]
	copied.Counters.Add(counter.PlusOnePlusOne, 3)
	if effectivePower(branch, copied) != 7 || effectiveController(branch, copied) != game.Player2 {
		t.Fatal("standalone reads retained pre-mutation characteristics")
	}
	if before := permanentValuesBeforeLayer(branch, copied, game.LayerPowerToughnessModify); before.power != 4 {
		t.Fatal("before-layer read included later counters or retained a stale board")
	}
	copied.FaceDown = true
	if effectivePower(branch, copied) != 5 {
		t.Fatal("standalone read retained face-up characteristics")
	}
	copied.FaceDown = false
	if effectivePower(branch, copied) != 7 ||
		effectivePower(g, permanent) != 4 || effectiveController(g, permanent) != game.Player1 {
		t.Fatal("mutation or derived values leaked between clones")
	}
	if g.InStaticSourceFrame() || branch.InStaticSourceFrame() {
		t.Fatal("standalone reads left a frame on the game or clone")
	}
}

func TestStandaloneRuleEffectsReflectSourceDeparture(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Read-frame prohibition", Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind: game.RuleEffectCantAttack, AffectedController: game.ControllerOpponent,
			PermanentTypes: []types.Card{types.Creature},
		}}}},
	}})
	original := activeRuleEffects(g)
	if len(original) != 1 || original[0].SourceObjectID != source.ObjectID {
		t.Fatal("fixture did not expose its static rule effect")
	}
	branch := g.Clone()
	clonedSource, ok := permanentByObjectID(branch, source.ObjectID)
	if !ok || !movePermanentToZone(branch, clonedSource, zone.Graveyard) {
		t.Fatal("could not remove the cloned source")
	}
	if len(activeRuleEffects(branch)) != 0 || len(activeRuleEffects(g)) != 1 {
		t.Fatal("rule-effect cache survived source departure or crossed clones")
	}
	if !movePermanentToZone(g, source, zone.Graveyard) || len(activeRuleEffects(g)) != 0 {
		t.Fatal("standalone rule query retained a departed source")
	}
	if len(original) != 1 || original[0].SourceObjectID != source.ObjectID {
		t.Fatal("later reads mutated the earlier rule-effect snapshot")
	}
}
