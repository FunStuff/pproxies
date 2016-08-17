package proxy

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCyber(t *testing.T) {
	proxies, err := CyberSrc(10 * time.Second)
	if err != nil {
		t.Error(err)
	}
	assert.NotZero(t, len(proxies))
}
