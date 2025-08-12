// Code generated from Pkl module `Diameter`. DO NOT EDIT.
package diameter

type Cmd struct {
	Code uint32 `pkl:"code"`

	Name string `pkl:"name"`

	Short string `pkl:"short"`

	Flags uint8 `pkl:"flags"`

	Request []*AvpRule `pkl:"request"`

	Answer []*AvpRule `pkl:"answer"`
}
