package grpc_client

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	pb "github.com/100bench/infr_training/pkg/proto/gen/notification/v1"
)

type notificationClient struct {
	client pb.NotificationServiceClient
}

func NewNotificationClient(conn *grpc.ClientConn) *notificationClient {
	return &notificationClient{
		client: pb.NewNotificationServiceClient(conn),
	}
}

func (n *notificationClient) Send(ctx context.Context, id, title string, notifyType pb.EventType) error {
	event:= &pb.TaskEvent{Id: id, Title: title, EventType: notifyType}

	_, err:= n.client.Send(ctx, event)

	if err != nil{
		return fmt.Errorf("%w: notificationLCient.Send()", err)
	}

	return nil
}
