package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
)

func TestLowerChooseNewTargetsSpell(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		oracleText   string
		wantKinds    []game.StackObjectKind
		wantOptional bool
	}{
		{
			name:         "target spell",
			oracleText:   "You may choose new targets for target spell.",
			wantKinds:    []game.StackObjectKind{game.StackSpell},
			wantOptional: true,
		},
		{
			name:       "target spell or ability",
			oracleText: "You may choose new targets for target spell or ability.",
			wantKinds: []game.StackObjectKind{
				game.StackSpell, game.StackActivatedAbility, game.StackTriggeredAbility,
			},
			wantOptional: true,
		},
		{
			name:         "change the target of target spell",
			oracleText:   "Change the target of target spell with a single target.",
			wantKinds:    []game.StackObjectKind{game.StackSpell},
			wantOptional: false,
		},
		{
			name:       "change the target of target spell or ability",
			oracleText: "Change the target of target spell or ability with a single target.",
			wantKinds: []game.StackObjectKind{
				game.StackSpell, game.StackActivatedAbility, game.StackTriggeredAbility,
			},
			wantOptional: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Retarget",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: test.oracleText,
			})
			if !face.SpellAbility.Exists {
				t.Fatal("spell ability missing")
			}
			ability := face.SpellAbility.Val
			if len(ability.Modes) != 1 {
				t.Fatalf("modes = %d, want 1", len(ability.Modes))
			}
			mode := ability.Modes[0]
			if len(mode.Targets) != 1 {
				t.Fatalf("targets = %d, want 1", len(mode.Targets))
			}
			target := mode.Targets[0]
			if target.Allow != game.TargetAllowStackObject {
				t.Fatalf("target allow = %v, want stack object", target.Allow)
			}
			if !slices.Equal(target.Predicate.StackObjectKinds, test.wantKinds) {
				t.Fatalf("stack object kinds = %+v, want %+v", target.Predicate.StackObjectKinds, test.wantKinds)
			}
			if len(mode.Sequence) != 1 {
				t.Fatalf("sequence = %d, want 1", len(mode.Sequence))
			}
			instr := mode.Sequence[0]
			if instr.Optional != test.wantOptional {
				t.Fatalf("instruction Optional = %v, want %v", instr.Optional, test.wantOptional)
			}
			retarget, ok := instr.Primitive.(game.ChooseNewTargets)
			if !ok {
				t.Fatalf("primitive = %T, want game.ChooseNewTargets", instr.Primitive)
			}
			if retarget.Object.Kind() != game.ObjectReferenceTargetStackObject || retarget.Object.TargetIndex() != 0 {
				t.Fatalf("retarget object = %+v, want target stack object 0", retarget.Object)
			}
		})
	}
}

// TestLowerImpsMischiefRetargetThenLoseLife proves the two-effect redirect body
// "Change the target of target spell with a single target. You lose life equal
// to that spell's mana value." (Imp's Mischief) lowers to a ChooseNewTargets
// over the targeted spell followed by a controller life loss whose amount reads
// that spell's live mana value through DynamicAmountObjectManaValue.
func TestLowerImpsMischiefRetargetThenLoseLife(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Imp's Mischief",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{1}{B}",
		OracleText: "Change the target of target spell with a single target. You lose life equal to that spell's mana value.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability missing")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || mode.Targets[0].Allow != game.TargetAllowStackObject {
		t.Fatalf("targets = %#v, want one stack-object target", mode.Targets)
	}
	if !slices.Equal(mode.Targets[0].Predicate.StackObjectKinds, []game.StackObjectKind{game.StackSpell}) {
		t.Fatalf("stack object kinds = %+v, want spell", mode.Targets[0].Predicate.StackObjectKinds)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d, want 2", len(mode.Sequence))
	}
	retarget, ok := mode.Sequence[0].Primitive.(game.ChooseNewTargets)
	if !ok || retarget.Object != game.TargetStackObjectReference(0) {
		t.Fatalf("first primitive = %#v, want ChooseNewTargets over target 0", mode.Sequence[0].Primitive)
	}
	loseLife, ok := mode.Sequence[1].Primitive.(game.LoseLife)
	if !ok {
		t.Fatalf("second primitive = %T, want game.LoseLife", mode.Sequence[1].Primitive)
	}
	if loseLife.Player != game.ControllerReference() {
		t.Fatalf("lose life player = %#v, want controller", loseLife.Player)
	}
	dynamic := loseLife.Amount.DynamicAmount()
	if !dynamic.Exists {
		t.Fatalf("lose life amount = %#v, want dynamic", loseLife.Amount)
	}
	if dynamic.Val.Kind != game.DynamicAmountObjectManaValue ||
		dynamic.Val.Object != game.TargetStackObjectReference(0) {
		t.Fatalf("lose life dynamic = %#v, want object mana value of target stack object 0", dynamic.Val)
	}
}

