package cardgen

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerDrawTriggerYou(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Draw Sentinel",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Creature — Wizard",
		OracleText: "Whenever you draw a card, you gain 1 life.",
		Colors:     []string{"U"},
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	got := face.TriggeredAbilities[0]
	if got.Trigger.Type != game.TriggerWhenever {
		t.Errorf("Trigger.Type = %v, want TriggerWhenever", got.Trigger.Type)
	}
	if got.Trigger.Pattern.Event != game.EventCardDrawn {
		t.Errorf("Pattern.Event = %v, want EventCardDrawn", got.Trigger.Pattern.Event)
	}
	if got.Trigger.Pattern.Player != game.TriggerPlayerYou {
		t.Errorf("Pattern.Player = %v, want TriggerPlayerYou", got.Trigger.Pattern.Player)
	}
	if got.Trigger.Pattern.OneOrMore {
		t.Error("Pattern.OneOrMore = true, want false")
	}
}

func TestLowerDrawTriggerOpponent(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Opponent Draw Watcher",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Creature — Wizard",
		OracleText: "Whenever an opponent draws a card, that player loses 1 life.",
		Colors:     []string{"U"},
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	got := face.TriggeredAbilities[0]
	if got.Trigger.Pattern.Event != game.EventCardDrawn {
		t.Errorf("Pattern.Event = %v, want EventCardDrawn", got.Trigger.Pattern.Event)
	}
	if got.Trigger.Pattern.Player != game.TriggerPlayerOpponent {
		t.Errorf("Pattern.Player = %v, want TriggerPlayerOpponent", got.Trigger.Pattern.Player)
	}
}

func TestLowerDrawTriggerAnyPlayer(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Universal Draw Watcher",
		Layout:     "normal",
		ManaCost:   "{3}{U}",
		TypeLine:   "Creature — Wizard",
		OracleText: "Whenever a player draws a card, you gain 1 life.",
		Colors:     []string{"U"},
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	if face.TriggeredAbilities[0].Trigger.Pattern.Player != game.TriggerPlayerAny {
		t.Errorf("Pattern.Player = %v, want TriggerPlayerAny", face.TriggeredAbilities[0].Trigger.Pattern.Player)
	}
}

func TestLowerDiscardTriggerYou(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Discard Reactor",
		Layout:     "normal",
		ManaCost:   "{1}{B}",
		TypeLine:   "Creature — Rogue",
		OracleText: "Whenever you discard a card, you lose 1 life.",
		Colors:     []string{"B"},
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	got := face.TriggeredAbilities[0]
	if got.Trigger.Pattern.Event != game.EventCardDiscarded {
		t.Errorf("Pattern.Event = %v, want EventCardDiscarded", got.Trigger.Pattern.Event)
	}
	if got.Trigger.Pattern.Player != game.TriggerPlayerYou {
		t.Errorf("Pattern.Player = %v, want TriggerPlayerYou", got.Trigger.Pattern.Player)
	}
	if got.Trigger.Pattern.OneOrMore {
		t.Error("Pattern.OneOrMore = true, want false")
	}
}

func TestLowerDiscardOneOrMoreTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Discard Engine",
		Layout:     "normal",
		ManaCost:   "{2}{B}",
		TypeLine:   "Creature — Specter",
		OracleText: "Whenever you discard one or more cards, you lose 1 life.",
		Colors:     []string{"B"},
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	got := face.TriggeredAbilities[0]
	if got.Trigger.Pattern.Event != game.EventCardDiscarded {
		t.Errorf("Pattern.Event = %v, want EventCardDiscarded", got.Trigger.Pattern.Event)
	}
	if got.Trigger.Pattern.Player != game.TriggerPlayerYou {
		t.Errorf("Pattern.Player = %v, want TriggerPlayerYou", got.Trigger.Pattern.Player)
	}
	if !got.Trigger.Pattern.OneOrMore {
		t.Error("Pattern.OneOrMore = false, want true")
	}
}

