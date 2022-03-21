package uviews

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/usfsci/ustore"
)

func TestMessage(t *testing.T) {
	b, _ := ustore.NewBase()

	c := ustore.Client{
		Base: *b,
		Name: "A Name",
	}

	m := struct {
		Timestamp int64       `json:"timestamp"`
		Data      interface{} `json:"data,omitempty"`
	}{
		Timestamp: time.Now().In(time.UTC).Unix(),
		Data:      &c,
	}
	jm, err := json.Marshal(m)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("%s\n", jm)

	rm := Message{}
	if err := json.Unmarshal(jm, &rm); err != nil {
		t.Error(err)
		return
	}

	nc := &ustore.Client{}
	if err := json.Unmarshal(rm.Data, &nc); err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("Timestamp: %d\n", rm.Timestamp)

	fmt.Printf("Client:\n%+v\n", nc)
}
