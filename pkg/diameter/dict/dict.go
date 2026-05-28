//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: diameter.go
// Description: Diameter pkg: dictionary handling
//

package dict

import (
	"context"
	"fmt"
	"iter"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"

	"tgdp/pkg/diameter/diwe"
)

// Consts
const (
	FormatPkl  = iota // FormatPkl loads from Pkl configuration files
	FormatJson        // FormatJson loads from JSON files (not implemented)
	FormatYaml        // FormatYaml loads from YAML files (not implemented)
)

// Types
//
// Dict is the main dictionary type that provides thread-safe access to Diameter
// applications, commands, and AVPs. It uses read-write locks for concurrent access
// and maintains lookup caches for fast O(1) lookups by ID, code, or name.
type Dict struct {
	mu    sync.RWMutex
	core  Core        // core contains the raw dictionary data
	cache lookupCache // cache provides fast lookups
	flags flagsNames  // flags maps bit flags to names for debugging
}

// lookupCache provides fast O(1) lookup for dictionary entries by various keys.
// It is built from core data and kept in sync with the dictionary.
type lookupCache struct {
	appCacheById   map[uint32]*App
	appCacheByName map[string]*App
	cmdCacheByCode map[uint32]map[uint32]*Command // keyed by appId -> cmdCode
	cmdCacheByName map[uint32]map[string]*Command // keyed by appId -> cmdName
	avpCacheByCode map[uint32]*Avp
	avpCacheByName map[string]*Avp
	enumCache      map[uint32]map[string]int32 // keyed by avpCode -> itemName
}

// flagsNames maps bit flags to human-readable names for debugging/logging.
type flagsNames struct {
	avp map[uint8]string // AVP bit flag names
	cmd map[uint8]string // Command bit flag names
}

// loaderFunc is a function type for loading dictionary data from a file.
// It receives the Dict instance and the file path to load from.
type loaderFunc func(*Dict, string) error

// Variables
var (
	// loaders maps format constants to their respective loader functions.
	loaders = map[int]loaderFunc{
		FormatPkl:  loadFromPkl,
		FormatJson: loadFromJson,
		FormatYaml: loadFromYaml,
	}
)

// Constructor
//
// New creates a new Dict with the given core dictionary data.
// It initializes the lookup caches and flag name mappings.
// The returned Dict is ready for use and is thread-safe.
func New(core Core) *Dict {
	d := &Dict{core: core}
	d.fillLookupCache()
	d.fillFlagsNames()
	return d
}

// Methods
//

// GetApp retrieves an application by its ID (uint32) or name (string).
// It first attempts to parse string inputs as numeric IDs, then falls back to name lookup.
func (d *Dict) GetApp(appId any) (*App, error) {
	switch id := OID(appId).(type) {
	case uint32:
		return d.GetAppById(uint32(id))
	case string:
		return d.GetAppByName(id)
	case *App:
		return id, nil
	default:
		return nil, &diwe.ErrUnknownApp{AppId: appId}
	}
}

// GetAppById returns an application by its numeric ID.
// Returns ErrUnknownApp if the application is not found.
func (d *Dict) GetAppById(appId uint32) (*App, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if app, ok := d.cache.appCacheById[appId]; ok {
		return app, nil
	}

	return nil, &diwe.ErrUnknownApp{AppId: appId}
}

// GetAppByName returns an application by its name (case-insensitive).
// Returns ErrUnknownApp if the application is not found.
func (d *Dict) GetAppByName(appName string) (*App, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if app, ok := d.cache.appCacheByName[strings.ToLower(appName)]; ok {
		return app, nil
	}

	return nil, &diwe.ErrUnknownApp{AppId: appName}
}

// GetCmd retrieves a command by its code (uint32) or name (string) within a specific application.
// The app parameter is required to scope the command lookup to that application.
// It first attempts to parse string inputs as numeric codes, then falls back to name lookup.
func (d *Dict) GetCmd(cmdId any, app *App) (*Command, error) {
	if app == nil {
		return nil, &diwe.ErrUnknownApp{AppId: "nil"}
	}

	switch id := OID(cmdId).(type) {
	case uint32:
		return d.GetCmdByCode(id, app)
	case string:
		return d.GetCmdByName(id, app)
	case *Command:
		return id, nil
	default:
		return nil, &diwe.ErrUnknownCmd{App: app.Name, CmdId: cmdId}
	}
}

