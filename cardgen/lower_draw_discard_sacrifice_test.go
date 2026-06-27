package cardgen

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
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

// TestLowerUpkeepDrawIfGreatestToughness verifies that Abzan Beastmaster's
// "draw a card if you control the creature with the greatest toughness or tied
// for the greatest toughness" gates the upkeep draw on the greatest-toughness
// condition.
func TestLowerUpkeepDrawIfGreatestToughness(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Beastmaster",
		Layout:     "normal",
		ManaCost:   "{2}{G}",
		TypeLine:   "Creature — Dog Shaman",
		OracleText: "At the beginning of your upkeep, draw a card if you control the creature with the greatest toughness or tied for the greatest toughness.",
		Colors:     []string{"G"},
		Power:      new("2"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	instruction := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0]
	if _, ok := instruction.Primitive.(game.Draw); !ok {
		t.Fatalf("primitive = %#v, want Draw", instruction.Primitive)
	}
	if !instruction.Condition.Exists ||
		!instruction.Condition.Val.Condition.Val.ControllerControlsGreatestToughnessCreature {
		t.Fatalf("draw was not gated on the greatest-toughness condition: %#v", instruction.Condition)
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

func TestLowerDrawTriggerEventPlayerMayPayFailureCreatesTreasure(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Smothering Tithe",
		Layout:     "normal",
		ManaCost:   "{3}{W}",
		TypeLine:   "Enchantment",
		OracleText: "Whenever an opponent draws a card, that player may pay {2}. If the player doesn't, you create a Treasure token.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Optional ||
		trigger.Trigger.Pattern.Event != game.EventCardDrawn ||
		trigger.Trigger.Pattern.Player != game.TriggerPlayerOpponent {
		t.Fatalf("trigger = %#v", trigger)
	}
	sequence := trigger.Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence = %#v, want payment and consequence", sequence)
	}
	pay, ok := sequence[0].Primitive.(game.Pay)
	if !ok || !pay.Payment.Payer.Exists ||
		pay.Payment.Payer.Val != game.EventPlayerReference() ||
		!pay.Payment.ManaCost.Exists ||
		!slices.Equal(pay.Payment.ManaCost.Val, cost.Mana{cost.O(2)}) {
		t.Fatalf("payment = %#v", sequence[0])
	}
	if sequence[0].PublishResult != unlessPaidResultKey {
		t.Fatalf("payment result = %q", sequence[0].PublishResult)
	}
	consequence := sequence[1]
	if consequence.Optional || !consequence.ResultGate.Exists ||
		consequence.ResultGate.Val.Key != unlessPaidResultKey ||
		consequence.ResultGate.Val.Succeeded != game.TriFalse {
		t.Fatalf("consequence envelope = %#v", consequence)
	}
	create, ok := consequence.Primitive.(game.CreateToken)
	if !ok || create.Amount != game.Fixed(1) || create.Recipient.Exists {
		t.Fatalf("consequence = %#v, want one controller Treasure", consequence.Primitive)
	}
	token, ok := create.Source.TokenDefRef()
	if !ok || token.Name != string(types.Treasure) ||
		len(token.ManaAbilities) != 1 ||
		len(token.ActivatedAbilities) != 0 {
		t.Fatalf("Treasure definition = %#v", token)
	}
	manaAbility := token.ManaAbilities[0]
	if len(manaAbility.AdditionalCosts) != 2 ||
		manaAbility.AdditionalCosts[0].Kind != cost.AdditionalTap ||
		manaAbility.AdditionalCosts[1].Kind != cost.AdditionalSacrificeSource ||
		len(manaAbility.Content.Modes) != 1 ||
		len(manaAbility.Content.Modes[0].Sequence) != 2 {
		t.Fatalf("Treasure mana ability = %#v", manaAbility)
	}
	if _, ok := manaAbility.Content.Modes[0].Sequence[0].Primitive.(game.Choose); !ok {
		t.Fatalf("Treasure instruction 0 = %T, want color choice", manaAbility.Content.Modes[0].Sequence[0].Primitive)
	}
	if _, ok := manaAbility.Content.Modes[0].Sequence[1].Primitive.(game.AddMana); !ok {
		t.Fatalf("Treasure instruction 1 = %T, want AddMana", manaAbility.Content.Modes[0].Sequence[1].Primitive)
	}
}

func TestLowerDrawTriggerEventPlayerMayPayFailureRejectsUnsafeForms(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		"Whenever an opponent draws a card, that player may pay 2 life. If the player doesn't, you create a Treasure token.",
		"Whenever an opponent draws a card, that player may sacrifice a creature. If the player doesn't, you create a Treasure token.",
		"Whenever an opponent draws a card, you may pay {2}. If you don't, you create a Treasure token.",
		"Whenever an opponent draws a card, that player may pay {X}. If the player doesn't, you create a Treasure token.",
		"Whenever an opponent draws a card, that player may pay {2}. If the player doesn't, target player creates a Treasure token.",
		"Whenever an opponent draws a card, that player may pay {2}. If the player doesn't, each opponent creates a Treasure token.",
		"Whenever an opponent draws a card, that player may pay {2}. If the player doesn't, instead you create a Treasure token.",
		"Whenever an opponent draws a card, that player may pay {2}. If the player doesn't, you create a Treasure token and gain 1 life.",
		"Whenever an opponent draws a card, that player may pay {2}. If the player doesn't, you create a Treasure token. This ability triggers only once each turn.",
	} {
		t.Run(oracle, func(t *testing.T) {
			t.Parallel()
			faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Unsafe Tithe",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: oracle,
			})
			if len(diagnostics) == 0 {
				t.Fatalf("expected unsupported diagnostic for %q", oracle)
			}
			if len(faces) > 0 && len(faces[0].TriggeredAbilities) != 0 {
				t.Fatalf("unexpected supported trigger for %q", oracle)
			}
		})
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
	if got := trigger.InterveningCondition.Val.Aggregates; len(got) != 1 || got[0].Aggregate != game.AggregateControllerLife || got[0].Op != compare.GreaterOrEqual || got[0].Value != 5 {
		t.Errorf("condition = %+v, want controller life >= 5", trigger.InterveningCondition.Val)
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

func TestLowerSacrificeSpellEachOtherPlayerCreature(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Grave Pact",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Whenever a creature you control dies, each other player sacrifices a creature of their choice.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
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

// TestLowerSacrificeSpellThatPlayerEventPlayerChoice covers the "that player
// sacrifices ... of their choice" edict on a phase trigger (Sheoldred,
// Whispering One): the player named by the triggering event chooses, lowered to
// game.EventPlayerReference.
func TestLowerSacrificeSpellThatPlayerEventPlayerChoice(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Praetor",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Praetor",
		OracleText: "At the beginning of each opponent's upkeep, that player sacrifices a creature of their choice.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 0 {
		t.Fatalf("targets = %d, want none", len(mode.Targets))
	}
	prim, ok := mode.Sequence[0].Primitive.(game.SacrificePermanents)
	if !ok {
		t.Fatalf("primitive = %T, want game.SacrificePermanents", mode.Sequence[0].Primitive)
	}
	if prim.Player.Kind() != game.PlayerReferenceEventPlayer {
		t.Fatalf("player = %v, want event player", prim.Player.Kind())
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

// TestLowerSacrificeSpellEachPlayerAllColored covers All Is Dust's mass edict
// "Each player sacrifices all permanents they control that are one or more
// colors." It lowers to a single SacrificePermanents with All set, the
// all-players group, and a Colored permanent selection so colorless permanents
// survive. The bounded chosen-amount path is untouched (no Amount).
func TestLowerSacrificeSpellEachPlayerAllColored(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "All Is Dust",
		Layout:     "normal",
		TypeLine:   "Kindred Sorcery — Eldrazi",
		OracleText: "Each player sacrifices all permanents they control that are one or more colors.",
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
	if !prim.All {
		t.Fatal("prim.All = false, want true")
	}
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceAllPlayers {
		t.Fatalf("player group = %v, want all players", prim.PlayerGroup.Kind)
	}
	if prim.Player.Kind() != game.PlayerReferenceNone {
		t.Fatalf("player = %v, want none", prim.Player.Kind())
	}
	if prim.Amount.IsDynamic() || prim.Amount.Value() != 0 {
		t.Fatalf("amount = %#v, want zero", prim.Amount)
	}
	if !prim.Selection.Colored {
		t.Fatalf("selection = %#v, want Colored", prim.Selection)
	}
	if len(prim.Selection.RequiredTypes) != 0 {
		t.Fatalf("selection RequiredTypes = %#v, want any permanent", prim.Selection.RequiredTypes)
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

// TestLowerEdictWithInabilityDiscardFallback covers Plaguecrafter's edict plus
// "Each player who can't discards a card." rider folding into one
// SacrificePermanents instruction carrying a discard fallback.
func TestLowerEdictWithInabilityDiscardFallback(t *testing.T) {
	t.Parallel()
	power, toughness := "3", "2"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Plaguecrafter",
		Layout:     "normal",
		TypeLine:   "Creature — Human Shaman",
		OracleText: "When this creature enters, each player sacrifices a creature or planeswalker of their choice. Each player who can't discards a card.",
		Power:      &power,
		Toughness:  &toughness,
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence length = %d, want 1 folded instruction", len(mode.Sequence))
	}
	prim, ok := mode.Sequence[0].Primitive.(game.SacrificePermanents)
	if !ok {
		t.Fatalf("primitive = %T, want game.SacrificePermanents", mode.Sequence[0].Primitive)
	}
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceAllPlayers {
		t.Fatalf("player group = %v, want all players", prim.PlayerGroup.Kind)
	}
	if prim.Fallback.Kind != game.SacrificeFallbackDiscard {
		t.Fatalf("fallback kind = %v, want SacrificeFallbackDiscard", prim.Fallback.Kind)
	}
	if prim.Fallback.Amount.Value() != 1 {
		t.Fatalf("fallback amount = %d, want 1", prim.Fallback.Amount.Value())
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

func TestLowerSacrificeSpellEachPlayerNonTokenCreature(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Accursed Edict",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Each player sacrifices a nontoken creature of their choice.",
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
	if !slices.Equal(prim.Selection.RequiredTypes, []types.Card{types.Creature}) {
		t.Fatalf("selection = %#v, want creature filter", prim.Selection)
	}
	if !prim.Selection.NonToken {
		t.Fatalf("selection = %#v, want NonToken", prim.Selection)
	}
}

func TestLowerSacrificeSpellCreatureOrPlaneswalker(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Plaguecrafter Edict",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Each player sacrifices a creature or planeswalker of their choice.",
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
	if !slices.Equal(prim.Selection.RequiredTypesAny, []types.Card{types.Creature, types.Planeswalker}) {
		t.Fatalf("selection = %#v, want creature-or-planeswalker filter", prim.Selection)
	}
}

func TestLowerAccursedMarauderEnterTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Accursed Marauder",
		Layout:     "normal",
		ManaCost:   "{1}{B}",
		TypeLine:   "Creature — Zombie",
		OracleText: "When this creature enters, each player sacrifices a nontoken creature of their choice.",
		Colors:     []string{"B"},
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want one", len(face.TriggeredAbilities))
	}
	prim, ok := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.SacrificePermanents)
	if !ok {
		t.Fatalf("primitive = %T, want game.SacrificePermanents", face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive)
	}
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceAllPlayers {
		t.Fatalf("player group = %v, want all players", prim.PlayerGroup.Kind)
	}
	if !slices.Equal(prim.Selection.RequiredTypes, []types.Card{types.Creature}) || !prim.Selection.NonToken {
		t.Fatalf("selection = %#v, want nontoken creature filter", prim.Selection)
	}
}

func TestLowerSacrificeSpellRejectsTappedQualifier(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Tapped Edict",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Each player sacrifices a tapped creature of their choice.",
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("tapped-creature sacrifice unexpectedly lowered without diagnostic")
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

func TestLowerSacrificeSpellControllerSelfSacrifice(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Sacrifice a creature.",
		"You sacrifice a creature.",
		"Sacrifice two permanents.",
	} {
		source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:       "Forced Tribute",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: oracleText,
		}, "c")
		if err != nil {
			t.Fatal(err)
		}
		if len(diagnostics) != 0 {
			t.Fatalf("controller self-sacrifice %q unexpectedly failed: %v", oracleText, diagnostics)
		}
		if !strings.Contains(source, "game.SacrificePermanents") ||
			!strings.Contains(source, "game.ControllerReference()") {
			t.Fatalf("controller self-sacrifice %q did not lower to a controller SacrificePermanents:\n%s", oracleText, source)
		}
	}
}

func TestLowerSacrificeSpellRejectsUnknownActorPattern(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Forced Tribute",
		Layout:   "normal",
		TypeLine: "Sorcery",
		// A single unspecified player actor is not a supported actor pattern.
		OracleText: "A player sacrifices a creature.",
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
		Name:       "Reprocess",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Sacrifice any number of artifacts, creatures, and/or lands. Draw a card for each permanent sacrificed this way.",
	}, "r")
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

