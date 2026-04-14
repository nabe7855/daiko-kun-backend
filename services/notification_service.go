package services

import (
	"fmt"
	"time"
)

// SendNotification simulates sending a push notification via FCM
// In the future, this will use the Firebase Admin SDK
func SendNotification(token string, title string, body string, data map[string]string) error {
	if token == "" {
		return fmt.Errorf("FCM token is empty")
	}

	// For now, we just log the notification to the console
	fmt.Printf("\n--- [PUSH NOTIFICATION SENT] ---\n")
	fmt.Printf("Time:  %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("To:    %s\n", token)
	fmt.Printf("Title: %s\n", title)
	fmt.Printf("Body:  %s\n", body)
	if len(data) > 0 {
		fmt.Printf("Data:  %v\n", data)
	}
	fmt.Printf("--------------------------------\n\n")

	return nil
}

// NotifyCustomerWhenDriverAccepted notifies the customer that their ride has been accepted
func NotifyCustomerWhenDriverAccepted(customerToken string, driverName string) error {
	title := "ドライバーが決定しました"
	body := fmt.Sprintf("%s ドライバーがあなたの依頼を受諾しました。到着までしばらくお待ちください。", driverName)
	return SendNotification(customerToken, title, body, map[string]string{"type": "ride_accepted"})
}

// NotifyCustomerWhenDriverArrived notifies the customer that the driver has arrived
func NotifyCustomerWhenDriverArrived(customerToken string) error {
	title := "ドライバーが到着しました"
	body := "ドライバーが指定の場所に到着しました。合流をお願いします。"
	return SendNotification(customerToken, title, body, map[string]string{"type": "driver_arrived"})
}
