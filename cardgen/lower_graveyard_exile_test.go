package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerTargetedGraveyardExile(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Wretch",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Exile target card from a graveyard.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v, want one", mode.Targets)
	}
	target := mode.Targets[0]
	if target.MinTargets != 1 || target.MaxTargets != 1 ||
		target.Allow != game.TargetAllowCard || target.TargetZone != zone.Graveyard ||
		!target.Selection.Val.Empty() {
		t.Fatalf("target = %#v", target)
	}
	move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
	if !ok {
		t.Fatalf("primitive = %T, want game.MoveCard", mode.Sequence[0].Primitive)
	}
	if move.Card.Kind != game.CardReferenceTarget || move.Card.TargetIndex != 0 ||
		move.FromZone != zone.Graveyard || move.Destination != zone.Exile {
		t.Fatalf("move = %#v", move)
	}
}

func TestLowerTargetedGraveyardExileTypedFromOpponent(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mummy",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Exile target creature card from an opponent's graveyard.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	target := mode.Targets[0]
	if target.Allow != game.TargetAllowCard || target.TargetZone != zone.Graveyard {
		t.Fatalf("target = %#v", target)
	}
	selection := target.Selection.Val
	if !slices.Equal(selection.RequiredTypes, []types.Card{types.Creature}) ||
		selection.Controller != game.ControllerOpponent {
		t.Fatalf("selection = %#v", selection)
	}
	move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
	if !ok || move.FromZone != zone.Graveyard || move.Destination != zone.Exile {
		t.Fatalf("move = %#v", mode.Sequence[0].Primitive)
	}
}

func TestLowerTargetedGraveyardExileUpToOne(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Gryff",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Exile up to one target card from a graveyard.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	target := mode.Targets[0]
	if target.MinTargets != 0 || target.MaxTargets != 1 ||
		target.Allow != game.TargetAllowCard || target.TargetZone != zone.Graveyard {
		t.Fatalf("target = %#v", target)
	}
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(mode.Sequence))
	}
	move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
	if !ok || move.Card.TargetIndex != 0 || move.Destination != zone.Exile {
		t.Fatalf("move = %#v", mode.Sequence[0].Primitive)
	}
}

// TestLowerTargetedGraveyardExileSingleGraveyard lowers the "from a single
// graveyard" variant, which restricts every chosen card to one shared graveyard.
// The constraint lowers to TargetSpec.SameGraveyard on the card target so the
// runtime enumerates only target sets drawn from a single graveyard.
func TestLowerTargetedGraveyardExileSingleGraveyard(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Decay",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Exile up to three target cards from a single graveyard.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v, want one", mode.Targets)
	}
	target := mode.Targets[0]
	if target.MinTargets != 0 || target.MaxTargets != 3 ||
		target.Allow != game.TargetAllowCard || target.TargetZone != zone.Graveyard ||
		!target.SameGraveyard || !target.Selection.Val.Empty() {
		t.Fatalf("target = %#v", target)
	}
	move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
	if !ok || move.Card.Kind != game.CardReferenceTarget ||
		move.FromZone != zone.Graveyard || move.Destination != zone.Exile {
		t.Fatalf("move = %#v", mode.Sequence[0].Primitive)
	}
}

// TestLowerTargetedGraveyardExileAnyGraveyards lowers the plural "from
// graveyards" variant, which lets each chosen card lie in a different graveyard.
// Unlike the "from a single graveyard" form it carries no SameGraveyard
// restriction, so the runtime may draw targets from distinct graveyards.
func TestLowerTargetedGraveyardExileAnyGraveyards(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Macabre",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Exile up to two target cards from graveyards.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v, want one", mode.Targets)
	}
	target := mode.Targets[0]
	if target.MinTargets != 0 || target.MaxTargets != 2 ||
		target.Allow != game.TargetAllowCard || target.TargetZone != zone.Graveyard ||
		target.SameGraveyard || !target.Selection.Val.Empty() {
		t.Fatalf("target = %#v", target)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want two moves", mode.Sequence)
	}
	for i := range mode.Sequence {
		move, ok := mode.Sequence[i].Primitive.(game.MoveCard)
		if !ok || move.Card.Kind != game.CardReferenceTarget ||
			move.Card.TargetIndex != i ||
			move.FromZone != zone.Graveyard || move.Destination != zone.Exile {
			t.Fatalf("move[%d] = %#v", i, mode.Sequence[i].Primitive)
		}
	}
}

