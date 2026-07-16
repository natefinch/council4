package rules

import (
	"maps"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// PlayerObservation is the fog-of-war filtered view a player has of the game.
//
// It is a read-only window over the live *game.Game plus the observing player's
// ID, not a deep copy. Accessor methods expose only information the observing
// player is allowed to see: their own hand, all public zones, life totals,
// commander state, and the effective characteristics of battlefield
// permanents. Opponents' hand contents, library order, and the hidden identity
// of face-down permanents are never exposed.
//
// Agents must treat the observation as read-only; the returned view values are
// copies and do not allow mutating game state.
type PlayerObservation struct {
	g      *game.Game
	Player game.PlayerID
	Turn   TurnObservation
}

// TurnObservation describes the public turn state relevant to action choice.
type TurnObservation struct {
	TurnNumber     int
	ActivePlayer   game.PlayerID
	PriorityPlayer game.PlayerID
	Phase          game.Phase
	Step           game.Step
}

// CardView is a read-only view of a card in a publicly known position, such as
// the observing player's own hand or any player's graveyard, exile, or command
// zone.
type CardView struct {
	CardInstanceID id.ID
	Name           string
	Owner          game.PlayerID
	Types          []types.Card
	ManaValue      int
	// Colors is the card's colors (CR 105). Empty for colorless cards.
	Colors []color.Color
	// ProducesColors lists the colors of mana this card's mana abilities can
	// add. It is empty for cards with no mana ability and for sources that only
	// produce colorless mana, so an agent can sequence colored sources without
	// inspecting ability internals.
	ProducesColors []color.Color
	// ProducesMana reports that the card has a mana ability (a mana rock or mana
	// dork), including colorless-only sources, so an agent can value ramp without
	// inspecting ability internals.
	ProducesMana bool
	// RampsLand reports that casting or resolving the card puts a land onto the
	// battlefield (land ramp such as Rampant Growth or a land-fetching enters
	// trigger), derived from the scorable-effect IR, so an agent can prioritize
	// ramp.
	RampsLand bool
	// EntersTapped reports that the card always enters the battlefield tapped (an
	// unconditional tapland), so an agent can sequence a tapland into a turn it
	// does not need the mana. Conditional or pay-to-untap lands are not flagged,
	// since they may enter untapped.
	EntersTapped bool
}

// PlayerView is a read-only view of one player's public state. Hand and library
// are reported only as sizes; their contents and order are hidden for
// opponents (and the library order is hidden even for the observer).
type PlayerView struct {
	ID                game.PlayerID
	Life              int
	PoisonCounters    int
	HandSize          int
	LibrarySize       int
	GraveyardSize     int
	Eliminated        bool
	IsMonarch         bool
	HasInitiative     bool
	DungeonsCompleted int
	// InDungeon reports whether the player is currently in a dungeon; DungeonName
	// and DungeonRoom then name the dungeon and the player's current room. Dungeon
	// position is public information (CR 309.4).
	InDungeon           bool
	DungeonName         string
	DungeonRoom         string
	CommanderInstanceID id.ID
	CommanderCastCount  int
	CommanderDamage     map[id.ID]int
}

// PermanentView is a read-only view of one battlefield permanent with its
// current effective characteristics (after continuous effects, counters, and
// temporary modifiers). Face-down permanents report their public 2/2 nameless
// characteristics, never their hidden identity.
type PermanentView struct {
	ObjectID       id.ID
	CardInstanceID id.ID
	Name           string
	Controller     game.PlayerID
	Owner          game.PlayerID
	Power          int
	Toughness      int
	HasToughness   bool
	Tapped         bool
	SummoningSick  bool
	FaceDown       bool
	PhasedOut      bool
	Types          []types.Card
	Subtypes       []types.Sub
	// EntryChoices is a copy of the public persistent choices stored on the
	// permanent.
	EntryChoices map[game.ChoiceKey]game.ResolutionChoiceResult
	// ProducesMana reports whether this permanent has any mana ability, so an
	// agent can count its untapped mana sources.
	ProducesMana bool
	// ProducesColors lists the colors of mana this permanent's mana abilities
	// can add. It is empty for non-producers and for sources that only produce
	// colorless mana.
	ProducesColors []color.Color
	keywords       map[game.Keyword]bool
}

// HasKeyword reports whether the permanent currently has the given keyword.
func (v PermanentView) HasKeyword(keyword game.Keyword) bool {
	return v.keywords[keyword]
}

// NewObservation builds the fog-of-war observation that playerID has of g. It is
// the same read-only view the engine passes to agents, exposed so callers
// outside the priority loop (simulations, reports, and agent tests) can build an
// observation for a given player and game.
func NewObservation(g *game.Game, playerID game.PlayerID) PlayerObservation {
	return PlayerObservation{
		g:      g,
		Player: playerID,
		Turn: TurnObservation{
			TurnNumber:     g.Turn.TurnNumber,
			ActivePlayer:   g.Turn.ActivePlayer,
			PriorityPlayer: g.Turn.PriorityPlayer,
			Phase:          g.Turn.Phase,
			Step:           g.Turn.Step,
		},
	}
}

// StackObjectView is a read-only view of one spell or ability on the stack.
// A face-down spell reports its public characteristics only: FaceDown is true
// and Name is empty so its hidden identity is never exposed.
type StackObjectView struct {
	ID         id.ID
	Kind       game.StackObjectKind
	Controller game.PlayerID
	Name       string
	FaceDown   bool
	Targets    []game.Target
	// ManaValue, Types, and Colors describe a spell on the stack so an agent can
	// decide whether it is worth countering. They are unset for face-down spells
	// (hidden identity) and for non-spell stack objects (abilities).
	ManaValue int
	Types     []types.Card
	Colors    []color.Color
}

// Players returns a public view of every player, in seat order.
func (o PlayerObservation) Players() []PlayerView {
	views := make([]PlayerView, 0, len(o.g.Players))
	for _, player := range o.g.Players {
		if o.g.Mode == game.RunModeGoldfish && player.Eliminated {
			continue
		}
		views = append(views, playerView(player))
	}
	return views
}

// PlayerState returns the public view of a single player.
func (o PlayerObservation) PlayerState(playerID game.PlayerID) PlayerView {
	return playerView(o.g.Players[playerID])
}

// Life returns a player's current life total.
func (o PlayerObservation) Life(playerID game.PlayerID) int {
	return o.g.Players[playerID].Life
}

// Hand returns the observing player's own hand. Opponents' hands are never
// revealed; use PlayerState(p).HandSize for an opponent's hand count.
func (o PlayerObservation) Hand() []CardView {
	return o.cardViews(&o.g.Players[o.Player].Hand)
}

// Graveyard returns a player's graveyard, which is public.
func (o PlayerObservation) Graveyard(playerID game.PlayerID) []CardView {
	return o.cardViews(&o.g.Players[playerID].Graveyard)
}

// Exile returns a player's exiled cards. Face-down exiled cards (e.g. suspended
// or foretold) are reported without their hidden identity.
func (o PlayerObservation) Exile(playerID game.PlayerID) []CardView {
	return o.cardViews(&o.g.Players[playerID].Exile)
}

// CommandZone returns a player's command zone cards, which are public.
func (o PlayerObservation) CommandZone(playerID game.PlayerID) []CardView {
	return o.cardViews(&o.g.Players[playerID].CommandZone)
}

// LibraryTopRevealed returns the top card of a player's library when that player
// plays with the top card of their library revealed (Oracle of Mul Daya, Future
// Sight). It reports false when the player has no such effect or an empty
// library; the library's order is otherwise hidden from every observer.
func (o PlayerObservation) LibraryTopRevealed(playerID game.PlayerID) (CardView, bool) {
	if !playerPlaysWithTopCardRevealed(o.g, playerID) {
		return CardView{}, false
	}
	return o.libraryTopView(playerID)
}

// LibraryTopLookable returns the top card of the observing player's own library
// when they may privately look at it at any time ("You may look at the top card
// of your library any time.", Bolas's Citadel). It reports false for any other
// player or when the observer has no such effect, since the permission is
// private to its controller.
func (o PlayerObservation) LibraryTopLookable(playerID game.PlayerID) (CardView, bool) {
	if o.Player != playerID || !playerCanLookAtTopCardAnyTime(o.g, playerID) {
		return CardView{}, false
	}
	return o.libraryTopView(playerID)
}

// Battlefield returns a view of every permanent on the shared battlefield, with
// effective characteristics.
func (o PlayerObservation) Battlefield() []PermanentView {
	views := make([]PermanentView, 0, len(o.g.Battlefield))
	for _, permanent := range o.g.Battlefield {
		if permanent == nil {
			continue
		}
		views = append(views, o.permanentView(permanent))
	}
	return views
}

// Stack returns the stack from bottom (resolves last) to top (resolves next).
func (o PlayerObservation) Stack() []StackObjectView {
	objects := o.g.Stack.Objects()
	views := make([]StackObjectView, 0, len(objects))
	for _, obj := range objects {
		views = append(views, o.stackObjectView(obj))
	}
	return views
}

// libraryTopView returns a view of the top card of playerID's library, or false
// when the library is empty. Callers gate it on the appropriate visibility
// permission.
func (o PlayerObservation) libraryTopView(playerID game.PlayerID) (CardView, bool) {
	library := &o.g.Players[playerID].Library
	top, ok := library.Top()
	if !ok {
		return CardView{}, false
	}
	single := zone.New(library.Type)
	single.Add(top)
	views := o.cardViews(&single)
	if len(views) == 0 {
		return CardView{}, false
	}
	return views[0], true
}

func playerView(player *game.Player) PlayerView {
	var commanderDamage map[id.ID]int
	if len(player.CommanderDamage) > 0 {
		commanderDamage = make(map[id.ID]int, len(player.CommanderDamage))
		maps.Copy(commanderDamage, player.CommanderDamage)
	}
	view := PlayerView{
		ID:                  player.ID,
		Life:                player.Life,
		PoisonCounters:      player.PoisonCounters,
		HandSize:            player.Hand.Size(),
		LibrarySize:         player.Library.Size(),
		GraveyardSize:       player.Graveyard.Size(),
		Eliminated:          player.Eliminated,
		IsMonarch:           player.IsMonarch,
		HasInitiative:       player.HasInitiative,
		DungeonsCompleted:   player.DungeonsCompleted,
		CommanderInstanceID: player.CommanderInstanceID,
		CommanderCastCount:  player.CommanderCastCount,
		CommanderDamage:     commanderDamage,
	}
	if player.Dungeon.Exists {
		if def, ok := game.DungeonByID(player.Dungeon.Val.Dungeon); ok {
			view.InDungeon = true
			view.DungeonName = def.Name
			if room, ok := def.Room(player.Dungeon.Val.Room); ok {
				view.DungeonRoom = room.Name
			}
		}
	}
	return view
}

// cardViews builds CardViews for a zone whose contents are visible to the
// observer. Face-down cards are reported without their hidden identity.
func (o PlayerObservation) cardViews(z *zone.Zone) []CardView {
	views := make([]CardView, 0, z.Size())
	for _, cardID := range z.All() {
		view := CardView{CardInstanceID: cardID}
		if z.IsFaceDown(cardID) {
			views = append(views, view)
			continue
		}
		card, ok := o.g.GetCardInstance(cardID)
		if !ok {
			views = append(views, view)
			continue
		}
		view.Name = card.Def.Name
		view.Owner = card.Owner
		view.Types = append([]types.Card(nil), card.Def.Types...)
		view.ManaValue = card.Def.ManaValue()
		view.Colors = append([]color.Color(nil), card.Def.Colors...)
		view.ProducesColors = cardFaceManaColors(&card.Def.CardFace, commanderIdentityColors(o.g, card.Owner))
		view.ProducesMana = len(card.Def.ManaAbilities) > 0
		view.RampsLand = cardFaceRampsLand(&card.Def.CardFace)
		view.EntersTapped = cardFaceEntersTapped(&card.Def.CardFace)
		views = append(views, view)
	}
	return views
}

func (o PlayerObservation) permanentView(permanent *game.Permanent) PermanentView {
	values := effectivePermanentValues(o.g, permanent)
	var keywords map[game.Keyword]bool
	values.keywords.each(func(keyword game.Keyword) {
		if keywords == nil {
			keywords = make(map[game.Keyword]bool)
		}
		keywords[keyword] = true
	})
	producesMana, producesColors := abilitiesManaProduction(values.abilities, permanent.EntryChoices, commanderIdentityColors(o.g, effectiveController(o.g, permanent)))
	var entryChoices map[game.ChoiceKey]game.ResolutionChoiceResult
	if !permanent.FaceDown {
		entryChoices = maps.Clone(permanent.EntryChoices)
	}
	return PermanentView{
		ObjectID:       permanent.ObjectID,
		CardInstanceID: permanent.CardInstanceID,
		Name:           values.name,
		Controller:     effectiveController(o.g, permanent),
		Owner:          permanent.Owner,
		Power:          values.power,
		Toughness:      values.toughness,
		HasToughness:   values.toughnessOK,
		Tapped:         permanent.Tapped,
		SummoningSick:  permanent.SummoningSick,
		FaceDown:       permanent.FaceDown,
		PhasedOut:      permanent.PhasedOut,
		Types:          append([]types.Card(nil), values.types...),
		Subtypes:       append([]types.Sub(nil), values.subtypes...),
		EntryChoices:   entryChoices,
		ProducesMana:   producesMana,
		ProducesColors: producesColors,
		keywords:       keywords,
	}
}

func (o PlayerObservation) stackObjectView(obj *game.StackObject) StackObjectView {
	view := StackObjectView{
		ID:         obj.ID,
		Kind:       obj.Kind,
		Controller: obj.Controller,
		Name:       o.stackObjectName(obj),
		FaceDown:   obj.FaceDown,
		Targets:    append([]game.Target(nil), obj.Targets...),
	}
	if obj.FaceDown || obj.Kind != game.StackSpell {
		return view
	}
	if card, ok := o.g.GetCardInstance(obj.SourceID); ok {
		view.ManaValue = card.Def.ManaValue()
		view.Types = append([]types.Card(nil), card.Def.Types...)
		view.Colors = append([]color.Color(nil), card.Def.Colors...)
	}
	return view
}

// stackObjectName resolves a display name for a stack object from its source
// card or token definition, falling back to an empty string. A face-down spell
// keeps its real source card on the stack (CR 712 / morph), so its name is
// suppressed to avoid exposing the hidden identity.
func (o PlayerObservation) stackObjectName(obj *game.StackObject) string {
	if obj.FaceDown {
		return ""
	}
	if card, ok := o.g.GetCardInstance(obj.SourceID); ok {
		return card.Def.Name
	}
	if card, ok := o.g.GetCardInstance(obj.SourceCardID); ok {
		return card.Def.Name
	}
	if obj.SourceTokenDef != nil {
		return obj.SourceTokenDef.Name
	}
	return ""
}
