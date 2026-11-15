// capture fetches the URL-HEADING-1 golden-set pages and stores raw bodies as
// NN.html snapshots next to expected.json. Run manually: go run ./capture.
// Files under testdata/ are ignored by the Go tool, so this never compiles
// into the backend.
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var pages = []struct {
	id  string
	url string
}{
	{"01", "https://developers.onelogin.com/openid-connect/guides/auth-flow-pkce"},
	{"02", "https://www.keycloak.org/docs-api/latest/rest-api/index.html"},
	{"03", "https://jwt.io/"},
	{"04", "https://developers.onelogin.com/openid-connect/api/authorization-code-grant"},
	{"05", "https://visiology-doc.atlassian.net/wiki/spaces/trouble/pages/158728193/Keycloak+LDAP"},
	{"06", "https://openid.net/specs/openid-connect-registration-1_0.html"},
	{"07", "https://jwt.io/introduction"},
	{"08", "https://jsonformatter.org/"},
	{"09", "https://cheloveki.zenclass.ru/public/products?code"},
	{"10", "https://habr.com/ru/companies/ruvds/articles/512862/"},
	{"11", "https://practicum.yandex.ru/blog/devtools-instrumenty-razrabotchika/"},
	{"12", "https://skillbox.ru/media/code/chto-mozhno-delat-v-chrome-devtools-5-poleznykh-funktsiy-dlya-nachinayushchikh/"},
	{"13", "https://developer.mozilla.org/ru/docs/Learn/Common_questions/Tools_and_setup/What_are_browser_developer_tools"},
	{"14", "https://www.unisender.com/ru/blog/gid-po-devtools-chrome-i-drugih-brauzerov/"},
	{"15", "https://babok-school.ru/blog/typical-errors-in-openapi-spec/"},
	{"16", "https://stepik.org/lesson/741960/step/1?after_pass_reset=true&unit=743634"},
	{"17", "https://docs.id.itmo.pro/auth-oidc/requests/"},
	{"18", "https://www.manning.com/books/api-design-patterns"},
	{"19", "https://karpov.courses/systemdesign"},
}

func main() {
	client := &http.Client{Timeout: 20 * time.Second}
	for _, p := range pages {
		req, err := http.NewRequest(http.MethodGet, p.url, nil)
		if err != nil {
			fmt.Printf("%s ERR %v\n", p.id, err)
			continue
		}
		req.Header.Set("User-Agent", "KnowledgeGraphBot/1.0")
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("%s ERR %v\n", p.id, err)
			continue
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		resp.Body.Close()
		if err != nil {
			fmt.Printf("%s ERR %v\n", p.id, err)
			continue
		}
		name := p.id + ".html"
		if werr := os.WriteFile("../"+name, body, 0o644); werr != nil {
			fmt.Printf("%s ERR %v\n", p.id, werr)
			continue
		}
		fmt.Printf("%s %d %d bytes %s -> %s\n", p.id, resp.StatusCode, len(body), p.url, resp.Request.URL)
	}
}
