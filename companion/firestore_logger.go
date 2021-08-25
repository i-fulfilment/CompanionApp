package companion

import (
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	"time"
)

func NewFirestoreLogger(reference string, client *firestore.Client) *FirestoreLogger {
	return &FirestoreLogger{
		client:    client,
		reference: reference,
	}
}

type FirestoreLogger struct {
	client    *firestore.Client
	reference string
}

func (f FirestoreLogger) Write(p []byte) (int, error) {

	if f.client == nil || f.reference == "" {
		return len(string(p)), nil
	}

	record := make(map[string]interface{})
	err := json.Unmarshal(p, &record)

	record["timestamp"] = time.Now().Unix()

	if err != nil {
		return 0, err
	}

	_, _, err = f.client.Collection("CompanionApps").Doc(f.reference).Collection("Logs").Add(context.Background(), record)

	if err != nil {
		return 0, nil
	}

	return len(string(p)), nil
}
