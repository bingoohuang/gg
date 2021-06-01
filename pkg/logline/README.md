# 日志行解析

## 解析的设计思路

本解析的设计，为说人话的方式进行，避免诸如[Logstash Grok Patterns](https://coralogix.com/blog/logstash-grok-tutorial-with-examples/) 等的复杂形式。

日志行解析设计思路的五大原则：

1. 对照原则：样本与模式对照书写，模式中的#对应的样本字符为锚定符
2. 锚定原则：需要捕获锚定符之间的值时，给定一个标识符（例如ip,time)，如果不需要取值则使用空格略过
3. 命名原则：值名称为time时表示日期时间，对应的样本中的时间值，要修改成golang的[时间格式 layout](https://golang.org/src/time/format.go)
4. 转换原则：竖线表示转换过滤器，目前仅支持path过滤器，就是从uri(带query)中取出path(不带query)
5. 类型原则：捕获标识符对应的样本值为整数时会解析成int类型，为小数时会解析成float64类型

### 示例1

```go
// pattern="%h %l %u %t %r %s %b %S %D %T %F %{Referer}i %{X-Forwarded-For}i %{User-Agent}i %{X-Real-IP}i"
const samplee = `127.0.0.1 - - [02/Jan/2006:15:04:05 -0700] GET    /path?indent=true HTTP/1.1 200  41824     - 8      0.008   6 - - Nginx/1.1`
const pattern = `ip       # # ##time                      ##method#uri|path         #        #code#bytesSent#-#millis#seconds#`
```

对于上面的样本（samplee）与模式（pattern），上下是对照的。在模式中使用`#`来指定样本对对应的锚定字符，然后在锚定字符之间，通过命名来获取对应的样本中的信息。 获取的值可以通过`|`
符号，建立转换规则，对取值进行转换处理。取值类型由样本中对应的示例值给出（目前支持字符串、整型、日期时间、浮点四种）

### 示例2

```go
// pattern: '%h %l %u %t "%r" %s %b "%{Referer}i" "%{User-Agent}i" %D'
const samplee := "10.1.6.1 - - [02/Jan/2006:15:04:05 -0700] !HEAD   /         HTTP/1.0! 200  94        !-! !-! 0     "
const pattern := "ip      # #  #time                      # #method#path|path#        ##code#bytesSent## # # ##millis"
```

## [tomcat access log 格式设置](https://qsli.github.io/2016/12/23/tomcat-access-log/)

### Tomcat access log 日志格式

1. 文件位置: conf/server.xml
2. 默认配置

```xml
<!-- Access log processes all example.
     Documentation at: /docs/config/valve.html
     Note: The pattern used is equivalent to using pattern="common" -->
<Valve className="org.apache.catalina.valves.AccessLogValve" directory="logs"
       prefix="localhost_access_log." suffix=".txt"
       pattern="%h %l %u %t &quot;%r&quot; %s %b"/>
```

名称|含义
---|---
%a|Remote IP address
%A|Local IP address
%b|Bytes sent, excluding HTTP headers, or ‘-‘ if zero
%B|Bytes sent, excluding HTTP headers
%h|Remote host name (or IP address if enableLookups for the connector is false)
%H|Request protocol
%l|Remote logical username from identd (always returns ‘-‘)
%m|Request method (GET, POST, etc.)
%p|Local port on which this request was received
%q|Query string (prepended with a ‘?’ if it exists)
%r|First line of the request (method and request URI)
%s|HTTP status code of the response
%S|User session ID
%t|Date and time, in Common Log Format
%u|Remote user that was authenticated (if any), else ‘-‘
%U|Requested URL path
%v|Local server name
%D|Time taken to process the request, in millis
%T|Time taken to process the request, in seconds
%F|Time taken to commit the response, in millis
%I|Current request thread name (can compare later with stacktraces

默认的配置打出来的access日志如下：

> 127.0.0.1 - - [07/Oct/2016:22:31:56 +0800] "GET /dubbo/ HTTP/1.1" 404 963

> 远程IP logicalUsername remoteUser 时间和日期 http请求的第一行 状态码 除去http头的发送大小

### header、cookie、session其他字段的支持

> There is also support to write information incoming or outgoing headers, cookies, session or request attributes and special timestamp formats. It is modeled after the Apache HTTP Server log configuration syntax:

名称|含义
---|---
%{xxx}i|for incoming headers
%{xxx}o|for outgoing response headers
%{xxx}c|for a specific cookie
%{xxx}r|xxx is an attribute in the ServletRequest
%{xxx}s|xxx is an attribute in the HttpSession
%{xxx}t|xxx is an enhanced SimpleDateFormat pattern

例如： `%{X-Forwarded-For}i` 即可打印出实际访问的ip地址（考虑到ng的反向代理）

HTTP头一般格式如下:

`X-Forwarded-For: client1, proxy1, proxy2`

> 其中的值通过一个 逗号+空格 把多个IP地址区分开, 最左边（client1）是最原始客户端的IP地址, 代理服务器每成功收到一个请求，就把请求来源IP地址添加到右边。 在上面这个例子中，这个请求成功通过了三台代理服务器：proxy1, proxy2 及 proxy3。请求由client1发出，到达了proxy3（proxy3可能是请求的终点）。请求刚从client1中发出时，XFF是空的，请求被发往proxy1； 通过proxy1的时候，client1被添加到XFF中，之后请求被发往proxy2;通过proxy2的时候，proxy1被添加到XFF中，之后请求被发往proxy3； 通过proxy3时，proxy2被添加到XFF中，之后请求的的去向不明，如果proxy3不是请求终点，请求会被继续转发。

> 鉴于伪造这一字段非常容易，应该谨慎使用X-Forwarded-For字段。正常情况下XFF中最后一个IP地址是最后一个代理服务器的IP地址, 这通常是一个比较可靠的信息来源。

### 参考

1. [Apache Tomcat 7 The Valve Component](http://tomcat.apache.org/tomcat-7.0-doc/config/valve.html)
1. [X-Forwarded-For](https://zh.wikipedia.org/wiki/X-Forwarded-For)
