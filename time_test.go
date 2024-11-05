package clock_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("std Time", func() {
	Describe("Now", func() {
		It("return increasing time.Time", func() {
			t1 := time.Now()
			time.Sleep(ms)
			t2 := time.Now()
			Expect(t2.After(t1)).To(BeTrue())
		})
	})

	Describe("NewTimer", func() {
		It("fired by the timer", func() {
			t := time.NewTimer(200 * ms)
			Consistently(t.C, 80*ms, 10*ms).
				ShouldNot(Receive())

			const d = 20 * ms
			Expect(t.Reset(d)).To(BeTrue(), "1st reset")
			t1 := time.Now()
			ct := <-t.C
			t2 := time.Now()
			Expect(ct).To(BeTemporally("~", t2, th), "ct")
			Expect(t2).To(BeTemporally("~", t1.Add(d), th), "t2 - t1")

			Expect(t.Reset(50 * ms)).To(BeFalse())
			Expect(t.Stop()).To(BeTrue(), "stop")
			Consistently(t.C).ShouldNot(Receive())
		})
	})

	Describe("NewTicker", func() {
		It("tick in regular order", func() {
			const d1 = 100 * ms
			const d2 = 50 * ms

			t := time.NewTicker(d1)
			ct1 := <-t.C
			t1 := time.Now()
			ct2 := <-t.C
			t2 := time.Now()
			t.Reset(d2)
			ct3 := <-t.C
			t3 := time.Now()
			t.Stop()
			Consistently(t.C).ShouldNot(Receive())

			Expect(ct1).To(BeTemporally("~", t1, th), "ct1")
			Expect(ct2).To(BeTemporally("~", t2, th), "ct2")
			Expect(ct3).To(BeTemporally("~", t3, th), "ct3")

			Expect(ct1.Before(ct2)).To(BeTrue(), "ct1 < ct2")
			Expect(ct2).To(BeTemporally("~", ct1.Add(d1), th), "ct2 - ct1")
			Expect(ct2.Before(ct3)).To(BeTrue(), "ct2 < ct3")
			Expect(ct3).To(BeTemporally("~", ct2.Add(d2), th), "ct3 - ct2")
		})
	})
})
