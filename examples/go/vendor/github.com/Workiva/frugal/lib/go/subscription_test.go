/*
 * Copyright 2017 Workiva
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package frugal

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockFScopeTransport struct {
	mockTTransport
	mock.Mock
}

func (m *mockFScopeTransport) LockTopic(topic string) error {
	return m.Called(topic).Error(0)
}

func (m *mockFScopeTransport) UnlockTopic() error {
	return m.Called().Error(0)
}

func (m *mockFScopeTransport) Subscribe(topic string, callback FAsyncCallback) error {
	return m.Called(topic, callback).Error(0)
}

func (m *mockFScopeTransport) Unsubscribe() error {
	return m.Called().Error(0)
}

func (m *mockFScopeTransport) DiscardFrame() {
	m.Called()
}

func (m *mockFScopeTransport) IsSubscribed() bool {
	return m.Called().Bool(0)
}

// Ensures Unsubscribe closes the transport and returns nil on success.
func TestSubscriptionUnsubscribe(t *testing.T) {
	mockTransport := new(mockFScopeTransport)
	mockTransport.On("Unsubscribe").Return(nil)
	sub := NewFSubscription("foo", mockTransport)
	assert.Nil(t, sub.Unsubscribe())
	mockTransport.AssertExpectations(t)
}

// Ensures Unsubscribe returns an error if the underlying transport close
// fails.
func TestSubscriptionUnsubscribeError(t *testing.T) {
	mockTransport := new(mockFScopeTransport)
	err := errors.New("error")
	mockTransport.On("Unsubscribe").Return(err)
	sub := NewFSubscription("foo", mockTransport)
	assert.Equal(t, err, sub.Unsubscribe())
	mockTransport.AssertExpectations(t)
}

// Ensures Topic returns the correct topic string.
func TestSubscriptionTopic(t *testing.T) {
	sub := NewFSubscription("foo", nil)
	assert.Equal(t, "foo", sub.Topic())
}
