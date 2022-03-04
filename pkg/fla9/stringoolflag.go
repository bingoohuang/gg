package fla9

type StringBool struct {
	Val    string
	Exists bool
}

func (i *StringBool) String() string   { return i.Val }
func (i *StringBool) Get() interface{} { return i.Val }
func (i *StringBool) Set(value string) error {
	i.Val = value
	i.Exists = true
	return nil
}

func (i *StringBool) SetExists(b bool) { i.Exists = b }
