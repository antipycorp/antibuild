package site

import (
	"fmt"
	"reflect"
	"testing"
)

var testData = Data{
	"Data": "n0thing",
	"Data2": Data{
		"more": "data",
	},
}

func TestToFromv1(t *testing.T) {
	values := make(map[uint64]Value)
	keys := make(map[uint64]Key)

	_, err := getmapv1(values, keys, &testData)
	if err != nil {
		t.FailNow()
	}

	data, err := tomapv1(values, keys)
	if err != nil {
		t.FailNow()
	}

	if !reflect.DeepEqual(data, testData) {
		t.FailNow()
	}
}

func TestEncDecv1(t *testing.T) {
	bytes := encodev1(&testData)
	data := decodev1(bytes)
	if !reflect.DeepEqual(data, testData) {
		fmt.Println(data)
		fmt.Println(testData)
		t.FailNow()
	}
}
