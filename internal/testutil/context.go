/*
Apache Score
Copyright 2022 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package testutil

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
)

type testContextKey string

const (
	testContextTestKey = testContextKey("_key_")
	testContextTestVal = "_val_"
)

func TestContext() context.Context {
	return context.WithValue(context.Background(), testContextTestKey, testContextTestVal)
}

func WithTestContext() gomock.Matcher {
	return ContextWithValue(testContextTestKey, testContextTestVal)
}

type contextWithValueMatcher struct {
	key testContextKey
	val string
}

func ContextWithValue(key testContextKey, value string) gomock.Matcher {
	return &contextWithValueMatcher{key: key, val: value}
}

func (m *contextWithValueMatcher) Matches(x interface{}) bool {
	if ctx, ok := x.(context.Context); ok {
		return ctx.Value(m.key) == m.val
	}
	return false
}

func (m *contextWithValueMatcher) String() string {
	return fmt.Sprintf("has value `%s`=`%s`", m.key, m.val)
}
