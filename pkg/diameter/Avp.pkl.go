// Code generated from Pkl module `Diameter`. DO NOT EDIT.
package diameter

type Avp struct {
	Code uint32 `pkl:"code"`

	Name string `pkl:"name"`

	Flags uint8 `pkl:"flags"`

	VndId uint32 `pkl:"vnd_id"`

	Length uint32 `pkl:"length"`

	Type int `pkl:"type"`

	Enum *Enum `pkl:"enum"`

	Group *Group `pkl:"group"`

	Data *AvpData `pkl:"data"`
}
