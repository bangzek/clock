package clock_test

import (
	"time"

	. "github.com/bangzek/clock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mock", func() {
	var c *Mock
	tm := time.Date(2021, time.February, 1, 23, 24, 25, 0, time.UTC)
	BeforeEach(func() { c = new(Mock) })

	Describe("Now", func() {
		Context("unscripted", func() {
			It("append now with DefaultScriptNow duration to calls", func() {
				c.Start(tm)
				xt := tm.Add(DefaultScriptNow)
				Expect(c.Now()).To(Equal(xt))
				c.Stop()
				Expect(c.Calls()).To(Equal([]string{"now"}))
				Expect(c.Times()).To(Equal([]time.Time{xt}))
			})
		})

		Context("scripted", func() {
			It("append now with the durations to calls", func() {
				const d = 100 * ms
				c.NowScripts = []time.Duration{d, 0}
				t1 := tm.Add(d)
				t2 := tm.Add(d + DefaultScriptNow)
				t3 := tm.Add(d + 2*DefaultScriptNow)
				c.Start(tm)
				Expect(c.Now()).To(Equal(t1), "#1")
				Expect(c.Now()).To(Equal(t2), "#2")
				Expect(c.Now()).To(Equal(t3), "#3")
				c.Stop()
				Expect(c.Calls()).To(Equal([]string{"now", "now", "now"}))
				Expect(c.Times()).To(Equal([]time.Time{t1, t2, t3}))
			})
		})
	})

	Describe("Timer", func() {
		Context("unscripted", func() {
			It("append timer to calls", func() {
				const d = time.Second / 2
				c.Start(tm)
				t1 := c.NewTimer(time.Second)
				t2 := c.NewTimer(2 * time.Second)
				Expect(t2.Reset(d)).To(BeTrue())
				Expect(t1.Stop()).To(BeTrue())

				Eventually(t2.C).Should(Receive())
				c.Stop()
				Expect(c.Calls()).To(Equal([]string{
					"timer 1s",
					"timer 2s",
					"timer-2.reset " + d.String(),
					"timer-1.stop",
				}))
				Expect(c.Times()).To(HaveExactElements(
					tm.Add(DefaultScriptNow),
					tm.Add(2*DefaultScriptNow),
					tm.Add(3*DefaultScriptNow),
					BeTemporally("~", tm.Add(3*DefaultScriptNow+d), th),
				))
			})

			It("runs like real timer", func() {
				const d1 = 200 * ms
				const d2 = 20 * ms
				const d3 = 50 * ms
				const dr = DefaultScriptRatio
				c.Start(tm)
				t := c.NewTimer(d1)
				Consistently(t.C, d1/dr*9/10, d1/dr/10).ShouldNot(Receive())

				Expect(t.Reset(d2)).To(BeTrue(), "1st reset")
				t1 := c.Now()
				ct := <-t.C
				t2 := c.Now()
				Expect(ct).To(BeTemporally("~", t2, th), "ct")
				Expect(t2).To(BeTemporally("~", t1.Add(d2), th), "t2 - t1")

				Expect(t.Reset(d3)).To(BeFalse())
				Expect(t.Stop()).To(BeTrue(), "stop")
				Consistently(t.C).ShouldNot(Receive())
				c.Stop()

				Expect(c.Calls()).To(Equal([]string{
					"timer " + d1.String(),
					"timer-1.reset " + d2.String(),
					"now",
					"now",
					"timer-1.reset " + d3.String(),
					"timer-1.stop",
				}))
				const dn = DefaultScriptNow
				Expect(c.Times()).To(HaveExactElements(
					tm.Add(dn),
					tm.Add(2*dn),
					tm.Add(3*dn),
					BeTemporally("~", tm.Add(3*dn+d2), th),
					BeTemporally("~", tm.Add(4*dn+d2), th),
					BeTemporally("~", tm.Add(5*dn+d2), th),
				))
			})
		})

		Context("scripted", func() {
			It("append timer to calls", func() {
				const d = 2 * ms
				c.TimerScripts = [][]Script{
					nil,
					{{d, 20}, {0, 0}},
				}
				c.Start(tm)
				t1 := c.NewTimer(time.Second)
				t2 := c.NewTimer(2 * time.Second)
				Expect(t2.Reset(time.Second / 2)).To(BeTrue())
				Expect(t2.Reset(ms)).To(BeTrue())
				Expect(t1.Stop()).To(BeTrue())

				Eventually(t2.C).Should(Receive())
				c.Stop()
				Expect(c.Calls()).To(Equal([]string{
					"timer 1s",
					"timer 2s",
					"timer-2.reset 500ms",
					"timer-2.reset 1ms",
					"timer-1.stop",
				}))
				const dn = DefaultScriptNow
				Expect(c.Times()).To(HaveExactElements(
					tm.Add(dn),
					tm.Add(dn+d),
					tm.Add(dn+d+dn),
					tm.Add(dn+d+2*dn),
					BeTemporally("~", tm.Add(dn+d+3*dn), th),
				))
			})
		})
	})

	Describe("Ticker", func() {
		Context("unscripted", func() {
			It("append ticker to calls", func() {
				const d = time.Second / 2
				c.Start(tm)
				t1 := c.NewTicker(time.Second)
				t2 := c.NewTicker(2 * time.Second)
				t2.Reset(d)
				t1.Stop()

				Eventually(t2.C).Should(Receive())
				c.Stop()
				Expect(c.Calls()).To(Equal([]string{
					"ticker 1s",
					"ticker 2s",
					"ticker-2.reset " + d.String(),
					"ticker-1.stop",
				}))
				const dn = DefaultScriptNow
				Expect(c.Times()).To(HaveExactElements(
					tm.Add(dn),
					tm.Add(2*dn),
					tm.Add(3*dn),
					BeTemporally("~", tm.Add(4*dn+d), th),
				))
			})

			It("runs like real ticker", func() {
				const d1 = 100 * ms
				const d2 = 50 * ms

				c.Start(tm)
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

				Expect(ct1).To(BeTemporally("~", t1, th), "ct1")
				Expect(ct2).To(BeTemporally("~", t2, th), "ct2")
				Expect(ct3).To(BeTemporally("~", t3, th), "ct3")

				Expect(ct1.Before(ct2)).To(BeTrue(), "ct1 < ct2")
				Expect(ct2).To(BeTemporally("~", ct1.Add(d1), 3*ms),
					"ct2 - ct1")
				Expect(ct2.Before(ct3)).To(BeTrue(), "ct2 < ct3")
				Expect(ct3).To(BeTemporally("~", ct2.Add(d2), 3*ms),
					"ct3 - ct2")

				c.Stop()
				Expect(c.Calls()).To(Equal([]string{
					"ticker " + d1.String(),
					"now",
					"now",
					"ticker-1.reset " + d2.String(),
					"now",
					"ticker-1.stop",
				}))
				const dn = DefaultScriptNow
				Expect(c.Times()).To(HaveExactElements(
					tm.Add(dn),
					BeTemporally("~", tm.Add(dn+d1), th),
					BeTemporally("~", tm.Add(2*dn+d1), th),
					BeTemporally("~", tm.Add(dn+2*d1), th),
					BeTemporally("~", tm.Add(2*dn+2*d1), th),
					BeTemporally("~", tm.Add(3*dn+2*d1), th),
					BeTemporally("~", tm.Add(3*dn+2*d1+d2), th),
					BeTemporally("~", tm.Add(4*dn+2*d1+d2), th),
				))
			})
		})

		Context("scripted", func() {
			It("append ticker to calls", func() {
				const d = 2 * ms
				const d2 = 3 * ms / 2
				c.TickerScripts = [][]Script{
					nil,
					{{d, 20}, {0, 0}},
				}
				c.Start(tm)
				t1 := c.NewTicker(time.Second)
				t2 := c.NewTicker(2 * time.Second)
				t2.Reset(time.Second / 2)
				t2.Reset(d2)
				t1.Stop()

				Eventually(t2.C, d2+d2/200, d2/200).Should(Receive())
				c.Stop()
				Expect(c.Calls()).To(Equal([]string{
					"ticker 1s",
					"ticker 2s",
					"ticker-2.reset 500ms",
					"ticker-2.reset " + d2.String(),
					"ticker-1.stop",
				}))
				const dn = DefaultScriptNow
				// sometime it's 5 sometime it's 6
				list := c.Times()
				Expect(len(list)).To(BeNumerically(">=", 5))
				Expect(list[0]).To(Equal(tm.Add(dn)), "times[0]")
				Expect(list[1]).To(Equal(tm.Add(dn+d)), "times[1]")
				Expect(list[2]).To(Equal(tm.Add(dn+d+dn)), "times[2]")
				Expect(list[3]).To(Equal(tm.Add(dn+d+dn+dn)), "times[3]")
				Expect(list[4]).To(
					BeTemporally("~", tm.Add(dn+d+dn+dn+d2), th),
					"times[4]")
			})
		})
	})

	const (
		stopFirst  = "clock.Mock must be Stop() first"
		startFirst = "clock.Mock must be Start() first"
	)

	Context("forget to Start()", func() {
		Describe("Now", func() {
			It("should panic", func() {
				Expect(func() { c.Now() }).To(PanicWith(startFirst))
			})
		})
		Describe("NewTimer", func() {
			It("should panic", func() {
				Expect(func() { c.NewTimer(time.Second) }).
					To(PanicWith(startFirst))
			})
		})
		Describe("NewTicker", func() {
			It("should panic", func() {
				Expect(func() { c.NewTicker(time.Second) }).
					To(PanicWith(startFirst))
			})
		})
		Describe("Calls", func() {
			It("should panic", func() {
				Expect(func() { c.Calls() }).To(PanicWith(stopFirst))
			})
		})
		Describe("Times", func() {
			It("should panic", func() {
				Expect(func() { c.Times() }).To(PanicWith(stopFirst))
			})
		})
	})

	Context("forget to Stop()", func() {
		BeforeEach(func() { c.Start(tm) })
		Describe("Calls", func() {
			It("should panic", func() {
				Expect(func() { c.Calls() }).To(PanicWith(stopFirst))
			})
		})
		Describe("Times", func() {
			It("should panic", func() {
				Expect(func() { c.Times() }).To(PanicWith(stopFirst))
			})
		})
	})
})
