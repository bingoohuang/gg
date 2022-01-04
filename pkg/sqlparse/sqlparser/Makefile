MAKEFLAGS = -s

sql.go: sql.y
	goyacc -o sql.go sql.y
	gofmt -w sql.go

clean:
	rm -f y.output sql.go