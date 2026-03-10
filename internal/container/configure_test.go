package container

import (
	"errors"
	"reflect"
	"testing"
)

func TestSupervisorctlCmd(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "start openclaw uses explicit config file",
			args: []string{"start", "openclaw"},
			want: []string{"supervisorctl", "-c", "/etc/supervisor/supervisord.conf", "start", "openclaw"},
		},
		{
			name: "stop openclaw uses explicit config file",
			args: []string{"stop", "openclaw"},
			want: []string{"supervisorctl", "-c", "/etc/supervisor/supervisord.conf", "stop", "openclaw"},
		},
		{
			name: "restart openclaw uses explicit config file",
			args: []string{"restart", "openclaw"},
			want: []string{"supervisorctl", "-c", "/etc/supervisor/supervisord.conf", "restart", "openclaw"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := supervisorctlCmd(tt.args...); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("supervisorctlCmd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateConfigureParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		params  ConfigureParams
		wantErr error
	}{
		{
			name: "provider required",
			params: ConfigureParams{
				APIKey: "test-key",
			},
			wantErr: errors.New("provider is required"),
		},
		{
			name: "api key required",
			params: ConfigureParams{
				Provider: "anthropic",
			},
			wantErr: errors.New("api key is required"),
		},
		{
			name: "channel requires token",
			params: ConfigureParams{
				Provider: "anthropic",
				APIKey:   "test-key",
				Channel:  "telegram",
			},
			wantErr: errors.New("channel token is required when channel is set"),
		},
		{
			name: "token requires channel",
			params: ConfigureParams{
				Provider:     "anthropic",
				APIKey:       "test-key",
				ChannelToken: "123456:ABC",
			},
			wantErr: errors.New("channel is required when channel token is set"),
		},
		{
			name: "base provider config is valid",
			params: ConfigureParams{
				Provider: "anthropic",
				APIKey:   "test-key",
			},
		},
		{
			name: "channel config is valid",
			params: ConfigureParams{
				Provider:     "anthropic",
				APIKey:       "test-key",
				Channel:      "telegram",
				ChannelToken: "123456:ABC",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateConfigureParams(tt.params)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("ValidateConfigureParams() unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("ValidateConfigureParams() error = nil, want %q", tt.wantErr)
			}
			if err.Error() != tt.wantErr.Error() {
				t.Fatalf("ValidateConfigureParams() error = %q, want %q", err.Error(), tt.wantErr.Error())
			}
		})
	}
}

func TestConfiguredChannel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		channels  map[string]openClawChannelConfig
		wantName  string
		wantToken string
	}{
		{
			name: "ignores enabled plugin without token",
			channels: map[string]openClawChannelConfig{
				"telegram": {Token: ""},
			},
		},
		{
			name: "prefers bot token when present",
			channels: map[string]openClawChannelConfig{
				"telegram": {BotToken: "123456:ABC"},
			},
			wantName:  "telegram",
			wantToken: "123456:ABC",
		},
		{
			name: "falls back to token field",
			channels: map[string]openClawChannelConfig{
				"discord": {Token: "discord-token"},
			},
			wantName:  "discord",
			wantToken: "discord-token",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotName, gotToken := configuredChannel(tt.channels)
			if gotName != tt.wantName || gotToken != tt.wantToken {
				t.Fatalf("configuredChannel() = (%q, %q), want (%q, %q)", gotName, gotToken, tt.wantName, tt.wantToken)
			}
		})
	}
}
