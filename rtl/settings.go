package rtl

import (
    "fmt"
    "gopkg.in/ini.v1"
)

var defaultSectionName string = "Looker"

type ApiSettings struct {
    BaseUrl      string `ini:"base_url"`
    VerifySsl    bool   `ini:"verify_ssl"`
    Timeout      int32  `ini:"timeout"`
    AgentTag     string `ini:"agent_tag"`
    FileName     string `ini:"file_name"`
    ClientId     string `ini:"client_id"`
    ClientSecret string `ini:"client_secret"`
    ApiVersion   string `ini:"api_version"`
}

func NewSettings() *ApiSettings {
    return &ApiSettings{}
}

func (a *ApiSettings) WithBaseUrl(BaseUrl string) *ApiSettings {
    a.BaseUrl = BaseUrl
    return a.withApiVersion()
}
func (a *ApiSettings) WithVerifySsl(VerifySsl bool) *ApiSettings {
    a.VerifySsl = VerifySsl
    return a
}
func (a *ApiSettings) WithTimeout(Timeout int32) *ApiSettings {
    a.Timeout = Timeout
    return a
}
func (a *ApiSettings) WithClientId(ClientId string) *ApiSettings {
    a.ClientId = ClientId
    return a
}
func (a *ApiSettings) WithClientSecret(ClientSecret string) *ApiSettings {
    a.ClientSecret = ClientSecret
    return a
}
func (a *ApiSettings) withApiVersion() *ApiSettings {
    a.ApiVersion = DefaultApiVersion
    return a
}

func NewSettingsFromFile(file string, section *string) (ApiSettings, error) {
    if section == nil {
        section = &defaultSectionName
    }

    s := ApiSettings{
        VerifySsl:  true,
        ApiVersion: DefaultApiVersion,
    }

    cfg, err := ini.Load(file)
    if err != nil {
        return s, fmt.Errorf("error reading ini file: %w", err)
    }

    err = cfg.Section(*section).MapTo(&s)
    return s, err

}
