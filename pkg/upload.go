package insights

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

func Upload(client *Client, file string, collector string) error {
	URL, err := url.Parse(client.baseURL)
	if err != nil {
		return err
	}
	URL.Path = path.Join(URL.Path, "/ingress/v1/upload")

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			"file", filepath.Base(file)))
	if collector != "" {
		h.Set("Content-Type",
			fmt.Sprintf("application/vnd.redhat.%s.collection+tgz", collector))
	} else {
		h.Set("Content-Type", "application/octet-stream")
	}
	pw, err := w.CreatePart(h)
	if err != nil {
		return err
	}

	if _, err := io.Copy(pw, f); err != nil {
		return err
	}

	facts, err := json.Marshal(getCanonicalFacts())
	if err != nil {
		return err
	}

	if err := w.WriteField("metadata", string(facts)); err != nil {
		return err
	}
	w.Close()

	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, URL.String(), &buf)
	req.Header.Add("Content-Type", w.FormDataContentType())

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	switch res.StatusCode {
	case http.StatusAccepted:
		break
	default:
		return fmt.Errorf("%v: %v", http.StatusText(res.StatusCode), string(data))
	}

	return nil
}
