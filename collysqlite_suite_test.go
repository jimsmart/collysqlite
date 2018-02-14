package collysqlite_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCollySqlite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CollySqlite Suite")
}
