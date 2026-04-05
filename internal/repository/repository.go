package repository

import (
	"context"
	"fmt"
)

type MockTemplateRepo struct {
	data map[string]string
}

func NewMockTemplateRepo() *MockTemplateRepo {
	return &MockTemplateRepo{
		data: map[string]string{
			"welcome": "Hi {{.Name}}!",
			"alert":   "Critical: {{.Msg}}",
		},
	}
}

func (m *MockTemplateRepo) GetTemplate(_ context.Context, name string) (string, error) {
	if t, ok := m.data[name]; ok {
		return t, nil
	}
	return "", fmt.Errorf("template %s not found in mock store", name)
}