// TestGenerateSingleGraveyardExileSource proves the SameGraveyard restriction
// survives rendering into the generated card source (not just the in-memory
// lowering), so the executable card actually enforces one shared graveyard.
func TestGenerateSingleGraveyardExileSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Decay",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{B}",
		OracleText: "Exile up to three target cards from a single graveyard.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(spaceCollapsed(source), spaceCollapsed("SameGraveyard: true,")) {
		t.Fatalf("source missing SameGraveyard: true:\n%s", source)
	}
}

// TestLowerPlayerGraveyardExile lowers the whole-graveyard exile "Exile target
// player's graveyard." to a single target-player TargetSpec plus the player-zone
// group form of MoveCard.
func TestLowerPlayerGraveyardExile(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bog",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Exile target player's graveyard.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v, want one", mode.Targets)
	}
	target := mode.Targets[0]
	if target.MinTargets != 1 || target.MaxTargets != 1 ||
		target.Allow != game.TargetAllowPlayer ||
		target.Selection.Val.Player != game.PlayerAny {
		t.Fatalf("target = %#v", target)
	}
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want one", mode.Sequence)
	}
	move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
	if !ok {
		t.Fatalf("primitive = %T, want game.MoveCard", mode.Sequence[0].Primitive)
	}
	if move.Player.Kind() != game.PlayerReferenceTargetPlayer || move.Player.TargetIndex() != 0 ||
		move.Card.Kind != game.CardReferenceNone ||
		move.FromZone != zone.Graveyard || move.Destination != zone.Exile {
		t.Fatalf("move = %#v", move)
	}
}

// TestLowerOpponentGraveyardExile lowers the opponent-restricted variant, which
// only differs in carrying the opponent target predicate.
func TestLowerOpponentGraveyardExile(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Nightmare",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Exile target opponent's graveyard.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	target := mode.Targets[0]
	if target.Allow != game.TargetAllowPlayer || target.Selection.Val.Player != game.PlayerOpponent {
		t.Fatalf("target = %#v", target)
	}
	move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
	if !ok || move.Player.Kind() != game.PlayerReferenceTargetPlayer ||
		move.FromZone != zone.Graveyard || move.Destination != zone.Exile {
		t.Fatalf("move = %#v", mode.Sequence[0].Primitive)
	}
}

// TestLowerPlayerGraveyardExileBojukaBog confirms the anchor land lowers all
// three of its abilities — enters tapped, the enters-trigger graveyard exile,
// and the mana ability — with the graveyard exile as a target-player MoveCard.
func TestLowerPlayerGraveyardExileBojukaBog(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Bojuka Bog",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "This land enters tapped.\nWhen this land enters, exile target player's graveyard.\n{T}: Add {B}.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want one", len(face.TriggeredAbilities))
	}
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("mana abilities = %d, want one", len(face.ManaAbilities))
	}
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("replacement abilities = %d, want one (enters tapped)", len(face.ReplacementAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
	if !ok || move.Player.Kind() != game.PlayerReferenceTargetPlayer ||
		move.FromZone != zone.Graveyard || move.Destination != zone.Exile {
		t.Fatalf("trigger move = %#v", mode.Sequence[0].Primitive)
	}
}

// TestLowerAllGraveyardExile lowers the non-targeted whole-graveyard wipe to the
// player-group form of MoveCard over every player's graveyard, carrying no
// target. The "each player's graveyard." synonym lowers identically.
func TestLowerAllGraveyardExile(t *testing.T) {
	t.Parallel()
	for _, text := range []string{"Exile all graveyards.", "Exile each player's graveyard."} {
		face := lowerSingleFace(t, &ScryfallCard{
			Name:       "Test Wipe",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: text,
		})
		mode := face.SpellAbility.Val.Modes[0]
		if len(mode.Targets) != 0 {
			t.Fatalf("targets = %#v, want none for %q", mode.Targets, text)
		}
		move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
		if !ok || move.PlayerGroup.Kind != game.PlayerGroupReferenceAllPlayers ||
			move.Player.Kind() != game.PlayerReferenceNone ||
			move.Card.Kind != game.CardReferenceNone ||
			move.FromZone != zone.Graveyard || move.Destination != zone.Exile {
			t.Fatalf("move = %#v for %q", mode.Sequence[0].Primitive, text)
		}
	}
}

// TestLowerControllerGraveyardChoiceExile lowers the non-target "Exile a <filter>
// card from your graveyard" wording to a single game.ExileFromGraveyard
// instruction whose Selection carries the type filter and the controller scope.
func TestLowerControllerGraveyardChoiceExile(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Reclaimer",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Exile a creature card from your graveyard.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 {
		t.Fatalf("targets = %#v, want none", mode.Targets)
	}
	exile, ok := mode.Sequence[0].Primitive.(game.ChooseFromZone)
	if !ok {
		t.Fatalf("primitive = %T, want game.ChooseFromZone", mode.Sequence[0].Primitive)
	}
	if exile.Player.Kind() != game.PlayerReferenceController ||
		exile.Quantity.IsDynamic() || exile.Quantity.Value() != 1 {
		t.Fatalf("exile = %#v", exile)
	}
	if !slices.Equal(exile.Filter.RequiredTypes, []types.Card{types.Creature}) ||
		exile.Filter.Controller != game.ControllerYou {
		t.Fatalf("selection = %#v", exile.Filter)
	}
}

