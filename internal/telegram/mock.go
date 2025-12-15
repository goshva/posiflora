package telegram

import "context"

type MockClient struct {
	Sent []string
}

func (m *MockClient) SendMessage(ctx context.Context, botToken, chatID, text string) error {
	m.Sent = append(m.Sent, text)
	return nil
}
