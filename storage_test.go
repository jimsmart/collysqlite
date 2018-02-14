package collysqlite_test

import (
	"github.com/jimsmart/collysqlite"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Storage", func() {

	It("should Init and Destroy", func() {
		name := "test-db-" + randomName()
		s := collysqlite.NewStorage(name)
		Expect(s.Init()).To(BeNil())
		filename1 := name + "-cookies.sqlite"
		filename2 := name + "-visits.sqlite"
		filename3 := name + "-cache.sqlite"
		Expect(filename1).To(BeAnExistingFile())
		Expect(filename2).To(BeAnExistingFile())
		Expect(filename3).To(BeAnExistingFile())
		Expect(s.Destroy()).To(BeNil())
		Expect(filename1).NotTo(BeAnExistingFile())
		Expect(filename2).NotTo(BeAnExistingFile())
		Expect(filename3).NotTo(BeAnExistingFile())
	})
})