// TestLowerOptionalGraveyardExileThenGatedEffect lowers the "you may exile a
// <filter> card from your graveyard. If you do, <Y>." wrapper (Masked Vandal):
// the exile X-action is marked Optional and publishes its result, and the gated
// Y-effect resolves only when a card was exiled.
func TestLowerOptionalGraveyardExileThenGatedEffect(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Vandal",
		Layout:     "normal",
		TypeLine:   "Creature",
		OracleText: "When this creature enters, you may exile a creature card from your graveyard. If you do, draw a card.",
	})
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", mode.Sequence)
	}
	exile, ok := mode.Sequence[0].Primitive.(game.ChooseFromZone)
	if !ok {
		t.Fatalf("primitive[0] = %T, want game.ChooseFromZone", mode.Sequence[0].Primitive)
	}
	if !slices.Equal(exile.Filter.RequiredTypes, []types.Card{types.Creature}) {
		t.Fatalf("selection = %#v", exile.Filter)
	}
	if !mode.Sequence[0].Optional ||
		mode.Sequence[0].PublishResult != game.ResultKey("if-you-do") {
		t.Fatalf("exile instruction = %#v, want optional publishing if-you-do", mode.Sequence[0])
	}
	gate := mode.Sequence[1].ResultGate
	if !gate.Exists || gate.Val.Key != game.ResultKey("if-you-do") ||
		gate.Val.Succeeded != game.TriTrue {
		t.Fatalf("draw gate = %#v, want gated on if-you-do success", mode.Sequence[1].ResultGate)
	}
}

// TestLowerControllerGraveyardChoiceExileFailsClosedTargeted documents that the
// targeted "exile target ... from your graveyard" form keeps lowering to a card
// target (lowerTargetedGraveyardExile) rather than the choose-at-resolution
// game.ExileFromGraveyard primitive.
func TestLowerControllerGraveyardChoiceExileFailsClosedTargeted(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Targeter",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Exile target creature card from your graveyard.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if _, ok := mode.Sequence[0].Primitive.(game.ChooseFromZone); ok {
		t.Fatal("targeted graveyard exile must not lower to game.ChooseFromZone")
	}
}

// TestLowerPlayerGraveyardExileFailsClosed documents that the referenced-player
// wording stays unsupported rather than lowering to a whole-graveyard wipe,
// since its semantics are not represented.
func TestLowerPlayerGraveyardExileFailsClosed(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Exile that player's graveyard.",
	} {
		_, diagnostics := lowerExecutableFaces(&ScryfallCard{
			Name:       "Test Closed",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: text,
		})
		if len(diagnostics) == 0 {
			t.Fatalf("expected unsupported diagnostic for %q", text)
		}
	}
}
