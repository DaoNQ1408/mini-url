package repository

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type URLRecord struct {
	ShortID     string `dynamodbav:"short_id"`
	OriginalURL string `dynamodbav:"original_url"`
}

func SaveURL(client *dynamodb.Client, tableName, shortID, longURL string) error {
	item := URLRecord{ShortID: shortID, OriginalURL: longURL}
	av, _ := attributevalue.MarshalMap(item)

	_, err := client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &tableName,
		Item:      av,
	})
	return err
}

func GetURL(client *dynamodb.Client, tableName, shortID string) (string, error) {
	input := &dynamodb.GetItemInput{
		TableName: &tableName,
		Key: map[string]types.AttributeValue{
			"short_id": &types.AttributeValueMemberS{Value: shortID},
		},
	}

	// 2. Gọi AWS SDK để lấy item
	result, err := client.GetItem(context.TODO(), input)
	if err != nil {
		return "", err
	}

	if result.Item == nil {
		return "", nil
	}

	var record URLRecord
	err = attributevalue.UnmarshalMap(result.Item, &record)
	if err != nil {
		return "", err
	}

	return record.OriginalURL, nil
}
