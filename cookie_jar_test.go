package collysqlite_test

import (
	"net/http"
	"net/url"
	"time"

	"github.com/jimsmart/collysqlite"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CookieJar", func() {

	It("should Init and Destroy", func() {
		Context("with a vanilla name (no path)", func() {
			name := "test-db-" + randomName()
			j := collysqlite.NewCookieJar(name)
			Expect(j.Init()).To(BeNil())
			filename := name + ".sqlite"
			Expect(filename).To(BeAnExistingFile())
			Expect(j.Destroy()).To(BeNil())
			Expect(filename).NotTo(BeAnExistingFile())
		})

		Context("with a ./name", func() {
			name := "test-db-" + randomName()
			j := collysqlite.NewCookieJar(name)
			Expect(j.Init()).To(BeNil())
			filename := name + ".sqlite"
			Expect(filename).To(BeAnExistingFile())
			Expect(j.Destroy()).To(BeNil())
			Expect(filename).NotTo(BeAnExistingFile())
		})

		Context("with a ./subfolder/name", func() {
			name := "test-db-" + randomName()
			j := collysqlite.NewCookieJar(name)
			Expect(j.Init()).To(BeNil())
			filename := name + ".sqlite"
			Expect(filename).To(BeAnExistingFile())
			Expect(j.Destroy()).To(BeNil())
			Expect(filename).NotTo(BeAnExistingFile())
		})
	})

	It("should set and get cookies", func() {
		name := "test-db-" + randomName()
		j := collysqlite.NewCookieJar(name)
		Expect(j.Init()).To(BeNil())
		defer j.Destroy()

		// SetCookies.
		url, _ := url.Parse("http://example.org")
		cookies := []*http.Cookie{
			&http.Cookie{
				Name:   "cookie1_name",
				Value:  "cookie1_value",
				Path:   "/",
				Domain: ".example.org",
			},
			&http.Cookie{
				Name:   "cookie2_name",
				Value:  "cookie2_value",
				Path:   "/",
				Domain: ".example.org",
			},
		}
		Expect(j.SetCookies(url, cookies)).To(BeNil())
		// Get existing.
		got, err := j.Cookies(url)
		Expect(err).To(BeNil())
		// Expect(got).To(Equal(cookies)) // This doesn't work: some fields get populated during the roundtrip to the db/etc.

		// TODO(js) For a while this test was producing nonsense: the cookie serialisation in redisstorage has bugs :/
		// TODO(js) Write tests to highight existing bugs, submit patches... as soon as I'm not on the arse end of a coding marathon.
		Expect(got).To(HaveLen(2))
		sgot := toStrings(got)
		Expect(sgot).To(ContainElement("cookie1_name=cookie1_value; Path=/; Domain=example.org"))
		Expect(sgot).To(ContainElement(cookies[0].String()))
		Expect(sgot).To(ContainElement(cookies[1].String()))
	})

	It("should add cookies to existing cookies", func() {
		name := "test-db-" + randomName()
		j := collysqlite.NewCookieJar(name)
		Expect(j.Init()).To(BeNil())
		defer j.Destroy()

		// SetCookies.
		url, _ := url.Parse("http://example.org")
		cookies := []*http.Cookie{
			&http.Cookie{
				Name:   "cookie1_name",
				Value:  "cookie1_value",
				Path:   "/",
				Domain: ".example.org",
			},
			&http.Cookie{
				Name:   "cookie2_name",
				Value:  "cookie2_value",
				Path:   "/",
				Domain: ".example.org",
			},
		}
		Expect(j.SetCookies(url, cookies)).To(BeNil())

		// Add another.
		more := []*http.Cookie{
			&http.Cookie{
				Name:   "cookie3_name",
				Value:  "cookie3_value",
				Path:   "/",
				Domain: ".example.org",
			},
		}
		Expect(j.SetCookies(url, more)).To(BeNil())
		// Get existing.
		got, err := j.Cookies(url)
		Expect(err).To(BeNil())
		Expect(got).To(HaveLen(3))
		sgot := toStrings(got)
		Expect(sgot).To(ContainElement(more[0].String()))
	})

	It("should drop expired cookies", func() {
		name := "test-db-" + randomName()
		j := collysqlite.NewCookieJar(name)
		Expect(j.Init()).To(BeNil())
		defer j.Destroy()

		// SetCookies.
		url, _ := url.Parse("http://example.org")
		cookies := []*http.Cookie{
			&http.Cookie{
				Name:   "cookie1_name",
				Value:  "cookie1_value",
				Path:   "/",
				Domain: ".example.org",
			},
			&http.Cookie{
				Name:   "cookie2_name",
				Value:  "cookie2_value",
				Path:   "/",
				Domain: ".example.org",
			},
		}
		Expect(j.SetCookies(url, cookies)).To(BeNil())
		// Get existing.
		got, err := j.Cookies(url)
		Expect(err).To(BeNil())
		Expect(got).To(HaveLen(2))

		// Expire a cookie.
		expired := []*http.Cookie{
			&http.Cookie{
				Name:    "cookie1_name",
				Path:    "/",
				Domain:  ".example.org",
				Expires: time.Now(),
			},
		}
		Expect(j.SetCookies(url, expired)).To(BeNil())
		got, err = j.Cookies(url)
		Expect(err).To(BeNil())
		Expect(got).To(HaveLen(1))
		Expect(got[0].String()).To(Equal("cookie2_name=cookie2_value; Path=/; Domain=example.org"))
	})

	It("should drop secure cookies if not over https", func() {
		name := "test-db-" + randomName()
		j := collysqlite.NewCookieJar(name)
		Expect(j.Init()).To(BeNil())
		defer j.Destroy()

		// SetCookies - one is marked secure.
		url, _ := url.Parse("https://example.org")
		cookies := []*http.Cookie{
			&http.Cookie{
				Name:   "cookie1_name",
				Value:  "cookie1_value",
				Path:   "/",
				Domain: ".example.org",
				Secure: true,
			},
			&http.Cookie{
				Name:   "cookie2_name",
				Value:  "cookie2_value",
				Path:   "/",
				Domain: ".example.org",
			},
		}
		Expect(j.SetCookies(url, cookies)).To(BeNil())
		// Get existing.
		got, err := j.Cookies(url)
		Expect(err).To(BeNil())
		Expect(got).To(HaveLen(2))
		// Get for http.
		url, _ = url.Parse("http://example.org")
		got, err = j.Cookies(url)
		Expect(err).To(BeNil())
		Expect(got).To(HaveLen(1))
		Expect(got[0].String()).To(Equal("cookie2_name=cookie2_value; Path=/; Domain=example.org"))
	})

	It("should not get cookies for an unknown domain", func() {
		name := "test-db-" + randomName()
		j := collysqlite.NewCookieJar(name)
		Expect(j.Init()).To(BeNil())
		defer j.Destroy()

		// Get non-existing.
		url, _ := url.Parse("http://no-such-domain.org")
		got, err := j.Cookies(url)
		Expect(err).To(BeNil())
		Expect(got).To(HaveLen(0))
	})

	It("should handle cookies containing a newline", func() {
		name := "test-db-" + randomName()
		j := collysqlite.NewCookieJar(name)
		Expect(j.Init()).To(BeNil())
		defer j.Destroy()

		// SetCookies.
		url, _ := url.Parse("http://example.org")
		cookies := []*http.Cookie{
			&http.Cookie{
				Name:   "cookie1_name",
				Value:  "cookie1_\n_value",
				Path:   "/",
				Domain: ".example.org",
			},
		}
		Expect(j.SetCookies(url, cookies)).To(BeNil())
		// Get existing.
		got, err := j.Cookies(url)
		Expect(err).To(BeNil())
		Expect(got).To(HaveLen(1))
		// It turns out that this is ok, net/http handles this,
		// and emits a warning  ('dropping invalid bytes') to the console when it does so.
		Expect(got[0].String()).To(Equal("cookie1_name=cookie1__value; Path=/; Domain=example.org"))
		Expect(got[0].String()).To(Equal(cookies[0].String()))
	})

})

func toStrings(cookies []*http.Cookie) []string {
	s := make([]string, len(cookies))
	for i, c := range cookies {
		s[i] = c.String()
	}
	return s
}
