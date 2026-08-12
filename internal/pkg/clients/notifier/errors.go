package notifier

import "errors"

// ErrUnavailable indicates the notification service could not be reached. It is
// a transient failure: callers may retry.
var ErrUnavailable = errors.New("notifier unavailable")
