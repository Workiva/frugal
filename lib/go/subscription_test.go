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

func (m *mockFScopeTransport) Subscribe(topic string) error {
	return m.Called(topic).Error(0)
}

func (m *mockFScopeTransport) DiscardFrame() {
	m.Called()
}

// Ensures Unsubscribe closes the transport and returns nil on success.
func TestSubscriptionUnsubscribe(t *testing.T) {
	mockTransport := new(mockFScopeTransport)
	mockTransport.mockTTransport.On("Close").Return(nil)
	sub := NewFSubscription("foo", mockTransport)
	assert.Nil(t, sub.Unsubscribe())
	mockTransport.AssertExpectations(t)
}

// Ensures Unsubscribe returns an error if the underlying transport close
// fails.
func TestSubscriptionUnsubscribeError(t *testing.T) {
	mockTransport := new(mockFScopeTransport)
	err := errors.New("error")
	mockTransport.mockTTransport.On("Close").Return(err)
	sub := NewFSubscription("foo", mockTransport)
	assert.Equal(t, err, sub.Unsubscribe())
	mockTransport.AssertExpectations(t)
}

// Ensures Topic returns the correct topic string.
func TestSubscriptionTopic(t *testing.T) {
	sub := NewFSubscription("foo", nil)
	assert.Equal(t, "foo", sub.Topic())
}
