package eval

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func contentOf(primitives ...game.Primitive) game.AbilityContent {
	sequence := make([]game.Instruction, 0, len(primitives))
	for _, primitive := range primitives {
		sequence = append(sequence, game.Instruction{Primitive: primitive})
	}
	return game.AbilityContent{Modes: []game.Mode{{Sequence: sequence}}}
}

func TestScorableEffectClassifiesValueDominantPrimitives(t *testing.T) {
	you := game.ControllerReference()
	cases := []struct {
		name      string
		primitive game.Primitive
		want      EffectAtom
	}{
		{"draw", game.Draw{Amount: game.Fixed(2), Player: you}, EffectAtom{Kind: EffectCardsDrawn, Amount: 2, Affected: AffectedYou}},
		{"discard", game.Discard{Amount: game.Fixed(3), Player: you}, EffectAtom{Kind: EffectCardsLost, Amount: 3, Affected: AffectedYou}},
		{"mill", game.Mill{Amount: game.Fixed(2), Player: you}, EffectAtom{Kind: EffectCardsLost, Amount: 2, Affected: AffectedYou}},
		{"gain life", game.GainLife{Amount: game.Fixed(4), Player: you}, EffectAtom{Kind: EffectLifeGained, Amount: 4, Affected: AffectedYou}},
		{"lose life", game.LoseLife{Amount: game.Fixed(1), Player: you}, EffectAtom{Kind: EffectLifeLost, Amount: 1, Affected: AffectedYou}},
		{"damage", game.Damage{Amount: game.Fixed(3)}, EffectAtom{Kind: EffectDamageDealt, Amount: 3, Affected: AffectedTarget}},
		{"destroy", game.Destroy{}, EffectAtom{Kind: EffectPermanentRemoved, Affected: AffectedTarget}},
		{"exile", game.Exile{}, EffectAtom{Kind: EffectPermanentRemoved, Affected: AffectedTarget}},
		{"bounce", game.Bounce{}, EffectAtom{Kind: EffectPermanentRemoved, Affected: AffectedTarget}},
		{"tap", game.Tap{}, EffectAtom{Kind: EffectPermanentTapped, Affected: AffectedTarget}},
		{"fight", game.Fight{}, EffectAtom{Kind: EffectDamageDealt, Affected: AffectedTarget}},
		{"monstrosity", game.Monstrosity{Amount: game.Fixed(3)}, EffectAtom{Kind: EffectCounterAdded, Amount: 3, Affected: AffectedUnknown}},
		{"add mana", game.AddMana{Amount: game.Fixed(2)}, EffectAtom{Kind: EffectManaAdded, Amount: 2, Affected: AffectedYou}},
		{"create token", game.CreateToken{Amount: game.Fixed(1)}, EffectAtom{Kind: EffectTokenCreated, Amount: 1, Affected: AffectedYou}},
		{"search", game.Search{Amount: game.Fixed(1)}, EffectAtom{Kind: EffectCardTutored, Amount: 1, Affected: AffectedYou}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			atoms := ScorableEffect(contentOf(c.primitive))
			if len(atoms) != 1 || atoms[0] != c.want {
				t.Fatalf("ScorableEffect(%s) = %#v, want [%#v]", c.name, atoms, c.want)
			}
		})
	}
}

func TestScorableEffectSummarizesBazaarAsNetCardLoss(t *testing.T) {
	you := game.ControllerReference()
	atoms := ScorableEffect(contentOf(
		game.Draw{Amount: game.Fixed(2), Player: you},
		game.Discard{Amount: game.Fixed(3), Player: you},
	))
	if len(atoms) != 2 {
		t.Fatalf("atoms = %#v, want draw + discard", atoms)
	}
	drawn, lost := 0, 0
	for _, atom := range atoms {
		if atom.Affected != AffectedYou {
			t.Fatalf("atom %#v not affecting you", atom)
		}
		switch atom.Kind {
		case EffectCardsDrawn:
			drawn += atom.Amount
		case EffectCardsLost:
			lost += atom.Amount
		default:
		}
	}
	if drawn-lost != -1 {
		t.Fatalf("net cards = %d, want -1 (draw 2, discard 3)", drawn-lost)
	}
}

