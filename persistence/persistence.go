package persistence

import (
	"fmt"
)

type Conversation struct {
	Id       int64
	Messages []Message
}

type Message struct {
	Id      int64
	Role    string
	Content string
}

// Creates an empty conversation in the database.
func (c *DatabaseClient) InitializeConversation() (
	Conversation,
	error,
) {
	var convo Conversation

	_, err := c.Conn.Exec(
		"insert into conversations default values",
	)

	if err != nil {
		return convo, err
	}

	convo, err = c.FetchRecentConversation()

	if err != nil {
		return convo, err
	}

	return convo, nil
}

// Simply fetches the conversation that was created most recently.
func (c *DatabaseClient) FetchRecentConversation() (
	Conversation,
	error,
) {
	var convo Conversation

	row := c.Conn.QueryRow(
		`select max(id) from conversations`,
	)

	var conversationId *int64
	if err := row.Scan(&conversationId); err != nil {
		return convo, err
	}

	if conversationId == nil {
		return convo, fmt.Errorf("No conversations.")
	}

	convo.Id = *conversationId

	messageRows, err := c.Conn.Query(`
		select
			id,
			role,
			content
		from messages where conversation_id = ?
	`, conversationId)

	if err != nil {
		return convo, err
	}

	for {
		if messageRows.Next() == false {
			break
		}

		var messageId *int64
		var role string
		var content string

		if err = messageRows.Scan(&messageId, &role, &content); err != nil {
			return convo, err
		}

		convo.Messages = append(convo.Messages, Message{
			Id:      *messageId,
			Role:    role,
			Content: content,
		})

	}

	return convo, nil
}

func (c *DatabaseClient) InsertMessageIntoConversation(
	conversationId int64,
	role string,
	content string,
) (Message, error) {
	var message Message

	result, err := c.Conn.Exec(`
		insert into messages(conversation_id, role, content) values(?, ?, ?)
	`, conversationId, role, content)

	if err != nil {
		return message, err
	}

	messageId, err := result.LastInsertId()

	if err != nil {
		return message, err
	}

	row := c.Conn.QueryRow("select id, role, content from messages where id = ?", messageId)

	var theId int64
	var theRole string
	var theContent string

	if err = row.Scan(&theId, &theRole, &theContent); err != nil {
		return message, err
	}

	message = Message{
		Id:      theId,
		Role:    theRole,
		Content: theContent,
	}

	return message, nil
}

// Extends the existing content of a message with the provided text (used for
// recording streaming responses from an LLM chat).
func (c *DatabaseClient) WriteToMessage(messageId int64, text string) error {
	_, err := c.Conn.Exec(
		"update messages set content = concat(content, ?) where id = ?",
		text,
		messageId,
	)

	return err
}

func (c *DatabaseClient) PruneOldConversations(maxHistory int) error {
	_, err := c.Conn.Exec(`
		delete from conversations
		where id not in (
			select id from conversations
			order by id desc
			limit ?
		)
	`, maxHistory)

	if err != nil {
		return err
	}

	return nil
}