func TestLowerDiscardTriggerDrawForEachCardDiscarded(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Discard Scholar",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Creature — Wizard",
		OracleText: "Whenever you discard one or more cards, draw a card for each card discarded this way.",
		Colors:     []string{"U"},
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Event != game.EventCardDiscarded {
		t.Fatalf("Pattern.Event = %v, want EventCardDiscarded", ta.Trigger.Pattern.Event)
	}
	if !ta.Trigger.Pattern.OneOrMore {
		t.Error("Pattern.OneOrMore = false, want true")
	}
	draw, ok := ta.Content.Modes[0].Sequence[0].Primitive.(game.Draw)
	if !ok {
		t.Fatalf("primitive = %+v, want game.Draw", ta.Content.Modes[0].Sequence[0].Primitive)
	}
	dynamic := draw.Amount.DynamicAmount()
	if !dynamic.Exists {
		t.Fatalf("draw.Amount = %+v, want dynamic amount", draw.Amount)
	}
	if dynamic.Val.Kind != game.DynamicAmountEventCardCount {
		t.Fatalf("draw.Amount kind = %v, want DynamicAmountEventCardCount", dynamic.Val.Kind)
	}
	if dynamic.Val.Multiplier != 1 {
		t.Fatalf("draw.Amount multiplier = %d, want 1", dynamic.Val.Multiplier)
	}
}

