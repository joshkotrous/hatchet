package slack

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/slack-go/slack"

	"github.com/hatchet-dev/hatchet/pkg/integrations"
)

type SlackIntegration struct {
	api    *slack.Client
	teamId string
}

func NewSlackIntegration(authToken string, teamId string, debug bool) *SlackIntegration {
	api := slack.New(authToken, slack.OptionDebug(debug))

	return &SlackIntegration{
		api:    api,
		teamId: teamId,
	}
}

func (s *SlackIntegration) GetId() string {
	return "slack"
}

func (s *SlackIntegration) Actions() []string {
	return []string{
		"create-channel",
		"send-message",
		"add-users-to-channel",
	}
}

func (s *SlackIntegration) ActionHandler(action string) any {
	switch action {
	case "create-channel":
		return s.createChannel
	case "add-users-to-channel":
		return s.addUsersToChannel
	case "send-message":
		return s.sendMessageToChannel
	default:
		return nil
	}
}

func (s *SlackIntegration) GetWebhooks() []integrations.IntegrationWebhook {
	return []integrations.IntegrationWebhook{}
}

type CreateChannelData struct {
	ChannelName string `json:"channelName"`
}

type CreateChannelOutput struct {
	ChannelId string `json:"channelId"`
}

// validateChannelName checks if the given channel name meets Slack's requirements
func validateChannelName(name string) error {
	// Check if the channel name is empty
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("channel name cannot be empty")
	}

	// Check length (Slack limits channel names to 80 characters)
	if len(name) > 80 {
		return fmt.Errorf("channel name cannot exceed 80 characters")
	}

	// Check if the name is lowercase
	if name != strings.ToLower(name) {
		return fmt.Errorf("channel name must be lowercase")
	}

	// Check if the first character is valid (must be alphanumeric)
	if len(name) > 0 && (name[0] == '-' || name[0] == '_') {
		return fmt.Errorf("channel name cannot start with a hyphen or underscore")
	}

	// Check for valid characters (alphanumeric, hyphens, and underscores)
	validChars := regexp.MustCompile(`^[a-z0-9_-]+$`)
	if !validChars.MatchString(name) {
		return fmt.Errorf("channel name can only contain lowercase letters, numbers, hyphens, and underscores")
	}

	return nil
}

func (s *SlackIntegration) createChannel(ctx context.Context, data *CreateChannelData) (*CreateChannelOutput, error) {
	// Validate the channel name
	if err := validateChannelName(data.ChannelName); err != nil {
		return nil, fmt.Errorf("invalid channel name: %w", err)
	}

	channel, err := s.api.CreateConversation(slack.CreateConversationParams{
		IsPrivate:   true,
		ChannelName: data.ChannelName,
		TeamID:      s.teamId,
	})

	if err != nil {
		return nil, fmt.Errorf("error creating slack channel: %w", err)
	}

	return &CreateChannelOutput{
		ChannelId: channel.ID,
	}, nil
}

type AddUsersToChannelData struct {
	ChannelID string   `json:"channelId"`
	UserIDs   []string `json:"userIds"`
}

type AddUsersToChannelOutput struct{}

func (s *SlackIntegration) addUsersToChannel(ctx context.Context, data *AddUsersToChannelData) error {
	_, err := s.api.InviteUsersToConversation(data.ChannelID, data.UserIDs...)

	if err != nil {
		return fmt.Errorf("error adding users to slack channel: %w", err)
	}

	return nil
}

type SendMessageToChannelData struct {
	ChannelID string `json:"channelId"`
	Message   string `json:"message"`
}

func (s *SlackIntegration) sendMessageToChannel(ctx context.Context, data *SendMessageToChannelData) error {
	_, _, err := s.api.PostMessage(data.ChannelID, slack.MsgOptionText(data.Message, false))

	if err != nil {
		return fmt.Errorf("error sending message to slack channel: %w", err)
	}

	return nil
}