package main

import (
	"strconv"
	"testing"
)

type PointForMap struct {
	tags map[string]string
}

type Label struct {
	k, v string
}

type PointForStruct struct {
	tags []*Label
}

func BenchmarkMap(b *testing.B) {
	p := &PointForMap{
		tags: make(map[string]string),
	}

	for i := 0; i < 10000; i++ {
		p.tags[strconv.Itoa(i)] = strconv.Itoa(i)
	}

	for i := 0; i < b.N; i++ {
		if k, found := p.tags["dc"]; found {
			p.tags[k] = "hello"
		}
	}
}

func BenchmarkStruct(b *testing.B) {
	p := &PointForStruct{
		tags: make([]*Label, 0, 1),
	}
	p.tags = append(p.tags, &Label{"dc", "bj"})
	for i := 0; i < 10000; i++ {
		p.tags = append(p.tags, &Label{strconv.Itoa(i), strconv.Itoa(i)})
	}

	for i := 0; i < b.N; i++ {
		for _, item := range p.tags {
			if item.k == "dc" {
				item.v = "bj"
			}
		}
	}
}