func TestLowerDiscardThisWayCountRejectedInSpell(t *testing.T) {
	t.Parallel()
	// Outside a draw/discard trigger there is no triggering card count, so the
	// "for each card discarded this way" amount must stay unsupported.
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Stray Draw",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Draw a card for each card discarded this way.",
	}, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("draw scaled by discard count unexpectedly lowered outside a trigger")
	}
}

func TestLowerDrawForEachCounterQualifiedCount(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Counter Scholar",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Draw a card for each creature you control with a +1/+1 counter on it.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	draw, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.Draw)
	if !ok {
		t.Fatalf("primitive = %+v, want game.Draw", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	dynamic := draw.Amount.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountCountSelector || dynamic.Val.Multiplier != 1 {
		t.Fatalf("draw.Amount dynamic = %+v, want count selector x1", dynamic)
	}
	selection := dynamic.Val.Group.Selection()
	if selection.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want ControllerYou", selection.Controller)
	}
	if len(selection.RequiredTypes) != 1 || selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("required types = %v, want [Creature]", selection.RequiredTypes)
	}
	if !selection.MatchCounter || selection.RequiredCounter != counter.PlusOnePlusOne {
		t.Fatalf("RequiredCounter = (%v,%v), want +1/+1", selection.MatchCounter, selection.RequiredCounter)
	}
}

