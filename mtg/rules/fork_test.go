package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func forkTestSpell(g *game.Game, controller game.PlayerID, def *game.CardDef) *game.StackObject {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: def, Owner: controller}
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   cardID,
		Controller: controller,
	}
	g.Stack.Push(obj)
	return obj
}

func resolveForkCopy(t *testing.T, g *game.Game, engine *Engine, original *game.StackObject, setColors []color.Color) *game.StackObject {
	t.Helper()
	addEffectSpellToStack(g, game.Player1, game.CopyStackObject{
		Object:              game.TargetStackObjectReference(0),
		MayChooseNewTargets: true,
		SetColors:           setColors,
	}, []game.Target{game.StackObjectTarget(original.ID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	copyObj, ok := g.Stack.Peek()
	if !ok || !copyObj.Copy || copyObj.ID == original.ID {
		t.Fatalf("copy not created: %+v", copyObj)
	}
	return copyObj
}

func TestForkCopyColorExceptionReplacesSourceColors(t *testing.T) {
	for _, test := range []struct {
		name   string
		colors []color.Color
	}{
		{name: "red", colors: []color.Color{color.Red}},
		{name: "colorless"},
		{name: "multicolored_color_indicator", colors: []color.Color{color.Blue, color.Green}},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			def := &game.CardDef{CardFace: game.CardFace{
				Name:   "Source Spell",
				Types:  []types.Card{types.Instant},
				Colors: append([]color.Color(nil), test.colors...),
			}}
			original := forkTestSpell(g, game.Player2, def)
			copyObj := resolveForkCopy(t, g, engine, original, []color.Color{color.Red})

			got, ok := stackObjectColors(g, copyObj)
			if !ok || !slices.Equal(got, []color.Color{color.Red}) {
				t.Fatalf("copy colors = %v, %v, want [Red], true", got, ok)
			}
			if !copyObj.CopyValues.Exists ||
				!slices.Equal(copyObj.CopyValues.Val.Colors, []color.Color{color.Red}) {
				t.Fatalf("copy values = %+v, want red copiable values", copyObj.CopyValues)
			}
			originalColors, ok := stackObjectColors(g, original)
			if !ok || !slices.Equal(originalColors, test.colors) {
				t.Fatalf("original colors = %v, want unchanged %v", originalColors, test.colors)
			}
			if !slices.Equal(def.Colors, test.colors) {
				t.Fatalf("source definition mutated to %v, want %v", def.Colors, test.colors)
			}

			g.Stack.Pop()
			resolvingColors, ok := stackObjectColors(g, copyObj)
			if !ok || !slices.Equal(resolvingColors, []color.Color{color.Red}) {
				t.Fatalf("resolving copy colors = %v, %v, want persistent [Red]", resolvingColors, ok)
			}
		})
	}
}

func TestForkCopyPreservesCopiableValuesAndCastChoices(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	ability := game.Mode{Sequence: []game.Instruction{{
		Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
	}}}.Ability()
	def := &game.CardDef{CardFace: game.CardFace{
		Name:         "Choice Spell",
		Types:        []types.Card{types.Sorcery},
		Colors:       []color.Color{color.Blue},
		OracleText:   "Choose one.",
		SpellAbility: opt.Val(ability),
	}}
	original := forkTestSpell(g, game.Player2, def)
	original.Controller = game.Player2
	original.ChosenModes = []int{0}
	original.XValue = 7
	original.KickerPaid = true
	original.KickerCount = 2
	original.AdditionalCostsPaid = []string{"discarded a card"}
	original.Targets = []game.Target{game.PlayerTarget(game.Player3)}
	original.TargetCounts = []int{1}

	copyObj := resolveForkCopy(t, g, engine, original, []color.Color{color.Red})
	if copyObj.Controller != game.Player1 ||
		copyObj.XValue != 7 ||
		!copyObj.KickerPaid ||
		copyObj.KickerCount != 2 ||
		!slices.Equal(copyObj.ChosenModes, []int{0}) ||
		!slices.Equal(copyObj.AdditionalCostsPaid, []string{"discarded a card"}) ||
		!slices.Equal(copyObj.Targets, original.Targets) ||
		!slices.Equal(copyObj.TargetCounts, []int{1}) {
		t.Fatalf("copy lost choices or controller: %+v", copyObj)
	}
	effective, ok := stackObjectSpellDef(g, copyObj)
	if !ok ||
		effective.Name != def.Name ||
		effective.OracleText != def.OracleText ||
		!effective.SpellAbility.Exists {
		t.Fatalf("effective copied definition = %+v, %v", effective, ok)
	}
}

func TestForkColorExceptionIsCopiedByLaterCopies(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	original := forkTestSpell(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:   "Blue Spell",
		Types:  []types.Card{types.Instant},
		Colors: []color.Color{color.Blue},
	}})
	redCopy := resolveForkCopy(t, g, engine, original, []color.Color{color.Red})

	addEffectSpellToStack(g, game.Player1, game.CopyStackObject{
		Object: game.TargetStackObjectReference(0),
	}, []game.Target{game.StackObjectTarget(redCopy.ID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	secondCopy, _ := g.Stack.Peek()
	got, ok := stackObjectColors(g, secondCopy)
	if !ok || !slices.Equal(got, []color.Color{color.Red}) {
		t.Fatalf("copy of Fork copy colors = %v, %v, want [Red], true", got, ok)
	}
	if !secondCopy.CopyValues.Exists ||
		!slices.Equal(secondCopy.CopyValues.Val.Colors, []color.Color{color.Red}) {
		t.Fatalf("copy of copy values = %+v, want red", secondCopy.CopyValues)
	}
}

func TestForkColorExceptionStillYieldsToDevoid(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	original := forkTestSpell(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:            "Devoid Spell",
		Types:           []types.Card{types.Instant},
		StaticAbilities: []game.StaticAbility{game.DevoidStaticBody},
	}})
	copyObj := resolveForkCopy(t, g, engine, original, []color.Color{color.Red})
	if !slices.Equal(copyObj.CopyValues.Val.Colors, []color.Color{color.Red}) {
		t.Fatalf("layer-1 copy colors = %v, want [Red]", copyObj.CopyValues.Val.Colors)
	}
	if got, ok := stackObjectColors(g, copyObj); !ok || len(got) != 0 {
		t.Fatalf("effective Devoid copy colors = %v, %v, want colorless", got, ok)
	}
}