func TestLowerDiscardTriggerCardTypeFilter(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Typed Discard Reactor",
		Layout:     "normal",
		ManaCost:   "{1}{B}",
		TypeLine:   "Creature — Rogue",
		OracleText: "Whenever you discard a creature card, you lose 1 life.\nWhenever you discard a noncreature, nonland card, you gain 1 life.",
		Colors:     []string{"B"},
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 2 {
		t.Fatalf("triggered abilities = %d, want 2", len(face.TriggeredAbilities))
	}
	creature := face.TriggeredAbilities[0]
	if creature.Trigger.Pattern.Event != game.EventCardDiscarded {
		t.Errorf("Pattern.Event = %v, want EventCardDiscarded", creature.Trigger.Pattern.Event)
	}
	wantRequired := game.Selection{RequiredTypes: []types.Card{types.Creature}}
	if !reflect.DeepEqual(creature.Trigger.Pattern.CardSelection, wantRequired) {
		t.Errorf("CardSelection = %#v, want %#v", creature.Trigger.Pattern.CardSelection, wantRequired)
	}
	noncreature := face.TriggeredAbilities[1]
	wantExcluded := game.Selection{ExcludedTypes: []types.Card{types.Creature, types.Land}}
	if !reflect.DeepEqual(noncreature.Trigger.Pattern.CardSelection, wantExcluded) {
		t.Errorf("CardSelection = %#v, want %#v", noncreature.Trigger.Pattern.CardSelection, wantExcluded)
	}
}

func TestLowerDiscardTriggerOpponent(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Specter Watcher",
		Layout:     "normal",
		ManaCost:   "{2}{B}",
		TypeLine:   "Creature — Specter",
		OracleText: "Whenever an opponent discards a card, you gain 1 life.",
		Colors:     []string{"B"},
		Power:      new("1"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	got := face.TriggeredAbilities[0]
	if got.Trigger.Pattern.Event != game.EventCardDiscarded {
		t.Errorf("Pattern.Event = %v, want EventCardDiscarded", got.Trigger.Pattern.Event)
	}
	if got.Trigger.Pattern.Player != game.TriggerPlayerOpponent {
		t.Errorf("Pattern.Player = %v, want TriggerPlayerOpponent", got.Trigger.Pattern.Player)
	}
}

func TestLowerDrawDiscardTriggerRejectsUnknownPhrase(t *testing.T) {
	t.Parallel()
	unknownPhrases := []string{
		"you draw two or more cards",
		"an opponent draws their second card in their draw step",
	}
	for _, phrase := range unknownPhrases {
		t.Run(phrase, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Unsupported Trigger Card",
				Layout:     "normal",
				ManaCost:   "{1}{U}",
				TypeLine:   "Creature — Wizard",
				OracleText: "Whenever " + phrase + ", you gain 1 life.",
				Colors:     []string{"U"},
				Power:      new("1"),
				Toughness:  new("1"),
			})
			if len(diagnostics) == 0 {
				t.Fatalf("expected diagnostic for phrase %q, got none", phrase)
			}
		})
	}
}

func TestLowerDrawDiscardTriggerSupportedInterveningCondition(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Conditional Draw Watcher",
		Layout:     "normal",
		ManaCost:   "{1}{U}",
		TypeLine:   "Creature — Wizard",
		OracleText: "Whenever you draw a card, if you have 5 or more life, you gain 1 life.",
		Colors:     []string{"U"},
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0].Trigger
	if trigger.Pattern.Event != game.EventCardDrawn {
		t.Errorf("event = %v, want EventCardDrawn", trigger.Pattern.Event)
	}
	if trigger.InterveningIf == "" || !trigger.InterveningCondition.Exists {
		t.Fatalf("trigger = %+v, want intervening condition", trigger)
	}
	if trigger.InterveningCondition.Val.ControllerLifeAtLeast != 5 {
		t.Errorf("condition = %+v, want ControllerLifeAtLeast 5", trigger.InterveningCondition.Val)
	}
}

