package utils

import "encoding/json"

// itemWrapper, used to wrap the items in a json items[] array (or retrieve/unwrap them)
type itemWrapper[T any] struct {
	Items []T `json:"items"`
}

// wrapItems will wrapp all the items in a json items[] array
// returns a itemWrapper, which is a json object with the items[] array
func WrapItems(item any) *itemWrapper[any] {
	iw := new(itemWrapper[any])
	iw.Items = append(iw.Items, item)
	return iw
}

// unwrapItems will unwrap the items from the json items[] array
// input is the json output from the API (GET) call as []byte
// returns a slice of items
func UnwrapItems[T any](jsonAsByte []byte) ([]*T, error) {
	var iw itemWrapper[T]
	err := json.Unmarshal(jsonAsByte, &iw)

	if err != nil {
		return nil, err
	}

	var items []*T
	for _, item := range iw.Items {
		i := item //Since the pointer on item moves
		items = append(items, &i)
	}

	return items, nil
}
