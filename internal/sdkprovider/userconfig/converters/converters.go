// This package provides functions that convert terraform values to JSON and back (from TF to DTO).
// Prerequisites:
// - in terraform a single _object_ is a list with a single element.
//   Hence, `{...}` must be turned into `[{...}]` when value comes _from_ JSON
//   and _unwrapped_ when it goes to JSON
// - terraform doesn't support multiple types per field.
//   Those fields are split into per-type fields: foo_string, foo_object.
//   This means those "virtual" fields must be mapped to a "real" field.
// - A list of objects might change its element order when data is fetched from Aiven.
//   In that case, values _must_ be sorted according to the state values.
//   Otherwise, TF will output a diff.
// - A userconfig values can't be removed on Aiven side.
//   They can be overridden only.

package converters

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const userConfigSuffix = "_user_config"

// Expand expands schema.ResourceData into a DTO map.
// It takes schema.Schema to know how to turn a TF item into json.
func Expand(kind string, s *schema.Schema, d *schema.ResourceData) (map[string]any, error) {
	key := kind + userConfigSuffix
	state := &stateCompose{
		key:      key,
		path:     key + ".0", // starts from root user config
		schema:   s,
		resource: d,
	}

	// When "configs" is empty, then we need to delete all arrays in it.
	// That's why it doesn't exit here.
	configs := d.GetRawConfig().GetAttr(key).AsValueSlice()
	if len(configs) > 0 {
		state.config = configs[0]
	}

	dto, err := expandObj(state)
	if err != nil {
		return nil, err
	}

	renameAliases(dto)
	return dto, nil
}

// stateCompose combines "raw state" and schema.ResourceData
// With the state it is possible to say "if value is null", hence if it is defined by user.
// With schema.ResourceData you get the value, that might be a zero-value.
type stateCompose struct {
	key      string // state attribute name or schema.ResourceData key
	path     string // schema.ResourceData path, i.e. foo.0.bar.0.baz to get the value
	schema   *schema.Schema
	config   cty.Value
	resource *schema.ResourceData
}

// setItems returns schema.Set values that has state.
// Supports int and float types only.
func (s *stateCompose) setItems() []any {
	result := make([]any, 0)
	if s.config.IsNull() {
		// Makes possible to send ip_filter=[] to clear the remote list.
		return result
	}

	// Builds elements hash map
	hashes := make(map[string]bool, s.config.LengthInt())
	for _, item := range s.config.AsValueSlice() {
		if item.Type() == cty.String {
			hashes[item.AsString()] = true
		} else {
			hashes[item.AsBigFloat().String()] = true
		}
	}

	// Picks up values with a state only
	for _, v := range s.value().(*schema.Set).List() {
		if hashes[fmt.Sprintf("%v", v)] {
			result = append(result, v)
		}
	}
	return result
}

// listItems returns a list of object's states
// Must not use it with scalar types, because "schema" expects to have Resource
func (s *stateCompose) listItems() (result []*stateCompose) {
	if s.config.IsNull() {
		return result
	}

	for i, v := range s.config.AsValueSlice() {
		c := &stateCompose{
			key:      s.key,
			path:     fmt.Sprintf("%s.%d", s.path, i),
			schema:   s.schema, // object is a list with one item, hence same schema
			config:   v,
			resource: s.resource,
		}
		result = append(result, c)
	}
	return result
}

// objectProperties returns object's properties states
func (s *stateCompose) objectProperties() map[string]*stateCompose {
	props := make(map[string]*stateCompose)
	res := s.schema.Elem.(*schema.Resource)
	for key, subSchema := range res.Schema {
		if subSchema.ForceNew && !s.resource.IsNewResource() {
			continue
		}

		var config cty.Value
		if !s.config.IsNull() {
			// Can't get value from nil
			config = s.config.GetAttr(key)
		}

		p := &stateCompose{
			key:      key,
			path:     fmt.Sprintf("%s.%s", s.path, key),
			resource: s.resource,
			config:   config,
			schema:   subSchema,
		}

		props[key] = p
	}
	return props
}
func (s *stateCompose) value() any {
	return s.resource.Get(s.path)
}

func (s *stateCompose) hasValue() bool {
	// By some reason iterable values are always not null
	return s.config.IsNull() || s.config.CanIterateElements() && s.config.LengthInt() == 0
}

func (s *stateCompose) hasChange() bool {
	return s.resource.HasChange(s.path)
}

func expandObj(state *stateCompose) (map[string]any, error) {
	m := make(map[string]any)
	for k, v := range state.objectProperties() {
		value, err := expandAttr(v)
		if err != nil {
			return nil, fmt.Errorf("%q field conversion error: %w", k, err)
		}
		if value != nil {
			m[k] = value
		}
	}
	return m, nil
}

func expandScalar(state *stateCompose) (any, error) {
	if state.hasValue() {
		// Null scalar, no value in the config
		return nil, nil
	}
	return state.value(), nil
}