func TestLowerDrawForEachPlainCountHasNoCounter(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Plain Scholar",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Draw a card for each creature you control.",
	})
	draw, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.Draw)
	if !ok {
		t.Fatalf("primitive = %+v, want game.Draw", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	if draw.Amount.DynamicAmount().Val.Group.Selection().MatchCounter {
		t.Fatal("RequiredCounter set for a count subject without a counter qualifier")
	}
}

func TestLowerSacrificeThenSearchSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Sac Ramp",
		Layout:     "normal",
		ManaCost:   "{2}{G}",
		TypeLine:   "Sorcery",
		OracleText: "Sacrifice a land. Search your library for up to two basic land cards, put them onto the battlefield tapped, then shuffle.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(mode.Sequence))
	}
	if _, ok := mode.Sequence[0].Primitive.(game.SacrificePermanents); !ok {
		t.Fatalf("first primitive = %#v, want SacrificePermanents", mode.Sequence[0].Primitive)
	}
	search, ok := mode.Sequence[1].Primitive.(game.Search)
	if !ok {
		t.Fatalf("second primitive = %#v, want Search", mode.Sequence[1].Primitive)
	}
	if !search.Spec.EntersTapped {
		t.Error("search spec should enter the battlefield tapped")
	}
	if mode.Sequence[1].Condition.Exists {
		t.Error("unconditional search should carry no gate condition")
	}
}

