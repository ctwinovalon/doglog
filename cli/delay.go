package cli

import "time"

// DelayForSeconds Sleep. We can't just spin on the Datadog call and there's no callback or interrupt options.
func DelayForSeconds(delay float64) {
	delayInMilliseconds := int(delay * 1000.0)
	time.Sleep(time.Duration(delayInMilliseconds) * time.Millisecond)
}

// MinDelay Minimum delay between calls to Datadog
const MinDelay = 10.0

// Maximum delay between calls to Datadog
const maxDelay = 30.0

// Back-off factor when increasing the delay.
const delayIncreaseFactor = 2.0

// AdjustDelay Adjust the delay between calls to Datadog, so we don't hammer it when no messages have
// arrived for a while.
func AdjustDelay(delay float64, found bool) float64 {
	if !found {
		if delay < maxDelay {
			delay *= delayIncreaseFactor
			if delay > maxDelay {
				delay = maxDelay
			}
		}
	} else {
		delay = MinDelay
	}
	return delay
}
