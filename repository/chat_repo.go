package repository

import (
	"fmt"
	"shareway/infra/db/migration"
	"shareway/schemas"
	"strings"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ChatRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewChatRepository(db *gorm.DB, redis *redis.Client) IChatRepository {
	return &ChatRepository{
		db:    db,
		redis: redis,
	}
}

type IChatRepository interface {
	SendMessage(req schemas.SendMessageRequest, userID uuid.UUID) (migration.Chat, error)
	GetChatRoomByUserIDs(userID1, userID2 uuid.UUID) (migration.Room, error)
	UploadImage(req schemas.SendImageRequest, userID uuid.UUID, imageURL string) (migration.Chat, error)
	GetAllChatRooms(userID uuid.UUID) ([]migration.Room, error)
	GetChatMessages(req schemas.GetChatMessagesRequest, userID uuid.UUID) ([]migration.Chat, error)
	UpdateCallStatus(req schemas.UpdateCallStatusRequest, userID uuid.UUID) (migration.Chat, error)
	InitiateCall(req schemas.InitiateCallRequest, userID uuid.UUID) (migration.Chat, error)
}

// GetChatRoomByUserIDs fetches a chat room by the user IDs
func (r *ChatRepository) GetChatRoomByUserIDs(userID1, userID2 uuid.UUID) (migration.Room, error) {
	var room migration.Room

	// Ensure user IDs are in a consistent order
	if userID1.String() > userID2.String() {
		userID1, userID2 = userID2, userID1
	}

	// Check if the chat room already exists
	err := r.db.Model(&migration.Room{}).
		Where("user1_id = ? AND user2_id = ?", userID1, userID2).
		Or("user1_id = ? AND user2_id = ?", userID2, userID1).
		First(&room).Error

	if err != nil {
		return migration.Room{}, err
	}

	return room, err
}

// SendMessage sends a message to a chat room
func (r *ChatRepository) SendMessage(req schemas.SendMessageRequest, userID uuid.UUID) (migration.Chat, error) {

	// Retrieve the chat room
	chat, err := r.GetChatRoomByUserIDs(userID, req.ReceiverID)
	if err != nil {
		return migration.Chat{}, err
	}

	// Create a new chat message
	newChat := migration.Chat{
		RoomID:      chat.ID,
		SenderID:    userID,
		ReceiverID:  req.ReceiverID,
		Message:     req.Message,
		MessageType: "text",
	}

	// Save the chat message
	err = r.db.Create(&newChat).Error
	if err != nil {
		return migration.Chat{}, err
	}

	// Update the chat room lastest message
	// Update the chat room with the last message ID and message content
	if err := r.db.Model(&migration.Room{}).Where("id = ?", chat.ID).Updates(map[string]interface{}{
		"last_message_id":   newChat.ID,
		"last_message_text": newChat.Message,
		"last_message_at":   newChat.CreatedAt,
	}).Error; err != nil {
		return migration.Chat{}, err
	}

	return newChat, nil
}

// UploadImage uploads an image to a chat room
func (r *ChatRepository) UploadImage(req schemas.SendImageRequest, userID uuid.UUID, imageURL string) (migration.Chat, error) {
	// Parse the receiver ID
	receiverID, err := uuid.Parse(req.ReceiverID)
	if err != nil {
		return migration.Chat{}, err
	}
	// Retrieve the chat room
	chat, err := r.GetChatRoomByUserIDs(userID, receiverID)
	if err != nil {
		return migration.Chat{}, err
	}

	// Create a new chat message
	newChat := migration.Chat{
		RoomID:      chat.ID,
		SenderID:    userID,
		ReceiverID:  receiverID,
		Message:     imageURL,
		MessageType: "image",
	}

	// Save the chat message
	err = r.db.Create(&newChat).Error
	if err != nil {
		return migration.Chat{}, err
	}

	// Update the chat room lastest message
	// Update the chat room with the last message ID and message content
	if err := r.db.Model(&migration.Room{}).Where("id = ?", chat.ID).Updates(map[string]interface{}{
		"last_message_id":   newChat.ID,
		"last_message_text": "Hình ảnh", // Display "Hình ảnh" for image messages not to show the image URL
		"last_message_at":   newChat.CreatedAt,
	}).Error; err != nil {
		return migration.Chat{}, err
	}

	return newChat, nil
}

// GetAllChatRooms fetches all chat rooms for a user
func (r *ChatRepository) GetAllChatRooms(userID uuid.UUID) ([]migration.Room, error) {
	var rooms []migration.Room

	// Fetch all chat rooms for the user
	err := r.db.Model(&migration.Room{}).
		Where("user1_id = ?", userID).
		Or("user2_id = ?", userID).
		Find(&rooms).Error

	if err != nil {
		return nil, err
	}

	return rooms, nil
}

// GetChatMessages fetches all messages in a chat room
func (r *ChatRepository) GetChatMessages(req schemas.GetChatMessagesRequest, userID uuid.UUID) ([]migration.Chat, error) {
	var messages []migration.Chat

	// Fetch all messages in the chat room
	err := r.db.Model(&migration.Chat{}).
		Where("room_id = ?", req.ChatRoomID).
		Order("created_at ASC"). // Order by created_at in ascending order to get older messages first
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	return messages, nil
}

// UpdateCallStatus updates the call status in a chat room
func (r *ChatRepository) UpdateCallStatus(req schemas.UpdateCallStatusRequest, userID uuid.UUID) (migration.Chat, error) {
	// Get the current message from the chat room use call_id
	var chat migration.Chat
	err := r.db.First(&chat, req.CallID).Error
	if err != nil {
		return migration.Chat{}, err
	}

	// Update the chat message wit necessary information
	chat.MessageType = req.CallType // call or missed_call

	// Handle call duration if provided from second to (giờ phút giây)
	if req.Duration > 0 {
		hours := req.Duration / 3600
		minutes := (req.Duration % 3600) / 60
		seconds := req.Duration % 60

		var timeStr string
		if hours > 0 {
			timeStr += fmt.Sprintf("%d giờ ", hours)
		}
		if minutes > 0 {
			timeStr += fmt.Sprintf("%d phút ", minutes)
		}
		if seconds > 0 {
			timeStr += fmt.Sprintf("%d giây", seconds)
		}
		if timeStr == "" {
			timeStr = "0 giây"
		}
		chat.Message = strings.TrimSpace(timeStr)
	}

	// Update the chat message
	if err := r.db.Save(&chat).Error; err != nil {
		return migration.Chat{}, err
	}

	// Update the chat room lastest message
	// Update the chat room with the last message ID and message content
	// Handle the message content based on the call type
	// If the call type is call, display "Cuộc gọi dến"
	// If the call type is missed_call, display "Cuộc gọi nhỡ"
	var messageContent string
	switch req.CallType {
	case "missed_call":
		messageContent = "Cuộc gọi nhỡ"
	case "call":
		messageContent = "Cuộc gọi đến"
	}
	if err := r.db.Model(&migration.Room{}).Where("id = ?", req.ChatRoomID).Updates(map[string]interface{}{
		"last_message_id":   chat.ID,
		"last_message_text": messageContent,
		"last_message_at":   chat.CreatedAt,
	}).Error; err != nil {
		return migration.Chat{}, err
	}

	return chat, nil
}

// InitiateCall initiates a call in a chat room
func (r *ChatRepository) InitiateCall(req schemas.InitiateCallRequest, userID uuid.UUID) (migration.Chat, error) {
	chatRoomUUID, err := uuid.Parse(req.ChatRoomID)
	if err != nil {
		return migration.Chat{}, err
	}

	receiverUUID, err := uuid.Parse(req.ReceiverID)
	if err != nil {
		return migration.Chat{}, err
	}

	newChat := migration.Chat{
		RoomID:      chatRoomUUID,
		SenderID:    userID,
		ReceiverID:  receiverUUID,
		Message:     "0",
		MessageType: "call", // Default call type is call
	}

	// Save the chat message
	err = r.db.Create(&newChat).Error
	if err != nil {
		return migration.Chat{}, err
	}

	// This is initiated call so no need to update the chat room lastest message
	return newChat, nil
}

var _ IChatRepository = (*ChatRepository)(nil)
