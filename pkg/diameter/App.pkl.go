// Code generated from Pkl module `Diameter`. DO NOT EDIT.
package diameter

type App struct {
	Id uint32 `pkl:"id"`

	Name string `pkl:"name"`

	Vnd string `pkl:"vnd"`

	VndId uint32 `pkl:"vnd_id"`

	Cmds []*Cmd `pkl:"cmds"`
}