// TestLowerRerouteRetargetAbilityThenDraw proves Reroute's two-paragraph body
// "Change the target of target activated ability with a single target. ... Draw
// a card." lowers to a ChooseNewTargets over the targeted activated ability,
// followed by a card draw merged in from the trailing paragraph.
func TestLowerRerouteRetargetAbilityThenDraw(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Reroute",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{1}{R}",
		OracleText: "Change the target of target activated ability with a single target. (Mana abilities can't be targeted.)\nDraw a card.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability missing")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || mode.Targets[0].Allow != game.TargetAllowStackObject {
		t.Fatalf("targets = %#v, want one stack-object target", mode.Targets)
	}
	if !slices.Equal(mode.Targets[0].Predicate.StackObjectKinds, []game.StackObjectKind{game.StackActivatedAbility}) {
		t.Fatalf("stack object kinds = %+v, want activated ability", mode.Targets[0].Predicate.StackObjectKinds)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d, want 2", len(mode.Sequence))
	}
	retarget, ok := mode.Sequence[0].Primitive.(game.ChooseNewTargets)
	if !ok || retarget.Object != game.TargetStackObjectReference(0) {
		t.Fatalf("first primitive = %#v, want ChooseNewTargets over target 0", mode.Sequence[0].Primitive)
	}
	if mode.Sequence[0].Optional {
		t.Fatal("retarget should be mandatory")
	}
	if _, ok := mode.Sequence[1].Primitive.(game.Draw); !ok {
		t.Fatalf("second primitive = %T, want game.Draw", mode.Sequence[1].Primitive)
	}
}

// TestLowerGoblinFlectomancerOptionalRetarget proves Goblin Flectomancer's
// sacrifice activated ability "You may change the targets of target instant or
// sorcery spell." lowers to an optional ChooseNewTargets over the targeted
// instant or sorcery spell, behind a Sacrifice additional cost.
func TestLowerGoblinFlectomancerOptionalRetarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Goblin Flectomancer",
		Layout:     "normal",
		TypeLine:   "Creature — Goblin Wizard",
		ManaCost:   "{U}{R}{R}",
		OracleText: "Sacrifice this creature: You may change the targets of target instant or sorcery spell.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	if len(ability.AdditionalCosts) != 1 || ability.AdditionalCosts[0].Kind != cost.AdditionalSacrificeSource {
		t.Fatalf("additional costs = %#v, want one sacrifice-source cost", ability.AdditionalCosts)
	}
	mode := ability.Content.Modes[0]
	if len(mode.Targets) != 1 || mode.Targets[0].Allow != game.TargetAllowStackObject {
		t.Fatalf("targets = %#v, want one stack-object target", mode.Targets)
	}
	if !slices.Equal(mode.Targets[0].Predicate.StackObjectKinds, []game.StackObjectKind{game.StackSpell}) {
		t.Fatalf("stack object kinds = %+v, want spell", mode.Targets[0].Predicate.StackObjectKinds)
	}
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %d, want 1", len(mode.Sequence))
	}
	if !mode.Sequence[0].Optional {
		t.Fatal("retarget should be optional")
	}
	retarget, ok := mode.Sequence[0].Primitive.(game.ChooseNewTargets)
	if !ok || retarget.Object != game.TargetStackObjectReference(0) {
		t.Fatalf("primitive = %#v, want ChooseNewTargets over target 0", mode.Sequence[0].Primitive)
	}
}