func TestLowerSacrificeThenConditionalInsteadSearchSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Entish Ramp",
		Layout:     "normal",
		ManaCost:   "{2}{G}",
		TypeLine:   "Sorcery",
		OracleText: "Sacrifice a land. Search your library for up to two basic land cards, put them onto the battlefield tapped, then shuffle. If you control a creature with power 4 or greater, instead search your library for up to three basic land cards, put them onto the battlefield tapped, then shuffle.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 3 {
		t.Fatalf("sequence length = %d, want 3", len(mode.Sequence))
	}
	base, ok := mode.Sequence[1].Primitive.(game.Search)
	if !ok {
		t.Fatalf("base primitive = %#v, want Search", mode.Sequence[1].Primitive)
	}
	if got := base.Amount; got != game.Fixed(2) {
		t.Errorf("base search amount = %#v, want Fixed(2)", got)
	}
	if !mode.Sequence[1].Condition.Exists || !mode.Sequence[1].Condition.Val.Condition.Val.Negate {
		t.Error("base search should be gated on the negated condition")
	}
	instead, ok := mode.Sequence[2].Primitive.(game.Search)
	if !ok {
		t.Fatalf("instead primitive = %#v, want Search", mode.Sequence[2].Primitive)
	}
	if got := instead.Amount; got != game.Fixed(3) {
		t.Errorf("instead search amount = %#v, want Fixed(3)", got)
	}
	if !mode.Sequence[2].Condition.Exists || mode.Sequence[2].Condition.Val.Condition.Val.Negate {
		t.Error("instead search should be gated on the non-negated condition")
	}
}

// TestLowerSacrificeSpellTokenSubtype verifies that "Sacrifice a <token
// subtype>." (Treasure, Food, ...) lowers the controller-sacrifice choice with
// a SubtypesAny selection filter rather than a card-type filter, so the runtime
// edict matches only permanents carrying that artifact-token subtype.
func TestLowerSacrificeSpellTokenSubtype(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Treasure Edict",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Sacrifice a Treasure.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not found")
	}
	mode := face.SpellAbility.Val.Modes[0]
	prim, ok := mode.Sequence[0].Primitive.(game.SacrificePermanents)
	if !ok {
		t.Fatalf("primitive = %T, want game.SacrificePermanents", mode.Sequence[0].Primitive)
	}
	if prim.Amount.Value() != 1 {
		t.Fatalf("amount = %d, want 1", prim.Amount.Value())
	}
	if !slices.Equal(prim.Selection.SubtypesAny, []types.Sub{types.Treasure}) {
		t.Fatalf("selection = %#v, want Treasure subtype filter", prim.Selection)
	}
	if len(prim.Selection.RequiredTypes) != 0 {
		t.Fatalf("selection RequiredTypes = %#v, want none for a subtype-only edict", prim.Selection.RequiredTypes)
	}
}

// TestLowerOptionalSacrificeTokenSubtype verifies the token-subtype sacrifice
// composes as the optional X action in "You may sacrifice a <token subtype>. If
// you do, <Y>.": the sacrifice is optional and publishes its result, and the
// benefit is gated on it.
func TestLowerOptionalSacrificeTokenSubtype(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Optional Food Sac",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "You may sacrifice a Food. If you do, draw a card.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not found")
	}
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", sequence)
	}
	prim, ok := sequence[0].Primitive.(game.SacrificePermanents)
	if !ok {
		t.Fatalf("instruction[0] = %T, want game.SacrificePermanents", sequence[0].Primitive)
	}
	if !slices.Equal(prim.Selection.SubtypesAny, []types.Sub{types.Food}) {
		t.Fatalf("selection = %#v, want Food subtype filter", prim.Selection)
	}
	if !sequence[0].Optional || sequence[0].PublishResult == "" {
		t.Fatalf("sacrifice must be optional and publish a result: %#v", sequence[0])
	}
	if !sequence[1].ResultGate.Exists || sequence[1].ResultGate.Val.Succeeded != game.TriTrue {
		t.Fatalf("draw must be gated on the sacrifice result: %#v", sequence[1])
	}
}
