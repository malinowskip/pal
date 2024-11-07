package persistence

import (
	"pal/testutil"
	"testing"
)

func TestInitializeConversation(t *testing.T) {
	projectPath := t.TempDir()
	client, err := StartClient(projectPath)

	if err != nil {
		t.Error(err)
	}

	convo, err := client.InitializeConversation()

	if err != nil {
		t.Error(err)
	}

	if err != nil {
		t.Error(err)
	}

	if len(convo.Messages) != 0 {
		t.Error("A new conversation should have zero messages")
	}

}

func TestInsertMessageIntoConversation(t *testing.T) {
	projectPath := t.TempDir()
	client, err := StartClient(projectPath)

	if err != nil {
		t.Error(err)
	}

	convo, err := client.InitializeConversation()

	if err != nil {
		t.Error(err)
	}

	_, err = client.InsertMessageIntoConversation(convo.Id, "assistant", "bar")

	if err != nil {
		t.Error(err)
	}

	convo, err = client.FetchRecentConversation()
	if err != nil {
		t.Error(err)
	}

	expectedMessages := []Message{
		{Id: 1, Role: "assistant", Content: "bar"},
	}

	testutil.AssertDeepEquals(t, convo.Messages, expectedMessages)
	testutil.AssertDeepEquals(t, convo.Id, int64(1))
}

func TestWriteToMessage(t *testing.T) {
	projectPath := t.TempDir()
	client, err := StartClient(projectPath)

	if err != nil {
		t.Error(err)
	}

	convo, err := client.InitializeConversation()

	if err != nil {
		t.Error(err)
	}

	message, err := client.InsertMessageIntoConversation(convo.Id, "assistant", "")

	if err != nil {
		t.Error(err)
	}

	components := []string{"Hello,", " world", "!"}

	for _, text := range components {
		client.WriteToMessage(message.Id, text)
	}

	convo, err = client.FetchRecentConversation()

	if err != nil {
		t.Error(err)
	}

	testutil.AssertDeepEquals(t, convo.Messages[0].Content, "Hello, world!")
}

func TestFetchRecentConversationReturnsErrIfNoRecentConversations(t *testing.T) {
	projectPath := t.TempDir()
	client, err := StartClient(projectPath)

	if err != nil {
		t.Error(err)
	}

	_, err = client.FetchRecentConversation()

	if err == nil {
		t.Error("If there are no recent conversations, err should be returned.")
	}
}

func TestPruneOldConversations(t *testing.T) {
	projectPath := t.TempDir()
	client, err := StartClient(projectPath)

	countConversations := func() int {
		var count int
		client.Conn.QueryRow("select count(*) from conversations").Scan(&count)
		return count
	}

	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 20; i++ {
		conv, err := client.InitializeConversation()
		if err != nil {
			t.Error(err)
		}
		_, err = client.InsertMessageIntoConversation(conv.Id, "user", "Hello, world!")
		if err != nil {
			t.Error(err)
		}
	}

	testutil.AssertDeepEquals(t, countConversations(), 20)

	if err = client.PruneOldConversations(10); err != nil {
		t.Error(err)
	}

	testutil.AssertDeepEquals(t, countConversations(), 10)

	lastConversation, err := client.FetchRecentConversation()
	if err != nil {
		t.Error(err)
	}

	testutil.AssertDeepEquals(t, lastConversation.Id, int64(20))

	var orphanMessageCount int
	err = client.Conn.QueryRow(`
		select count(*) as count from messages m
		join conversations c on c.id = m.conversation_id
		where c.id is null
	`).Scan(&orphanMessageCount)

	if err != nil {
		t.Error(err)
	}

	testutil.AssertDeepEquals(t, orphanMessageCount, 0)
}
