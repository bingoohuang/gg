MAKEFLAGS = -s

sql.go: sql.y
	# go install golang.org/x/tools/...@latest
	goyacc -o sql.go sql.y
	gofmt -w sql.go

clean:
	rm -f y.output sql.go
