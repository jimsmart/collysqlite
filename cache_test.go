package collysqlite_test

import (
	"encoding/hex"
	"log"
	"math/rand"

	"github.com/jimsmart/collysqlite"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cache", func() {

	It("should Init and Destroy", func() {
		Context("with a vanilla name (no path)", func() {
			name := "test-db-" + randomName()
			c := collysqlite.NewCache(name)
			Expect(c.Init()).To(BeNil())
			filename := name + ".sqlite"
			Expect(filename).To(BeAnExistingFile())
			Expect(c.Destroy()).To(BeNil())
			Expect(filename).NotTo(BeAnExistingFile())
		})

		Context("with a ./name", func() {
			name := "test-db-" + randomName()
			c := collysqlite.NewCache(name)
			Expect(c.Init()).To(BeNil())
			filename := name + ".sqlite"
			Expect(filename).To(BeAnExistingFile())
			Expect(c.Destroy()).To(BeNil())
			Expect(filename).NotTo(BeAnExistingFile())
		})

		Context("with a ./subfolder/name", func() {
			name := "test-db-" + randomName()
			c := collysqlite.NewCache(name)
			Expect(c.Init()).To(BeNil())
			filename := name + ".sqlite"
			Expect(filename).To(BeAnExistingFile())
			Expect(c.Destroy()).To(BeNil())
			Expect(filename).NotTo(BeAnExistingFile())
		})
	})

	It("should Put, Get and Remove", func() {
		name := "test-db-" + randomName()
		c := collysqlite.NewCache(name)
		Expect(c.Init()).To(BeNil())
		defer c.Destroy()

		// Put.
		url := "http://example.org"
		data := []byte{0, 1, 2, 3, 4, 5, 6, 7}
		Expect(c.Put(url, data)).To(BeNil())
		// Get existing.
		got, err := c.Get(url)
		Expect(err).To(BeNil())
		Expect(got).To(Equal(data))
		// Remove.
		Expect(c.Remove(url)).To(BeNil())
		// Get non-existing.
		got, err = c.Get(url)
		Expect(err).To(BeNil())
		Expect(got).To(BeNil())
		// Remove non-existing.
		Expect(c.Remove(url)).To(BeNil())
	})
})

func randomName() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	h := make([]byte, hex.EncodedLen(len(b)))
	hex.Encode(h, b)
	return string(h)
}
