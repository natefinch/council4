package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// genericAnthemEmblem is an emblem with "Creatures you control get +1/+1 and
// have flying", used to guard the emblem continuous-effect machinery
// independently of any specific card or dungeon room.
func genericAnthemEmblem() game.Ability {
	return &game.StaticAbility{
		Text: "Creatures you control get +1/+1 and have flying.",
		ContinuousEffects: []game.ContinuousEffect{
			{
				Layer:          game.LayerPowerToughnessModify,
				Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
				PowerDelta:     1,
				ToughnessDelta: 1,
			},
			{
				Layer:       game.LayerAbility,
				Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
				AddKeywords: []game.Keyword{game.Flying},
			},
		},
	}
}

func TestEmblemAnthemBuffsControllersCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mine := addCombatPermanent(g, game.Player1, namedCreatureDef("Mine"))
	obj := &game.StackObject{Controller: game.Player1}
	resolveInstruction(engine, g, obj, game.CreateEmblem{EmblemAbilities: []game.Ability{genericAnthemEmblem()}}, nil)

	permanent, ok := permanentByObjectID(g, mine.ObjectID)
	if !ok {
		t.Fatal("controlled creature disappeared")
	}
	values := effectivePermanentValues(g, permanent)
	if values.power != 2 || values.toughness != 2 {
		t.Fatalf("creature = %d/%d, want 2/2 (1/1 + emblem +1/+1)", values.power, values.toughness)
	}
	if !hasKeyword(g, permanent, game.Flying) {
		t.Fatal("emblem did not grant flying to its controller's creature")
	}
}

func TestEmblemAnthemIgnoresOpponentsCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	theirs := addCombatPermanent(g, game.Player2, namedCreatureDef("Theirs"))
	obj := &game.StackObject{Controller: game.Player1}
	resolveInstruction(engine, g, obj, game.CreateEmblem{EmblemAbilities: []game.Ability{genericAnthemEmblem()}}, nil)

	permanent, ok := permanentByObjectID(g, theirs.ObjectID)
	if !ok {
		t.Fatal("opponent creature disappeared")
	}
	values := effectivePermanentValues(g, permanent)
	if values.power != 1 || values.toughness != 1 {
		t.Fatalf("opponent creature = %d/%d, want 1/1 (emblem only buffs its controller's creatures)", values.power, values.toughness)
	}
	if hasKeyword(g, permanent, game.Flying) {
		t.Fatal("emblem granted flying to an opponent's creature")
	}
}

func TestEmblemAnthemAppliesToNewlyEnteredCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	obj := &game.StackObject{Controller: game.Player1}
	resolveInstruction(engine, g, obj, game.CreateEmblem{EmblemAbilities: []game.Ability{genericAnthemEmblem()}}, nil)

	// A creature that enters after the emblem already exists must still be buffed:
	// emblem sources are collected fresh each frame, not snapshotted at creation.
	later := addCombatPermanent(g, game.Player1, namedCreatureDef("Later"))
	permanent, ok := permanentByObjectID(g, later.ObjectID)
	if !ok {
		t.Fatal("controlled creature disappeared")
	}
	values := effectivePermanentValues(g, permanent)
	if values.power != 2 || values.toughness != 2 {
		t.Fatalf("later creature = %d/%d, want 2/2", values.power, values.toughness)
	}
	if !hasKeyword(g, permanent, game.Flying) {
		t.Fatal("emblem did not grant flying to a later-entering creature")
	}
}
