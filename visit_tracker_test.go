package collysqlite_test

import (
	"github.com/jimsmart/collysqlite"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VisitTracker", func() {

	It("should Init and Destroy", func() {

		Context("with a vanilla name (no path)", func() {
			name := "test-db-" + randomName()
			t := collysqlite.NewVisitTracker(name)
			Expect(t.Init()).To(BeNil())
			filename := name + ".sqlite"
			Expect(filename).To(BeAnExistingFile())
			Expect(t.Destroy()).To(BeNil())
			Expect(filename).NotTo(BeAnExistingFile())
		})

		Context("with a ./name", func() {
			name := "test-db-" + randomName()
			t := collysqlite.NewVisitTracker(name)
			Expect(t.Init()).To(BeNil())
			filename := name + ".sqlite"
			Expect(filename).To(BeAnExistingFile())
			Expect(t.Destroy()).To(BeNil())
			Expect(filename).NotTo(BeAnExistingFile())
		})

		Context("with a ./subfolder/name", func() {
			name := "test-db-" + randomName()
			t := collysqlite.NewVisitTracker(name)
			Expect(t.Init()).To(BeNil())
			filename := name + ".sqlite"
			Expect(filename).To(BeAnExistingFile())
			Expect(t.Destroy()).To(BeNil())
			Expect(filename).NotTo(BeAnExistingFile())
		})
	})

	It("should track visits", func() {
		name := "test-db-" + randomName()
		j := collysqlite.NewVisitTracker(name)
		Expect(j.Init()).To(BeNil())
		defer j.Destroy()

		// Visited.
		id := uint64(12345)
		Expect(j.Visited(id)).To(BeNil())
		// IsVisited.
		got, err := j.IsVisited(id)
		Expect(err).To(BeNil())
		Expect(got).To(BeTrue())
		// Get non-existing.
		got, err = j.IsVisited(123)
		Expect(err).To(BeNil())
		Expect(got).To(BeFalse())
	})

})
