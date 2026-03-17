package container

import "testing"

func TestChannelPolicyStepsSlackUsesBooleanAllowBots(t *testing.T) {
	steps := channelPolicySteps("slack", "channels.slack")

	var found bool
	for _, step := range steps {
		if step.path != "channels.slack.allowBots" {
			continue
		}
		found = true
		if step.value != "true" {
			t.Fatalf("expected Slack allowBots value true, got %q", step.value)
		}
		if !step.strictJSON {
			t.Fatal("expected Slack allowBots to use strict JSON")
		}
	}
	if !found {
		t.Fatal("expected Slack allowBots policy step")
	}
}

func TestChannelPolicyStepsDiscordUsesMentionsAllowBots(t *testing.T) {
	steps := channelPolicySteps("discord", "channels.discord")

	var found bool
	for _, step := range steps {
		if step.path != "channels.discord.allowBots" {
			continue
		}
		found = true
		if step.value != "mentions" {
			t.Fatalf("expected Discord allowBots value mentions, got %q", step.value)
		}
		if step.strictJSON {
			t.Fatal("did not expect Discord allowBots to use strict JSON")
		}
	}
	if !found {
		t.Fatal("expected Discord allowBots policy step")
	}
}

func TestChannelPolicyStepsTelegramIncludesGroupAllowFrom(t *testing.T) {
	steps := channelPolicySteps("telegram", "channels.telegram")

	var found bool
	for _, step := range steps {
		if step.path != "channels.telegram.groupAllowFrom" {
			continue
		}
		found = true
		if step.value != `["*"]` {
			t.Fatalf("expected Telegram groupAllowFrom wildcard, got %q", step.value)
		}
		if !step.strictJSON {
			t.Fatal("expected Telegram groupAllowFrom to use strict JSON")
		}
	}
	if !found {
		t.Fatal("expected Telegram groupAllowFrom policy step")
	}
}
