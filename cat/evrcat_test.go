package cat

import (
	"testing"
)

func Test_replaceSymbols(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"Symbol with known token",
			args{
				"[05-03-2024] [06:17:28]: [LEVELLOAD] Loading level '0xAC360E41E4EDE056'",
			},
			"[05-03-2024] [06:17:28]: [LEVELLOAD] Loading level 'mnu_master'",
		},
		{
			"Symbol with unknown token",
			args{
				"[05-03-2024] [06:17:28]: [LEVELLOAD] Loading level '0xFF360E41E4EDE056'",
			},
			"[05-03-2024] [06:17:28]: [LEVELLOAD] Loading level '0xFF360E41E4EDE056'",
		},
		{
			"updated symbols",
			args{
				"emote '0x667feb110569d3a3'",
			},
			"emote 'emote_vrml_a'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewEVRCat().ReplaceHashes(tt.args.line); got != tt.want {
				t.Errorf("replaceSymbols() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_replaceTokens(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"Token conversion",
			args{
				"mnu_master",
			},
			"0xac360e41e4ede056",
		},
		{
			"blank line",
			args{
				"",
			},
			"",
		},
		{
			"spaces",
			args{
				"   ",
			},
			"",
		},
		{
			"uppercase",
			args{
				"0xAC360E41E4EDE056",
			},
			"0xac360e41e4ede056",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewEVRCat().ReplaceTokens(tt.args.line, false, nil); got != tt.want {
				t.Errorf("replaceTokens() = %v, want %v", got, tt.want)
			}
		})
	}
}
