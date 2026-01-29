package main

import (
	"context"
	"fmt"
	"reflect"
)

type MockDB struct {
	beginFail  bool
	closeFail  bool
	commitFail bool
}

func (m *MockDB) Begin(ctx context.Context) (TX, error) {
	if m.beginFail {
		return nil, fmt.Errorf("begin mockDB failed")
	}
	return &MockTX{commitFail: m.commitFail}, nil
}

func (m *MockDB) Close() error {
	if m.closeFail {
		return fmt.Errorf("close mockDB failed")
	}
	return nil
}

type MockTX struct {
	commitFail bool
}

func (m *MockTX) Query() (Rows, error) {
	return &MockRows{
		Data:  [][]any{{"hello"}, {"world"}, {"!"}},
		index: 0,
	}, nil
}

func (m *MockTX) Commit() error {
	if m.commitFail {
		return fmt.Errorf("commit mockDB failed")
	}
	return nil
}

func (m *MockTX) Rollback() error {
	return nil
}

type MockRows struct {
	Data  [][]any
	index int
}

func (m *MockRows) Next() bool {
	return m.index < len(m.Data)
}

func (m *MockRows) Scan(dest ...any) error {
	row := m.Data[m.index]
	for i, val := range row {
		// 使用反射或类型断言将 val 赋值给 dest[i]
		reflect.ValueOf(dest[i]).Elem().Set(reflect.ValueOf(val))
	}
	m.index++
	return nil
}

func (*MockRows) Close() error {
	return nil
}
