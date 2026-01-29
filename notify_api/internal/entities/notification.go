package entities

type NotificationType int

const (
	Unknown     NotificationType = 0
	TaskCreated NotificationType = 1
	TaskDeleted NotificationType = 2
)

type Notification struct{
	Type NotificationType
	Payload map[string]any
}

func NewNotification(notificationType NotificationType, payload map[string]any) Notification {
	return Notification{
		Type:    notificationType,
		Payload: payload,
	}
}