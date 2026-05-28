// Code generated from Pkl module `Diameter`. DO NOT EDIT.
package dict

type Avp struct {
	Code uint32 `pkl:"code"`

	Name string `pkl:"name"`

	Flags uint8 `pkl:"flags"`

	VndId uint32 `pkl:"vnd_id"`

	Type int `pkl:"type"`

	Enum *Enum `pkl:"enum"`

	Group *Group `pkl:"group"`
}
