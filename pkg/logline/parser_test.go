package logline

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestParse(t *testing.T) {
	loglineSampl := `192.158.7.1 - - [26/May/2021:18:55:44 +0800] GET    /path?indent=true&wt=json HTTP/1.1 200  41824     - 8      0.008 6 - - Go-http-client/1.1`
	const pattern = `ip         # # ##date                      ##method#uri|path                 #        #code#bytesSent#-#millis#`

	p, err := NewPattern(loglineSampl, pattern)
	assert.Nil(t, err)

	line := `192.158.77.11 - - [26/May/2021:18:55:45 +0800] GET /solr/licenseIndex/select?indent=true&5-26T10rows=2500&sort=id+asc&start=0&wt=json HTTP/1.1 200 41824 - 8 0.008 6 - - Go-http-client/1.1`

	m := p.Parse(line)
	tt, _ := TimeValue().Convert("26/May/2021:18:55:45 +0800")
	assert.Equal(t, map[string]interface{}{
		"ip":        "192.158.77.11",
		"date":      tt,
		"method":    http.MethodGet,
		"uri":       "/solr/licenseIndex/select",
		"code":      200,
		"bytesSent": 41824,
		"millis":    8,
	}, m)

	// https://qsli.github.io/2016/12/23/tomcat-access-log/
}
