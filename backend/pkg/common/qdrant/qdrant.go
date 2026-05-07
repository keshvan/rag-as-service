package qdrant

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
)

type Chunk struct {
	Index int
	Text  string
}

type QdrantConfig struct {
	Host       string
	Port       int
	Collection string
	UseTLS     bool
}

type Qdrant struct {
	client     *qdrant.Client
	collection string
}

func NewQdrant(cfg QdrantConfig) (*Qdrant, error) {
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == 0 {
		cfg.Port = 6334
	}

	client, err := qdrant.NewClient(&qdrant.Config{
		Host:   cfg.Host,
		Port:   cfg.Port,
		UseTLS: cfg.UseTLS,
	})
	if err != nil {
		return nil, fmt.Errorf("create qdrant client: %w", err)
	}

	return &Qdrant{
		client:     client,
		collection: cfg.Collection,
	}, nil
}

func (q *Qdrant) EnsureCollection(ctx context.Context, dim uint64) error {
	if dim == 0 {
		return fmt.Errorf("qdrant vector dimension must be greater than zero")
	}

	exists, err := q.client.CollectionExists(ctx, q.collection)
	if err != nil {
		return fmt.Errorf("check collection exists: %w", err)
	}

	if exists {
		// При желании здесь можно добавить валидацию размерности через GetCollectionInfo
		return nil
	}

	err = q.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: q.collection,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     dim,
			Distance: qdrant.Distance_Cosine,
		}),
	})
	if err != nil {
		return fmt.Errorf("create collection: %w", err)
	}

	return nil
}

func (q *Qdrant) ReplaceDocumentChunks(
	ctx context.Context,
	orgID, docID, objectKey, contentType string,
	chunks []Chunk,
	vectors [][]float32,
) error {
	if len(chunks) == 0 {
		return fmt.Errorf("replace document chunks: chunks are required")
	}
	if len(chunks) != len(vectors) {
		return fmt.Errorf("replace document chunks: chunk/vector mismatch: got=%d want=%d", len(vectors), len(chunks))
	}

	// 1. Удаляем старые точки по фильтру organization_id + document_id
	filter := &qdrant.Filter{
		Must: []*qdrant.Condition{
			qdrant.NewMatch("organization_id", orgID),
			qdrant.NewMatch("document_id", docID),
		},
	}

	wait := true
	_, err := q.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: q.collection,
		Wait:           &wait,
		Points: &qdrant.PointsSelector{
			PointsSelectorOneOf: &qdrant.PointsSelector_Filter{
				Filter: filter,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("delete existing chunks: %w", err)
	}

	// 2. Формируем новые точки
	points := make([]*qdrant.PointStruct, 0, len(chunks))
	for i, chunk := range chunks {
		points = append(points, &qdrant.PointStruct{
			Id: &qdrant.PointId{
				PointIdOptions: &qdrant.PointId_Uuid{
					Uuid: q.pointID(orgID, docID, chunk.Index),
				},
			},
			Vectors: &qdrant.Vectors{
				VectorsOptions: &qdrant.Vectors_Vector{
					Vector: &qdrant.Vector{
						Data: vectors[i],
					},
				},
			},
			Payload: map[string]*qdrant.Value{
				"organization_id": {Kind: &qdrant.Value_StringValue{StringValue: orgID}},
				"document_id":     {Kind: &qdrant.Value_StringValue{StringValue: docID}},
				"object_key":      {Kind: &qdrant.Value_StringValue{StringValue: objectKey}},
				"content_type":    {Kind: &qdrant.Value_StringValue{StringValue: contentType}},
				"chunk_index":     {Kind: &qdrant.Value_IntegerValue{IntegerValue: int64(chunk.Index)}},
				"text":            {Kind: &qdrant.Value_StringValue{StringValue: chunk.Text}},
			},
		})
	}

	// 3. Вставляем новые точки через метод Upsert с запросом UpsertPoints
	_, err = q.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: q.collection,
		Wait:           &wait,
		Points:         points,
	})
	if err != nil {
		return fmt.Errorf("upsert chunks: %w", err)
	}

	return nil
}

var pointNamespace = uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

func (q *Qdrant) pointID(orgID, docID string, chunkIndex int) string {
	data := orgID + ":" + docID + ":" + strconv.Itoa(chunkIndex)
	return uuid.NewSHA1(pointNamespace, []byte(data)).String()
}