func TestForkRedCopyFiresRedSpellCopyTrigger(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	triggerSource := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:          game.EventSpellCast,
		Controller:     game.TriggerControllerYou,
		MatchSpellCopy: true,
		CardSelection: game.Selection{
			RequiredTypesAny: []types.Card{types.Instant, types.Sorcery},
			ColorsAny:        []color.Color{color.Red},
		},
	}, []game.Instruction{{Primitive: game.Draw{
		Amount: game.Fixed(1),
		Player: game.ControllerReference(),
	}}}, nil)
	original := forkTestSpell(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:   "Blue Spell",
		Types:  []types.Card{types.Instant},
		Colors: []color.Color{color.Blue},
	}})
	copyObj := resolveForkCopy(t, g, engine, original, []color.Color{color.Red})

	var copyEvent *game.Event
	for i := range g.Events {
		if g.Events[i].Kind == game.EventSpellCopied && g.Events[i].StackObjectID == copyObj.ID {
			copyEvent = &g.Events[i]
			break
		}
	}
	if copyEvent == nil || !slices.Equal(copyEvent.Colors, []color.Color{color.Red}) {
		t.Fatalf("copy event = %+v, want red EventSpellCopied", copyEvent)
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("red spell-copy trigger did not fire")
	}
	top, _ := g.Stack.Peek()
	if top.Kind != game.StackTriggeredAbility || top.SourceID != triggerSource.ObjectID {
		t.Fatalf("top = %+v, want red copy trigger from %v", top, triggerSource.ObjectID)
	}
}
