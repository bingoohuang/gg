# customized features

## using

in go.mod as [Using forked package import in Go](https://stackoverflow.com/a/56792766)

```
replace github.com/bingoohuang/go-yaml => github.com/bingoohuang/go-yaml v1.8.11-0.20210719040622-7e6a9879a76a
```

If you are using go modules. You could use replace directive

The replace directive allows you to supply another import path that might be another module located in VCS (GitHub or
elsewhere), or on your local filesystem with a relative or absolute file path. The new import path from the replace
directive is used without needing to update the import paths in the actual source code.

So you could do below in your go.mod file

`module github.com/yogeshlonkar/openapi-to-postman`

go 1.12

```
require (
    github.com/someone/repo v1.20.0
)


replace github.com/someone/repo => github.com/you/repo v3.2.1
```

where v3.2.1 is tag on your repo. Also can be done through CLI

go mod edit -replace="github.com/someone/repo@v0.0.0=github.com/you/repo@v1.1.1"

## decoding by struct field type

```go
type Duration struct {
	Dur time.Duration
}

func decodeDuration(node ast.Node, typ reflect.Type) (reflect.Value, error) {
	if v, ok := node.(*ast.StringNode); ok {
		d, err := time.ParseDuration(v.Value)
		return reflect.ValueOf(d), err
	}
	return reflect.Value{}, yaml.ErrContinue
}

func TestDecoderDuration(t *testing.T) {
	c := Duration{}
	decodeOption := yaml.TypeDecoder(reflect.TypeOf((*time.Duration)(nil)).Elem(), decodeDuration)
	decoder := yaml.NewDecoder(strings.NewReader(`dur: 10s`), decodeOption)
	err := decoder.Decode(&c)
	assert.Nil(t, err)
	assert.Equal(t, Duration{Dur: 10 * time.Second}, c)

	decoder = yaml.NewDecoder(strings.NewReader(`dur: 111`), decodeOption)
	err = decoder.Decode(&c)
	assert.Nil(t, err)
	assert.Equal(t, Duration{Dur: 111}, c)
}
```

## decoding function by label

```go
type Config struct {
	Size int64 `yaml:",label=size"`
}

func decodeSize(node ast.Node, typ reflect.Type) (reflect.Value, error) {
	if s, ok := node.(*ast.StringNode); ok {
		if v, err := man.ParseBytes(s.Value); err != nil {
			return reflect.Value{}, err
		} else {
			return yaml.CastUint64(v, typ)
		}
	}
	return reflect.Value{}, yaml.ErrContinue
}

func TestDecoderLabel(t *testing.T) {
	c := Config{}
	decodeOption := yaml.LabelDecoder("size", decodeSize)
	decoder := yaml.NewDecoder(strings.NewReader(`size: 10MiB`), decodeOption)
	err := decoder.Decode(&c)
	assert.Nil(t, err)
	assert.Equal(t, Config{Size: 10 * 1024 * 1024}, c)

	decoder = yaml.NewDecoder(strings.NewReader(`size: 1234`), decodeOption)
	err = decoder.Decode(&c)
	assert.Nil(t, err)
	assert.Equal(t, Config{Size: 1234}, c)
}
```