func TestLowerDrawDiscardTriggerInterveningIfFailsClosedOnUnsupportedCondition(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Overflowing Draw Watcher",
		Layout:     "normal",
		ManaCost:   "{1}{U}",
		TypeLine:   "Creature — Wizard",
		OracleText: "Whenever you draw a card, if you have seven or more cards in hand, draw a card.",
		Colors:     []string{"U"},
		Power:      new("1"),
		Toughness:  new("1"),
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("draw trigger with unsupported intervening condition unexpectedly lowered")
	}
	if !strings.Contains(diagnostics[0].Detail, "condition") {
		t.Fatalf("diagnostic = %#v, want condition detail", diagnostics[0])
	}
}

func TestLowerSacrificeSpellTargetPlayerCreature(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Diabolic Edict",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target player sacrifices a creature.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not found")
	}
	modes := face.SpellAbility.Val.Modes
	if len(modes) != 1 {
		t.Fatalf("modes = %d, want 1", len(modes))
	}
	mode := modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v, want one target spec", mode.Targets)
	}
	if mode.Targets[0].Allow != game.TargetAllowPlayer {
		t.Fatalf("target allow = %v, want TargetAllowPlayer", mode.Targets[0].Allow)
	}
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(mode.Sequence))
	}
	prim, ok := mode.Sequence[0].Primitive.(game.SacrificePermanents)
	if !ok {
		t.Fatalf("primitive = %T, want game.SacrificePermanents", mode.Sequence[0].Primitive)
	}
	if prim.Player.Kind() != game.PlayerReferenceTargetPlayer || prim.Player.TargetIndex() != 0 {
		t.Fatalf("player = %#v, want TargetPlayerReference(0)", prim.Player)
	}
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		t.Fatalf("player group = %v, want none", prim.PlayerGroup.Kind)
	}
	if prim.Amount.Value() != 1 {
		t.Fatalf("amount = %d, want 1", prim.Amount.Value())
	}
	if !slices.Equal(prim.Selection.RequiredTypes, []types.Card{types.Creature}) {
		t.Fatalf("selection = %#v, want creature filter", prim.Selection)
	}
}

func TestLowerSacrificeSpellEachOpponentCreature(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Each Opponent Edict",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Each opponent sacrifices a creature of their choice.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not found")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 {
		t.Fatalf("targets = %d, want none", len(mode.Targets))
	}
	prim, ok := mode.Sequence[0].Primitive.(game.SacrificePermanents)
	if !ok {
		t.Fatalf("primitive = %T, want game.SacrificePermanents", mode.Sequence[0].Primitive)
	}
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceOpponents {
		t.Fatalf("player group = %v, want opponents", prim.PlayerGroup.Kind)
	}
	if prim.Player.Kind() != game.PlayerReferenceNone {
		t.Fatalf("player = %v, want none", prim.Player.Kind())
	}
	if prim.Amount.Value() != 1 {
		t.Fatalf("amount = %d, want 1", prim.Amount.Value())
	}
	if !slices.Equal(prim.Selection.RequiredTypes, []types.Card{types.Creature}) {
		t.Fatalf("selection = %#v, want creature filter", prim.Selection)
	}
}

func TestLowerSacrificeSpellEachPlayerLand(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Tremble",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Each player sacrifices a land.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not found")
	}
	mode := face.SpellAbility.Val.Modes[0]
	prim, ok := mode.Sequence[0].Primitive.(game.SacrificePermanents)
	if !ok {
		t.Fatalf("primitive = %T, want game.SacrificePermanents", mode.Sequence[0].Primitive)
	}
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceAllPlayers {
		t.Fatalf("player group = %v, want all players", prim.PlayerGroup.Kind)
	}
	if !slices.Equal(prim.Selection.RequiredTypes, []types.Card{types.Land}) {
		t.Fatalf("selection = %#v, want land filter", prim.Selection)
	}
}

