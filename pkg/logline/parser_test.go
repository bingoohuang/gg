package logline

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestParse(t *testing.T) {
	const sample1 = `127.0.0.1 - - [02/Jan/2006:15:04:05 -0700] GET    /path?indent=true HTTP/1.1 200  41824     - 8      0.008   6 - - Nginx/1.1`
	const pattern = `ip       # # ##time                      ##method#uri|path         #        #code#bytesSent#-#millis#seconds#`

	p, err := NewPattern(sample1, pattern)
	assert.Nil(t, err)

	line := `192.158.77.11 - - [26/May/2021:18:55:45 +0800] GET /solr/licenseIndex/select?indent=true&5-26T10rows=2500&sort=id+asc&start=0&wt=json HTTP/1.1 200 41824 - 8 0.008 6 - - Go-http-client/1.1`

	m := p.Parse(line)
	tt, _ := TimeValue(`02/Jan/2006:15:04:05 -0700`).Convert("26/May/2021:18:55:45 +0800")
	assert.Equal(t, map[string]interface{}{
		"ip":        "192.158.77.11",
		"time":      tt,
		"method":    http.MethodGet,
		"uri":       "/solr/licenseIndex/select",
		"code":      200,
		"bytesSent": 41824,
		"millis":    8,
		"seconds":   0.008,
	}, m)

	// https://qsli.github.io/2016/12/23/tomcat-access-log/
	//f, _ := os.Open(`~/Downloads/localhost_access_log2021-05-21.txt`)
	//defer f.Close()
	//
	//out, _ := os.OpenFile(`~/Downloads/localhost_access_log2021-05-21.parsed.txt`, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	//defer out.Close()
	//
	//scanner := bufio.NewScanner(f)
	//
	//lineNo := 0
	//for scanner.Scan() {
	//	line := scanner.Bytes()
	//	lineNo++
	//	m := p.ParseBytes(line)
	//
	//	fmt.Fprintf(out, "%v\n", m)
	//}
	//
	//fmt.Printf("total lines: %d\n", lineNo)
}
