package proxy

import (
	"net/http"
	"net/url"
	"time"

	logger "github.com/blendlabs/go-logger"
)

// RequestListener is a listener for request events.
type RequestListener func(writer *logger.Writer, ts logger.TimeSource, target *url.URL, req *http.Request, statusCode, contentLengthBytes int, elapsed time.Duration)

// NewRequestListener returns a new handler for request events.
func NewRequestListener(listener RequestListener) logger.EventListener {
	return func(writer *logger.Writer, ts logger.TimeSource, eventFlag logger.Event, state ...interface{}) {
		if len(state) < 3 {
			return
		}

		target, err := stateAsURL(state[0])
		if err != nil {
			return
		}

		req, err := stateAsRequest(state[1])
		if err != nil {
			return
		}

		statusCode, err := stateAsInteger(state[2])
		if err != nil {
			return
		}

		contentLengthBytes, err := stateAsInteger(state[3])
		if err != nil {
			return
		}

		elapsed, err := stateAsDuration(state[4])
		if err != nil {
			return
		}

		listener(writer, ts, target, req, statusCode, contentLengthBytes, elapsed)
	}
}

func stateAsURL(state interface{}) (*url.URL, error) {
	if typed, isTyped := state.(*url.URL); isTyped {
		return typed, nil
	}
	return nil, logger.ErrTypeConversion
}

func stateAsRequest(state interface{}) (*http.Request, error) {
	if typed, isTyped := state.(*http.Request); isTyped {
		return typed, nil
	}
	return nil, logger.ErrTypeConversion
}

func stateAsInteger(state interface{}) (int, error) {
	if typed, isTyped := state.(int); isTyped {
		return typed, nil
	}
	return 0, logger.ErrTypeConversion
}

func stateAsDuration(state interface{}) (time.Duration, error) {
	if typed, isTyped := state.(time.Duration); isTyped {
		return typed, nil
	}
	return 0, logger.ErrTypeConversion
}