func TestLowerSacrificeSpellTargetPlayerTwoCreatures(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Two Creature Edict",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target player sacrifices two creatures.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not found")
	}
	mode := face.SpellAbility.Val.Modes[0]
	prim, ok := mode.Sequence[0].Primitive.(game.SacrificePermanents)
	if !ok {
		t.Fatalf("primitive = %T, want game.SacrificePermanents", mode.Sequence[0].Primitive)
	}
	if prim.Amount.Value() != 2 {
		t.Fatalf("amount = %d, want 2", prim.Amount.Value())
	}
	if !slices.Equal(prim.Selection.RequiredTypes, []types.Card{types.Creature}) {
		t.Fatalf("selection = %#v, want creature filter", prim.Selection)
	}
}

func TestLowerSacrificeSpellEachPlayerPermanent(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Forced Pact",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Each player sacrifices a permanent.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not found")
	}
	mode := face.SpellAbility.Val.Modes[0]
	prim, ok := mode.Sequence[0].Primitive.(game.SacrificePermanents)
	if !ok {
		t.Fatalf("primitive = %T, want game.SacrificePermanents", mode.Sequence[0].Primitive)
	}
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceAllPlayers {
		t.Fatalf("player group = %v, want all players", prim.PlayerGroup.Kind)
	}
	if prim.Amount.Value() != 1 {
		t.Fatalf("amount = %d, want 1", prim.Amount.Value())
	}
	if !prim.Selection.Empty() {
		t.Fatalf("selection = %#v, want zero selection (any permanent)", prim.Selection)
	}
}

func TestLowerSacrificeSpellRejectsPronounReference(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Conditional Sacrifice",
		Layout:     "normal",
		ManaCost:   "{1}{B}",
		TypeLine:   "Creature — Zombie",
		OracleText: "When this creature enters, its controller sacrifices a creature.",
		Colors:     []string{"B"},
		Power:      new("2"),
		Toughness:  new("2"),
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("pronoun 'its controller' sacrifice unexpectedly lowered without diagnostic")
	}
}

func TestLowerSacrificeSpellRejectsUnknownActorPattern(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Forced Tribute",
		Layout:   "normal",
		TypeLine: "Sorcery",
		// "You sacrifice" is not a supported actor pattern
		OracleText: "You sacrifice a creature.",
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("unsupported sacrifice actor pattern unexpectedly lowered without diagnostic")
	}
}

func TestLowerSacrificeSpellRejectsQualifiedPermanentChoice(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Crackling Doom",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Crackling Doom deals 2 damage to each opponent. Each opponent sacrifices a creature with the greatest power among creatures that player controls.",
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("qualified sacrifice choice unexpectedly lowered without diagnostic")
	}
}

func TestLowerSacrificeSpellRejectsOrderedEffectSequence(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Wildfire",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Each player sacrifices four lands of their choice. Wildfire deals 4 damage to each creature.",
	}, "w")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("ordered sacrifice effect sequence unexpectedly lowered without diagnostic")
	}
}

func TestLowerDrawTriggerMaxTriggersPerTurn(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Draw Sentinel",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Creature — Wizard",
		OracleText: "Whenever you draw a card, you gain 1 life. This ability triggers only once each turn.",
		Colors:     []string{"U"},
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	if got := face.TriggeredAbilities[0].MaxTriggersPerTurn; got != 1 {
		t.Errorf("MaxTriggersPerTurn = %d, want 1", got)
	}
}

func TestLowerDrawTriggerTwiceEachTurn(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Draw Sentinel",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Creature — Wizard",
		OracleText: "Whenever you draw a card, you gain 1 life. This ability triggers only twice each turn.",
		Colors:     []string{"U"},
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	if got := face.TriggeredAbilities[0].MaxTriggersPerTurn; got != 2 {
		t.Errorf("MaxTriggersPerTurn = %d, want 2", got)
	}
}
