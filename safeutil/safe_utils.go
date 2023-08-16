package safeutil

import (
	"encoding/json"
	"io"
	"net/http"
)

func Read(resp *http.Response) ([]byte, error) {
	if resp == nil {
		return nil, nil
	}
	if resp.Body == nil {
		return nil, nil
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func Unmarshal(resp *http.Response, target any) error {
	bs, err := Read(resp)
	if err != nil {
		return err
	}

	return json.Unmarshal(bs, target)
}
