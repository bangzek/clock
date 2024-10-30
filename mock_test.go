package clock_test

import (
	"time"

	. "github.com/bangzek/clock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mock", func() {
	var c *Mock
	t := time.Date(2021, time.February, 1, 23, 24, 25, 0, time.UTC)
	BeforeEach(func() {
		c = NewMock(t)
	})

	Describe("Now", func() {
		Context("unscripted", func() {
			It("append now with DefaultScriptNow duration to calls", func() {
				Expect(c.Now()).To(Equal(t.Add(DefaultScriptNow)))
				Expect(c.Calls).
					To(Equal([]string{
						t.Add(DefaultScriptNow).
							Format("now " + time.RFC3339Nano),
					}))
			})
		})

		Context("scripted", func() {
			It("append now with the durations to calls", func() {
				d := 100 * time.Millisecond
				c.NowScripts = []time.Duration{d, 0}
				t1 := t.Add(d)
				t2 := t.Add(d + DefaultScriptNow)
				t3 := t.Add(d + 2*DefaultScriptNow)
				Expect(c.Now()).To(Equal(t1), "#1")
				Expect(c.Now()).To(Equal(t2), "#2")
				Expect(c.Now()).To(Equal(t3), "#3")
				Expect(c.Calls).To(Equal([]string{
					t1.Format("now " + time.RFC3339Nano),
					t2.Format("now " + time.RFC3339Nano),
					t3.Format("now " + time.RFC3339Nano),
				}))
			})
		})
	})

	Describe("Timer", func() {
		Context("unscripted", func() {
			It("append ticker to calls", func() {
				t1 := c.NewTimer(time.Second)
				t2 := c.NewTimer(2 * time.Second)
				Expect(t2.Reset(time.Second / 2)).To(BeTrue())
				Expect(t1.Stop()).To(BeTrue())

				Eventually(t2.C).Should(Receive())
				Expect(c.Calls).To(Equal([]string{
					"timer 1s",
					"timer 2s",
					"timer-2.reset 500ms",
					"timer-1.stop",
				}))
			})

			It("runs like real timer", func() {
				// so it don't fired too early
				t := c.NewTimer(200 * time.Millisecond * DefaultScriptRatio)
				Consistently(t.C, 80*time.Millisecond, 10*time.Millisecond).
					ShouldNot(Receive())

				d := 20 * time.Millisecond
				Expect(t.Reset(d)).To(BeTrue(), "1st reset")
				t1 := c.Now()
				ct := <-t.C
				t2 := c.Now()
				Expect(ct).To(BeTemporally("~", t2, mockThreshold), "ct")
				Expect(t2).To(BeTemporally("~", t1.Add(d), mockThreshold),
					"t2 - t1")

				Expect(t.Reset(50 * time.Millisecond)).To(BeFalse())
				Expect(t.Stop()).To(BeTrue(), "stop")
				Consistently(t.C).ShouldNot(Receive())
			})
		})

		Context("scripted", func() {
			It("append ticker to ops", func() {
				c.TimerScripts = [][]Script{
					nil,
					{{2 * time.Millisecond, 200}, {0, 0}},
				}
				t1 := c.NewTimer(time.Second)
				t2 := c.NewTimer(2 * time.Second)
				Expect(t2.Reset(time.Second / 2)).To(BeTrue())
				Expect(t2.Reset(time.Millisecond)).To(BeTrue())
				Expect(t1.Stop()).To(BeTrue())

				Eventually(t2.C).Should(Receive())
				Expect(c.Calls).To(Equal([]string{
					"timer 1s",
					"timer 2s",
					"timer-2.reset 500ms",
					"timer-2.reset 1ms",
					"timer-1.stop",
				}))
			})
		})
	})

	Describe("Ticker", func() {
		Context("unscripted", func() {
			It("append ticker to calls", func() {
				t1 := c.NewTicker(time.Second)
				t2 := c.NewTicker(2 * time.Second)
				t2.Reset(time.Second / 2)
				t1.Stop()

				Eventually(t2.C).Should(Receive())
				Expect(c.Calls).To(Equal([]string{
					"ticker 1s",
					"ticker 2s",
					"ticker-2.reset 500ms",
					"ticker-1.stop",
				}))
			})

			It("runs like real ticker", func() {
				d1 := 100 * time.Millisecond * mockRatio
				d2 := 50 * time.Millisecond * mockRatio

				t := c.NewTicker(d1)
				ct1 := <-t.C
				t1 := c.Now()
				ct2 := <-t.C
				t2 := c.Now()
				t.Reset(d2)
				ct3 := <-t.C
				t3 := c.Now()
				t.Stop()
				Consistently(t.C).ShouldNot(Receive())

				Expect(ct1).To(BeTemporally("~", t1, mockThreshold), "ct1")
				Expect(ct2).To(BeTemporally("~", t2, mockThreshold), "ct2")
				Expect(ct3).To(BeTemporally("~", t3, mockThreshold), "ct3")

				Expect(ct1.Before(ct2)).To(BeTrue(), "ct1 < ct2")
				Expect(ct2).To(
					BeTemporally("~", ct1.Add(d1), mockThreshold),
					"ct2 - ct1")
				Expect(ct2.Before(ct3)).To(BeTrue(), "ct2 < ct3")
				Expect(ct3).To(
					BeTemporally("~", ct2.Add(d2), mockThreshold),
					"ct3 - ct2")
			})
		})

		Context("scripted", func() {
			It("append ticker to calls", func() {
				c.TickerScripts = [][]Script{
					nil,
					{{2 * time.Millisecond, 200}, {0, 0}},
				}
				t1 := c.NewTicker(time.Second)
				t2 := c.NewTicker(2 * time.Second)
				t2.Reset(time.Second / 2)
				t2.Reset(time.Millisecond)
				t1.Stop()

				Eventually(t2.C).Should(Receive())
				Expect(c.Calls).To(Equal([]string{
					"ticker 1s",
					"ticker 2s",
					"ticker-2.reset 500ms",
					"ticker-2.reset 1ms",
					"ticker-1.stop",
				}))
			})
		})
	})
})
