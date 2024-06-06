package feed

import (
	"parserver/lock"
)

// Feed represents a user's twitter feed
type Feed interface {
	Add(body string, timestamp float64)
	Remove(timestamp float64) bool
	Contains(timestamp float64) bool
	GetFeeds() []map[string]any
}

// feed is the internal representation of a user's twitter feed (hidden from outside packages)
type feed struct {
	start   *post // a pointer to the beginning post
	context *lock.RWMutex
}

// post is the internal representation of a post on a user's twitter feed (hidden from outside packages)
type post struct {
	body      string  // the text of the post
	timestamp float64 // Unix timestamp of the post
	next      *post   // the next post in the feed
}

// NewPost creates and returns a new post value given its body and timestamp
func newPost(body string, timestamp float64, next *post) *post {
	return &post{body, timestamp, next}
}

// NewFeed creates a empty user feed
func NewFeed() Feed {
	ct := lock.NewRWMutex(32) // Set maximum parallel read threads to be 32
	return &feed{start: nil, context: ct}
}

// GetFeeds iterate the whole linked list and build a copy of feeds and return
func (f *feed) GetFeeds() []map[string]any {
	f.context.RLock()
	defer f.context.RUnlock()
	curr := f.start
	feeds := []map[string]any{}
	for curr != nil {
		feeds = append(feeds, map[string]any{"body": curr.body, "timestamp": curr.timestamp})
		curr = curr.next
	}
	return feeds
}

// Add inserts a new post to the feed. The feed is always ordered by the timestamp where
// the most recent timestamp is at the beginning of the feed followed by the second most
// recent timestamp, etc.
func (f *feed) Add(body string, timestamp float64) {
	f.context.Lock()
	defer f.context.Unlock()
	newPost := newPost(body, timestamp, nil)
	if f.start == nil || f.start.timestamp < timestamp {
		newPost.next = f.start
		f.start = newPost
		return
	}

	curr := f.start
	for curr.next != nil && curr.next.timestamp > timestamp {
		curr = curr.next
	}

	newPost.next = curr.next
	curr.next = newPost
}

// Remove deletes the post with the given timestamp. If the timestamp
// is not included in a post of the feed then the feed remains
// unchanged. Return true if the deletion was a success, otherwise return false
func (f *feed) Remove(timestamp float64) bool {
	f.context.Lock()
	defer f.context.Unlock()
	if f.start == nil {
		return false
	}
	if f.start.timestamp == timestamp {
		f.start = f.start.next
		return true
	}

	curr := f.start
	for curr.next != nil && curr.next.timestamp != timestamp {
		curr = curr.next
	}

	if curr.next != nil {
		curr.next = curr.next.next
		return true
	}
	return false
}

// Contains determines whether a post with the given timestamp is
// inside a feed. The function returns true if there is a post
// with the timestamp, otherwise, false.
func (f *feed) Contains(timestamp float64) bool {
	f.context.RLock()
	defer f.context.RUnlock()
	curr := f.start
	for curr != nil {
		if curr.timestamp == timestamp {
			return true
		}
		curr = curr.next
	}
	return false
}
