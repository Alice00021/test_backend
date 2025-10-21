package v1_test

import (
	"context"
	"testing"
)

type mockUserUsecase struct{}

func (m *mockUserUsecase) UpdateRating(ctx context.Context, userID string, rating float32) error {
	// имитация выполнения запроса в БД
	return nil
}

func BenchmarkUpdateRating(b *testing.B) {
	mockUC := &mockUserUsecase{}
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		err := mockUC.UpdateRating(ctx, "user123", 10.5)
		if err != nil {
			b.Fatal(err)
		}
	}
}
