package data

import (
	"github.com/medubin/gonzo/utils/url"
)

type Data struct {
	InStruct  bool
	InServer  bool
	Structs   []*Struct
	Endpoints []*Endpoint
	Variables []*Variable
}

type Endpoint struct {
	Verb   string
	Url    string
	Name   string
	Body   string
	Return string
}

type Struct struct {
	Name   string
	Type   string
	Fields []string
}

type Variable struct {
	Name string
	Type string
}

func (o *Data) AddStruct(name string) {
	o.Structs = append(o.Structs, &Struct{
		Name: name,
		Type: "struct {",
	})
}

func (o *Data) AddVariable(name string, typeName string) {
	o.Variables = append(o.Variables, &Variable{
		Name: name,
		Type: typeName,
	})
}

func (o *Data) AddStructField(name string, typeName string) {
	lastStruct := o.Structs[len(o.Structs)-1]
	lastStruct.Fields = append(lastStruct.Fields, name, typeName)
}

func (o *Data) AddEndpoint(e *Endpoint) {
	o.Endpoints = append(o.Endpoints, e)

	matches := url.GetKeys(e.Url)
	fields := make([]string, len(matches)*2)
	for i, match := range matches {
		fields[i*2] = match
		fields[i*2+1] = "string"
	}

	o.Structs = append(o.Structs, &Struct{
		Name:   e.Name + "Url",
		Type:   "struct {",
		Fields: fields,
	})
}

func (o *Data) FinishVariable() {
	o.InStruct = false
	o.InServer = false
}
