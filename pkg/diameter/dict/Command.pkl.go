// Code generated from Pkl module `Diameter`. DO NOT EDIT.
package dict

type Command struct {
	Code uint32 `pkl:"code"`

	Name string `pkl:"name"`

	Short string `pkl:"short"`

	Flags uint8 `pkl:"flags"`

	Request []AvpRule `pkl:"request"`

	Answer []AvpRule `pkl:"answer"`
}