func TestScorableEffectFlagsDynamicAmount(t *testing.T) {
	atoms := ScorableEffect(contentOf(game.Draw{Amount: game.Dynamic(game.DynamicAmount{}), Player: game.ControllerReference()}))
	if len(atoms) != 1 || !atoms[0].IsDynamic || atoms[0].Amount != 0 {
		t.Fatalf("dynamic draw atom = %#v, want IsDynamic with zero amount", atoms)
	}
}

func TestScorableEffectIgnoresUnmodeledPrimitive(t *testing.T) {
	if atoms := ScorableEffect(contentOf(game.Scry{})); len(atoms) != 0 {
		t.Fatalf("unmodeled primitive produced atoms %#v, want none", atoms)
	}
}

func TestScorableEffectModesScoresOnlyChosenMode(t *testing.T) {
	you := game.ControllerReference()
	content := game.AbilityContent{
		MinModes: 1,
		MaxModes: 1,
		Modes: []game.Mode{
			{Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(2), Player: you}}}},
			{Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(5), Player: you}}}},
		},
	}

	chosen := ScorableEffectModes(content, []int{1})
	if len(chosen) != 1 || chosen[0].Kind != EffectLifeGained || chosen[0].Amount != 5 {
		t.Fatalf("chosen-mode atoms = %#v, want a single gain-life-5 atom", chosen)
	}

	unioned := ScorableEffectModes(content, nil)
	if len(unioned) != 2 {
		t.Fatalf("mode-unaware atoms = %#v, want both modes unioned", unioned)
	}
}

func TestScorableEffectFlagsConditionalInstructionDynamic(t *testing.T) {
	you := game.ControllerReference()
	content := game.AbilityContent{Modes: []game.Mode{{Sequence: []game.Instruction{
		{Primitive: game.Draw{Amount: game.Fixed(2), Player: you}},
		{
			Primitive: game.GainLife{Amount: game.Fixed(3), Player: you},
			Condition: opt.Val(game.EffectCondition{}),
		},
	}}}}

	atoms := ScorableEffectModes(content, nil)
	if len(atoms) != 2 {
		t.Fatalf("atoms = %#v, want draw + conditional gain-life", atoms)
	}
	if atoms[0].IsDynamic {
		t.Fatalf("unconditional draw atom %#v should not be dynamic", atoms[0])
	}
	if !atoms[1].IsDynamic {
		t.Fatalf("conditional gain-life atom %#v should be flagged dynamic", atoms[1])
	}
}

func TestScorableEffectFlagsOptionalInstructionDynamic(t *testing.T) {
	content := game.AbilityContent{Modes: []game.Mode{{Sequence: []game.Instruction{
		{Primitive: game.AddMana{Amount: game.Fixed(2)}, Optional: true},
	}}}}

	atoms := ScorableEffectModes(content, nil)
	if len(atoms) != 1 || !atoms[0].IsDynamic {
		t.Fatalf("optional atom = %#v, want IsDynamic", atoms)
	}
}

func TestScorableEffectClassifiesLandRampSearch(t *testing.T) {
	ramp := game.Search{
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:  zone.Library,
			Destination: zone.Battlefield,
			Filter:      game.Selection{RequiredTypes: []types.Card{types.Land}},
		},
	}
	atoms := ScorableEffect(contentOf(ramp))
	if len(atoms) != 1 || atoms[0].Kind != EffectLandRamp {
		t.Fatalf("land-fetch search atoms = %#v, want a single EffectLandRamp", atoms)
	}

	tutor := game.Search{
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:  zone.Library,
			Destination: zone.Hand,
			Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
		},
	}
	atoms = ScorableEffect(contentOf(tutor))
	if len(atoms) != 1 || atoms[0].Kind != EffectCardTutored {
		t.Fatalf("creature tutor atoms = %#v, want a single EffectCardTutored", atoms)
	}
}
