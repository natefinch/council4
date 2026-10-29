package rules

import (
	"maps"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PlayerObservation is the fog-of-war filtered view a player has of the game.
//
// It is a read-only window over the live *game.Game plus the observing player's
// ID, not a deep copy. Accessor methods expose only information the viewing
// player is allowed to see: their own and controlled players' hands, public zones, life totals,
// commander state, and the effective characteristics of battlefield
// permanents. Opponents' hand contents, library order, and face-down identities
// are hidden; the controller of a face-down permanent receives its private card
// views as allowed by CR 708.6.
//
// Agents must treat the observation as read-only; the returned view values are
// copies and do not allow mutating game state.
type PlayerObservation struct {
	g      *game.Game
	Player game.PlayerID
	// DecisionMaker is the viewer's seat. Player remains the decision's actor.
	DecisionMaker game.PlayerID
	viewer        game.PlayerID
	Turn          TurnObservation
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
	// EvaluationUnknown means authored resolving effects require runtime search,
	// not a trustworthy static zero. Identity remains visible only via normal
	// CardView permissions; this flag never reveals another player's hidden hand.
	EvaluationUnknown bool
	CardInstanceID    id.ID
	Name              string
	Owner             game.PlayerID
	Types             []types.Card
	ManaValue         int
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
// temporary modifiers). Face-down permanents report only their public
// characteristics; their controller also receives the underlying card views in
// FaceDownCards.
type PermanentView struct {
	ObjectID       id.ID
	CardInstanceID id.ID
	// CommanderInstanceIDs lists every commander card represented by this
	// permanent, including cards merged beneath its top component. Commander
	// designation is public even while the permanent is face down.
	CommanderInstanceIDs []id.ID
	// FaceDownCards contains the underlying card views only for the player who
	// controls this face-down permanent. Other observers receive an empty slice.
	FaceDownCards []CardView
	Name          string
	Controller    game.PlayerID
	Owner         game.PlayerID
	Power         int
	Toughness     int
	HasToughness  bool
	Tapped        bool
	SummoningSick bool
	FaceDown      bool
	PhasedOut     bool
	Types         []types.Card
	Subtypes      []types.Sub
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
	// PendingDelayedTriggers lists the pending delayed triggered abilities whose
	// source is this permanent or that are bound to it (CR 603.7), such as an
	// end-step sacrifice of a token or a dashed creature's return to hand.
	// Delayed triggers are created by public spells and abilities, so every
	// observer sees them; hidden face-down identities stay redacted.
	PendingDelayedTriggers []DelayedTriggerView
	keywords               map[game.Keyword]bool
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
		g:             g,
		Player:        playerID,
		DecisionMaker: playerID,
		viewer:        playerID,
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
	// FaceDownCard is private information for the spell's controller and any
	// controller of that player. Public Name/Types remain face-down values.
	FaceDownCard opt.V[CardView]
	Targets      []game.Target
	// ManaValue, Types, and Colors describe a spell on the stack so an agent can
	// decide whether it is worth countering. They are unset for face-down spells
	// (hidden identity) and for non-spell stack objects (abilities).
	ManaValue int
	Types     []types.Card
	Colors    []color.Color
	// Text is the public rules text of an ability on the stack (CR 602.2a,
	// 603.3), including a fired delayed trigger. It is empty for spells and for
	// abilities whose definition records no text. An ability of a face-down
	// permanent the viewer may not look at keeps only text granted to it
	// publicly (such as ward), never its hidden card's text.
	Text string
	// SourceObjectID is the battlefield permanent an ability came from while
	// that permanent remains on the battlefield (CR 113.7), else zero.
	SourceObjectID id.ID
	// BoundObjectIDs and BoundCardID identify the permanents and card a fired
	// delayed trigger captured when it was created (CR 603.7c), so two delayed
	// triggers from same-name sources can be told apart. BoundCardID is zero
	// unless that card is still in the same zone and visible to the viewer.
	BoundObjectIDs []id.ID
	BoundCardID    id.ID
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

// Hand returns the decision actor's hand, if visible to the viewer. HandOf also
// exposes other hands the viewer may see; unrelated hands remain private.
func (o PlayerObservation) Hand() []CardView {
	return o.HandOf(o.Player)
}

// HandOf exposes only a hand the viewer may see, including controlled players.
// A controlled player never gains reciprocal access to their controller's hand.
func (o PlayerObservation) HandOf(player game.PlayerID) []CardView {
	if !o.g.CanSeePlayerInformation(o.viewer, player) {
		return nil
	}
	return o.cardViews(&o.g.Players[player].Hand)
}

// LibraryLookable exposes a library only during an actual library search.
// Turn control alone does not make the library's contents or order visible.
func (o PlayerObservation) LibraryLookable(player game.PlayerID) []CardView {
	if !o.g.LookingAtLibrary(o.viewer, player) {
		return nil
	}
	return o.cardViews(&o.g.Players[player].Library)
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

// LibraryTopLookable returns the top card of a visible player's library
// when they may privately look at it at any time ("You may look at the top card
// of your library any time.", Bolas's Citadel). Turn control shares this private
// permission with the controller, but never with uninvolved players.
func (o PlayerObservation) LibraryTopLookable(playerID game.PlayerID) (CardView, bool) {
	if !o.g.CanSeePlayerInformation(o.viewer, playerID) || !playerCanLookAtTopCardAnyTime(o.g, playerID) {
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

// FidelityLimitations is public policy/evaluation metadata, never a hidden-card
// census. It is invariant across seats and hypothetical library/hand samples.
func (PlayerObservation) FidelityLimitations() []Limitation {
	return ActionGenerationLimitations()
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
		card, ok := o.g.GetCardInstance(cardID)
		if z.IsFaceDown(cardID) && !o.canLookAtFaceDownExile(z, card) {
			views = append(views, view)
			continue
		}
		if !ok {
			views = append(views, view)
			continue
		}
		view.Name = card.Def.Name
		view.EvaluationUnknown = len(card.Def.ProgramReferences()) != 0
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
		ObjectID:               permanent.ObjectID,
		CardInstanceID:         permanent.CardInstanceID,
		CommanderInstanceIDs:   permanentCommanderInstanceIDs(o.g, permanent),
		FaceDownCards:          o.faceDownCardViews(permanent),
		Name:                   values.name,
		Controller:             effectiveController(o.g, permanent),
		Owner:                  permanent.Owner,
		Power:                  values.power,
		Toughness:              values.toughness,
		HasToughness:           values.toughnessOK,
		Tapped:                 permanent.Tapped,
		SummoningSick:          permanent.SummoningSick,
		FaceDown:               permanent.FaceDown,
		PhasedOut:              permanent.PhasedOut,
		Types:                  append([]types.Card(nil), values.types...),
		Subtypes:               append([]types.Sub(nil), values.subtypes...),
		EntryChoices:           entryChoices,
		ProducesMana:           producesMana,
		ProducesColors:         producesColors,
		PendingDelayedTriggers: o.permanentDelayedTriggers(permanent.ObjectID),
		keywords:               keywords,
	}
}

func (o PlayerObservation) faceDownCardViews(permanent *game.Permanent) []CardView {
	if !o.g.CanSeePlayerInformation(o.viewer, effectiveController(o.g, permanent)) {
		return nil
	}
	var views []CardView
	if permanent.FaceDown {
		if def, ok := physicalPermanentFaceDef(o.g, permanent, permanent.FaceDownFace); ok {
			views = append(views, o.cardViewForDef(permanent.CardInstanceID, permanent.Owner, def))
		}
	}
	for _, component := range permanent.MergedCards {
		if !permanent.FaceDown && !component.FaceDown {
			continue
		}
		face := component.Face
		if component.FaceDown {
			face = component.FaceDownFace
		}
		var def *game.CardDef
		var ok bool
		if component.TokenDef != nil {
			def, ok = component.TokenDef.FaceDefView(face)
		} else if card, exists := o.g.GetCardInstance(component.CardInstanceID); exists {
			def, ok = cardFaceDef(card, face)
		}
		if ok {
			views = append(views, o.cardViewForDef(component.CardInstanceID, component.Owner, def))
		}
	}
	return views
}

func physicalPermanentFaceDef(g *game.Game, permanent *game.Permanent, face game.FaceIndex) (*game.CardDef, bool) {
	if permanent.Token {
		if permanent.TokenDef == nil {
			return nil, false
		}
		return permanent.TokenDef.FaceDefView(face)
	}
	card, ok := g.GetCardInstance(permanent.CardInstanceID)
	if !ok {
		return nil, false
	}
	return cardFaceDef(card, face)
}

func (o PlayerObservation) cardViewForDef(cardID id.ID, owner game.PlayerID, def *game.CardDef) CardView {
	return CardView{
		EvaluationUnknown: len(def.ProgramReferences()) != 0,
		CardInstanceID:    cardID,
		Name:              def.Name,
		Owner:             owner,
		Types:             append([]types.Card(nil), def.Types...),
		ManaValue:         def.ManaValue(),
		Colors:            append([]color.Color(nil), def.Colors...),
		ProducesColors:    cardFaceManaColors(&def.CardFace, commanderIdentityColors(o.g, owner)),
		ProducesMana:      len(def.ManaAbilities) > 0,
		RampsLand:         cardFaceRampsLand(&def.CardFace),
		EntersTapped:      cardFaceEntersTapped(&def.CardFace),
	}
}

func (o PlayerObservation) canLookAtFaceDownExile(z *zone.Zone, card *game.CardInstance) bool {
	if z.Type != zone.Exile || card == nil || !o.g.CanSeePlayerInformation(o.viewer, card.Owner) {
		return false
	}
	// Foretell explicitly grants its owner permission to look (CR 702.143).
	_, foretold := o.g.ForetoldCards[card.ID]
	return foretold
}

func permanentCommanderInstanceIDs(g *game.Game, permanent *game.Permanent) []id.ID {
	var ids []id.ID
	if isCommanderCardID(g, permanent.CardInstanceID) {
		ids = append(ids, permanent.CardInstanceID)
	}
	for _, component := range permanent.MergedCards {
		if isCommanderCardID(g, component.CardInstanceID) {
			ids = append(ids, component.CardInstanceID)
		}
	}
	return ids
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
	if obj.Kind != game.StackSpell {
		o.fillAbilityStackObjectView(&view, obj)
	}
	if obj.FaceDown && o.g.CanSeePlayerInformation(o.viewer, obj.Controller) {
		if card, exists := o.g.GetCardInstance(obj.SourceID); exists {
			view.FaceDownCard = opt.Val(o.cardViewForDef(card.ID, card.Owner, card.Def))
		}
	}
	if obj.FaceDown || obj.Kind != game.StackSpell {
		return view
	}
	if def, ok := stackObjectSpellDef(o.g, obj); ok {
		view.ManaValue = stackManaValue(def, obj.XValue)
		view.Types = stackObjectCardTypes(obj, def)
		view.Colors = spellColors(def)
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
	if obj.Kind == game.StackSpell {
		if def, ok := stackObjectSpellDef(o.g, obj); ok {
			return def.Name
		}
	} else if o.hiddenFaceDownPermanent(obj.SourceID) {
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