// GetCmdByCode returns a command by its numeric code within the given application.
// The app must not be nil. Returns ErrUnknownCmd if not found.
func (d *Dict) GetCmdByCode(cmdCode uint32, app *App) (*Command, error) {
	if app == nil {
		return nil, &diwe.ErrUnknownApp{AppId: "nil"}
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	if cmds, ok := d.cache.cmdCacheByCode[app.Id]; ok {
		if cmd, ok := cmds[cmdCode]; ok {
			return cmd, nil
		}
	}

	return nil, &diwe.ErrUnknownCmd{App: app.Name, CmdId: cmdCode}
}

// GetCmdByName returns a command by its name (case-insensitive) within the given application.
// The app must not be nil. Returns ErrUnknownCmd if not found.
func (d *Dict) GetCmdByName(cmdName string, app *App) (*Command, error) {
	if app == nil {
		return nil, &diwe.ErrUnknownApp{AppId: "nil"}
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	if cmds, ok := d.cache.cmdCacheByName[app.Id]; ok {
		if cmd, ok := cmds[strings.ToLower(cmdName)]; ok {
			return cmd, nil
		}
	}

	return nil, &diwe.ErrUnknownCmd{App: app.Name, CmdId: cmdName}
}

// GetAvp retrieves an AVP by its code (uint32) or name (string).
// It first attempts to parse string inputs as numeric codes, then falls back to name lookup.
func (d *Dict) GetAvp(avpId any) (*Avp, error) {
	switch id := OID(avpId).(type) {
	case uint32:
		return d.GetAvpByCode(id)
	case string:
		return d.GetAvpByName(id)
	case *Avp:
		return id, nil
	default:
		return nil, &diwe.ErrUnknownAvp{Avp: avpId}
	}
}

// GetAvpByCode returns an AVP by its numeric code.
// Returns ErrUnknownAvp if the AVP is not found.
func (d *Dict) GetAvpByCode(avpCode uint32) (*Avp, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if avp, ok := d.cache.avpCacheByCode[avpCode]; ok {
		return avp, nil
	}

	return nil, &diwe.ErrUnknownAvp{Avp: avpCode}
}

// GetAvpByName returns an AVP by its name (case-insensitive).
// Returns ErrUnknownAvp if the AVP is not found.
func (d *Dict) GetAvpByName(avpName string) (*Avp, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if avp, ok := d.cache.avpCacheByName[strings.ToLower(avpName)]; ok {
		return avp, nil
	}

	return nil, &diwe.ErrUnknownAvp{Avp: avpName}
}

// GetEnumCode returns the code for an enumerated item by its name (case-insensitive).
// Returns ErrUnknownEnumItem if the AVP is not enumerated or the name is not found.
func (d *Dict) GetEnumCode(avpCode uint32, itemName string) (int32, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if items, ok := d.cache.enumCache[avpCode]; ok {
		if code, ok := items[strings.ToLower(itemName)]; ok {
			return code, nil
		}
	}

	return 0, &diwe.ErrUnknownEnumItem{Avp: fmt.Sprintf("%d", avpCode), Value: itemName}
}

// These methods return the flag and type definitions from the dictionary core.
// They are read-only and return copies or values rather than pointers.

// AvpFlag returns the AVP bit flag definitions (V, M, P flags).
func (d *Dict) AvpFlag() AvpBitFlags {
	return d.core.GetAvpFlags()
}

// CmdFlag returns the Command bit flag definitions (R, P, E, T flags).
func (d *Dict) CmdFlag() CmdBitFlags {
	return d.core.GetCmdFlags()
}

// AvpDataType returns the AVP data type definitions (e.g., Integer32, UTF8String, Grouped).
func (d *Dict) AvpDataType() AvpDataTypes {
	return d.core.GetAvpTypes()
}

// Iterators
//
// These methods return sequences for iterating over dictionary items.
// They hold a read lock for the duration of iteration.
// Note: The lock is held while the yield function executes, so avoid
// calling back into Dict methods from within the iterator.

// AppIter returns a sequence that yields all applications in the dictionary.
func (d *Dict) AppIter() iter.Seq[*App] {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return func(yield func(*App) bool) {
		for i := range d.core.GetApps() {
			if !yield(&d.core.GetApps()[i]) {
				break
			}
		}
	}
}

// AppIter2 returns a sequence that yields index-application pairs for all applications.
func (d *Dict) AppIter2() iter.Seq2[int, *App] {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return func(yield func(int, *App) bool) {
		for i := range d.core.GetApps() {
			if !yield(i, &d.core.GetApps()[i]) {
				break
			}
		}
	}
}

// CmdIter returns a sequence that yields all commands for the given application.
// Returns an empty sequence if app is nil.
func (d *Dict) CmdIter(app *App) iter.Seq[*Command] {
	if app == nil {
		return func(yield func(*Command) bool) {}
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	return func(yield func(*Command) bool) {
		for i := range app.Cmds {
			if !yield(&app.Cmds[i]) {
				break
			}
		}
	}
}

// CmdIter2 returns a sequence that yields index-command pairs for the given application.
func (d *Dict) CmdIter2(app *App) iter.Seq2[int, *Command] {
	if app == nil {
		return func(yield func(int, *Command) bool) {}
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	return func(yield func(int, *Command) bool) {
		for i := range app.Cmds {
			if !yield(i, &app.Cmds[i]) {
				break
			}
		}
	}
}

// AvpIter returns a sequence that yields all AVPs in the dictionary.
func (d *Dict) AvpIter() iter.Seq[*Avp] {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return func(yield func(*Avp) bool) {
		for i := range d.core.GetAvps() {
			if !yield(&d.core.GetAvps()[i]) {
				break
			}
		}
	}
}

// AvpIter2 returns a sequence that yields index-AVP pairs for all AVPs.
func (d *Dict) AvpIter2() iter.Seq2[int, *Avp] {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return func(yield func(int, *Avp) bool) {
		for i := range d.core.GetAvps() {
			if !yield(i, &d.core.GetAvps()[i]) {
				break
			}
		}
	}
}

// Data information helpers
//
// These methods provide human-readable names for flags and types.

// CmdFlagName returns the human-readable name for a command bit flag.
// Returns "Unknown" if the flag value is not recognized.
func (d *Dict) CmdFlagName(flag uint8) string {
	if name, exists := d.flags.cmd[flag]; exists {
		return name
	}
	return "Unknown"
}

// AvpFlagName returns the human-readable name for an AVP bit flag.
// Returns "-" if the flag value is not recognized.
func (d *Dict) AvpFlagName(flag uint8) string {
	if name, exists := d.flags.avp[flag]; exists {
		return name
	}
	return "-"
}

// AvpDataTypeName returns the human-readable name for an AVP data type ID.
// Uses reflection to map type IDs to field names in AvpDataTypes.
// Returns "Unknown" if the type ID is out of range.
func (d *Dict) AvpDataTypeName(id int) string {
	t := reflect.TypeOf(AvpDataTypes{})
	if id <= t.NumField() {
		return t.Field(id - 1).Name
	} else {
		return "Unknown"
	}
}

// Debug helpers
//
// These methods provide human-readable output for debugging purposes.
// They are not thread-efficient and should be used for debugging only.

// Show prints a compact summary of all applications and their commands to stdout.
func (d *Dict) Show() {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, app := range d.core.GetApps() {
		fmt.Printf("Application: %s (%d)\n", app.Name, app.Id)
		for _, cmd := range app.Cmds {
			fmt.Printf("  %s: %s (%d)\n", cmd.Short, cmd.Name, cmd.Code)
		}
		fmt.Println()
	}
}

// Dump prints a detailed dump of the entire dictionary to stdout,
// including all applications, commands, AVPs, and their nested structures
// (enumerations and grouped AVP members).
// The optional shift parameter is unused but kept for API compatibility.
func (d *Dict) Dump(shift ...int) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	fmt.Println("\n>>> Applications")
	for _, app := range d.core.GetApps() {
		fmt.Printf("id=%d \"%s\"\n", app.Id, app.Name)
		for _, cmd := range app.Cmds {
			fmt.Printf("\tcode=%d \"%s\" \"%s\"\n", cmd.Code, cmd.Short, cmd.Name)
			fmt.Println("\t\tRequest")
			for _, avp := range cmd.Request {
				fmt.Printf("\t\t\t\"%s\"\n", avp.Name)
			}
			fmt.Println("\t\tAnswer")
			for _, avp := range cmd.Answer {
				fmt.Printf("\t\t\t\"%s\"\n", avp.Name)
			}
		}
	}

	fmt.Println("\n>>> AVPs")
	for _, avp := range d.core.GetAvps() {
		fmt.Printf("code=%d name=\"%s\" flags=%d type=%d\n", avp.Code, avp.Name, avp.Flags, avp.Type)
		if avp.Type == d.AvpDataType().Enumerated {
			for _, item := range avp.Enum.Items {
				fmt.Printf("\tcode=%d name=\"%s\"\n", item.Code, item.Name)
			}
		}
		if avp.Type == d.AvpDataType().Grouped {
			for _, member := range avp.Group.Members {
				fmt.Printf("\tname=\"%s\"\n", member.Name)
			}
		}
	}
}

// AddMember adds a new member rule to a grouped AVP.
// The name specifies the member AVP name, required indicates if it's mandatory,
// and max specifies the maximum occurrence count (uses pointer to heap, see FIXME).
func (avp *Avp) AddMember(name string, required bool, max int) {
	member := AvpRule{Name: name, Required: required, Max: &max} // FIXME: esc to heap :(
	avp.Group.Members = append(avp.Group.Members, member)
}

// RemoveMember removes a member rule from a grouped AVP by name (case-insensitive).
// If multiple members with the same name exist, only the first one is removed.
func (avp *Avp) RemoveMember(name string) {
	for i, member := range avp.Group.Members {
		if strings.EqualFold(member.Name, name) {
			avp.Group.Members = slices.Delete(avp.Group.Members, i, i+1)
		}
	}
}

// Loaders
//
// These methods load dictionary data from external files in various formats.

// LoadFromFile loads dictionary data from a file in the specified format.
// Supported formats are FormatPkl, FormatJson, and FormatYaml.
// This method acquires an exclusive lock and rebuilds the lookup caches.
func (d *Dict) LoadFromFile(file string, format int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if loader, exists := loaders[format]; exists {
		return loader(d, file)
	}

	return &diwe.ErrUnknownFmt{Fmt: format}
}

// loadFromPkl loads dictionary data from a Pkl file.
// It reads the core data, replaces the existing core, and rebuilds caches.
func loadFromPkl(d *Dict, pklFile string) error {
	core, err := LoadFromPath(context.Background(), pklFile)
	if err != nil {
		return err
	}
	d.core = core
	d.fillLookupCache()
	d.fillFlagsNames()

	return nil
}

// loadFromJson loads dictionary data from a JSON file (not implemented).
func loadFromJson(d *Dict, jsonFile string) error {
	return &diwe.ErrNotImplemented{}
}

// loadFromYaml loads dictionary data from a YAML file (not implemented).
func loadFromYaml(d *Dict, yamlFile string) error {
	return &diwe.ErrNotImplemented{}
}

//  Private methods
//

// fillLookupCache rebuilds all lookup caches from the core dictionary data.
// This must be called after modifying the core data.
func (d *Dict) fillLookupCache() {
	d.cache.appCacheById = make(map[uint32]*App)
	d.cache.appCacheByName = make(map[string]*App)
	d.cache.cmdCacheByCode = make(map[uint32]map[uint32]*Command)
	d.cache.cmdCacheByName = make(map[uint32]map[string]*Command)
	d.cache.avpCacheByCode = make(map[uint32]*Avp)
	d.cache.avpCacheByName = make(map[string]*Avp)
	d.cache.enumCache = make(map[uint32]map[string]int32)

	for i := range d.core.GetApps() {
		app := &d.core.GetApps()[i]
		d.cache.appCacheById[app.Id] = app
		d.cache.appCacheByName[strings.ToLower(app.Name)] = app

		d.cache.cmdCacheByCode[app.Id] = make(map[uint32]*Command)
		d.cache.cmdCacheByName[app.Id] = make(map[string]*Command)

		for j := range d.core.GetApps()[i].Cmds {
			cmd := &d.core.GetApps()[i].Cmds[j]
			d.cache.cmdCacheByCode[app.Id][cmd.Code] = cmd
			d.cache.cmdCacheByName[app.Id][strings.ToLower(cmd.Short)] = cmd
		}
	}

	for i := range d.core.GetAvps() {
		avp := &d.core.GetAvps()[i]
		d.cache.avpCacheByCode[avp.Code] = avp
		d.cache.avpCacheByName[strings.ToLower(avp.Name)] = avp

		if avp.Type == d.core.GetAvpTypes().Enumerated && avp.Enum != nil {
			d.cache.enumCache[avp.Code] = make(map[string]int32)
			for _, item := range avp.Enum.Items {
				d.cache.enumCache[avp.Code][strings.ToLower(item.Name)] = item.Code
			}
		}
	}
}

// fillFlagsNames populates the flag name mappings from the core flag definitions.
// This must be called after the core data is loaded.
func (d *Dict) fillFlagsNames() {
	d.flags.avp = map[uint8]string{
		d.AvpFlag().V: "V",
		d.AvpFlag().M: "M",
		d.AvpFlag().P: "P",
	}

	d.flags.cmd = map[uint8]string{
		d.CmdFlag().R: "Request",
		d.CmdFlag().P: "Proxyable",
		d.CmdFlag().E: "Error",
		d.CmdFlag().T: "Retransmission",
	}
}

// Helpers
//

// OID makes the appropriate Diameter object ID type (uint32 or string) from a Go lang type.
func OID(id any) any {
	if id == nil {
		return nil
	}

	switch v := id.(type) {
	case int:
		return uint32(v)
	case uint:
		return uint32(v)
	case int8:
		return uint32(v)
	case uint8:
		return uint32(v)
	case int16:
		return uint32(v)
	case uint16:
		return uint32(v)
	case int32:
		return uint32(v)
	case uint32:
		return uint32(v)
	case int64:
		return uint32(v)
	case uint64:
		return uint32(v)
	case string:
		if n, err := strconv.ParseUint(v, 10, 32); err == nil {
			return uint32(n)
		}
		return v
	}

	return id
}
