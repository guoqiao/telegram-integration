package memogram

import (
	"testing"

	"github.com/go-telegram/bot/models"
	v1pb "github.com/usememos/memos/proto/gen/api/v1"
)

func TestSearchResultMessage(t *testing.T) {
	tests := []struct {
		name       string
		serverAddr string
		profile    *v1pb.InstanceProfile
		wantURL    string
	}{
		{
			name:       "public instance URL takes precedence",
			serverAddr: "http://internal:8081",
			profile:    &v1pb.InstanceProfile{InstanceUrl: "https://notes.example.com/"},
			wantURL:    "https://notes.example.com/memos/RfdGm9XSmDPcWWUEdqFeps",
		},
		{
			name:       "server URL without instance profile",
			serverAddr: "https://notes.example.com/",
			wantURL:    "https://notes.example.com/memos/RfdGm9XSmDPcWWUEdqFeps",
		},
		{
			name:       "empty instance URL falls back to server",
			serverAddr: "localhost:8081",
			profile:    &v1pb.InstanceProfile{},
			wantURL:    "http://localhost:8081/memos/RfdGm9XSmDPcWWUEdqFeps",
		},
		{
			name:       "legacy DNS server address",
			serverAddr: "dns:localhost:8081",
			wantURL:    "http://localhost:8081/memos/RfdGm9XSmDPcWWUEdqFeps",
		},
		{
			name:       "instance hosted under a path",
			serverAddr: "http://internal:8081",
			profile:    &v1pb.InstanceProfile{InstanceUrl: "https://example.com/notes/"},
			wantURL:    "https://example.com/notes/memos/RfdGm9XSmDPcWWUEdqFeps",
		},
	}

	memo := &v1pb.Memo{
		Name:    "memos/RfdGm9XSmDPcWWUEdqFeps",
		Content: "☕ Delonghi <Specialista> & Coffee Machine\n![label](/file/attachments/425HJAgffUtvuEFEa7Dolk)",
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				config:          &Config{ServerAddr: tt.serverAddr},
				instanceProfile: tt.profile,
			}
			message := s.searchResultMessage(123, memo)
			if message.ChatID != int64(123) {
				t.Fatalf("unexpected chat ID: %v", message.ChatID)
			}
			if want := memo.Name + "\n" + memo.Content; message.Text != want {
				t.Fatalf("unexpected message text: got %q, want %q", message.Text, want)
			}
			if message.ParseMode != "" {
				t.Fatalf("note content should be sent without markup parsing, got %q", message.ParseMode)
			}
			if len(message.Entities) != 1 {
				t.Fatalf("expected one link entity, got %v", message.Entities)
			}
			entity := message.Entities[0]
			if entity.Type != models.MessageEntityTypeTextLink || entity.Offset != 0 || entity.Length != len(memo.Name) || entity.URL != tt.wantURL {
				t.Fatalf("unexpected link entity: %+v", entity)
			}
		})
	}
}
