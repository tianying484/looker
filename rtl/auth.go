package rtl

import (
    "bytes"
    "crypto/tls"
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "net/url"
    "os"
    "reflect"
    "time"
)

type AccessToken struct {
    AccessToken string `json:"access_token"`
    TokenType   string `json:"token_type"`
    ExpiresIn   int32  `json:"expires_in"`
    ExpireTime  time.Time
}

func (t AccessToken) IsExpired() bool {
    return t.ExpireTime.IsZero() || time.Now().After(t.ExpireTime)
}

func NewAccessToken(js []byte) (AccessToken, error) {
    token := AccessToken{}
    if err := json.Unmarshal(js, &token); err != nil {
        return token, err
    }
    token.ExpireTime = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
    return token, nil
}

type AuthSession struct {
    Config ApiSettings
    token  AccessToken
}

func NewAuthSession(config ApiSettings) *AuthSession {
    return &AuthSession{
        Config: config,
    }
}

func (s *AuthSession) login(id *string) error {
    u := fmt.Sprintf("%s/api/%s/login", s.Config.BaseUrl, s.Config.ApiVersion)
    data := url.Values{
        "client_id":     {s.Config.ClientId},
        "client_secret": {s.Config.ClientSecret},
    }
    tran := &(*http.DefaultTransport.(*http.Transport))
    tran.TLSClientConfig = &tls.Config{InsecureSkipVerify: !s.Config.VerifySsl}
    cl := http.Client{
        Transport: tran,
        Timeout:   time.Duration(s.Config.Timeout) * time.Second,
    }
    res, err := cl.PostForm(u, data)
    if err != nil {
        return err
    }

    if res.StatusCode != http.StatusOK {
        return fmt.Errorf("status not OK: %s", res.Status)
    }

    defer res.Body.Close()
    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        return fmt.Errorf("error reading response Body: %w", err)
    }

    s.token, err = NewAccessToken(body)

    return err
}

// Authenticate checks if the token is expired (do the token refresh if so), and updates the request Header with Authorization
func (s *AuthSession) Authenticate(req *http.Request) error {
    if s.token.IsExpired() {
        if err := s.login(nil); err != nil {
            return err
        }
    }
    req.Header.Add("Authorization", fmt.Sprintf("token %s", s.token.AccessToken))
    return nil
}

func (s *AuthSession) Do(result interface{}, method, ver, path string, reqPars map[string]interface{}, body interface{}, options *ApiSettings) error {

    // prepare URL
    u := fmt.Sprintf("%s/api%s%s", s.Config.BaseUrl, ver, path)

    lookerRequest, err := serializeBody(body)
    if err != nil {
        return err
    }

    // create new request
    httpRequest, err := http.NewRequest(method, u, bytes.NewBufferString(lookerRequest.Body))
    if err != nil {
        return err
    }
    // add header with body request
    for k, v := range lookerRequest.Header {
        httpRequest.Header[k] = v
    }
    httpRequest.Header.Add("Accept", "application/json")
    httpRequest.Header.Add("Accept", "text")

    // set query params
    setQuery(httpRequest.URL, reqPars)

    // set auth Header
    if err = s.Authenticate(httpRequest); err != nil {
        return err
    }

    tran := &(*http.DefaultTransport.(*http.Transport))
    tran.TLSClientConfig = &tls.Config{InsecureSkipVerify: !s.Config.VerifySsl}
    httpClient := http.Client{
        Transport: tran,
        Timeout:   time.Duration(s.Config.Timeout) * time.Second,
    }

    // do the actual http call
    res, err := httpClient.Do(httpRequest)
    if err != nil {
        return err
    }

    if res.StatusCode < 200 || res.StatusCode > 226 {
        return fmt.Errorf("response error: %s", res.Status)
    }

    defer res.Body.Close()

    if strResult, ok := result.(*string); ok {
        return decodeQueryResult(strResult, res.Body)
    }

    //respBody, err := ioutil.ReadAll(res.Body)
    //if err != nil {
    //    return err
    //}
    //fmt.Println(string(respBody))
    //
    //return json.Unmarshal(respBody, result)
    return json.NewDecoder(res.Body).Decode(result)
}

func decodeQueryResult(result *string, body io.Reader) error {
    var (
        err      error
        respBody []byte
    )
    if respBody, err = ioutil.ReadAll(body); err != nil {
        return err
    }
    *result = string(respBody)
    return nil
}

type lookerRequest struct {
    Header http.Header
    Body   string
}

// serializeBody serializes Body to a json, if the Body is already string, it will just return it unchanged
func serializeBody(body interface{}) (ret lookerRequest, err error) {
    ret.Header = make(http.Header)
    if body == nil {
        return ret, nil
    }

    // get the `Body` type
    kind := reflect.TypeOf(body).Kind()
    value := reflect.ValueOf(body)

    // check if it is pointer
    if kind == reflect.Ptr {
        // if so, use the value kind
        kind = reflect.ValueOf(body).Elem().Kind()
        value = reflect.ValueOf(body).Elem()
    }

    // it is string, return it as it is
    if kind == reflect.String {
        ret.Body = fmt.Sprintf("%v", value)
        return ret, nil
    }

    bb, err := json.Marshal(body)
    if err != nil {
        _, _ = fmt.Fprintf(os.Stderr, "error serializing Body: %v", err)
    }
    ret.Header.Add("Content-Type", "application/json")
    ret.Body = string(bb)

    return ret, nil
}

// setQuery takes the provided parameter map and sets it as query parameters of the provided url
func setQuery(u *url.URL, pars map[string]interface{}) {

    if pars == nil || u == nil {
        return
    }

    q := u.Query()
    for k, v := range pars {
        // skip nil
        if v == nil || reflect.ValueOf(v).IsNil() {
            continue
        }
        // marshal the value to json
        jsn, err := json.Marshal(v)
        if err != nil {
            _, _ = fmt.Fprintf(os.Stderr, "error serializing parameter: %s, error: %v", k, err)
        }
        q.Add(k, string(jsn))
    }
    u.RawQuery = q.Encode()
}
