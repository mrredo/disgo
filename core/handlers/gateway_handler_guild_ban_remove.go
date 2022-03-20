package handlers

import (
	"github.com/DisgoOrg/disgo/core"
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
)

// gatewayHandlerGuildBanRemove handles discord.GatewayEventTypeGuildBanRemove
type gatewayHandlerGuildBanRemove struct{}

// EventType returns the discord.GatewayEventType
func (h *gatewayHandlerGuildBanRemove) EventType() discord.GatewayEventType {
	return discord.GatewayEventTypeGuildBanRemove
}

// New constructs a new payload receiver for the raw gateway event
func (h *gatewayHandlerGuildBanRemove) New() interface{} {
	return &discord.GuildBanRemoveGatewayEvent{}
}

// HandleGatewayEvent handles the specific raw gateway event
func (h *gatewayHandlerGuildBanRemove) HandleGatewayEvent(bot core.Bot, sequenceNumber discord.GatewaySequence, v interface{}) {
	payload := *v.(*discord.GuildBanRemoveGatewayEvent)

	bot.Caches().Users().Put(payload.User.ID, payload.User)

	bot.EventManager().Dispatch(&events.GuildUnbanEvent{
		GenericEvent: events.NewGenericEvent(bot, sequenceNumber),
		GuildID:      payload.GuildID,
		User:         payload.User,
	})
}