// expandAttr returns go value
func expandAttr(state *stateCompose) (any, error) {
	switch state.schema.Type {
	case schema.TypeSet, schema.TypeList:
	default:
		return expandScalar(state)
	}

	if state.hasValue() && !state.hasChange() {
		// A value that hasn't been sent by user yet.
		// But has been received from the API.
		return nil, nil
	}

	if state.schema.Type == schema.TypeSet {
		return state.setItems(), nil
	}

	// schema.TypeList
	_, isObjList := state.schema.Elem.(*schema.Resource)
	states := state.listItems()
	items := make([]any, 0, len(states))
	for i := range states {
		var exp any
		var err error
		if isObjList {
			exp, err = expandObj(states[i])
		} else {
			exp, err = expandScalar(states[i])
		}

		if err != nil {
			return nil, err
		}

		// If an object is not empty
		if exp != nil {
			items = append(items, exp)
		}
	}

	// If schema.TypeList && MaxItems == 1, then it is an object
	if isObjList && state.schema.MaxItems == 1 {
		switch len(items) {
		case 1:
			// A plain object (in TF a list with one object is an object)
			return items[0], nil
		case 0:
			// The object has no state or removed.
			// We can't remove objects from state, so send a nil.
			return nil, nil
		}
	}

	return items, nil
}

// Flatten flattens DTO into a terraform compatible object
func Flatten(kind string, s *schema.Schema, d *schema.ResourceData, dto map[string]any) ([]map[string]any, error) {
	prefix := fmt.Sprintf("%s%s.0.", kind, userConfigSuffix)

	// Renames ip_filter field
	if _, ok := dto["ip_filter"]; ok {
		assignAlias(d, prefix+"ip_filter", dto, "ip_filter", "network")
	}

	// Renames namespaces field
	if mapping, ok := drillKey(dto, "rules.0.mapping"); ok {
		assignAlias(d, prefix+"rules.0.mapping.0.namespaces", mapping.(map[string]any), "namespaces", "name")
	}

	// Copies "create only" fields from the original config.
	// Like admin_password, that is received only on POST request when service is created.
	for _, k := range createOnlyFields() {
		v, ok := d.GetOk(prefix + k)
		if ok {
			dto[k] = v
		}
	}

	r := s.Elem.(*schema.Resource)
	tfo, err := flattenObj(r.Schema, dto)
	if tfo == nil || err != nil {
		return nil, err
	}
	return []map[string]any{tfo}, nil
}

func flattenObj(s map[string]*schema.Schema, dto map[string]any) (map[string]any, error) {
	tfo := make(map[string]any)
	for k, v := range s {
		vv, ok := dto[k]
		if !ok {
			continue
		}

		if vv == nil {
			continue
		}

		value, err := flattenAttr(v, vv)
		if err != nil {
			return nil, fmt.Errorf("%q field conversion error: %w", k, err)
		}

		if value != nil {
			tfo[k] = value
		}
	}
	if len(tfo) == 0 {
		return nil, nil
	}
	return tfo, nil
}

func flattenAttr(s *schema.Schema, data any) (any, error) {
	switch s.Type {
	case schema.TypeString:
		return castType[string](data)
	case schema.TypeBool:
		return castType[bool](data)
	case schema.TypeInt:
		i, err := data.(json.Number).Int64()
		return int(i), err
	case schema.TypeFloat:
		return data.(json.Number).Float64()
	}

	// A set can contain scalars only
	scalar, scalarOk := s.Elem.(*schema.Schema)
	if scalarOk {
		switch s.Type {
		case schema.TypeList:
			return data.([]any), nil
		case schema.TypeSet:
			values := make([]any, 0)
			for _, v := range data.([]any) {
				val, err := flattenAttr(scalar, v)
				if err != nil {
					return nil, err
				}
				values = append(values, val)
			}
			return schema.NewSet(schema.HashSchema(scalar), values), nil
		}
	}

	// Single object or list of objects
	resource := s.Elem.(*schema.Resource)
	var list []any
	if o, isObject := data.(map[string]any); isObject {
		// Single object, but it is a list with one element for terraform
		if len(o) != 0 {
			list = append(list, o)
		}
	} else {
		// List of objects
		list = data.([]any)
	}

	return flattenList(resource.Schema, list)
}

func flattenList(s map[string]*schema.Schema, list []any) ([]any, error) {
	if len(list) == 0 {
		return nil, nil
	}

	items := make([]any, 0, len(list))
	for _, item := range list {
		v, err := flattenObj(s, item.(map[string]any))
		if err != nil {
			return nil, err
		}
		if v != nil {
			items = append(items, v)
		}
	}
	return items, nil
}

// assignAlias renames keys for multi-typed properties, i.e. ip_filter -> [ip_filter_string, ip_filter_object]
func assignAlias(d *schema.ResourceData, path string, dto map[string]any, key, sortBy string) {
	values, ok := dto[key].([]any)
	if !ok || len(values) == 0 {
		return
	}

	var suffix string
	const (
		str = "_string"
		obj = "_object"
	)

	// If DTO has objects, then it is foo_object
	if _, ok := values[0].(map[string]any); ok {
		suffix = obj

		// State objects have specific order.
		// Must sort DTO objects, otherwise diff shows changes.
		if inStateObjs, ok := d.GetOk(path + obj); ok {
			dto[key] = sortByKey(sortBy, inStateObjs, dto[key])
		}
	}

	// If the state has foo_string, the user has new key
	if _, ok := d.GetOk(path + str); ok {
		suffix = str
	}

	if suffix != "" {
		dto[key+suffix] = dto[key]
		delete(dto, key)
	}
}

// createOnlyFields these fields are received on POST request only
func createOnlyFields() []string {
	return []string{
		"admin_username",
		"admin_password",
	}
}

func aliasFieldsMap() map[string]string {
	return map[string]string{
		"ip_filter_string":                    "ip_filter",
		"ip_filter_object":                    "ip_filter",
		"rules.0.mapping.0.namespaces_string": "namespaces",
		"rules.0.mapping.0.namespaces_object": "namespaces",
	}
}
