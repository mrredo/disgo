package discord

import (
	"fmt"
	"regexp"

	"github.com/disgoorg/snowflake/v2"
)

type MentionType struct {
	*regexp.Regexp
}

var (
	MentionTypeUser      = MentionType{regexp.MustCompile(`<@!?(\d+)>`)}
	MentionTypeRole      = MentionType{regexp.MustCompile(`<@&(\d+)>`)}
	MentionTypeChannel   = MentionType{regexp.MustCompile(`<#(\d+)>`)}
	MentionTypeEmoji     = MentionType{regexp.MustCompile(`<a?:(\w+):(\d+)>`)}
	MentionTypeTimestamp = MentionType{regexp.MustCompile(`<t:(?P<time>-?\d{1,17})(?::(?P<format>[tTdDfFR]))?>`)}
	MentionTypeHere      = MentionType{regexp.MustCompile(`@here`)}
	MentionTypeEveryone  = MentionType{regexp.MustCompile(`@everyone`)}
)

type Mentionable interface {
	fmt.Stringer
	Mention() string
}

func ChannelMention(id snowflake.ID) string {
	return fmt.Sprintf("<#%s>", id)
}

func UserTag(username string, discriminator string) string {
	return fmt.Sprintf("%s#%s", username, discriminator)
}

func UserMention(id snowflake.ID) string {
	return fmt.Sprintf("<@%s>", id)
}

func RoleMention(id snowflake.ID) string {
	return fmt.Sprintf("<@&%s>", id)
}

func EmojiMention(id snowflake.ID, name string) string {
	return fmt.Sprintf("<:%s:%s>", name, id)
}

func AnimatedEmojiMention(id snowflake.ID, name string) string {
	return fmt.Sprintf("<a:%s:%s>", name, id)
}

func TimestampMention(timestamp int64) string {
	return fmt.Sprintf("<t:%d>", timestamp)
}

func FormattedTimestampMention(timestamp int64, style TimestampStyle) string {
	return fmt.Sprintf("<t:%d:%s>", timestamp, style)
}
