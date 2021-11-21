package discord

// UserFlags defines certain flags/badges a user can have (https://discord.com/developers/docs/resources/user#user-object-user-flags)
type UserFlags int

// All UserFlags
//goland:noinspection GoUnusedConst
const (
	UserFlagDiscordEmployee UserFlags = 1 << iota
	UserFlagPartneredServerOwner
	UserFlagHypeSquadEvents
	UserFlagBugHunterLevel1
	UserFlagHouseBravery
	UserFlagHouseBrilliance
	UserFlagHouseBalance
	UserFlagEarlySupporter
	UserFlagTeamUser
	UserFlagBugHunterLevel2
	UserFlagVerifiedBot
	UserFlagEarlyVerifiedBotDeveloper
	UserFlagDiscordCertifiedModerator
	UserFlagNone UserFlags = 0
)

var _ Mentionable = (*User)(nil)

// User is a struct for interacting with discord's users
type User struct {
	ID            Snowflake `json:"id"`
	Username      string    `json:"username"`
	Discriminator string    `json:"discriminator"`
	Avatar        *string   `json:"avatar"`
	Banner        *string   `json:"banner"`
	AccentColor   *int      `json:"accent_color"`
	IsBot         bool      `json:"bot"`
	IsSystem      bool      `json:"system"`
	PublicFlags   UserFlags `json:"public_flags"`
}

func (u User) String() string {
	return userMention(u.ID)
}

func (u User) Mention() string {
	return u.String()
}

func (u User) Tag() string {
	return userTag(u.Username, u.Discriminator)
}

// OAuth2User represents a full User returned by the oauth2 endpoints
type OAuth2User struct {
	User
	MfaEnabled  *bool   `json:"mfa_enabled"`
	Locale      *string `json:"locale"`
	Verified    *bool   `json:"verified"`
	Email       *string `json:"email"`
	Flags       *int    `json:"flags"`
	PremiumType *int    `json:"premium_type"`
}

// SelfUserUpdate is the payload used to update the OAuth2User
type SelfUserUpdate struct {
	Username string    `json:"username"`
	Avatar   *NullIcon `json:"avatar"`
}
