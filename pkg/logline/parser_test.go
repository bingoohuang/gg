package logline

import (
	"bufio"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	// 1. 样本与模式对照书写，模式中的#对应的样本字符为锚定符
	// 2. 需要捕获锚定符之间的值时，给定一个标识符（例如ip,time)，如果不需要取值则使用空格略过
	// 3. 值名称为time时表示日期时间，对应的样本中的时间值，要修改成golang的时间格式(layout)，参见 https://golang.org/src/time/format.go
	// 4. 竖线表示过滤器，目前仅支持path过滤器，就是从uri(带query)中取出path(不带query)
	// 5. 捕获标识符对应的样本值为整数时会解析成int类型，为小数时会解析成float64类型
	const samplee = `127.0.0.1 - - [02/Jan/2006:15:04:05 -0700] GET    /path?indent=true HTTP/1.1 200  41824     - 8      0.008   6 - - Nginx/1.1`
	const pattern = `ip       # # ##time                      ##method#uri|path         #        #code#bytesSent#-#millis#seconds#`

	p, err := NewPattern(samplee, pattern)
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

	// pattern: '%h %l %u %t "%r" %s %b "%{Referer}i" "%{User-Agent}i" %D'
	// %D	Time taken to process the request, in millis
	logSamplee := "10.1.6.1 - - [02/Jan/2006:15:04:05 -0700] !HEAD   /         HTTP/1.0! 200  94        !-! !-! 0     "
	logPattern := "ip      # #  #time                      # #method#path|path#        ##code#bytesSent## # # ##millis"
	p2, err2 := NewPattern(logSamplee, logPattern, WithReplace(`!`, `"`))
	assert.Nil(t, err2)

	line = `10.16.26.21 - - [19/May/2021:00:00:13 +0800] "POST /upload1 HTTP/1.1" 200 94 "-" "Apache-HttpClient/4.5.1 (Java/1.8.0_74)" 42`
	m2 := p2.Parse(line)
	tt, _ = TimeValue(`02/Jan/2006:15:04:05 -0700`).Convert("19/May/2021:00:00:13 +0800")
	assert.Equal(t, map[string]interface{}{
		"ip":        "10.16.26.21",
		"time":      tt,
		"method":    "POST",
		"path":      "/upload1",
		"code":      200,
		"bytesSent": 94,
		"millis":    42,
	}, m2)

	// https://qsli.github.io/2016/12/23/tomcat-access-log/
	//parseFile(`/Users/bingoobjca/Downloads/localhost_access_log2021-05-21.txt`, p)
	//parseFile(`/Users/bingoobjca/Downloads/scaffold_access_log.2021-05-19.log`, p2)
}

func parseFile(file string, p *Pattern) {
	f, _ := os.Open(file)
	defer f.Close()

	out, _ := os.OpenFile(file+".parsed", os.O_CREATE|os.O_WRONLY, os.ModePerm)
	defer out.Close()

	scanner := bufio.NewScanner(f)

	lineNo := 0
	for scanner.Scan() {
		line := scanner.Bytes()
		lineNo++
		m := p.ParseBytes(line)

		fmt.Fprintf(out, "%v\n", m)
	}

	fmt.Printf("total lines: %d\n", lineNo)
}
